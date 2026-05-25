//! # Git Handler
//!
//! Handles Git data operation requests received over the Unix domain socket:
//! - `Clone` / `Push` / `Pull` — pack-based operations.
//! - `ListRefs` / `ListBranches` / `CreateBranch` / `DeleteBranch` / `ListTags` — ref operations.
//! - `ListCommits` / `GetCommitDetail` / `GetDiff` — commit operations.
//! - `GetObject` — raw object access.

use std::path::Path;

use anyhow::{Context, Result};
use tracing::instrument;

use crate::config::Config;
use crate::socket::protocol::{GitOperation, RepoRequest, RepoResponse};

/// Dispatch a Git data operation request to the appropriate function.
#[instrument(skip(config))]
pub async fn handle(request: &RepoRequest, config: &Config) -> Result<RepoResponse> {
    let repos_dir = Path::new(&config.repos_path);

    let repo_name = request.repo_name.as_deref().unwrap_or("");

    match request.operation {
        GitOperation::Clone => handle_clone(request, repos_dir, repo_name).await,
        GitOperation::Push => handle_push(request, repos_dir, repo_name),
        GitOperation::Pull => handle_pull(request, repos_dir, repo_name),
        GitOperation::ListRefs => handle_list_refs(request, repos_dir, repo_name),
        GitOperation::GetObject => handle_get_object(request, repos_dir, repo_name),
        GitOperation::ListBranches => handle_list_branches(request, repos_dir, repo_name),
        GitOperation::CreateBranch => handle_create_branch(request, repos_dir, repo_name),
        GitOperation::DeleteBranch => handle_delete_branch(request, repos_dir, repo_name),
        GitOperation::ListTags => handle_list_tags(request, repos_dir, repo_name),
        GitOperation::ListCommits => handle_list_commits(request, repos_dir, repo_name),
        GitOperation::GetCommitDetail => handle_get_commit_detail(request, repos_dir, repo_name),
        GitOperation::GetDiff => handle_get_diff(request, repos_dir, repo_name),
        _ => Ok(RepoResponse::err(
            request.request_id.clone(),
            format!("unsupported operation for git_handler: {:?}", request.operation),
        )),
    }
}

/// Handle `Clone`: clone/mirror a remote repository into the local store.
async fn handle_clone(
    request: &RepoRequest,
    repos_dir: &Path,
    repo_name: &str,
) -> Result<RepoResponse> {
    let url = request
        .params
        .get("url")
        .context("'url' param is required for Clone")?;

    // Clone in a blocking task since git2 is synchronous
    let repos_dir = repos_dir.to_path_buf();
    let repo_name = repo_name.to_string();
    let url = url.clone();
    let request_id = request.request_id.clone();

    let result = tokio::task::spawn_blocking(move || {
        let dest = repos_dir.join(format!("{repo_name}.git"));
        let mut builder = git2::build::RepoBuilder::new();
        builder.bare(true);
        builder.clone(&url, &dest)
    })
    .await
    .context("clone task panicked")?;

    match result {
        Ok(_repo) => {
            let data = serde_json::json!({
                "cloned": true,
                "name": repo_name,
            });
            Ok(RepoResponse::ok(request_id, data))
        }
        Err(e) => Ok(RepoResponse::err(request_id, format!("clone failed: {e}"))),
    }
}

/// Handle `Push`: receive pack data from the client.
fn handle_push(
    request: &RepoRequest,
    repos_dir: &Path,
    repo_name: &str,
) -> Result<RepoResponse> {
    let pack_data_b64 = request
        .params
        .get("pack_data")
        .context("'pack_data' param is required for Push")?;

    // Pack data is expected as base64-encoded bytes in the JSON message
    // TODO: Implement proper base64 decoding; for now, treat as raw UTF-8 bytes
    let pack_data = pack_data_b64.as_bytes();

    let result = crate::git::pack::receive_pack(repos_dir, repo_name, pack_data)
        .map_err(|e| anyhow::anyhow!("{e}"))?;

    let data = serde_json::json!({
        "objects_received": result.objects_received,
        "refs_updated": result.refs_updated,
    });

    Ok(RepoResponse::ok(request.request_id.clone(), data))
}

/// Handle `Pull`: generate a pack file for the client.
fn handle_pull(
    request: &RepoRequest,
    repos_dir: &Path,
    repo_name: &str,
) -> Result<RepoResponse> {
    let want_oids: Vec<String> = request
        .params
        .get("want_oids")
        .map(|s| serde_json::from_str(s).unwrap_or_default())
        .unwrap_or_default();

    let have_oids: Vec<String> = request
        .params
        .get("have_oids")
        .map(|s| serde_json::from_str(s).unwrap_or_default())
        .unwrap_or_default();

    let result = crate::git::pack::send_pack(repos_dir, repo_name, &want_oids, &have_oids)
        .map_err(|e| anyhow::anyhow!("{e}"))?;

    let data = serde_json::json!({
        "object_count": result.object_count,
        "pack_size": result.pack_data.len(),
        // TODO: base64-encode pack_data for JSON transport
    });

    Ok(RepoResponse::ok(request.request_id.clone(), data))
}

/// Handle `ListRefs`: list all references (branches + tags).
fn handle_list_refs(
    request: &RepoRequest,
    repos_dir: &Path,
    repo_name: &str,
) -> Result<RepoResponse> {
    let branches = crate::git::refs::list_branches(repos_dir, repo_name)
        .map_err(|e| anyhow::anyhow!("{e}"))?;
    let tags = crate::git::refs::list_tags(repos_dir, repo_name)
        .map_err(|e| anyhow::anyhow!("{e}"))?;

    let data = serde_json::json!({
        "branches": serde_json::to_value(&branches)?,
        "tags": serde_json::to_value(&tags)?,
    });

    Ok(RepoResponse::ok(request.request_id.clone(), data))
}

