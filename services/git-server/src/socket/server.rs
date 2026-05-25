//! # Socket Server
//!
//! Unix domain socket server that accepts connections, reads JSON-delimited
//! messages ([`RepoRequest`]), dispatches them to the appropriate handler,
//! and writes back [`RepoResponse`] messages.
//!
//! The server runs on Tokio and spawns a task per connection.

use std::path::Path;

use anyhow::{Context, Result};
use tokio::io::{AsyncBufReadExt, AsyncWriteExt, BufReader};
use tokio::net::UnixListener;
use tracing::{debug, error, info, instrument, warn};

use crate::config::Config;
use crate::socket::protocol::{GitOperation, RepoRequest, RepoResponse};

/// Maximum allowed message size in bytes (1 MiB).
const MAX_MESSAGE_SIZE: usize = 1024 * 1024;

/// Start the Unix domain socket server.
///
/// Binds to `config.socket_path`, then loops accepting connections.
/// Each connection is handled in a separate Tokio task.
///
/// # Errors
/// Returns an error if the socket cannot be bound.
pub async fn start(config: Config) -> Result<()> {
    let socket_path = Path::new(&config.socket_path);

    // Remove stale socket file if it exists
    if socket_path.exists() {
        tokio::fs::remove_file(socket_path)
            .await
            .context("failed to remove stale socket file")?;
    }

    // Ensure parent directory exists
    if let Some(parent) = socket_path.parent() {
        tokio::fs::create_dir_all(parent)
            .await
            .context("failed to create socket directory")?;
    }

    let listener = UnixListener::bind(socket_path)
        .context("failed to bind Unix domain socket")?;

    info!(path = %config.socket_path, "Socket server listening");

    loop {
        match listener.accept().await {
            Ok((stream, _addr)) => {
                let cfg = config.clone();
                tokio::spawn(async move {
                    if let Err(e) = handle_connection(stream, cfg).await {
                        error!(error = %e, "Connection handler error");
                    }
                });
            }
            Err(e) => {
                warn!(error = %e, "Failed to accept connection");
            }
        }
    }
}

/// Handle a single client connection.
///
/// Reads newline-delimited JSON messages, dispatches each to the appropriate
/// handler, and writes back the JSON response.
#[instrument(skip_all)]
async fn handle_connection(stream: tokio::net::UnixStream, config: Config) -> Result<()> {
    let (reader, mut writer) = stream.into_split();
    let mut buf_reader = BufReader::new(reader);
    let mut line = String::new();

    loop {
        line.clear();
        let bytes_read = buf_reader
            .read_line(&mut line)
            .await
            .context("failed to read from socket")?;

        if bytes_read == 0 {
            debug!("Client disconnected");
            break;
        }

        if line.len() > MAX_MESSAGE_SIZE {
            let resp = RepoResponse::err(
                String::new(),
                format!("message too large: {} bytes", line.len()),
            );
            write_response(&mut writer, &resp).await?;
            continue;
        }

        let request: RepoRequest = match serde_json::from_str(line.trim()) {
            Ok(req) => req,
            Err(e) => {
                warn!(error = %e, "Invalid request JSON");
                let resp = RepoResponse::err(String::new(), format!("invalid JSON: {e}"));
                write_response(&mut writer, &resp).await?;
                continue;
            }
        };

        debug!(
            request_id = %request.request_id,
            operation = ?request.operation,
            "Processing request"
        );

        let response = dispatch_request(&request, &config).await;
        write_response(&mut writer, &response).await?;
    }

    Ok(())
}

/// Dispatch a request to the correct handler based on the operation.
async fn dispatch_request(request: &RepoRequest, config: &Config) -> RepoResponse {
    let result = match request.operation {
        // Repository CRUD
        GitOperation::InitRepo
        | GitOperation::DeleteRepo
        | GitOperation::ListRepos => {
            crate::handlers::repo_handler::handle(request, config).await
        }

        // Git data operations
        GitOperation::Clone
        | GitOperation::Push
        | GitOperation::Pull
        | GitOperation::ListRefs
        | GitOperation::GetObject
        | GitOperation::ListBranches
        | GitOperation::CreateBranch
        | GitOperation::DeleteBranch
        | GitOperation::ListTags
        | GitOperation::ListCommits
        | GitOperation::GetCommitDetail
        | GitOperation::GetDiff => {
            crate::handlers::git_handler::handle(request, config).await
        }

        // File browsing
        GitOperation::BrowseTree
        | GitOperation::ReadFile => {
            crate::handlers::browse_handler::handle(request, config).await
        }
    };

    match result {
        Ok(resp) => resp,
        Err(e) => RepoResponse::err(request.request_id.clone(), format!("{e:#}")),
    }
}

/// Serialize and write a response as a newline-delimited JSON message.
async fn write_response(
    writer: &mut tokio::net::unix::OwnedWriteHalf,
    response: &RepoResponse,
) -> Result<()> {
    let mut json = serde_json::to_string(response)
        .context("failed to serialize response")?;
    json.push('\n');
    writer
        .write_all(json.as_bytes())
        .await
        .context("failed to write response to socket")?;
    Ok(())
}
