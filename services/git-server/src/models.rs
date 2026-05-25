//! # Data Models
//!
//! Shared data structures used throughout the git-server service.
//! All models derive `Serialize` and `Deserialize` for JSON transport
//! over the Unix domain socket.

use serde::{Deserialize, Serialize};

/// Metadata about a Git repository.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct RepoInfo {
    /// Repository name (directory name without `.git` suffix).
    pub name: String,

    /// Absolute path to the bare repository on disk.
    pub path: String,

    /// Whether this is a bare repository.
    pub is_bare: bool,

    /// The default branch name (e.g. "main", "master"), if any.
    pub default_branch: Option<String>,

    /// ISO-8601 timestamp of the most recent commit, if available.
    pub last_commit_at: Option<String>,

    /// Total number of branches.
    pub branch_count: usize,

    /// Total number of tags.
    pub tag_count: usize,
}

/// Information about a Git branch.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct BranchInfo {
    /// Branch name (e.g. "main", "feature/login").
    pub name: String,

    /// Whether this is the HEAD branch.
    pub is_head: bool,

    /// OID of the commit the branch points to.
    pub commit_id: String,

    /// Commit message of the tip commit (first line).
    pub commit_summary: Option<String>,

    /// ISO-8601 timestamp of the tip commit.
    pub committed_at: Option<String>,
}

/// Information about a single Git commit.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CommitInfo {
    /// Full hexadecimal OID of the commit.
    pub id: String,

    /// First line of the commit message.
    pub summary: String,

    /// Full commit message body.
    pub message: String,

    /// Author name.
    pub author_name: String,

    /// Author email.
    pub author_email: String,

    /// ISO-8601 timestamp of when the commit was authored.
    pub authored_at: String,

    /// Committer name (may differ from author in cherry-picks/rebases).
    pub committer_name: String,

    /// Committer email.
    pub committer_email: String,

    /// ISO-8601 timestamp of when the commit was committed.
    pub committed_at: String,

    /// OIDs of parent commits.
    pub parent_ids: Vec<String>,

    /// OID of the tree object.
    pub tree_id: String,
}

/// An entry in a Git tree (directory listing).
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct TreeEntry {
    /// Entry name (filename or subdirectory name).
    pub name: String,

    /// Object type: "blob", "tree", or "commit" (submodule).
    pub object_type: TreeEntryKind,

    /// OID of the referenced object.
    pub oid: String,

    /// UNIX file mode (e.g. 0o100644 for regular file).
    pub filemode: u32,

    /// Size in bytes (only populated for blobs).
    pub size: Option<u64>,
}

/// The kind of object a tree entry points to.
#[derive(Debug, Clone, Serialize, Deserialize, PartialEq, Eq)]
#[serde(rename_all = "snake_case")]
pub enum TreeEntryKind {
    /// Regular file.
    Blob,
    /// Subdirectory.
    Tree,
    /// Submodule reference.
    Commit,
}

/// Represents a file entry for the file browser view.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct FileEntry {
    /// Relative path from the repository root.
    pub path: String,

    /// File name.
    pub name: String,

    /// Whether this entry is a directory.
    pub is_dir: bool,

    /// Size in bytes (only for files, not directories).
    pub size: Option<u64>,

    /// OID of the blob or tree.
    pub oid: String,
}

/// Information about a Git tag.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct TagInfo {
    /// Tag name.
    pub name: String,

    /// OID that the tag points to.
    pub target_id: String,

    /// Whether this is an annotated tag (vs lightweight).
    pub is_annotated: bool,

    /// Tag message (only for annotated tags).
    pub message: Option<String>,

    /// Tagger name (only for annotated tags).
    pub tagger_name: Option<String>,

    /// ISO-8601 timestamp (only for annotated tags).
    pub tagged_at: Option<String>,
}

/// Diff information between two commits or a commit and its parent.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct DiffInfo {
    /// Number of files changed.
    pub files_changed: usize,

    /// Total insertions across all files.
    pub insertions: usize,

    /// Total deletions across all files.
    pub deletions: usize,

    /// Per-file diff entries.
    pub files: Vec<DiffFileEntry>,
}

/// A single file's diff information.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct DiffFileEntry {
    /// File path (new path if renamed).
    pub path: String,

    /// Old file path (if renamed or moved).
    pub old_path: Option<String>,

    /// Status: "added", "deleted", "modified", "renamed".
    pub status: String,

    /// Number of added lines.
    pub insertions: usize,

    /// Number of removed lines.
    pub deletions: usize,

    /// Unified diff patch text.
    pub patch: Option<String>,
}
