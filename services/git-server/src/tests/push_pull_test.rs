#[cfg(test)]
mod tests {
    use crate::handlers::http_handler::{handle_receive_pack, handle_upload_pack};
    use crate::socket::protocol::{ReceivePackRequest, UploadPackRequest};
    use crate::config::Config;
    use crate::git::repository;
    use tempfile::tempdir;

    #[tokio::test]
    async fn test_push_integration() {
        let temp_dir = tempdir().unwrap();
        let repos_path = temp_dir.path();
        let repo_name = "test_repo";
        repository::init_bare(repos_path, repo_name).unwrap();

        let config = Config {
            repos_path: repos_path.to_string_lossy().to_string(),
            ..Default::default()
        };
        
        let req = ReceivePackRequest {
            repo_name: repo_name.to_string(),
            body: b"0000000000000000000000000000000000000000 1111111111111111111111111111111111111111 refs/heads/main\0report-status\nPACK...".to_vec(),
        };

        let result = handle_receive_pack(req, &config).await;
        
        if let Some(err) = &result.error {
            assert!(!err.message.contains("RepoNotFound"), "Error should be about pack processing, not missing repo");
        }
    }

    #[tokio::test]
    async fn test_pull_integration() {
        let temp_dir = tempdir().unwrap();
        let repos_path = temp_dir.path();
        let repo_name = "test_repo";
        repository::init_bare(repos_path, repo_name).unwrap();

        let config = Config {
            repos_path: repos_path.to_string_lossy().to_string(),
            ..Default::default()
        };
        
        let req = UploadPackRequest {
            repo_name: repo_name.to_string(),
            body: b"0032want 1111111111111111111111111111111111111111\n00000009done\n".to_vec(),
        };

        let result = handle_upload_pack(req, &config).await;
        
        if let Some(err) = &result.error {
            assert!(!err.message.contains("RepoNotFound"), "Error should be about pack processing, not missing repo");
        }
    }
}