/// Handle `GetObject`: read a raw Git object by OID.
fn handle_get_object(
    request: &RepoRequest,
    repos_dir: &Path,
    repo_name: &str,
) -> Result<RepoResponse> {
    let oid = request
        .params
        .get("oid")
        .context("'oid' param is required for GetObject")?;

    // Try reading as commit, then tree, then blob
    if let Ok(commit) = crate::git::objects::read_commit(repos_dir, repo_name, oid) {
        let data = serde_json::to_value(&commit)?;
        return Ok(RepoResponse::ok(request.request_id.clone(), data));
    }

    if let Ok(tree) = crate::git::objects::read_tree(repos_dir, repo_name, oid) {
        let data = serde_json::to_value(&tree)?;
        return Ok(RepoResponse::ok(request.request_id.clone(), data));
    }

    if let Ok(blob) = crate::git::objects::read_blob(repos_dir, repo_name, oid) {
        let data = serde_json::json!({
            "type": "blob",
            "size": blob.len(),
            // TODO: base64-encode binary content
            "content": String::from_utf8_lossy(&blob),
        });
        return Ok(RepoResponse::ok(request.request_id.clone(), data));
    }

    Ok(RepoResponse::err(
        request.request_id.clone(),
        format!("object not found: {oid}"),
    ))
}

/// Handle `ListBranches`.
fn handle_list_branches(
    request: &RepoRequest,
    repos_dir: &Path,
    repo_name: &str,
) -> Result<RepoResponse> {
    let branches = crate::git::refs::list_branches(repos_dir, repo_name)
        .map_err(|e| anyhow::anyhow!("{e}"))?;
    let data = serde_json::to_value(&branches)?;
    Ok(RepoResponse::ok(request.request_id.clone(), data))
}

/// Handle `CreateBranch`.
fn handle_create_branch(
    request: &RepoRequest,
    repos_dir: &Path,
    repo_name: &str,
) -> Result<RepoResponse> {
    let branch_name = request
        .params
        .get("branch_name")
        .context("'branch_name' param is required")?;
    let target = request
        .params
        .get("target_commit")
        .context("'target_commit' param is required")?;

    let branch = crate::git::refs::create_branch(repos_dir, repo_name, branch_name, target)
        .map_err(|e| anyhow::anyhow!("{e}"))?;
    let data = serde_json::to_value(&branch)?;
    Ok(RepoResponse::ok(request.request_id.clone(), data))
}

/// Handle `DeleteBranch`.
fn handle_delete_branch(
    request: &RepoRequest,
    repos_dir: &Path,
    repo_name: &str,
) -> Result<RepoResponse> {
    let branch_name = request
        .params
        .get("branch_name")
        .context("'branch_name' param is required")?;

    crate::git::refs::delete_branch(repos_dir, repo_name, branch_name)
        .map_err(|e| anyhow::anyhow!("{e}"))?;

    let data = serde_json::json!({ "deleted": true, "branch": branch_name });
    Ok(RepoResponse::ok(request.request_id.clone(), data))
}

/// Handle `ListTags`.
fn handle_list_tags(
    request: &RepoRequest,
    repos_dir: &Path,
    repo_name: &str,
) -> Result<RepoResponse> {
    let tags = crate::git::refs::list_tags(repos_dir, repo_name)
        .map_err(|e| anyhow::anyhow!("{e}"))?;
    let data = serde_json::to_value(&tags)?;
    Ok(RepoResponse::ok(request.request_id.clone(), data))
}

/// Handle `ListCommits`.
fn handle_list_commits(
    request: &RepoRequest,
    repos_dir: &Path,
    repo_name: &str,
) -> Result<RepoResponse> {
    let start_oid = request.params.get("start_oid").map(|s| s.as_str());
    let limit: usize = request
        .params
        .get("limit")
        .and_then(|s| s.parse().ok())
        .unwrap_or(50);
    let offset: usize = request
        .params
        .get("offset")
        .and_then(|s| s.parse().ok())
        .unwrap_or(0);

    let commits = crate::git::commits::list_commits(repos_dir, repo_name, start_oid, limit, offset)
        .map_err(|e| anyhow::anyhow!("{e}"))?;
    let data = serde_json::to_value(&commits)?;
    Ok(RepoResponse::ok(request.request_id.clone(), data))
}

/// Handle `GetCommitDetail`.
fn handle_get_commit_detail(
    request: &RepoRequest,
    repos_dir: &Path,
    repo_name: &str,
) -> Result<RepoResponse> {
    let oid = request
        .params
        .get("commit_id")
        .context("'commit_id' param is required")?;

    let commit = crate::git::commits::get_commit_detail(repos_dir, repo_name, oid)
        .map_err(|e| anyhow::anyhow!("{e}"))?;
    let data = serde_json::to_value(&commit)?;
    Ok(RepoResponse::ok(request.request_id.clone(), data))
}

/// Handle `GetDiff`.
fn handle_get_diff(
    request: &RepoRequest,
    repos_dir: &Path,
    repo_name: &str,
) -> Result<RepoResponse> {
    let to_oid = request
        .params
        .get("to_commit")
        .context("'to_commit' param is required")?;
    let from_oid = request.params.get("from_commit").map(|s| s.as_str());

    let diff = crate::git::commits::get_diff(repos_dir, repo_name, from_oid, to_oid)
        .map_err(|e| anyhow::anyhow!("{e}"))?;
    let data = serde_json::to_value(&diff)?;
    Ok(RepoResponse::ok(request.request_id.clone(), data))
}
