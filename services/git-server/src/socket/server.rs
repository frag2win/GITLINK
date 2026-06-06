//! # Socket Server
//!
//! Unix domain socket server that accepts connections, reads length-delimited
//! protobuf messages ([`GitCommandRequest`]), dispatches them to the appropriate handler,
//! and writes back [`GitCommandResponse`] messages.
//!
//! The server runs on Tokio and spawns a task per connection.

use std::path::Path;

use anyhow::{Context, Result};
use futures::{SinkExt, StreamExt};
use prost::Message;
use tokio::net::UnixListener;
use tokio_util::codec::{Framed, LengthDelimitedCodec};
use tracing::{debug, error, info, instrument, warn};

use crate::config::Config;
use crate::socket::protocol::{GitCommandRequest, GitCommandResponse, RequestCommand};

/// Start the Unix domain socket server.
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

    info!(path = %config.socket_path, "Socket server listening (Protobuf)");

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

/// Handle a single client connection with length-delimited Protobuf messages.
#[instrument(skip_all)]
async fn handle_connection(stream: tokio::net::UnixStream, config: Config) -> Result<()> {
    // Use tokio_util LengthDelimitedCodec for framing the binary protobuf payloads
    let mut framed = Framed::new(stream, LengthDelimitedCodec::new());

    while let Some(result) = framed.next().await {
        let bytes = result.context("failed to read frame from socket")?;
        
        let request = match GitCommandRequest::decode(bytes) {
            Ok(req) => req,
            Err(e) => {
                warn!(error = %e, "Failed to decode protobuf message");
                let resp = GitCommandResponse {
                    success: false,
                    error_message: format!("protobuf decode error: {e}"),
                    result: None,
                };
                write_response(&mut framed, &resp).await?;
                continue;
            }
        };

        debug!("Processing protobuf request");
        let response = dispatch_request(request, &config).await;
        write_response(&mut framed, &response).await?;
    }

    debug!("Client disconnected");
    Ok(())
}

/// Dispatch a request to the correct handler based on the command type.
async fn dispatch_request(request: GitCommandRequest, config: &Config) -> GitCommandResponse {
    let command = match request.command {
        Some(c) => c,
        None => return GitCommandResponse {
            success: false,
            error_message: "Missing command in request".to_string(),
            result: None,
        },
    };

    let result = match command {
        // Repo CRUD
        RequestCommand::CreateRepo(req) => crate::handlers::repo_handler::handle_create(req, config).await,
        RequestCommand::DeleteRepo(req) => crate::handlers::repo_handler::handle_delete(req, config).await,
        RequestCommand::ListRepos(req) => crate::handlers::repo_handler::handle_list(req, config).await,

        // Git data
        RequestCommand::Push(req) => crate::handlers::git_handler::handle_push(req, config).await,
        RequestCommand::Pull(req) => crate::handlers::git_handler::handle_pull(req, config).await,
        RequestCommand::ListCommits(req) => crate::handlers::git_handler::handle_list_commits(req, config).await,
        RequestCommand::GetCommit(req) => crate::handlers::git_handler::handle_get_commit(req, config).await,
        RequestCommand::GetDiff(req) => crate::handlers::git_handler::handle_get_diff(req, config).await,

        // Browse
        RequestCommand::GetFile(req) => crate::handlers::browse_handler::handle_get_file(req, config).await,
        RequestCommand::GetTree(req) => crate::handlers::browse_handler::handle_get_tree(req, config).await,
    };

    result
}

/// Serialize and write a response back to the framed stream.
async fn write_response(
    framed: &mut Framed<tokio::net::UnixStream, LengthDelimitedCodec>,
    response: &GitCommandResponse,
) -> Result<()> {
    let mut resp_bytes = bytes::BytesMut::new();
    response.encode(&mut resp_bytes).context("failed to encode response")?;
    framed.send(resp_bytes.freeze()).await.context("failed to write response to socket")?;
    Ok(())
}
