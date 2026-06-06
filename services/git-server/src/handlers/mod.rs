//! # Handlers Module
//!
//! Request handlers that bridge socket protocol messages to Git operations.
//!
//! ## Submodules
//! - [`repo_handler`] — Repository CRUD operations (init, list, delete).
//! - [`git_handler`] — Git data operations (clone, push, pull, branches, commits, etc.).
//! - [`browse_handler`] — File browsing (tree listing, file content).

pub mod browse_handler;
pub mod git_handler;
pub mod repo_handler;
pub mod http_handler;
