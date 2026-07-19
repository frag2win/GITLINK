//! # Repository Lock
//!
//! Provides a simple, atomic filesystem-based lock for Git repositories
//! to prevent concurrent mutations (e.g., branching, merging, pushing).

use std::fs::{File, OpenOptions};
use std::path::{Path, PathBuf};
use tracing::{debug, instrument};

use crate::error::GitError;

/// A lock guard that removes the lock file when dropped.
pub struct RepoLock {
    path: PathBuf,
}

impl Drop for RepoLock {
    fn drop(&mut self) {
        if let Err(e) = std::fs::remove_file(&self.path) {
            tracing::warn!(path = %self.path.display(), "Failed to release repository lock: {}", e);
        } else {
            debug!(path = %self.path.display(), "Released repository lock");
        }
    }
}

/// Acquire an exclusive lock on the repository.
/// 
/// This creates a `.gitlink.lock` file in the bare repository directory.
/// Returns a `RepoLock` guard that automatically deletes the file when dropped.
/// 
/// # Errors
/// Returns `GitError::RepoLocked` if another process holds the lock.
#[instrument(skip(repo_path))]
pub fn acquire(repo_path: &Path) -> Result<RepoLock, GitError> {
    let lock_path = repo_path.join(".gitlink.lock");

    // Atomic creation: create_new(true) fails if the file already exists.
    match OpenOptions::new().write(true).create_new(true).open(&lock_path) {
        Ok(_file) => {
            debug!(path = %lock_path.display(), "Acquired repository lock");
            Ok(RepoLock { path: lock_path })
        }
        Err(e) if e.kind() == std::io::ErrorKind::AlreadyExists => {
            Err(GitError::RepoLocked { path: repo_path.display().to_string() })
        }
        Err(e) => {
            Err(GitError::Other(format!("Failed to create lock file: {}", e)))
        }
    }
}
