//! # Browse Handler
//!
//! Handles file browsing requests received over the Unix domain socket:
//! - `BrowseTree` — List entries of a tree (directory) at a given commit/path.
//! - `ReadFile` — Read the content of a specific file (blob).
//!
//! These operations power the web UI's repository file browser.

use std::path::Path;

use anyhow::{Context, Result};
use tracing::instrument;

use crate::config::Config;
use crate::socket::protocol::{GitOperation, RepoRequest, RepoResponse};

/// Dispatch a file browsing request.
#[instrument(skip(config))]
pub async fn handle(request: &RepoRequest, config: &Config) -> Result<RepoResponse> {
    let repos_dir = Path::new(&config.repos_path);
    let repo_name = request
        .repo_name
        .as_deref()
        .context("repo_name is required for browse operations")?;

    match request.operation {
        GitOperation::BrowseTree => handle_browse_tree(request, repos_dir, repo_name),
        GitOperation::ReadFile => handle_read_file(request, repos_dir, repo_name),
        _ => Ok(RepoResponse::err(
            request.request_id.clone(),
            format!("unsupported operation for browse_handler: {:?}", request.operation),
        )),
    }
}

/// Handle `BrowseTree`: list directory entries at a given path and commit.
///
/// # Expected Parameters
/// - `commit_id` — OID of the commit to browse (defaults to HEAD if absent).
/// - `path` — Subdirectory path to list (defaults to root `""`).
fn handle_browse_tree(
    request: &RepoRequest,
    repos_dir: &Path,
    repo_name: &str,
) -> Result<RepoResponse> {
    // Resolve the tree OID: either from an explicit commit_id or HEAD
    let tree_oid_str = resolve_tree_oid(repos_dir, repo_name, request)?;

    let path = request.params.get("path").map(|s| s.as_str()).unwrap_or("");

    if path.is_empty() {
        // List the root tree
        let entries = crate::git::objects::read_tree(repos_dir, repo_name, &tree_oid_str)
            .map_err(|e| anyhow::anyhow!("{e}"))?;
        let data = serde_json::to_value(&entries)?;
        Ok(RepoResponse::ok(request.request_id.clone(), data))
    } else {
        // Walk to the specified subdirectory
        let entries = crate::git::objects::walk_tree(repos_dir, repo_name, &tree_oid_str, path)
            .map_err(|e| anyhow::anyhow!("{e}"))?;
        let data = serde_json::to_value(&entries)?;
        Ok(RepoResponse::ok(request.request_id.clone(), data))
    }
}

/// Handle `ReadFile`: read the content of a file (blob) by OID or path.
///
/// # Expected Parameters
/// - `oid` — Direct blob OID (preferred, fast path).
/// - OR `commit_id` + `path` — Resolve the blob OID by walking the tree.
fn handle_read_file(
    request: &RepoRequest,
    repos_dir: &Path,
    repo_name: &str,
) -> Result<RepoResponse> {
    // Fast path: direct OID
    if let Some(oid) = request.params.get("oid") {
        let content = crate::git::objects::read_blob(repos_dir, repo_name, oid)
            .map_err(|e| anyhow::anyhow!("{e}"))?;

        let data = build_file_response(&content);
        return Ok(RepoResponse::ok(request.request_id.clone(), data));
    }

    // Slow path: resolve from commit + path
    let tree_oid_str = resolve_tree_oid(repos_dir, repo_name, request)?;
    let file_path = request
        .params
        .get("path")
        .context("either 'oid' or 'path' param is required for ReadFile")?;

    // Walk the tree to find the blob at the given path
    let entries =
        crate::git::objects::walk_tree(repos_dir, repo_name, &tree_oid_str, "")
            .map_err(|e| anyhow::anyhow!("{e}"))?;

    let file_entry = entries
        .iter()
        .find(|e| e.path == *file_path && !e.is_dir)
        .context(format!("file not found at path: {file_path}"))?;

    let content = crate::git::objects::read_blob(repos_dir, repo_name, &file_entry.oid)
        .map_err(|e| anyhow::anyhow!("{e}"))?;

    let data = build_file_response(&content);
    Ok(RepoResponse::ok(request.request_id.clone(), data))
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

/// Resolve the root tree OID from a commit_id param or HEAD.
fn resolve_tree_oid(
    repos_dir: &Path,
    repo_name: &str,
    request: &RepoRequest,
) -> Result<String> {
    let commit_id = match request.params.get("commit_id") {
        Some(id) => id.clone(),
        None => {
            // Resolve HEAD
            let repo = crate::git::repository::open(repos_dir, repo_name)
                .map_err(|e| anyhow::anyhow!("{e}"))?;
            let head = repo
                .head()
                .context("repository has no HEAD")?;
            head.target()
                .context("HEAD is not a direct reference")?
                .to_string()
        }
    };

    let commit_info = crate::git::objects::read_commit(repos_dir, repo_name, &commit_id)
        .map_err(|e| anyhow::anyhow!("{e}"))?;

    Ok(commit_info.tree_id)
}

/// Build a JSON response for file content.
///
/// Attempts UTF-8 decoding; falls back to indicating binary content.
fn build_file_response(content: &[u8]) -> serde_json::Value {
    let is_binary = content.contains(&0);

    if is_binary {
        serde_json::json!({
            "is_binary": true,
            "size": content.len(),
            "content": null,
        })
    } else {
        serde_json::json!({
            "is_binary": false,
            "size": content.len(),
            "content": String::from_utf8_lossy(content),
        })
    }
}
