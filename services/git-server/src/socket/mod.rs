//! # Socket Module
//!
//! Handles Unix domain socket communication for the git-server service.
//! This is the **only** communication interface — no TCP/HTTP networking.
//!
//! ## Submodules
//! - [`server`] — Socket server: bind, accept, read/write loop.
//! - [`protocol`] — Request/Response message types for the JSON protocol.

pub mod protocol;
pub mod server;
