#[cfg(test)]
mod tests {
    use std::path::Path;
    use tempfile::tempdir;
    use crate::git::{repository, merge};
    use crate::error::GitError;

    #[test]
    fn test_merge_up_to_date() {
        let temp_dir = tempdir().unwrap();
        let repos_path = temp_dir.path();
        let repo_name = "test-merge-repo";

        // Init repo
        let repo_info = repository::init_bare(repos_path, repo_name).unwrap();
        let repo = repository::open(repos_path, repo_name).unwrap();

        let merge_engine = merge::MergeEngine::new(&repo);
        
        // No branches exist yet, so execute should fail with RefNotFound
        let err = merge_engine.execute("main", "feature", "Test Author", "test@test.com", "Merge message").unwrap_err();
        match err {
            GitError::RefNotFound { .. } => {},
            _ => panic!("Expected RefNotFound, got {:?}", err),
        }
    }
}
