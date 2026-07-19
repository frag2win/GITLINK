#[cfg(test)]
mod tests {
    use std::path::Path;
    use tempfile::tempdir;
    use crate::git::{repository, commits, objects};
    use crate::error::GitError;

    #[test]
    fn test_commit_and_tree_errors() {
        let temp_dir = tempdir().unwrap();
        let repos_path = temp_dir.path();
        let repo_name = "test-commit-repo";

        repository::init_bare(repos_path, repo_name).unwrap();

        // Get commit that doesn't exist
        let err = commits::get_commit_detail(repos_path, repo_name, "0000000000000000000000000000000000000000").unwrap_err();
        match err {
            GitError::ObjectNotFound { .. } => {},
            _ => panic!("Expected ObjectNotFound, got {:?}", err),
        }

        // Tree Traversal that doesn't exist
        let err_tree = objects::read_tree_by_path(repos_path, repo_name, "HEAD", "").unwrap_err();
        match err_tree {
            GitError::RefNotFound { .. } => {},
            _ => panic!("Expected RefNotFound (no HEAD), got {:?}", err_tree),
        }
    }
}
