//! # Socket Protocol
//!
//! Defines the JSON message types exchanged over the Unix domain socket.
//! Every request is a [`RepoRequest`] and every response is a [`RepoResponse`].
//!
//! The protocol is line-delimited JSON (newline-terminated).

use serde::{Deserialize, Serialize};

// ---------------------------------------------------------------------------
// Git Operations Enum
// ---------------------------------------------------------------------------

/// Enumerates all Git operations the server can perform.
#[derive(Debug, Clone, Serialize, Deserialize, PartialEq, Eq)]
#[serde(rename_all = "snake_case")]
pub enum GitOperation {
    /// Clone/mirror a repository from a URL into the local store.
    Clone,
    /// Receive a pack file (push).
    Push,
    /// Send a pack file (pull / fetch).
    Pull,
    /// List all references (branches + tags).
    ListRefs,
    /// Read a specific Git object by OID.
    GetObject,
    /// Initialize a new bare repository.
    InitRepo,
    /// Delete a repository.
    DeleteRepo,
    /// List all repositories.
    ListRepos,
    /// List branches.
    ListBranches,
    /// Create a branch.
    CreateBranch,
    /// Delete a branch.
    DeleteBranch,
    /// List tags.
    ListTags,
    /// List commits (walk history).
    ListCommits,
    /// Get detailed commit information.
    GetCommitDetail,
    /// Get diff between commits.
    GetDiff,
    /// Browse a tree (directory listing).
    BrowseTree,
    /// Read file content (blob).
    ReadFile,
}

// ---------------------------------------------------------------------------
// Request
// ---------------------------------------------------------------------------

/// A request sent to the git-server over the Unix domain socket.
///
/// The `operation` field determines which handler processes the request.
/// The `params` map carries operation-specific key-value arguments.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct RepoRequest {
    /// Unique identifier for this request (for correlating responses).
    pub request_id: String,

    /// The Git operation to perform.
    pub operation: GitOperation,

    /// Repository name (required for most operations).
    pub repo_name: Option<String>,

    /// Operation-specific parameters as a flat key-value map.
    ///
    /// Examples:
    /// - `branch_name`: target branch for create/delete branch operations.
    /// - `commit_id`: OID for commit detail / diff operations.
    /// - `path`: file path for browse/read operations.
    /// - `from_commit` / `to_commit`: range for diff operations.
    /// - `limit` / `offset`: pagination for list operations.
    pub params: std::collections::HashMap<String, String>,
}

// ---------------------------------------------------------------------------
// Response
// ---------------------------------------------------------------------------

/// A response sent back to the client over the Unix domain socket.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct RepoResponse {
    /// The request_id this response corresponds to.
    pub request_id: String,

    /// Whether the operation succeeded.
    pub success: bool,

    /// Human-readable error message (populated when `success == false`).
    pub error: Option<String>,

    /// JSON-encoded payload of the response data.
    /// The structure depends on the operation performed.
    pub data: Option<serde_json::Value>,
}

impl RepoResponse {
    /// Create a successful response with data.
    pub fn ok(request_id: String, data: serde_json::Value) -> Self {
        Self {
            request_id,
            success: true,
            error: None,
            data: Some(data),
        }
    }

    /// Create an error response.
    pub fn err(request_id: String, error: impl Into<String>) -> Self {
        Self {
            request_id,
            success: false,
            error: Some(error.into()),
            data: None,
        }
    }
}
