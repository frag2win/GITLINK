//! # Repository Handler
//!
//! Handles repository CRUD requests received over the Unix domain socket:
//! - `InitRepo` — Create a new bare repository.
//! - `ListRepos` — List all repositories.
//! - `DeleteRepo` — Remove a repository.

use std::path::Path;

use anyhow::{Context, Result};
use tracing::instrument;

use crate::config::Config;
use crate::socket::protocol::{GitOperation, RepoRequest, RepoResponse};

/// Dispatch a repository CRUD request to the appropriate Git operation.
///
/// # Supported Operations
/// - [`GitOperation::InitRepo`] — requires `repo_name`.
/// - [`GitOperation::ListRepos`] — no additional params.
/// - [`GitOperation::DeleteRepo`] — requires `repo_name`.
#[instrument(skip(config))]
pub async fn handle(request: &RepoRequest, config: &Config) -> Result<RepoResponse> {
    let repos_dir = Path::new(&config.repos_path);

    match request.operation {
        GitOperation::InitRepo => handle_init(request, repos_dir),
        GitOperation::ListRepos => handle_list(request, repos_dir),
        GitOperation::DeleteRepo => handle_delete(request, repos_dir),
        _ => Ok(RepoResponse::err(
            request.request_id.clone(),
            format!("unsupported operation for repo_handler: {:?}", request.operation),
        )),
    }
}

/// Handle `InitRepo`: create a new bare repository.
fn handle_init(request: &RepoRequest, repos_dir: &Path) -> Result<RepoResponse> {
    let repo_name = request
        .repo_name
        .as_deref()
        .context("repo_name is required for InitRepo")?;

    let repo_info = crate::git::repository::init_bare(repos_dir, repo_name)
        .map_err(|e| anyhow::anyhow!("{e}"))?;

    let data = serde_json::to_value(&repo_info)
        .context("failed to serialize RepoInfo")?;

    Ok(RepoResponse::ok(request.request_id.clone(), data))
}

/// Handle `ListRepos`: enumerate all repositories.
fn handle_list(request: &RepoRequest, repos_dir: &Path) -> Result<RepoResponse> {
    let repos = crate::git::repository::list_repos(repos_dir)
        .map_err(|e| anyhow::anyhow!("{e}"))?;

    let data = serde_json::to_value(&repos)
        .context("failed to serialize repo list")?;

    Ok(RepoResponse::ok(request.request_id.clone(), data))
}

/// Handle `DeleteRepo`: remove a repository from disk.
fn handle_delete(request: &RepoRequest, repos_dir: &Path) -> Result<RepoResponse> {
    let repo_name = request
        .repo_name
        .as_deref()
        .context("repo_name is required for DeleteRepo")?;

    crate::git::repository::delete_repo(repos_dir, repo_name)
        .map_err(|e| anyhow::anyhow!("{e}"))?;

    let data = serde_json::json!({
        "deleted": true,
        "name": repo_name,
    });

    Ok(RepoResponse::ok(request.request_id.clone(), data))
}
