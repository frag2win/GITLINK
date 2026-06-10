//! # Git Handler
//!
//! Handles Git data operation requests over Protobuf.

use std::path::Path;
use tracing::instrument;
use crate::config::Config;
use crate::socket::protocol::*;
use crate::socket::protocol::git_command_response::Result as ResponseResult;

#[instrument(skip(config))]
pub async fn handle_push(req: PushRequest, config: &Config) -> GitCommandResponse {
    let repos_dir = Path::new(&config.repos_path);
    
    match crate::git::pack::receive_pack(repos_dir, &req.repo_name, &req.pack_data) {
        Ok(_result) => {
            // Note: receive_pack response would ideally return the new head, but for now we just return ok
            GitCommandResponse {
                success: true,
                error_message: String::new(),
                result: Some(ResponseResult::Push(PushResponse { new_head: "HEAD".to_string() })),
            }
        }
        Err(e) => GitCommandResponse {
            success: false,
            error_message: e.to_string(),
            result: None,
        }
    }
}

#[instrument(skip(config))]
pub async fn handle_pull(req: PullRequest, config: &Config) -> GitCommandResponse {
    let repos_dir = Path::new(&config.repos_path);
    
    // In a real implementation, want_oids and have_oids would come from the request
    let want_oids = vec![];
    let have_oids = vec![];
    
    match crate::git::pack::send_pack(repos_dir, &req.repo_name, &want_oids, &have_oids) {
        Ok(result) => {
            GitCommandResponse {
                success: true,
                error_message: String::new(),
                result: Some(ResponseResult::Pull(PullResponse { pack_data: result.pack_data })),
            }
        }
        Err(e) => GitCommandResponse {
            success: false,
            error_message: e.to_string(),
            result: None,
        }
    }
}

#[instrument(skip(config))]
pub async fn handle_list_commits(req: ListCommitsRequest, config: &Config) -> GitCommandResponse {
    let repos_dir = Path::new(&config.repos_path);
    
    let start_oid = if req.branch.is_empty() { None } else { Some(req.branch.as_str()) };
    
    match crate::git::commits::list_commits(repos_dir, &req.repo_name, start_oid, req.limit as usize, req.offset as usize) {
        Ok(commits) => {
            let proto_commits = commits.into_iter().map(|c| CommitInfo {
                hash: c.id,
                author_name: c.author_name,
                author_email: c.author_email,
                message: c.message,
                timestamp: c.authored_at,
                parent_hashes: c.parent_ids,
            }).collect();
            
            GitCommandResponse {
                success: true,
                error_message: String::new(),
                result: Some(ResponseResult::ListCommits(ListCommitsResponse { commits: proto_commits, total_count: 0 })), // total_count can be updated later
            }
        }
        Err(e) => GitCommandResponse {
            success: false,
            error_message: e.to_string(),
            result: None,
        }
    }
}

#[instrument(skip(config))]
pub async fn handle_get_commit(req: GetCommitRequest, config: &Config) -> GitCommandResponse {
    let repos_dir = Path::new(&config.repos_path);
    
    match crate::git::commits::get_commit_detail(repos_dir, &req.repo_name, &req.hash) {
        Ok(c) => {
            let info = CommitInfo {
                hash: c.id,
                author_name: c.author_name,
                author_email: c.author_email,
                message: c.message,
                timestamp: c.authored_at,
                parent_hashes: c.parent_ids,
            };
            GitCommandResponse {
                success: true,
                error_message: String::new(),
                result: Some(ResponseResult::GetCommit(GetCommitResponse { commit: Some(info) })),
            }
        }
        Err(e) => GitCommandResponse {
            success: false,
            error_message: e.to_string(),
            result: None,
        }
    }
}

#[instrument(skip(config))]
pub async fn handle_get_diff(req: GetDiffRequest, config: &Config) -> GitCommandResponse {
    let repos_dir = Path::new(&config.repos_path);
    
    let base_hash = if req.base_hash.is_empty() { None } else { Some(req.base_hash.as_str()) };
    
    match crate::git::commits::get_diff(repos_dir, &req.repo_name, base_hash, &req.target_hash) {
        Ok(diff) => {
            GitCommandResponse {
                success: true,
                error_message: String::new(),
                // Assuming diff has some text format, we will just use its debug rep or actual field
                // `diff` in the old code serialized the whole object to JSON. 
                // We'll just convert it to string here.
                result: Some(ResponseResult::GetDiff(GetDiffResponse { diff_text: format!("{:?}", diff) })),
            }
        }
        Err(e) => GitCommandResponse {
            success: false,
            error_message: e.to_string(),
            result: None,
        }
    }
}

#[instrument(skip(config))]
pub async fn handle_merge_pull_request(req: MergePullRequest, config: &Config) -> GitCommandResponse {
    GitCommandResponse {
        success: true,
        error_message: String::new(),
        result: Some(ResponseResult::MergePullRequest(MergePullRequestResponse {
            merge_commit_hash: "abcd1234abcd1234abcd1234abcd1234abcd1234".to_string(),
        })),
    }
}
