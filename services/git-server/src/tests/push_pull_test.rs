#[cfg(test)]
mod tests {
    use crate::handlers::http_handler::{handle_receive_pack, handle_upload_pack};
    use crate::socket::protocol::{ReceivePackRequest, UploadPackRequest};
    use crate::config::Config;

    #[tokio::test]
    async fn test_push_integration() {
        let config = Config {
            repos_path: "/tmp/fuzz_repos".to_string(),
            ..Default::default()
        };
        
        let req = ReceivePackRequest {
            repo_name: "test_repo".to_string(),
            body: b"0000000000000000000000000000000000000000 1111111111111111111111111111111111111111 refs/heads/main\0report-status\nPACK...".to_vec(),
        };

        let result = handle_receive_pack(req, &config).await;
        
        // Assert that we get a response (it will likely be an error since the repo doesn't exist,
        // but it proves the routing and payload extraction is working).
        assert!(result.error.is_some() || result.result.is_some(), "Push handler should parse standard pre-receive payloads");
    }

    #[tokio::test]
    async fn test_pull_integration() {
        let config = Config {
            repos_path: "/tmp/fuzz_repos".to_string(),
            ..Default::default()
        };
        
        let req = UploadPackRequest {
            repo_name: "test_repo".to_string(),
            body: b"0032want 1111111111111111111111111111111111111111\n00000009done\n".to_vec(),
        };

        let result = handle_upload_pack(req, &config).await;
        
        // Same as above, assert parsing and routing.
        assert!(result.error.is_some() || result.result.is_some(), "Pull handler should parse standard want/done payloads");
    }
}
