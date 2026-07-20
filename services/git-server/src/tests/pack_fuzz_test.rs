#[cfg(test)]
mod tests {
    use crate::git::pack::receive_pack;
    use crate::git::repository;
    use tempfile::tempdir;

    #[test]
    fn fuzz_pack_parser_corrupt_header() {
        let temp_dir = tempdir().unwrap();
        let repos_path = temp_dir.path();
        let repo_name = "test_repo";
        repository::init_bare(repos_path, repo_name).unwrap();

        // Simulating a fuzzing test checking resilience against corrupt headers
        let corrupt_data = vec![0x50, 0x41, 0x43, 0x4B, 0x00, 0x00, 0x00, 0x03, 0x00, 0x00, 0x00, 0x00];
        let result = receive_pack(repos_path, repo_name, &corrupt_data);
        
        let err = result.unwrap_err();
        assert!(!err.to_string().contains("RepoNotFound"), "Error should be about parsing, not missing repo");
    }

    #[test]
    fn fuzz_pack_parser_truncated_body() {
        let temp_dir = tempdir().unwrap();
        let repos_path = temp_dir.path();
        let repo_name = "test_repo";
        repository::init_bare(repos_path, repo_name).unwrap();

        let truncated_data = vec![0x50, 0x41, 0x43, 0x4B, 0x00, 0x00, 0x00, 0x02, 0x00, 0x00, 0x00, 0x01, 0x90];
        let result = receive_pack(repos_path, repo_name, &truncated_data);
        
        let err = result.unwrap_err();
        assert!(!err.to_string().contains("RepoNotFound"), "Error should be about parsing, not missing repo");
    }

    #[test]
    fn fuzz_pack_parser_invalid_checksum() {
        let temp_dir = tempdir().unwrap();
        let repos_path = temp_dir.path();
        let repo_name = "test_repo";
        repository::init_bare(repos_path, repo_name).unwrap();

        let invalid_checksum_data = vec![
            0x50, 0x41, 0x43, 0x4B, 0x00, 0x00, 0x00, 0x02, 0x00, 0x00, 0x00, 0x00,
            0xAA, 0xBB, 0xCC, 0xDD, 0xEE, 0xFF, 0x00, 0x11, 0x22, 0x33, 0x44, 0x55, 
            0x66, 0x77, 0x88, 0x99, 0xAA, 0xBB, 0xCC, 0xDD
        ];
        let result = receive_pack(repos_path, repo_name, &invalid_checksum_data);
        
        let err = result.unwrap_err();
        assert!(!err.to_string().contains("RepoNotFound"), "Error should be about parsing, not missing repo");
    }
}
