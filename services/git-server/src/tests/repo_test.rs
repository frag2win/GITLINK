#[cfg(test)]
mod tests {
    use std::path::Path;
    use tempfile::tempdir;
    use crate::git::repository;
    use crate::error::GitError;

    #[test]
    fn test_repo_crud_idempotency() {
        let temp_dir = tempdir().unwrap();
        let repos_path = temp_dir.path();
        let repo_name = "test-repo";

        // 1. Create Repo
        let info = repository::init_bare(repos_path, repo_name).expect("Failed to init repo");
        assert_eq!(info.name, repo_name);
        assert!(info.is_bare);

        // 2. Idempotency Check: Create Again
        let err = repository::init_bare(repos_path, repo_name).unwrap_err();
        match err {
            GitError::RepoAlreadyExists { .. } => {},
            _ => panic!("Expected RepoAlreadyExists, got {:?}", err),
        }

        // 3. Open Repo
        let _repo = repository::open(repos_path, repo_name).expect("Failed to open repo");

        // 4. Delete Repo
        repository::delete_repo(repos_path, repo_name).expect("Failed to delete repo");

        // 5. Delete Again
        let del_err = repository::delete_repo(repos_path, repo_name).unwrap_err();
        match del_err {
            GitError::RepoNotFound { .. } => {},
            _ => panic!("Expected RepoNotFound, got {:?}", del_err),
        }
    }
}
