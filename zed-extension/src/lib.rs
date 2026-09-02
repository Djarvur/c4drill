//! c4drill Zed extension — the WASM shim that wires Zed to the native
//! `c4drill` language server (`c4drill serve --lsp`, issue #32).
//!
//! The shim intentionally contains no language logic: diagnostics,
//! completion, hover, definition, document symbols, formatting and the
//! custom `c4drill/renderDiagram` preview request all live in the shared
//! Go LSP server. The shim's job is the standard Zed extension contract:
//!
//! 1. locate the native `c4drill` binary (settings override, then a
//!    `which`/`where` probe through the host, then bare-name PATH
//!    resolution by the Zed host process);
//! 2. launch it with `serve --lsp`;
//! 3. forward the user's `lsp.c4drill.settings` as workspace
//!    configuration.
//!
//! See the extension README for the user-facing story (TOML scoping,
//! the diagram preview fallback, packaging).

mod config;
mod discovery;

use zed_extension_api::{self as zed, settings::LspSettings, Command, Worktree};

struct C4drillExtension;

impl zed::Extension for C4drillExtension {
    fn new() -> Self {
        Self
    }

    /// Returns the command that starts the c4drill language server.
    ///
    /// Resolution order ([`discovery::resolve_launch`]):
    /// 1. `lsp.c4drill.binary.path` (settings override, verbatim);
    /// 2. `which c4drill` / `where c4drill` executed through the host;
    /// 3. the bare name `c4drill`, resolved against `PATH` by the Zed
    ///    host when it spawns the process.
    fn language_server_command(
        &mut self,
        language_server_id: &zed::LanguageServerId,
        worktree: &Worktree,
    ) -> zed::Result<Command> {
        let server_config = self.server_config(language_server_id, worktree);

        let probe_result = if server_config.binary_path.is_some() {
            None // settings override wins; skip the probe entirely
        } else {
            self.probe_binary()
        };

        let launch = discovery::resolve_launch(
            server_config.binary_path.as_deref(),
            server_config.binary_arguments.as_deref(),
            probe_result.as_deref(),
        );

        Ok(Command {
            command: launch.program,
            args: launch.args,
            env: server_config.binary_env,
        })
    }

    /// Forwards the user's `lsp.c4drill.settings` block to the server as
    /// workspace configuration. The current server ignores it (the #32
    /// surface has no configuration keys yet); sending it costs nothing
    /// and keeps client-side configuration working once the server grows
    /// any.
    fn language_server_workspace_configuration(
        &mut self,
        language_server_id: &zed::LanguageServerId,
        worktree: &Worktree,
    ) -> zed::Result<Option<serde_json::Value>> {
        let server_config = self.server_config(language_server_id, worktree);

        Ok(config::workspace_configuration(&server_config))
    }
}

impl C4drillExtension {
    /// Reads the user's `lsp.c4drill` settings for `worktree`.
    fn server_config(
        &mut self,
        language_server_id: &zed::LanguageServerId,
        worktree: &Worktree,
    ) -> config::ServerConfig {
        let raw_settings = LspSettings::for_worktree(language_server_id.as_ref(), worktree)
            .ok()
            .and_then(|settings| serde_json::to_string(&settings).ok());

        match raw_settings {
            Some(raw) => config::parse_server_config(&raw),
            None => config::ServerConfig::default(),
        }
    }

    /// Runs the platform's `which`-style probe through the host process.
    /// Any failure (missing probe tool, non-zero exit) degrades to `None`,
    /// falling back to bare-name resolution by the Zed host.
    fn probe_binary(&mut self) -> Option<String> {
        let windows = matches!(zed::current_platform(), (zed::Os::Windows, _));

        let (program, args) = discovery::probe_command(discovery::SERVER_PROGRAM, windows);

        let output = zed::process::Command::new(program)
            .args(args)
            .output()
            .ok()?;

        // Exit code 0 = found. On unix a signal-killed probe reports
        // status None — treat that as failure too.
        if output.status != Some(0) {
            return None;
        }

        discovery::parse_probe_output(&String::from_utf8_lossy(&output.stdout))
    }
}

zed::register_extension!(C4drillExtension);
