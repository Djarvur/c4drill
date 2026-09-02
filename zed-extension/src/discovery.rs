//! Binary discovery for the `c4drill` language server.
//!
//! The extension runs inside a WASM sandbox with no filesystem or
//! environment access, so discovery cannot scan `PATH` directly. The
//! strategy, in priority order (mirroring the standard Zed extension
//! pattern):
//!
//! 1. An explicit user settings override (`lsp.c4drill.binary.path`) —
//!    used verbatim.
//! 2. A `which c4drill` (unix) / `where c4drill` (windows) probe executed
//!    through the host process — the first stdout line is the resolved
//!    absolute path.
//! 3. The bare program name `c4drill` — the Zed host process resolves it
//!    against `PATH` when spawning the language server, so discovery still
//!    works on systems without `which`/`where`.
//!
//! Everything here is pure logic over strings so it is unit-testable on the
//! host (the WASM-only pieces live in [`crate::C4drillExtension`]).

/// The language server program name (the `c4drill` binary).
pub const SERVER_PROGRAM: &str = "c4drill";

/// The default arguments launching the language server with:
/// `c4drill serve --lsp` (LSP over stdio, issue #32).
pub const DEFAULT_SERVER_ARGS: [&str; 2] = ["serve", "--lsp"];

/// How the server binary was resolved.
#[derive(Debug, Clone, PartialEq, Eq)]
pub enum BinarySource {
    /// The user configured an explicit path in settings.
    SettingsOverride,
    /// Resolved through `which`/`where` on the host.
    PathProbe,
    /// The bare program name; the Zed host resolves it against `PATH`.
    BareName,
}

/// The resolved launch specification for the language server.
#[derive(Debug, Clone, PartialEq, Eq)]
pub struct LaunchSpec {
    pub program: String,
    pub args: Vec<String>,
    pub source: BinarySource,
}

/// Returns the platform's `which`-style probe invocation for `program`.
pub fn probe_command(program: &str, windows: bool) -> (String, Vec<String>) {
    if windows {
        // `where` is a cmd.exe builtin AND %SystemRoot%\System32\where.exe;
        // invoking it through cmd keeps the probe working when PATH lacks
        // System32.
        (
            "cmd".to_string(),
            vec!["/C".to_string(), "where".to_string(), program.to_string()],
        )
    } else {
        // `command -v` is a POSIX shell builtin — more portable than the
        // external `which` binary (not installed on some minimal images).
        (
            "sh".to_string(),
            vec![
                "-c".to_string(),
                format!("command -v -- {}", shell_quote(program)),
            ],
        )
    }
}

/// Extracts the first plausible path from `which`/`where` stdout output.
/// Returns `None` for empty output or probe failure noise.
pub fn parse_probe_output(stdout: &str) -> Option<String> {
    stdout
        .lines()
        .map(str::trim)
        .find(|line| !line.is_empty())
        .map(str::to_string)
}

/// Builds the launch spec from the settings override (if any) and the
/// probe result (if any). Priority: settings override > probe > bare name.
///
/// `arguments` from settings REPLACE the default `serve --lsp` arguments
/// (the user takes over the whole command line, the standard Zed behavior).
pub fn resolve_launch(
    settings_path: Option<&str>,
    settings_arguments: Option<&[String]>,
    probe_result: Option<&str>,
) -> LaunchSpec {
    let args: Vec<String> = match settings_arguments {
        Some(args) => args.to_vec(),
        None => DEFAULT_SERVER_ARGS.iter().map(|s| s.to_string()).collect(),
    };

    if let Some(path) = settings_path.map(str::trim).filter(|p| !p.is_empty()) {
        return LaunchSpec {
            program: path.to_string(),
            args,
            source: BinarySource::SettingsOverride,
        };
    }

    if let Some(path) = probe_result.map(str::trim).filter(|p| !p.is_empty()) {
        return LaunchSpec {
            program: path.to_string(),
            args,
            source: BinarySource::PathProbe,
        };
    }

    LaunchSpec {
        program: SERVER_PROGRAM.to_string(),
        args,
        source: BinarySource::BareName,
    }
}

