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
#[cfg(unix)]
use tokio::net::UnixStream;
use tokio::net::TcpStream;

use tokio_util::codec::{Framed, LengthDelimitedCodec};
use tracing::{debug, error, info, warn};

use crate::config::Config;
use crate::socket::protocol::{GitCommandRequest, GitCommandResponse, RequestCommand};

/// Start the Unix domain socket server.
pub async fn start(config: Config) -> Result<()> {
    if config.ipc_network == "unix" {
        #[cfg(unix)]
        {
            let socket_path = Path::new(&config.ipc_address);
            if socket_path.exists() {
                tokio::fs::remove_file(socket_path).await.context("failed to remove stale socket file")?;
            }
            if let Some(parent) = socket_path.parent() {
                tokio::fs::create_dir_all(parent).await.context("failed to create socket directory")?;
            }
            
            let listener = tokio::net::UnixListener::bind(socket_path).context("failed to bind Unix domain socket")?;
            std::os::unix::fs::PermissionsExt::set_mode(&mut std::fs::Permissions::from_mode(0o777), 0o777); // Simplified for now
            // Just use a simpler permissions setter
            std::fs::set_permissions(socket_path, std::os::unix::fs::PermissionsExt::from_mode(0o777))
                .context("failed to set socket permissions")?;
            
            info!(path = %config.ipc_address, "Socket server listening (Unix Protobuf)");
            loop {
                match listener.accept().await {
                    Ok((stream, _addr)) => {
                        let cfg = config.clone();
                        tokio::spawn(async move {
                            if let Err(e) = handle_unix_connection(stream, cfg).await {
                                error!(error = %e, "Connection handler error");
                            }
                        });
                    }
                    Err(e) => warn!(error = %e, "Failed to accept connection"),
                }
            }
        }
        #[cfg(not(unix))]
        {
            anyhow::bail!("Unix sockets are not supported on this platform");
        }
    } else if config.ipc_network == "tcp" {
        let listener = tokio::net::TcpListener::bind(&config.ipc_address).await.context("failed to bind TCP socket")?;
        info!(address = %config.ipc_address, "Socket server listening (TCP Protobuf)");
        loop {
            match listener.accept().await {
                Ok((stream, _addr)) => {
                    let cfg = config.clone();
                    tokio::spawn(async move {
                        if let Err(e) = handle_tcp_connection(stream, cfg).await {
                            error!(error = %e, "Connection handler error");
                        }
                    });
                }
                Err(e) => warn!(error = %e, "Failed to accept connection"),
            }
        }
    } else {
        anyhow::bail!("Unsupported IPC network: {}", config.ipc_network);
    }
}

#[cfg(unix)]
async fn handle_unix_connection(stream: UnixStream, config: Config) -> Result<()> {
    let mut framed = Framed::new(stream, LengthDelimitedCodec::new());
    handle_framed_connection(&mut framed, config).await
}

async fn handle_tcp_connection(stream: TcpStream, config: Config) -> Result<()> {
    let mut framed = Framed::new(stream, LengthDelimitedCodec::new());
    handle_framed_connection(&mut framed, config).await
}

async fn handle_framed_connection<T: tokio::io::AsyncRead + tokio::io::AsyncWrite + Unpin>(
    framed: &mut Framed<T, LengthDelimitedCodec>, 
    config: Config
) -> Result<()> {

    while let Some(result) = framed.next().await {
        let bytes = result.context("failed to read frame from socket")?;
        
        let request = match GitCommandRequest::decode(bytes) {
            Ok(req) => {
                if req.protocol_version != 1 {
                    warn!(version = req.protocol_version, "Unsupported protocol version");
                    let resp = GitCommandResponse {
                        protocol_version: 1,
                        error: Some(crate::socket::protocol::GitError { code: "UnsupportedProtocol".to_string(), message: format!("unsupported protocol version: {}", req.protocol_version) }),
                        result: None,
                    };
                    write_response(framed, &resp).await?;
                    continue;
                }
                req
            },
            Err(e) => {
                warn!(error = %e, "Failed to decode protobuf message");
                let resp = GitCommandResponse {
                    protocol_version: 1,
                    error: Some(crate::socket::protocol::GitError { code: "DecodeError".to_string(), message: format!("protobuf decode error: {e}") }),
                    result: None,
                };
                write_response(framed, &resp).await?;
                continue;
            }
        };

        let response = dispatch_request(request, &config).await;
        write_response(framed, &response).await?;
    }

    debug!("Client disconnected");
    Ok(())
}

/// Dispatch a request to the correct handler based on the command type.
async fn dispatch_request(request: GitCommandRequest, config: &Config) -> GitCommandResponse {
    let command = match request.command {
        Some(c) => c,
        None => return GitCommandResponse {
            protocol_version: 1,
            error: Some(crate::socket::protocol::GitError { code: "MissingCommand".to_string(), message: "Missing command in request".to_string() }),
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

        // Smart HTTP Bridge
        RequestCommand::InfoRefs(req) => crate::handlers::http_handler::handle_info_refs(req, config).await,
        RequestCommand::UploadPack(req) => crate::handlers::http_handler::handle_upload_pack(req, config).await,
        RequestCommand::ReceivePack(req) => crate::handlers::http_handler::handle_receive_pack(req, config).await,

        // Pull Requests
        RequestCommand::MergePullRequest(req) => crate::handlers::git_handler::handle_merge_pull_request(req, config).await,

        // Branch Operations
        RequestCommand::CreateBranch(req) => crate::handlers::git_handler::handle_create_branch(req, config).await,
        RequestCommand::DeleteBranch(req) => crate::handlers::git_handler::handle_delete_branch(req, config).await,
        RequestCommand::ListBranches(req) => crate::handlers::git_handler::handle_list_branches(req, config).await,
        RequestCommand::GetBranch(req) => crate::handlers::git_handler::handle_get_branch(req, config).await,

        // Tag Operations
        RequestCommand::CreateTag(req) => crate::handlers::git_handler::handle_create_tag(req, config).await,
        RequestCommand::DeleteTag(req) => crate::handlers::git_handler::handle_delete_tag(req, config).await,
        RequestCommand::ListTags(req) => crate::handlers::git_handler::handle_list_tags(req, config).await,
    };

    result
}

/// Serialize and write a response back to the framed stream.
async fn write_response<T: tokio::io::AsyncWrite + Unpin>(
    framed: &mut Framed<T, LengthDelimitedCodec>,
    response: &GitCommandResponse,
) -> Result<()> {
    let mut resp_bytes = bytes::BytesMut::new();
    response.encode(&mut resp_bytes).context("failed to encode response")?;
    framed.send(resp_bytes.freeze()).await.context("failed to write response to socket")?;
    Ok(())
}
