//! # Git Server — Entry Point
//!
//! This is the main entry point for the git-server service.
//! It initializes structured logging, loads configuration, and starts the
//! Unix Domain Socket server that listens for Git operation requests.
//!
//! **No network exposure** — all communication happens via UDS at `/socket/git.sock`.

mod config;
mod error;
mod git;
mod handlers;
mod models;
mod socket;

#[cfg(test)]
mod tests;

use anyhow::Result;
use tracing::{info, error};

/// Application entry point.
///
/// Sets up tracing, loads configuration, and starts the socket server.
/// The server runs indefinitely, accepting connections on the Unix domain socket.
#[tokio::main]
async fn main() -> Result<()> {
    // Initialize structured logging with the configured log level
    let cfg = config::Config::load();
    cfg.validate().map_err(|e| anyhow::anyhow!("Configuration validation failed: {}", e))?;
    init_tracing(&cfg.log_level);

    info!(
        repos_path = %cfg.repos_path,
        ipc_address = %cfg.ipc_address,
        "Starting git-server"
    );

    // Ensure the repos directory exists
    tokio::fs::create_dir_all(&cfg.repos_path).await?;

    // Start the Unix domain socket server
    match socket::server::start(cfg).await {
        Ok(()) => {
            info!("git-server shut down gracefully");
            Ok(())
        }
        Err(e) => {
            error!(error = %e, "git-server encountered a fatal error");
            Err(e)
        }
    }
}

/// Initialize the tracing subscriber with the given log level filter.
///
/// Outputs structured JSON logs to stdout for container-friendly logging.
fn init_tracing(log_level: &str) {
    use tracing_subscriber::{fmt, EnvFilter};

    let filter = EnvFilter::try_from_default_env()
        .unwrap_or_else(|_| EnvFilter::new(log_level));

    fmt()
        .with_env_filter(filter)
        .with_target(true)
        .with_thread_ids(true)
        .json()
        .init();
}
