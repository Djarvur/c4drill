#!/bin/sh
# bootstrap-grammar.sh — make the in-repo tree-sitter-c4d grammar
# installable as a Zed dev extension.
#
# Why this exists: Zed always builds extension grammars by git-cloning
# `repository` at `rev` from extension.toml (see zed-industries/zed,
# crates/extension/src/extension_builder.rs, compile_grammar/checkout_repo).
# It requires the grammar directory (grammars/c4d) to be a git clone whose
# origin URL equals the manifest repository — a plain in-repo directory
# fails with "grammar directory already exists, but is not a git clone".
#
# Two modes:
#
#   scripts/bootstrap-grammar.sh --local      (default; works offline)
#       Rewrites extension.toml's grammar repository to a file:// URL
#       pointing at this checkout and turns grammars/c4d into a
#       self-hosting clone. Fully offline. The extension.toml change is a
#       LOCAL tweak — revert it before committing.
#
#   scripts/bootstrap-grammar.sh --canonical
#       Sets grammars/c4d up as a clone of the committed canonical URL
#       (https://github.com/Djarvur/tree-sitter-c4d). Works once the
#       grammar has been pushed to that repository (the extraction step);
#       then restore the canonical repository line in extension.toml.
#
# After editing the grammar, re-run the script so Zed's shallow fetch sees
# the new revision, then re-run "Install as Dev Extension".

set -eu

mode="local"
for arg in "$@"; do
    case "$arg" in
        --local) mode="local" ;;
        --canonical) mode="canonical" ;;
        *) echo "usage: $0 [--local|--canonical]" >&2; exit 2 ;;
    esac
done

here=$(cd "$(dirname "$0")" && pwd)
ext_dir=$(cd "$here/.." && pwd)
grammar_dir="$ext_dir/grammars/c4d"
manifest="$ext_dir/extension.toml"
canonical_url="https://github.com/Djarvur/tree-sitter-c4d"
rev="master"

[ -f "$grammar_dir/grammar.js" ] || {
    echo "error: $grammar_dir/grammar.js not found" >&2
    exit 1
}
command -v git >/dev/null || { echo "error: git is required" >&2; exit 1; }

# ---- reset the grammar dir into a fresh single-commit repo -------------
rm -rf "$grammar_dir/.git"
git -C "$grammar_dir" init --quiet
git -C "$grammar_dir" checkout --quiet -b snapshot 2>/dev/null || true
git -C "$grammar_dir" add -A
git -C "$grammar_dir" \
    -c user.name=bootstrap -c user.email=bootstrap@localhost \
    commit --quiet -m "tree-sitter-c4d local snapshot" || true
sha=$(git -C "$grammar_dir" rev-parse HEAD)

if [ "$mode" = "local" ]; then
    url="file://$grammar_dir"
    # publish the snapshot as the branch Zed fetches; 'master' is not the
    # checked-out branch, so pushing into this same repo is allowed
    git -C "$grammar_dir" push --quiet "$grammar_dir" "HEAD:refs/heads/$rev" 2>/dev/null
    git -C "$grammar_dir" remote add origin "$url"
    git -C "$grammar_dir" fetch --quiet origin "$rev" 2>/dev/null || true

    # point the manifest at the local clone (a local, uncommitted tweak)
    if grep -q "repository = \"$canonical_url\"" "$manifest"; then
        sed -i.bak "s|repository = \"$canonical_url\"|repository = \"$url\"|" "$manifest"
        rm -f "$manifest.bak"
    fi
    echo "grammar bootstrapped (offline --local mode):"
    echo "  repository: $url   (written into extension.toml — a local tweak,"
    echo "                                       do not commit it)"
else
    git -C "$grammar_dir" remote add origin "$canonical_url"
    if ! grep -q "repository = \"$url\"" "$manifest" 2>/dev/null && \
       ! grep -q "repository = \"$canonical_url\"" "$manifest"; then
        echo "error: extension.toml no longer pins $canonical_url" >&2
        exit 1
    fi
    echo "grammar bootstrapped (canonical mode):"
    echo "  repository: $canonical_url"
    echo "  Zed will shallow-fetch it at install time — works once the"
    echo "  grammar has been pushed to that repository (git push from"
    echo "  $grammar_dir: git push origin HEAD:refs/heads/master)."
fi

echo "  rev:        $rev ($sha)"
echo
echo "Now (re-)run Zed: Extensions > Install as Dev Extension."
