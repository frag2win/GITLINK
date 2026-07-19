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
                protocol_version: 1, error: None,
                
                result: Some(ResponseResult::Push(PushResponse { new_head: "HEAD".to_string() })),
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
pub async fn handle_pull(req: PullRequest, config: &Config) -> GitCommandResponse {
    let repos_dir = Path::new(&config.repos_path);
    
    // In a real implementation, want_oids and have_oids would come from the request
    let want_oids = vec![];
    let have_oids = vec![];
    
    match crate::git::pack::send_pack(repos_dir, &req.repo_name, &want_oids, &have_oids) {
        Ok(result) => {
            GitCommandResponse {
                protocol_version: 1, error: None,
                
                result: Some(ResponseResult::Pull(PullResponse { pack_data: result.pack_data })),
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
                protocol_version: 1, error: None,
                
                result: Some(ResponseResult::ListCommits(ListCommitsResponse { commits: proto_commits, total_count: 0 })), // total_count can be updated later
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
                protocol_version: 1, error: None,
                
                result: Some(ResponseResult::GetCommit(GetCommitResponse { commit: Some(info) })),
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
pub async fn handle_get_diff(req: GetDiffRequest, config: &Config) -> GitCommandResponse {
    let repos_dir = Path::new(&config.repos_path);
    
    let base_hash = if req.base_hash.is_empty() { None } else { Some(req.base_hash.as_str()) };
    
    match crate::git::commits::get_diff(repos_dir, &req.repo_name, base_hash, &req.target_hash, &crate::git::commits::DiffConfig::default()) {
        Ok(diff_text) => {
            GitCommandResponse {
                protocol_version: 1, error: None,
                
                result: Some(ResponseResult::GetDiff(GetDiffResponse { diff_text })),
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
pub async fn handle_merge_pull_request(req: MergePullRequest, config: &Config) -> GitCommandResponse {
    let repos_dir = Path::new(&config.repos_path);
    
    let repo = match crate::git::repository::open(repos_dir, &req.repo_name) {
        Ok(r) => r,
        Err(e) => return GitCommandResponse { protocol_version: 1, error: Some(crate::socket::protocol::GitError { code: "GitError".to_string(), message: e.to_string() }), result: None },
    };

    let repo_path = crate::git::repository::repo_full_path(repos_dir, &req.repo_name);
    let _lock = match crate::git::lock::acquire(&repo_path) {
        Ok(l) => l,
        Err(e) => return GitCommandResponse { protocol_version: 1, error: Some(crate::socket::protocol::GitError { code: "GitError".to_string(), message: e.to_string() }), result: None },
    };

    let engine = crate::git::merge::MergeEngine::new(&repo);
    match engine.execute(&req.base_branch, &req.head_branch, &req.author_name, &req.author_email, &req.commit_message) {
        Ok(crate::git::merge::MergeResult::Success { new_head }) => {
            GitCommandResponse {
                protocol_version: 1, error: None,
                
                result: Some(ResponseResult::MergePullRequest(MergePullRequestResponse {
                    outcome: Some(merge_pull_request_response::Outcome::MergeCommitHash(new_head)),
                })),
            }
        }
        Ok(crate::git::merge::MergeResult::Conflicts(conflicts)) => {
            let proto_conflicts = conflicts.into_iter().map(|c| MergeConflict {
                path: c.path,
                conflict_type: format!("{:?}", c.conflict_type),
                base_oid: c.base_oid,
                ours_oid: c.ours_oid,
                theirs_oid: c.theirs_oid,
                is_binary: c.is_binary,
            }).collect();

            GitCommandResponse {
                protocol_version: 1, error: None,
                result: Some(ResponseResult::MergePullRequest(MergePullRequestResponse {
                    outcome: Some(merge_pull_request_response::Outcome::Conflicts(MergeConflictList {
                        conflicts: proto_conflicts,
                    })),
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
pub async fn handle_create_branch(req: CreateBranchRequest, config: &Config) -> GitCommandResponse {
    let repos_dir = Path::new(&config.repos_path);
    if let Err(e) = crate::git::sanitize::validate_repo_name(&req.repo_name) {
        return GitCommandResponse { protocol_version: 1, error: Some(crate::socket::protocol::GitError { code: "InvalidRepoName".to_string(), message: e.to_string() }), result: None };
    }
    if let Err(e) = crate::git::sanitize::validate_branch_name(&req.branch_name) {
        return GitCommandResponse { protocol_version: 1, error: Some(crate::socket::protocol::GitError { code: "InvalidBranchName".to_string(), message: e.to_string() }), result: None };
    }
    if let Err(e) = crate::git::sanitize::validate_revision(&req.target_commit_id) {
        return GitCommandResponse { protocol_version: 1, error: Some(crate::socket::protocol::GitError { code: "InvalidRevision".to_string(), message: e.to_string() }), result: None };
    }

    match crate::git::refs::create_branch(repos_dir, &req.repo_name, &req.branch_name, &req.target_commit_id) {
        Ok(b) => {
            let info = BranchInfo {
                name: b.name,
                is_head: b.is_head,
                commit_id: b.commit_id,
                commit_summary: b.commit_summary.unwrap_or_default(),
                committed_at: b.committed_at.unwrap_or_default(),
            };
            GitCommandResponse {
                protocol_version: 1, error: None,
                result: Some(ResponseResult::CreateBranch(CreateBranchResponse { branch: Some(info) })),
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
pub async fn handle_delete_branch(req: DeleteBranchRequest, config: &Config) -> GitCommandResponse {
    let repos_dir = Path::new(&config.repos_path);
    if let Err(e) = crate::git::sanitize::validate_repo_name(&req.repo_name) {
        return GitCommandResponse { protocol_version: 1, error: Some(crate::socket::protocol::GitError { code: "InvalidRepoName".to_string(), message: e.to_string() }), result: None };
    }
    if let Err(e) = crate::git::sanitize::validate_branch_name(&req.branch_name) {
        return GitCommandResponse { protocol_version: 1, error: Some(crate::socket::protocol::GitError { code: "InvalidBranchName".to_string(), message: e.to_string() }), result: None };
    }

    match crate::git::refs::delete_branch(repos_dir, &req.repo_name, &req.branch_name) {
        Ok(_) => GitCommandResponse {
            protocol_version: 1, error: None,
            result: Some(ResponseResult::DeleteBranch(DeleteBranchResponse {})),
        },
        Err(e) => GitCommandResponse {
            protocol_version: 1,
            error: Some(crate::socket::protocol::GitError { code: "GitError".to_string(), message: e.to_string() }),
            result: None,
        }
    }
}

#[instrument(skip(config))]
pub async fn handle_list_branches(req: ListBranchesRequest, config: &Config) -> GitCommandResponse {
    let repos_dir = Path::new(&config.repos_path);
    if let Err(e) = crate::git::sanitize::validate_repo_name(&req.repo_name) {
        return GitCommandResponse { protocol_version: 1, error: Some(crate::socket::protocol::GitError { code: "InvalidRepoName".to_string(), message: e.to_string() }), result: None };
    }

    match crate::git::refs::list_branches(repos_dir, &req.repo_name) {
        Ok(branches) => {
            let proto_branches = branches.into_iter().map(|b| BranchInfo {
                name: b.name,
                is_head: b.is_head,
                commit_id: b.commit_id,
                commit_summary: b.commit_summary.unwrap_or_default(),
                committed_at: b.committed_at.unwrap_or_default(),
            }).collect();
            GitCommandResponse {
                protocol_version: 1, error: None,
                result: Some(ResponseResult::ListBranches(ListBranchesResponse { branches: proto_branches })),
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
pub async fn handle_get_branch(req: GetBranchRequest, config: &Config) -> GitCommandResponse {
    let repos_dir = Path::new(&config.repos_path);
    if let Err(e) = crate::git::sanitize::validate_repo_name(&req.repo_name) {
        return GitCommandResponse { protocol_version: 1, error: Some(crate::socket::protocol::GitError { code: "InvalidRepoName".to_string(), message: e.to_string() }), result: None };
    }
    if let Err(e) = crate::git::sanitize::validate_branch_name(&req.branch_name) {
        return GitCommandResponse { protocol_version: 1, error: Some(crate::socket::protocol::GitError { code: "InvalidBranchName".to_string(), message: e.to_string() }), result: None };
    }

    match crate::git::refs::list_branches(repos_dir, &req.repo_name) {
        Ok(branches) => {
            if let Some(b) = branches.into_iter().find(|x| x.name == req.branch_name) {
                let info = BranchInfo {
                    name: b.name,
                    is_head: b.is_head,
                    commit_id: b.commit_id,
                    commit_summary: b.commit_summary.unwrap_or_default(),
                    committed_at: b.committed_at.unwrap_or_default(),
                };
                GitCommandResponse {
                    protocol_version: 1, error: None,
                    result: Some(ResponseResult::GetBranch(GetBranchResponse { branch: Some(info) })),
                }
            } else {
                GitCommandResponse {
                    protocol_version: 1,
                    error: Some(crate::socket::protocol::GitError { code: "BranchNotFound".to_string(), message: format!("branch '{}' not found", req.branch_name) }),
                    result: None,
                }
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
pub async fn handle_create_tag(req: CreateTagRequest, config: &Config) -> GitCommandResponse {
    let repos_dir = Path::new(&config.repos_path);
    if let Err(e) = crate::git::sanitize::validate_repo_name(&req.repo_name) {
        return GitCommandResponse { protocol_version: 1, error: Some(crate::socket::protocol::GitError { code: "InvalidRepoName".to_string(), message: e.to_string() }), result: None };
    }
    if let Err(e) = crate::git::sanitize::validate_tag_name(&req.tag_name) {
        return GitCommandResponse { protocol_version: 1, error: Some(crate::socket::protocol::GitError { code: "InvalidTagName".to_string(), message: e.to_string() }), result: None };
    }
    if let Err(e) = crate::git::sanitize::validate_revision(&req.target_commit_id) {
        return GitCommandResponse { protocol_version: 1, error: Some(crate::socket::protocol::GitError { code: "InvalidRevision".to_string(), message: e.to_string() }), result: None };
    }

    let msg_opt = if req.message.is_empty() { None } else { Some(req.message.as_str()) };
    let tagger_opt = if req.tagger_name.is_empty() { None } else { Some(req.tagger_name.as_str()) };
    let email_opt = if req.tagger_email.is_empty() { None } else { Some(req.tagger_email.as_str()) };

    match crate::git::refs::create_tag(repos_dir, &req.repo_name, &req.tag_name, &req.target_commit_id, msg_opt, tagger_opt, email_opt) {
        Ok(t) => {
            let info = TagInfo {
                name: t.name,
                target_id: t.target_id,
                is_annotated: t.is_annotated,
                message: t.message.unwrap_or_default(),
                tagger_name: t.tagger_name.unwrap_or_default(),
                tagged_at: t.tagged_at.unwrap_or_default(),
            };
            GitCommandResponse {
                protocol_version: 1, error: None,
                result: Some(ResponseResult::CreateTag(CreateTagResponse { tag: Some(info) })),
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
pub async fn handle_delete_tag(req: DeleteTagRequest, config: &Config) -> GitCommandResponse {
    let repos_dir = Path::new(&config.repos_path);
    if let Err(e) = crate::git::sanitize::validate_repo_name(&req.repo_name) {
        return GitCommandResponse { protocol_version: 1, error: Some(crate::socket::protocol::GitError { code: "InvalidRepoName".to_string(), message: e.to_string() }), result: None };
    }
    if let Err(e) = crate::git::sanitize::validate_tag_name(&req.tag_name) {
        return GitCommandResponse { protocol_version: 1, error: Some(crate::socket::protocol::GitError { code: "InvalidTagName".to_string(), message: e.to_string() }), result: None };
    }

    match crate::git::refs::delete_tag(repos_dir, &req.repo_name, &req.tag_name) {
        Ok(_) => GitCommandResponse {
            protocol_version: 1, error: None,
            result: Some(ResponseResult::DeleteTag(DeleteTagResponse {})),
        },
        Err(e) => GitCommandResponse {
            protocol_version: 1,
            error: Some(crate::socket::protocol::GitError { code: "GitError".to_string(), message: e.to_string() }),
            result: None,
        }
    }
}

#[instrument(skip(config))]
pub async fn handle_list_tags(req: ListTagsRequest, config: &Config) -> GitCommandResponse {
    let repos_dir = Path::new(&config.repos_path);
    if let Err(e) = crate::git::sanitize::validate_repo_name(&req.repo_name) {
        return GitCommandResponse { protocol_version: 1, error: Some(crate::socket::protocol::GitError { code: "InvalidRepoName".to_string(), message: e.to_string() }), result: None };
    }

    match crate::git::refs::list_tags(repos_dir, &req.repo_name) {
        Ok(tags) => {
            let proto_tags = tags.into_iter().map(|t| TagInfo {
                name: t.name,
                target_id: t.target_id,
                is_annotated: t.is_annotated,
                message: t.message.unwrap_or_default(),
                tagger_name: t.tagger_name.unwrap_or_default(),
                tagged_at: t.tagged_at.unwrap_or_default(),
            }).collect();
            GitCommandResponse {
                protocol_version: 1, error: None,
                result: Some(ResponseResult::ListTags(ListTagsResponse { tags: proto_tags })),
            }
        }
        Err(e) => GitCommandResponse {
            protocol_version: 1,
            error: Some(crate::socket::protocol::GitError { code: "GitError".to_string(), message: e.to_string() }),
            result: None,
        }
    }
}

