//! # Repository Operations
//!
//! Manages the lifecycle of bare Git repositories on disk:
//! - [`init_bare`] — Create a new bare repository.
//! - [`open`] — Open an existing repository.
//! - [`list_repos`] — Enumerate all repositories in the repos directory.
//! - [`delete_repo`] — Remove a repository from disk.

use std::path::{Path, PathBuf};

use git2::Repository;
use tracing::{debug, info, instrument};

use crate::error::GitError;
use crate::models::RepoInfo;

/// Initialize a new bare Git repository.
///
/// Creates a bare repo at `<repos_dir>/<name>.git`.
///
/// # Errors
/// Returns [`GitError::RepoAlreadyExists`] if the directory already exists.
#[instrument(skip(repos_dir))]
pub fn init_bare(repos_dir: &Path, name: &str) -> Result<RepoInfo, GitError> {
    let repo_path = repo_full_path(repos_dir, name);

    if repo_path.exists() {
        return Err(GitError::RepoAlreadyExists {
            path: repo_path.display().to_string(),
        });
    }

    info!(name = %name, path = %repo_path.display(), "Initializing bare repository");

    let repo = Repository::init_bare(&repo_path)?;

    // Install pre-receive hook
    let hook_path = repo_path.join("hooks").join("pre-receive");
    let hook_script = include_str!("../../scripts/pre-receive");
    std::fs::write(&hook_path, hook_script).map_err(|e| GitError::Other(e.to_string()))?;
    
    // Make the hook executable (Unix-specific)
    #[cfg(unix)]
    {
        use std::os::unix::fs::PermissionsExt;
        let mut perms = std::fs::metadata(&hook_path)
            .map_err(|e| GitError::Other(e.to_string()))?
            .permissions();
        perms.set_mode(0o755);
        std::fs::set_permissions(&hook_path, perms).map_err(|e| GitError::Other(e.to_string()))?;
    }

    Ok(build_repo_info(name, &repo_path, &repo))
}

/// Open an existing bare Git repository by name.
///
/// # Errors
/// Returns [`GitError::RepoNotFound`] if the directory does not exist.
#[instrument(skip(repos_dir))]
pub fn open(repos_dir: &Path, name: &str) -> Result<Repository, GitError> {
    let repo_path = repo_full_path(repos_dir, name);

    if !repo_path.exists() {
        return Err(GitError::RepoNotFound {
            path: repo_path.display().to_string(),
        });
    }

    debug!(name = %name, "Opening repository");
    let repo = Repository::open_bare(&repo_path)?;
    Ok(repo)
}

/// List all bare repositories in the repos directory.
///
/// Scans `repos_dir` for directories ending in `.git` and returns
/// metadata for each.
#[instrument(skip(repos_dir))]
pub fn list_repos(repos_dir: &Path) -> Result<Vec<RepoInfo>, GitError> {
    let mut repos = Vec::new();

    let entries = std::fs::read_dir(repos_dir).map_err(|e| GitError::Other(e.to_string()))?;

    for entry in entries {
        let entry = entry.map_err(|e| GitError::Other(e.to_string()))?;
        let path = entry.path();

        if path.is_dir() {
            // Try to open as a Git repo
            if let Ok(repo) = Repository::open_bare(&path) {
                let name = path
                    .file_name()
                    .and_then(|n| n.to_str())
                    .unwrap_or("unknown")
                    .trim_end_matches(".git")
                    .to_string();

                repos.push(build_repo_info(&name, &path, &repo));
            }
        }
    }

    debug!(count = repos.len(), "Listed repositories");
    Ok(repos)
}

/// Delete a repository from disk.
///
/// # Errors
/// Returns [`GitError::RepoNotFound`] if the repository does not exist.
#[instrument(skip(repos_dir))]
pub fn delete_repo(repos_dir: &Path, name: &str) -> Result<(), GitError> {
    let repo_path = repo_full_path(repos_dir, name);

    if !repo_path.exists() {
        return Err(GitError::RepoNotFound {
            path: repo_path.display().to_string(),
        });
    }

    info!(name = %name, path = %repo_path.display(), "Deleting repository");
    std::fs::remove_dir_all(&repo_path).map_err(|e| GitError::Other(e.to_string()))?;
    Ok(())
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

/// Construct the full filesystem path for a repository.
fn repo_full_path(repos_dir: &Path, name: &str) -> PathBuf {
    let dir_name = if name.ends_with(".git") {
        name.to_string()
    } else {
        format!("{name}.git")
    };
    repos_dir.join(dir_name)
}

/// Build a [`RepoInfo`] from an opened repository.
fn build_repo_info(name: &str, path: &Path, repo: &Repository) -> RepoInfo {
    let default_branch = repo
        .head()
        .ok()
        .and_then(|head| head.shorthand().map(String::from));

    let branch_count = repo
        .branches(Some(git2::BranchType::Local))
        .map(|branches| branches.count())
        .unwrap_or(0);

    let tag_count = repo
        .tag_names(None)
        .map(|tags| tags.len())
        .unwrap_or(0);

    let last_commit_at = repo
        .head()
        .ok()
        .and_then(|head| head.peel_to_commit().ok())
        .map(|commit| {
            let time = commit.time();
            // Simple ISO-8601 representation
            format!("{}+{:02}:{:02}", time.seconds(), time.offset_minutes() / 60, time.offset_minutes() % 60)
        });

    RepoInfo {
        name: name.to_string(),
        path: path.display().to_string(),
        is_bare: repo.is_bare(),
        default_branch,
        last_commit_at,
        branch_count,
        tag_count,
    }
}
