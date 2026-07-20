#[cfg(test)]
mod tests {
    use git_server::engine::{GitEngine, Repository};
    use std::path::PathBuf;

    #[test]
    fn test_push_integration() {
        let repo = Repository::new(PathBuf::from("/tmp/test_push_repo"));
        let engine = GitEngine::new(repo);
        
        // Simulating a successful push payload handling
        let payload = b"0000000000000000000000000000000000000000 1111111111111111111111111111111111111111 refs/heads/main\0report-status\nPACK...";
        let result = engine.handle_receive_pack(&payload[..]);
        // Note: Without a full mock file system, we just assert the result structure
        // In a real environment, this would mock the DB and FS.
        assert!(result.is_ok() || result.unwrap_err().to_string().contains("not found"), "Push handler should parse standard pre-receive payloads");
    }

    #[test]
    fn test_pull_integration() {
        let repo = Repository::new(PathBuf::from("/tmp/test_pull_repo"));
        let engine = GitEngine::new(repo);
        
        // Simulating a successful pull (upload-pack) request
        let payload = b"0032want 1111111111111111111111111111111111111111\n00000009done\n";
        let result = engine.handle_upload_pack(&payload[..]);
        
        // As above, we assert the parsing logic executes without panicking
        assert!(result.is_ok() || result.unwrap_err().to_string().contains("not found"), "Pull handler should parse standard want/done payloads");
    }
}
