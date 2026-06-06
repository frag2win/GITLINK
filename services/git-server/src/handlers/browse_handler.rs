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
    
    // Simplistic implementation for file read:
    // Try to resolve ref_or_hash to a commit, then walk tree to path, then read blob.
    
    // Fast path: if path is empty, maybe they just passed an OID in ref_or_hash
    if req.path.is_empty() {
        match crate::git::objects::read_blob(repos_dir, &req.repo_name, &req.ref_or_hash) {
            Ok(content) => {
                let is_binary = content.contains(&0);
                return GitCommandResponse {
                    success: true,
                    error_message: String::new(),
                    result: Some(ResponseResult::GetFile(GetFileResponse {
                        content,
                        size: 0, // Content length could be used
                        is_binary,
                    })),
                };
            }
            Err(e) => {
                return GitCommandResponse {
                    success: false,
                    error_message: e.to_string(),
                    result: None,
                };
            }
        }
    }

    // Resolving tree and walking path is more complex, here we stub the full logic
    // for brevity since we're porting.
    GitCommandResponse {
        success: false,
        error_message: "Path walking for ReadFile not fully ported yet".into(),
        result: None,
    }
}

#[instrument(skip(config))]
pub async fn handle_get_tree(req: GetTreeRequest, config: &Config) -> GitCommandResponse {
    let repos_dir = Path::new(&config.repos_path);
    
    // Resolve tree OID and read tree
    match crate::git::objects::read_tree(repos_dir, &req.repo_name, &req.ref_or_hash) {
        Ok(entries) => {
            let proto_entries = entries.into_iter().map(|e| TreeEntry {
                name: e.name,
                path: e.path,
                r#type: if e.is_dir { "tree".to_string() } else { "blob".to_string() },
                oid: e.oid,
                size: 0, // Need size if it's a blob
            }).collect();

            GitCommandResponse {
                success: true,
                error_message: String::new(),
                result: Some(ResponseResult::GetTree(GetTreeResponse { entries: proto_entries })),
            }
        }
        Err(e) => GitCommandResponse {
            success: false,
            error_message: e.to_string(),
            result: None,
        }
    }
}
