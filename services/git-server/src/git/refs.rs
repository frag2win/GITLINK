//! # Reference Management
//!
//! Operations on Git references — branches and tags.
//! - [`list_branches`] — Enumerate all local branches.
//! - [`create_branch`] — Create a new branch from a commit.
//! - [`delete_branch`] — Delete an existing branch.
//! - [`list_tags`] — Enumerate all tags (lightweight and annotated).

use std::path::Path;

use tracing::{debug, info, instrument};

use crate::error::GitError;
use crate::models::{BranchInfo, TagInfo};

/// List all local branches in the repository.
///
/// Returns a `Vec<BranchInfo>` with branch name, HEAD status, and tip commit info.
#[instrument(skip(repos_dir))]
pub fn list_branches(repos_dir: &Path, repo_name: &str) -> Result<Vec<BranchInfo>, GitError> {
    let repo = crate::git::repository::open(repos_dir, repo_name)?;

    let mut branches = Vec::new();
    let head_oid = repo.head().ok().and_then(|h| h.target());

    for branch_result in repo.branches(Some(git2::BranchType::Local))? {
        let (branch, _branch_type) = branch_result?;
        let reference = branch.get();

        let name = branch
            .name()?
            .unwrap_or("unnamed")
            .to_string();

        let commit_id = reference
            .target()
            .map(|oid| oid.to_string())
            .unwrap_or_default();

        let is_head = reference.target() == head_oid;

        let (commit_summary, committed_at) = reference
            .peel_to_commit()
            .ok()
            .map(|commit| {
                let summary = commit.summary().map(String::from);
                let time = commit.time();
                let timestamp = format!("{}", time.seconds());
                (summary, Some(timestamp))
            })
            .unwrap_or((None, None));

        branches.push(BranchInfo {
            name,
            is_head,
            commit_id,
            commit_summary,
            committed_at,
        });
    }

    debug!(count = branches.len(), repo = %repo_name, "Listed branches");
    Ok(branches)
}

/// Create a new branch pointing to the given commit OID.
///
/// # Parameters
/// - `branch_name` — Name for the new branch.
/// - `target_commit_id` — Hex OID of the commit to point to.
///
/// # Errors
/// Returns [`GitError::ObjectNotFound`] if the target commit does not exist.
#[instrument(skip(repos_dir))]
pub fn create_branch(
    repos_dir: &Path,
    repo_name: &str,
    branch_name: &str,
    target_commit_id: &str,
) -> Result<BranchInfo, GitError> {
    let repo = crate::git::repository::open(repos_dir, repo_name)?;

    let oid = git2::Oid::from_str(target_commit_id).map_err(|_| GitError::ObjectNotFound {
        oid: target_commit_id.to_string(),
    })?;

    let commit = repo.find_commit(oid).map_err(|_| GitError::ObjectNotFound {
        oid: target_commit_id.to_string(),
    })?;

    let _branch = repo.branch(branch_name, &commit, false)?;

    info!(
        repo = %repo_name,
        branch = %branch_name,
        target = %target_commit_id,
        "Created branch"
    );

    Ok(BranchInfo {
        name: branch_name.to_string(),
        is_head: false,
        commit_id: target_commit_id.to_string(),
        commit_summary: commit.summary().map(String::from),
        committed_at: Some(format!("{}", commit.time().seconds())),
    })
}

/// Delete a branch by name.
///
/// # Errors
/// Returns [`GitError::RefNotFound`] if the branch does not exist.
#[instrument(skip(repos_dir))]
pub fn delete_branch(
    repos_dir: &Path,
    repo_name: &str,
    branch_name: &str,
) -> Result<(), GitError> {
    let repo = crate::git::repository::open(repos_dir, repo_name)?;

    let mut branch = repo
        .find_branch(branch_name, git2::BranchType::Local)
        .map_err(|_| GitError::RefNotFound {
            name: branch_name.to_string(),
        })?;

    branch.delete()?;

    info!(repo = %repo_name, branch = %branch_name, "Deleted branch");
    Ok(())
}

/// List all tags in the repository (both lightweight and annotated).
#[instrument(skip(repos_dir))]
pub fn list_tags(repos_dir: &Path, repo_name: &str) -> Result<Vec<TagInfo>, GitError> {
    let repo = crate::git::repository::open(repos_dir, repo_name)?;

    let mut tags = Vec::new();
    let tag_names = repo.tag_names(None)?;

    for name in tag_names.iter().flatten() {
        let refname = format!("refs/tags/{name}");
        let reference = repo.find_reference(&refname)?;
        let target_id = reference
            .target()
            .map(|oid| oid.to_string())
            .unwrap_or_default();

        // Check if this is an annotated tag
        let (is_annotated, message, tagger_name, tagged_at) =
            if let Ok(tag) = reference.peel_to_tag() {
                (
                    true,
                    tag.message().map(String::from),
                    tag.tagger().and_then(|t| t.name().map(String::from)),
                    tag.tagger().map(|t| format!("{}", t.when().seconds())),
                )
            } else {
                (false, None, None, None)
            };

        tags.push(TagInfo {
            name: name.to_string(),
            target_id,
            is_annotated,
            message,
            tagger_name,
            tagged_at,
        });
    }

    debug!(count = tags.len(), repo = %repo_name, "Listed tags");
    Ok(tags)
}
