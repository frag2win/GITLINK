use std::path::Path;
use crate::config::Config;
use crate::git::http;
use crate::socket::protocol::{
    GitCommandResponse, InfoRefsRequest, InfoRefsResponse, ReceivePackRequest,
    ReceivePackResponse, UploadPackRequest, UploadPackResponse, ResponseResult,
};

pub async fn handle_info_refs(req: InfoRefsRequest, config: &Config) -> GitCommandResponse {
    let dir_name = if req.repo_name.ends_with(".git") {
        req.repo_name.clone()
    } else {
        format!("{}.git", req.repo_name)
    };
    let repo_path = Path::new(&config.repos_path).join(&dir_name);
    let path_str = repo_path.to_string_lossy();

    match http::info_refs(&path_str, &req.service) {
        Ok(output) => GitCommandResponse {
            protocol_version: 1, error: None,
            
            result: Some(ResponseResult::InfoRefs(InfoRefsResponse { output })),
        },
        Err(e) => GitCommandResponse {
            protocol_version: 1,
            error: Some(crate::socket::protocol::GitError { code: "GitError".to_string(), message: e.to_string() }),
            result: None,
        },
    }
}

pub async fn handle_upload_pack(req: UploadPackRequest, config: &Config) -> GitCommandResponse {
    let dir_name = if req.repo_name.ends_with(".git") {
        req.repo_name.clone()
    } else {
        format!("{}.git", req.repo_name)
    };
    let repo_path = Path::new(&config.repos_path).join(&dir_name);
    let path_str = repo_path.to_string_lossy();

    match http::upload_pack(&path_str, &req.body) {
        Ok(output) => GitCommandResponse {
            protocol_version: 1, error: None,
            
            result: Some(ResponseResult::UploadPack(UploadPackResponse { output })),
        },
        Err(e) => GitCommandResponse {
            protocol_version: 1,
            error: Some(crate::socket::protocol::GitError { code: "GitError".to_string(), message: e.to_string() }),
            result: None,
        },
    }
}

pub async fn handle_receive_pack(req: ReceivePackRequest, config: &Config) -> GitCommandResponse {
    let dir_name = if req.repo_name.ends_with(".git") {
        req.repo_name.clone()
    } else {
        format!("{}.git", req.repo_name)
    };
    let repo_path = Path::new(&config.repos_path).join(&dir_name);
    let path_str = repo_path.to_string_lossy();

    match http::receive_pack(&path_str, &req.body) {
        Ok(output) => GitCommandResponse {
            protocol_version: 1, error: None,
            
            result: Some(ResponseResult::ReceivePack(ReceivePackResponse { output })),
        },
        Err(e) => GitCommandResponse {
            protocol_version: 1,
            error: Some(crate::socket::protocol::GitError { code: "GitError".to_string(), message: e.to_string() }),
            result: None,
        },
    }
}
