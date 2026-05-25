//! # Commit Operations
//!
//! Higher-level commit operations that walk history and compute diffs:
//! - [`list_commits`] — Walk commit history from a starting point.
//! - [`get_commit_detail`] — Full commit metadata with parent info.
//! - [`get_diff`] — Compute the diff between two commits (or a commit and its parent).

use std::path::Path;

use tracing::{debug, instrument};

use crate::error::GitError;
use crate::models::{CommitInfo, DiffFileEntry, DiffInfo};

/// Walk commit history starting from `start_oid_str`.
///
/// Returns up to `limit` commits in reverse chronological order.
/// If `start_oid_str` is `None`, starts from HEAD.
///
/// # Parameters
/// - `limit` — Maximum number of commits to return.
/// - `offset` — Number of commits to skip before collecting.
#[instrument(skip(repos_dir))]
pub fn list_commits(
    repos_dir: &Path,
    repo_name: &str,
    start_oid_str: Option<&str>,
    limit: usize,
    offset: usize,
) -> Result<Vec<CommitInfo>, GitError> {
    let repo = crate::git::repository::open(repos_dir, repo_name)?;

    let start_oid = match start_oid_str {
        Some(s) => git2::Oid::from_str(s).map_err(|_| GitError::ObjectNotFound {
            oid: s.to_string(),
        })?,
        None => {
            // Default to HEAD
            repo.head()
                .map_err(|_| GitError::RefNotFound {
                    name: "HEAD".to_string(),
                })?
                .target()
                .ok_or_else(|| GitError::RefNotFound {
                    name: "HEAD".to_string(),
                })?
        }
    };

    let mut revwalk = repo.revwalk()?;
    revwalk.set_sorting(git2::Sort::TIME)?;
    revwalk.push(start_oid)?;

    let mut commits = Vec::with_capacity(limit);

    for (i, oid_result) in revwalk.enumerate() {
        if i < offset {
            continue;
        }
        if commits.len() >= limit {
            break;
        }

        let oid = oid_result?;
        let commit = repo.find_commit(oid)?;
        let author = commit.author();
        let committer = commit.committer();

        commits.push(CommitInfo {
            id: commit.id().to_string(),
            summary: commit.summary().unwrap_or("").to_string(),
            message: commit.message().unwrap_or("").to_string(),
            author_name: author.name().unwrap_or("").to_string(),
            author_email: author.email().unwrap_or("").to_string(),
            authored_at: format!("{}", author.when().seconds()),
            committer_name: committer.name().unwrap_or("").to_string(),
            committer_email: committer.email().unwrap_or("").to_string(),
            committed_at: format!("{}", committer.when().seconds()),
            parent_ids: commit.parent_ids().map(|id| id.to_string()).collect(),
            tree_id: commit.tree_id().to_string(),
        });
    }

    debug!(
        repo = %repo_name,
        count = commits.len(),
        offset = offset,
        limit = limit,
        "Listed commits"
    );

    Ok(commits)
}

/// Get detailed information about a specific commit.
///
/// This is a convenience wrapper around [`crate::git::objects::read_commit`].
#[instrument(skip(repos_dir))]
pub fn get_commit_detail(
    repos_dir: &Path,
    repo_name: &str,
    oid_str: &str,
) -> Result<CommitInfo, GitError> {
    crate::git::objects::read_commit(repos_dir, repo_name, oid_str)
}

/// Compute the diff between two commits.
///
/// If `from_oid_str` is `None`, diffs against the first parent of `to_oid_str`.
/// Returns a [`DiffInfo`] with file-level and aggregate statistics.
#[instrument(skip(repos_dir))]
pub fn get_diff(
    repos_dir: &Path,
    repo_name: &str,
    from_oid_str: Option<&str>,
    to_oid_str: &str,
) -> Result<DiffInfo, GitError> {
    let repo = crate::git::repository::open(repos_dir, repo_name)?;

    // Resolve the "to" commit and its tree
    let to_oid = git2::Oid::from_str(to_oid_str).map_err(|_| GitError::ObjectNotFound {
        oid: to_oid_str.to_string(),
    })?;
    let to_commit = repo.find_commit(to_oid)?;
    let to_tree = to_commit.tree()?;

    // Resolve the "from" tree (either explicit or first parent)
    let from_tree = match from_oid_str {
        Some(s) => {
            let from_oid = git2::Oid::from_str(s).map_err(|_| GitError::ObjectNotFound {
                oid: s.to_string(),
            })?;
            let from_commit = repo.find_commit(from_oid)?;
            Some(from_commit.tree()?)
        }
        None => to_commit.parent(0).ok().and_then(|p| p.tree().ok()),
    };

    let diff = repo.diff_tree_to_tree(
        from_tree.as_ref(),
        Some(&to_tree),
        None,
    )?;

    let stats = diff.stats()?;
    let mut files = Vec::new();

    diff.foreach(
        &mut |delta, _progress| {
            let path = delta
                .new_file()
                .path()
                .and_then(|p| p.to_str())
                .unwrap_or("")
                .to_string();

            let old_path = delta
                .old_file()
                .path()
                .and_then(|p| p.to_str())
                .map(String::from);

            let status = match delta.status() {
                git2::Delta::Added => "added",
                git2::Delta::Deleted => "deleted",
                git2::Delta::Modified => "modified",
                git2::Delta::Renamed => "renamed",
                git2::Delta::Copied => "copied",
                _ => "unknown",
            }
            .to_string();

            files.push(DiffFileEntry {
                path,
                old_path,
                status,
                insertions: 0,
                deletions: 0,
                patch: None,
            });

            true
        },
        None,
        None,
        None,
    )?;

    debug!(
        repo = %repo_name,
        to = %to_oid_str,
        files_changed = files.len(),
        "Computed diff"
    );

    Ok(DiffInfo {
        files_changed: stats.files_changed(),
        insertions: stats.insertions(),
        deletions: stats.deletions(),
        files,
    })
}
