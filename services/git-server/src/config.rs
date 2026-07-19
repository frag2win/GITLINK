//! # Configuration
//!
//! Provides the [`Config`] struct that holds all runtime configuration
//! for the git-server service. Values can be overridden via environment
//! variables; otherwise sensible defaults are used.

use serde::{Deserialize, Serialize};

/// Runtime configuration for the git-server service.
///
/// # Defaults
/// - `repos_path`: `/repos` — directory where bare Git repositories are stored.
/// - `socket_path`: `/socket/git.sock` — path to the Unix domain socket.
/// - `log_level`: `info` — tracing log level filter.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Config {
    /// Filesystem path to the directory containing bare Git repositories.
    pub repos_path: String,

    /// Filesystem path for the Unix domain socket file.
    pub ipc_network: String,
    pub ipc_address: String,

    /// Log level filter string (e.g. "debug", "info", "warn", "error").
    pub log_level: String,
}

impl Default for Config {
    fn default() -> Self {
        Self {
            repos_path: "/repos".to_string(),
            ipc_network: "unix".to_string(),
            ipc_address: "/socket/git.sock".to_string(),
            log_level: "info".to_string(),
        }
    }
}

impl Config {
    /// Load configuration from environment variables, falling back to defaults.
    ///
    /// # Environment Variables
    /// - `GIT_SERVER_REPOS_PATH` — overrides `repos_path`
    /// - `GIT_SERVER_SOCKET_PATH` — overrides `socket_path`
    /// - `GIT_SERVER_LOG_LEVEL` — overrides `log_level`
    pub fn load() -> Self {
        let defaults = Self::default();

        Self {
            repos_path: std::env::var("GIT_SERVER_REPOS_PATH")
                .unwrap_or(defaults.repos_path),
            ipc_network: std::env::var("GIT_SERVER_IPC_NETWORK")
                .unwrap_or(defaults.ipc_network),
            ipc_address: std::env::var("GIT_SERVER_IPC_ADDRESS")
                .unwrap_or(defaults.ipc_address),
            log_level: std::env::var("GIT_SERVER_LOG_LEVEL")
                .unwrap_or(defaults.log_level),
        }
    }

    /// Validate that the configuration values are sane.
    ///
    /// Returns an error if any required paths are empty.
    pub fn validate(&self) -> Result<(), crate::error::ConfigError> {
        if self.repos_path.is_empty() {
            return Err(crate::error::ConfigError::InvalidValue {
                field: "repos_path".into(),
                reason: "must not be empty".into(),
            });
        }
        if self.ipc_address.is_empty() {
            return Err(crate::error::ConfigError::InvalidValue {
                field: "ipc_address".into(),
                reason: "must not be empty".into(),
            });
        }
        Ok(())
    }
}
