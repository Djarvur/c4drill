//! Configuration plumbing between Zed user settings and the language
//! server launch.
//!
//! Users configure the extension through Zed's standard LSP settings
//! block (see the extension README for examples):
//!
//! ```json
//! {
//!   "lsp": {
//!     "c4drill": {
//!       "binary": {
//!         "path": "/opt/tools/c4drill",
//!         "arguments": ["serve", "--lsp"],
//!         "env": { "C4DRILL_LOG": "debug" }
//!       },
//!       "settings": { "format_on_save": true }
//!     }
//!   }
//! }
//! ```
//!
//! The binary block drives discovery ([`crate::discovery`]); the
//! `settings` block is forwarded to the server as workspace configuration.
//! The current c4drill server (issue #32) accepts but ignores workspace
//! configuration — forwarding it keeps the plumbing future-proof without
//! any server changes.

use serde_json::Value;

/// The pieces of the user's `lsp.c4drill` settings the shim consumes.
#[derive(Debug, Default, Clone, PartialEq, Eq)]
pub struct ServerConfig {
    /// `binary.path` — explicit binary path override.
    pub binary_path: Option<String>,
    /// `binary.arguments` — replaces the default `serve --lsp` arguments.
    pub binary_arguments: Option<Vec<String>>,
    /// `binary.env` — extra environment variables for the server process.
    pub binary_env: Vec<(String, String)>,
    /// `settings` — forwarded to the server as workspace configuration.
    pub server_settings: Option<Value>,
}

/// Extracts [`ServerConfig`] from the raw `lsp.c4drill` settings JSON (the
/// shape Zed's `get_settings` returns for the `"lsp"` category, already
/// scoped to the server key by the caller).
///
/// Tolerates malformed input by ignoring bad fields — a broken settings
/// file must never prevent the language server from launching with
/// defaults.
pub fn parse_server_config(raw: &str) -> ServerConfig {
    let mut config = ServerConfig::default();

    let value: Value = match serde_json::from_str(raw) {
        Ok(value) => value,
        Err(_) => return config,
    };

    if let Some(path) = value
        .get("binary")
        .and_then(|binary| binary.get("path"))
        .and_then(Value::as_str)
    {
        config.binary_path = Some(path.to_string());
    }

    if let Some(args) = value
        .get("binary")
        .and_then(|binary| binary.get("arguments"))
        .and_then(Value::as_array)
    {
        let parsed: Vec<String> = args
            .iter()
            .filter_map(|arg| arg.as_str().map(str::to_string))
            .collect();
        if !parsed.is_empty() {
            config.binary_arguments = Some(parsed);
        }
    }

    if let Some(env) = value
        .get("binary")
        .and_then(|binary| binary.get("env"))
        .and_then(Value::as_object)
    {
        config.binary_env = env
            .iter()
            .filter_map(|(key, value)| value.as_str().map(|value| (key.clone(), value.to_string())))
            .collect();
    }

    if let Some(settings) = value.get("settings") {
        if settings.is_object() {
            config.server_settings = Some(settings.clone());
        }
    }

    config
}

/// Returns the workspace configuration payload handed back to Zed (which
/// forwards it to the server). `None` means "nothing to send".
pub fn workspace_configuration(config: &ServerConfig) -> Option<Value> {
    config.server_settings.clone()
}

#[cfg(test)]
mod tests {
    use super::*;
    use serde_json::json;

    #[test]
    fn parse_server_config_empty_is_default() {
        assert_eq!(parse_server_config(""), ServerConfig::default());
        assert_eq!(parse_server_config("not json"), ServerConfig::default());
    }

    #[test]
    fn parse_server_config_reads_binary_block() {
        let raw = json!({
            "binary": {
                "path": "/opt/tools/c4drill",
                "arguments": ["serve", "--lsp", "--debug"],
                "env": { "C4DRILL_LOG": "debug" }
            }
        })
        .to_string();

        let config = parse_server_config(&raw);
        assert_eq!(config.binary_path.as_deref(), Some("/opt/tools/c4drill"));
        assert_eq!(
            config.binary_arguments,
            Some(vec![
                "serve".to_string(),
                "--lsp".to_string(),
                "--debug".to_string()
            ])
        );
        assert_eq!(
            config.binary_env,
            vec![("C4DRILL_LOG".to_string(), "debug".to_string())]
        );
    }

    #[test]
    fn parse_server_config_reads_settings_block() {
        let raw = json!({
            "settings": { "diagnostics": true },
            "initialization_options": { "ignored": true }
        })
        .to_string();

        let config = parse_server_config(&raw);
        assert_eq!(config.server_settings, Some(json!({ "diagnostics": true })));
    }

    #[test]
    fn parse_server_config_ignores_wrong_types() {
        let raw = json!({
            "binary": { "path": 42, "arguments": "serve", "env": { "K": 1 } },
            "settings": "not an object"
        })
        .to_string();

        let config = parse_server_config(&raw);
        assert_eq!(config.binary_path, None);
        assert_eq!(config.binary_arguments, None);
        assert_eq!(config.binary_env, Vec::<(String, String)>::new());
        assert_eq!(config.server_settings, None);
    }

    #[test]
    fn workspace_configuration_none_without_settings() {
        let config = ServerConfig::default();
        assert_eq!(workspace_configuration(&config), None);
    }

    #[test]
    fn workspace_configuration_passes_settings_through() {
        let config = ServerConfig {
            server_settings: Some(json!({ "key": "value" })),
            ..ServerConfig::default()
        };
        assert_eq!(
            workspace_configuration(&config),
            Some(json!({ "key": "value" }))
        );
    }
}