/// Minimal single-argument shell quoting for the `sh -c` probe. The argument
/// is a fixed constant today, but keep the probe injection-proof.
fn shell_quote(arg: &str) -> String {
    if arg
        .bytes()
        .all(|b| b.is_ascii_alphanumeric() || matches!(b, b'/' | b'_' | b'-' | b'.' | b'+'))
    {
        arg.to_string()
    } else {
        format!("'{}'", arg.replace('\'', "'\\''"))
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn probe_command_unix_uses_sh_command_v() {
        let (program, args) = probe_command("c4drill", false);
        assert_eq!(program, "sh");
        assert_eq!(args, vec!["-c", "command -v -- c4drill"]);
    }

    #[test]
    fn probe_command_windows_uses_cmd_where() {
        let (program, args) = probe_command("c4drill", true);
        assert_eq!(program, "cmd");
        assert_eq!(args, vec!["/C", "where", "c4drill"]);
    }

    #[test]
    fn probe_command_quotes_unsafe_program_names() {
        let (_, args) = probe_command("/opt/my tools/c4drill", false);
        assert_eq!(args, vec!["-c", "command -v -- '/opt/my tools/c4drill'"]);
    }

    #[test]
    fn parse_probe_output_takes_first_nonempty_line() {
        assert_eq!(
            parse_probe_output("\n/usr/local/bin/c4drill\n/opt/bin/c4drill\n"),
            Some("/usr/local/bin/c4drill".to_string())
        );
    }

    #[test]
    fn parse_probe_output_empty_is_none() {
        assert_eq!(parse_probe_output(""), None);
        assert_eq!(parse_probe_output("\n  \n"), None);
    }

    #[test]
    fn parse_probe_output_trims_whitespace() {
        assert_eq!(
            parse_probe_output("  /usr/bin/c4drill  \n"),
            Some("/usr/bin/c4drill".to_string())
        );
    }

    #[test]
    fn resolve_launch_default_is_bare_name_with_lsp_args() {
        let spec = resolve_launch(None, None, None);
        assert_eq!(spec.program, "c4drill");
        assert_eq!(spec.args, vec!["serve", "--lsp"]);
        assert_eq!(spec.source, BinarySource::BareName);
    }

    #[test]
    fn resolve_launch_probe_wins_over_bare_name() {
        let spec = resolve_launch(None, None, Some("/opt/homebrew/bin/c4drill"));
        assert_eq!(spec.program, "/opt/homebrew/bin/c4drill");
        assert_eq!(spec.args, vec!["serve", "--lsp"]);
        assert_eq!(spec.source, BinarySource::PathProbe);
    }

    #[test]
    fn resolve_launch_settings_override_wins_over_probe() {
        let spec = resolve_launch(
            Some("~/tools/c4drill"),
            None,
            Some("/opt/homebrew/bin/c4drill"),
        );
        assert_eq!(spec.program, "~/tools/c4drill");
        assert_eq!(spec.source, BinarySource::SettingsOverride);
    }

    #[test]
    fn resolve_launch_blank_settings_path_is_ignored() {
        let spec = resolve_launch(Some("   "), None, Some("/usr/bin/c4drill"));
        assert_eq!(spec.source, BinarySource::PathProbe);
    }

    #[test]
    fn resolve_launch_settings_arguments_replace_defaults() {
        let args = vec![
            "serve".to_string(),
            "--lsp".to_string(),
            "--debug".to_string(),
        ];
        let spec = resolve_launch(Some("/x/c4drill"), Some(&args), None);
        assert_eq!(spec.args, vec!["serve", "--lsp", "--debug"]);
    }

    #[test]
    fn resolve_launch_empty_probe_string_falls_back_to_bare_name() {
        let spec = resolve_launch(None, None, Some("   "));
        assert_eq!(spec.source, BinarySource::BareName);
    }
}
