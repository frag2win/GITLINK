//! # Repository Handler
//!
//! Handles repository CRUD requests over the Protobuf socket.

use std::path::Path;
use tracing::instrument;
use crate::config::Config;
use crate::socket::protocol::*;
use crate::socket::protocol::git_command_response::Result as ResponseResult;

#[instrument(skip(config))]
pub async fn handle_create(req: CreateRepoRequest, config: &Config) -> GitCommandResponse {
    let repos_dir = Path::new(&config.repos_path);
    
    match crate::git::repository::init_bare(repos_dir, &req.name) {
        Ok(repo_info) => {
            let info = RepoInfo {
                name: repo_info.name,
                path: repo_info.path,
                is_bare: repo_info.is_bare,
                default_branch: repo_info.default_branch.unwrap_or_default(),
                last_commit_at: repo_info.last_commit_at.unwrap_or_default(),
                branch_count: repo_info.branch_count as i32,
                tag_count: repo_info.tag_count as i32,
            };
            
            GitCommandResponse {
                success: true,
                error_message: String::new(),
                result: Some(ResponseResult::CreateRepo(CreateRepoResponse { repo: Some(info) })),
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
pub async fn handle_delete(req: DeleteRepoRequest, config: &Config) -> GitCommandResponse {
    let repos_dir = Path::new(&config.repos_path);
    
    match crate::git::repository::delete_repo(repos_dir, &req.name) {
        Ok(_) => GitCommandResponse {
            success: true,
            error_message: String::new(),
            result: Some(ResponseResult::DeleteRepo(DeleteRepoResponse {})),
        },
        Err(e) => GitCommandResponse {
            success: false,
            error_message: e.to_string(),
            result: None,
        }
    }
}

#[instrument(skip(config))]
pub async fn handle_list(_req: ListReposRequest, config: &Config) -> GitCommandResponse {
    let repos_dir = Path::new(&config.repos_path);
    
    match crate::git::repository::list_repos(repos_dir) {
        Ok(repos) => {
            let proto_repos = repos.into_iter().map(|r| RepoInfo {
                name: r.name,
                path: r.path,
                is_bare: r.is_bare,
                default_branch: r.default_branch.unwrap_or_default(),
                last_commit_at: r.last_commit_at.unwrap_or_default(),
                branch_count: r.branch_count as i32,
                tag_count: r.tag_count as i32,
            }).collect();
            
            GitCommandResponse {
                success: true,
                error_message: String::new(),
                result: Some(ResponseResult::ListRepos(ListReposResponse { repos: proto_repos })),
            }
        }
        Err(e) => GitCommandResponse {
            success: false,
            error_message: e.to_string(),
            result: None,
        }
    }
}
