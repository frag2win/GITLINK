//! # Browse Handler
//!
//! Handles file browsing requests over Protobuf.

use std::path::Path;
use tracing::instrument;
use crate::config::Config;
use crate::socket::protocol::*;
use crate::socket::protocol::git_command_response::Result as ResponseResult;

#[instrument(skip(config))]
pub async fn handle_get_file(req: GetFileRequest, config: &Config) -> GitCommandResponse {
    let repos_dir = Path::new(&config.repos_path);
    
    match crate::git::objects::read_file_by_path(repos_dir, &req.repo_name, &req.ref_or_hash, &req.path) {
        Ok(content) => {
            let is_binary = content.contains(&0);
            GitCommandResponse {
                protocol_version: 1, error: None,
                
                result: Some(ResponseResult::GetFile(GetFileResponse {
                    content,
                    size: 0,
                    is_binary,
                })),
            }
        }
        Err(e) => GitCommandResponse {
            protocol_version: 1,
            error: Some(crate::socket::protocol::GitError { code: "GitError".to_string(), message: e.to_string() }),
            result: None,
        }
    }
}

#[instrument(skip(config))]
pub async fn handle_get_tree(req: GetTreeRequest, config: &Config) -> GitCommandResponse {
    let repos_dir = Path::new(&config.repos_path);
    
    match crate::git::objects::read_tree_by_path(repos_dir, &req.repo_name, &req.ref_or_hash, &req.path) {
        Ok(entries) => {
            let proto_entries = entries.into_iter().map(|e| TreeEntry {
                name: e.name,
                path: "".to_string(),
                r#type: match e.object_type {
                    crate::models::TreeEntryKind::Blob => "blob".to_string(),
                    crate::models::TreeEntryKind::Tree => "tree".to_string(),
                    crate::models::TreeEntryKind::Commit => "commit".to_string(),
                },
                oid: e.oid,
                size: e.size.unwrap_or(0) as i64,
            }).collect();

            GitCommandResponse {
                protocol_version: 1, error: None,
                
                result: Some(ResponseResult::GetTree(GetTreeResponse { entries: proto_entries })),
            }
        }
        Err(e) => GitCommandResponse {
            protocol_version: 1,
            error: Some(crate::socket::protocol::GitError { code: "GitError".to_string(), message: e.to_string() }),
            result: None,
        }
    }
}
