//! # Object Operations
//!
//! Read Git objects from the repository's object database:
//! - [`read_blob`] — Read the content of a blob (file).
//! - [`read_tree`] — Read the entries of a tree (directory).
//! - [`read_commit`] — Read commit metadata by OID.
//! - [`walk_tree`] — Recursively walk a tree for file browser support.

use std::path::Path;

use tracing::{debug, instrument};

use crate::error::GitError;
use crate::models::{CommitInfo, FileEntry, TreeEntry, TreeEntryKind};

/// Read the raw content of a blob object by OID.
///
/// Returns the blob content as bytes. For text files, callers can
/// convert to UTF-8; binary detection is left to the caller.
///
/// # Errors
/// Returns [`GitError::ObjectNotFound`] if the OID does not exist.
#[instrument(skip(repos_dir))]
pub fn read_blob(repos_dir: &Path, repo_name: &str, oid_str: &str) -> Result<Vec<u8>, GitError> {
    let repo = crate::git::repository::open(repos_dir, repo_name)?;

    let oid = git2::Oid::from_str(oid_str).map_err(|_| GitError::ObjectNotFound {
        oid: oid_str.to_string(),
    })?;

    let blob = repo.find_blob(oid).map_err(|_| GitError::ObjectNotFound {
        oid: oid_str.to_string(),
    })?;

    debug!(oid = %oid_str, size = blob.size(), "Read blob");
    Ok(blob.content().to_vec())
}

/// Read the entries of a tree object by OID.
///
/// Returns a flat list of [`TreeEntry`] items (not recursive).
///
/// # Errors
/// Returns [`GitError::ObjectNotFound`] if the OID does not exist.
#[instrument(skip(repos_dir))]
pub fn read_tree(
    repos_dir: &Path,
    repo_name: &str,
    oid_str: &str,
) -> Result<Vec<TreeEntry>, GitError> {
    let repo = crate::git::repository::open(repos_dir, repo_name)?;

    let oid = git2::Oid::from_str(oid_str).map_err(|_| GitError::ObjectNotFound {
        oid: oid_str.to_string(),
    })?;

    let tree = repo.find_tree(oid).map_err(|_| GitError::ObjectNotFound {
        oid: oid_str.to_string(),
    })?;

    let mut entries = Vec::with_capacity(tree.len());

    for entry in tree.iter() {
        let object_type = match entry.kind() {
            Some(git2::ObjectType::Blob) => TreeEntryKind::Blob,
            Some(git2::ObjectType::Tree) => TreeEntryKind::Tree,
            Some(git2::ObjectType::Commit) => TreeEntryKind::Commit,
            _ => continue,
        };

        let size = if object_type == TreeEntryKind::Blob {
            repo.find_blob(entry.id()).ok().map(|b| b.size() as u64)
        } else {
            None
        };

        entries.push(TreeEntry {
            name: entry.name().unwrap_or("").to_string(),
            object_type,
            oid: entry.id().to_string(),
            filemode: entry.filemode() as u32,
            size,
        });
    }

    debug!(oid = %oid_str, count = entries.len(), "Read tree");
    Ok(entries)
}

/// Read commit metadata by OID.
///
/// # Errors
/// Returns [`GitError::ObjectNotFound`] if the commit does not exist.
#[instrument(skip(repos_dir))]
pub fn read_commit(
    repos_dir: &Path,
    repo_name: &str,
    oid_str: &str,
) -> Result<CommitInfo, GitError> {
    let repo = crate::git::repository::open(repos_dir, repo_name)?;

    let oid = git2::Oid::from_str(oid_str).map_err(|_| GitError::ObjectNotFound {
        oid: oid_str.to_string(),
    })?;

    let commit = repo.find_commit(oid).map_err(|_| GitError::ObjectNotFound {
        oid: oid_str.to_string(),
    })?;

    let author = commit.author();
    let committer = commit.committer();

    let parent_ids: Vec<String> = commit.parent_ids().map(|id| id.to_string()).collect();

    debug!(oid = %oid_str, "Read commit");

    Ok(CommitInfo {
        id: commit.id().to_string(),
        summary: commit.summary().unwrap_or("").to_string(),
        message: commit.message().unwrap_or("").to_string(),
        author_name: author.name().unwrap_or("").to_string(),
        author_email: author.email().unwrap_or("").to_string(),
        authored_at: format!("{}", author.when().seconds()),
        committer_name: committer.name().unwrap_or("").to_string(),
        committer_email: committer.email().unwrap_or("").to_string(),
        committed_at: format!("{}", committer.when().seconds()),
        parent_ids,
        tree_id: commit.tree_id().to_string(),
    })
}

/// Recursively walk a tree to produce a flat list of file entries.
///
/// Used by the file browser to display the full directory structure
/// of a repository at a given commit.
///
/// # Parameters
/// - `tree_oid_str` — OID of the root tree to walk.
/// - `base_path` — Prefix for relative paths (use `""` for the root).
#[instrument(skip(repos_dir))]
pub fn walk_tree(
    repos_dir: &Path,
    repo_name: &str,
    tree_oid_str: &str,
    base_path: &str,
) -> Result<Vec<FileEntry>, GitError> {
    let repo = crate::git::repository::open(repos_dir, repo_name)?;

    let oid = git2::Oid::from_str(tree_oid_str).map_err(|_| GitError::ObjectNotFound {
        oid: tree_oid_str.to_string(),
    })?;

    let tree = repo.find_tree(oid).map_err(|_| GitError::ObjectNotFound {
        oid: tree_oid_str.to_string(),
    })?;

    let mut file_entries = Vec::new();

    tree.walk(git2::TreeWalkMode::PreOrder, |dir, entry| {
        let name = entry.name().unwrap_or("").to_string();
        let path = if dir.is_empty() {
            if base_path.is_empty() {
                name.clone()
            } else {
                format!("{base_path}/{name}")
            }
        } else if base_path.is_empty() {
            format!("{dir}{name}")
        } else {
            format!("{base_path}/{dir}{name}")
        };

        let is_dir = entry.kind() == Some(git2::ObjectType::Tree);

        let size = if !is_dir {
            repo.find_blob(entry.id()).ok().map(|b| b.size() as u64)
        } else {
            None
        };

        file_entries.push(FileEntry {
            path,
            name,
            is_dir,
            size,
            oid: entry.id().to_string(),
        });

        git2::TreeWalkResult::Ok
    })
    .map_err(GitError::from)?;

    debug!(tree = %tree_oid_str, count = file_entries.len(), "Walked tree");
    Ok(file_entries)
}
