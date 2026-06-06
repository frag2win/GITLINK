//! # Error Types
//!
//! Defines custom error types for the git-server service:
//! - [`GitError`] — errors from Git operations (libgit2).
//! - [`SocketError`] — errors from Unix domain socket communication.
//! - [`ConfigError`] — errors from configuration loading/validation.
//!
//! All error types implement `std::error::Error` and convert into `anyhow::Error`.

use std::fmt;

// ---------------------------------------------------------------------------
// GitError
// ---------------------------------------------------------------------------

/// Errors arising from Git (libgit2) operations.
#[derive(Debug)]
pub enum GitError {
    /// The requested repository was not found at the given path.
    RepoNotFound { path: String },

    /// A repository already exists at the given path.
    RepoAlreadyExists { path: String },

    /// The repository name is invalid (e.g., path traversal).
    InvalidRepoName(String),

    /// A reference (branch, tag) was not found.
    RefNotFound { name: String },

    /// An object (commit, tree, blob) was not found.
    ObjectNotFound { oid: String },

    /// Wrapper around a raw `git2::Error`.
    Libgit2(git2::Error),

    /// Catch-all for other Git errors.
    Other(String),

    /// I/O error during Git operations (e.g. streaming packs).
    IoError(std::io::Error),

    /// A shell Git command failed.
    GitCommandFailed(String),
}

impl fmt::Display for GitError {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            Self::RepoNotFound { path } => write!(f, "repository not found: {path}"),
            Self::RepoAlreadyExists { path } => write!(f, "repository already exists: {path}"),
            Self::InvalidRepoName(msg) => write!(f, "invalid repository name: {msg}"),
            Self::RefNotFound { name } => write!(f, "reference not found: {name}"),
            Self::ObjectNotFound { oid } => write!(f, "object not found: {oid}"),
            Self::Libgit2(e) => write!(f, "libgit2 error: {e}"),
            Self::Other(msg) => write!(f, "git error: {msg}"),
            Self::IoError(e) => write!(f, "I/O error: {e}"),
            Self::GitCommandFailed(msg) => write!(f, "git command failed: {msg}"),
        }
    }
}

impl std::error::Error for GitError {
    fn source(&self) -> Option<&(dyn std::error::Error + 'static)> {
        match self {
            Self::Libgit2(e) => Some(e),
            _ => None,
        }
    }
}

impl From<git2::Error> for GitError {
    fn from(err: git2::Error) -> Self {
        Self::Libgit2(err)
    }
}

// ---------------------------------------------------------------------------
// SocketError
// ---------------------------------------------------------------------------

/// Errors arising from Unix domain socket communication.
#[derive(Debug)]
pub enum SocketError {
    /// Failed to bind or connect to the socket path.
    ConnectionFailed { path: String, reason: String },

    /// An I/O error occurred during socket read/write.
    Io(std::io::Error),

    /// The received message could not be deserialized.
    InvalidMessage { reason: String },

    /// The connection was closed unexpectedly.
    ConnectionClosed,

    /// The message exceeded the maximum allowed size.
    MessageTooLarge { size: usize, max: usize },
}

impl fmt::Display for SocketError {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            Self::ConnectionFailed { path, reason } => {
                write!(f, "socket connection failed at {path}: {reason}")
            }
            Self::Io(e) => write!(f, "socket I/O error: {e}"),
            Self::InvalidMessage { reason } => write!(f, "invalid message: {reason}"),
            Self::ConnectionClosed => write!(f, "socket connection closed unexpectedly"),
            Self::MessageTooLarge { size, max } => {
                write!(f, "message too large: {size} bytes (max {max})")
            }
        }
    }
}

impl std::error::Error for SocketError {
    fn source(&self) -> Option<&(dyn std::error::Error + 'static)> {
        match self {
            Self::Io(e) => Some(e),
            _ => None,
        }
    }
}

impl From<std::io::Error> for SocketError {
    fn from(err: std::io::Error) -> Self {
        Self::Io(err)
    }
}

impl From<serde_json::Error> for SocketError {
    fn from(err: serde_json::Error) -> Self {
        Self::InvalidMessage {
            reason: err.to_string(),
        }
    }
}

// ---------------------------------------------------------------------------
// ConfigError
// ---------------------------------------------------------------------------

/// Errors arising from configuration loading or validation.
#[derive(Debug)]
pub enum ConfigError {
    /// A required environment variable is missing.
    MissingEnvVar { name: String },

    /// A configuration value is invalid.
    InvalidValue { field: String, reason: String },
}

impl fmt::Display for ConfigError {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            Self::MissingEnvVar { name } => {
                write!(f, "missing required environment variable: {name}")
            }
            Self::InvalidValue { field, reason } => {
                write!(f, "invalid config value for '{field}': {reason}")
            }
        }
    }
}

impl std::error::Error for ConfigError {}

// ---------------------------------------------------------------------------
// Unified ServiceError (optional convenience wrapper)
// ---------------------------------------------------------------------------

/// Top-level service error that unifies all error categories.
#[derive(Debug)]
pub enum ServiceError {
    Git(GitError),
    Socket(SocketError),
    Config(ConfigError),
}

impl fmt::Display for ServiceError {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            Self::Git(e) => write!(f, "{e}"),
            Self::Socket(e) => write!(f, "{e}"),
            Self::Config(e) => write!(f, "{e}"),
        }
    }
}

impl std::error::Error for ServiceError {
    fn source(&self) -> Option<&(dyn std::error::Error + 'static)> {
        match self {
            Self::Git(e) => Some(e),
            Self::Socket(e) => Some(e),
            Self::Config(e) => Some(e),
        }
    }
}

impl From<GitError> for ServiceError {
    fn from(err: GitError) -> Self {
        Self::Git(err)
    }
}

impl From<SocketError> for ServiceError {
    fn from(err: SocketError) -> Self {
        Self::Socket(err)
    }
}

impl From<ConfigError> for ServiceError {
    fn from(err: ConfigError) -> Self {
        Self::Config(err)
    }
}
