#[cfg(test)]
mod tests {
    use std::path::Path;
    use tempfile::tempdir;
    use crate::git::{repository, refs};
    use crate::error::GitError;

    #[test]
    fn test_branch_management() {
        let temp_dir = tempdir().unwrap();
        let repos_path = temp_dir.path();
        let repo_name = "test-branch-repo";

        // Init repo
        let repo = repository::init_bare(repos_path, repo_name).unwrap();

        // Initially no branches (bare repo is empty)
        let branches = refs::list_branches(repos_path, repo_name).unwrap();
        assert_eq!(branches.len(), 0);

        // We can't easily test create_branch without an initial commit,
        // so we'd need to mock an empty commit or use standard git commands 
        // to setup the repo. Since this is a bare repo, let's just test
        // the error paths.

        let err = refs::create_branch(repos_path, repo_name, "my-branch", "0000000000000000000000000000000000000000").unwrap_err();
        match err {
            GitError::ObjectNotFound { .. } => {},
            _ => panic!("Expected ObjectNotFound, got {:?}", err),
        }

        let err_del = refs::delete_branch(repos_path, repo_name, "nonexistent").unwrap_err();
        match err_del {
            GitError::RefNotFound { .. } => {},
            _ => panic!("Expected RefNotFound, got {:?}", err_del),
        }
    }
}
