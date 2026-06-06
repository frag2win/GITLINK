//! # Socket Protocol
//!
//! Exposes the auto-generated Protobuf structs from `proto/git_commands.proto`.

// Include the generated protobuf code from prost
include!(concat!(env!("OUT_DIR"), "/git_commands.rs"));

// Re-export nested enums for easier access
pub use git_command_request::Command as RequestCommand;
pub use git_command_response::Result as ResponseResult;
