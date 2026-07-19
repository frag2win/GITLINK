//! # Git Module
//!
//! Core Git operations powered by `git2-rs` (libgit2 bindings).
//!
//! ## Submodules
//! - [`repository`] — Repository lifecycle: init, open, list, delete.
//! # Git Module
//!
//! Core Git operations powered by `git2-rs` (libgit2 bindings).
//!
//! ## Submodules
//! - [`repository`] — Repository lifecycle: init, open, list, delete.
//! - [`refs`] — Reference management: branches and tags.
//! - [`objects`] — Object access: blobs, trees, commits.
//! - [`commits`] — Commit history walking, detail, and diffs.
//! - [`pack`] — Pack file handling for push/pull operations.

pub mod commits;
pub mod http;
pub mod lock;
pub mod merge;
pub mod objects;
pub mod pack;
pub mod refs;
pub mod repository;
pub mod sanitize;
