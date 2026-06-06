use crate::error::GitError;

/// Validate and sanitize a repository name to prevent path traversal and ensure it only
/// contains safe characters.
///
/// # Errors
/// Returns [`GitError::InvalidRepoName`] if the name is empty, too long, or contains
/// invalid characters.
pub fn validate_repo_name(name: &str) -> Result<(), GitError> {
    if name.is_empty() {
        return Err(GitError::InvalidRepoName("Repository name cannot be empty".to_string()));
    }

    if name.len() > 100 {
        return Err(GitError::InvalidRepoName("Repository name too long (max 100 chars)".to_string()));
    }

    if name == "." || name == ".." {
        return Err(GitError::InvalidRepoName("Repository name cannot be '.' or '..'".to_string()));
    }

    // Only allow alphanumeric characters, dashes, underscores, and dots.
    for c in name.chars() {
        if !c.is_ascii_alphanumeric() && c != '-' && c != '_' && c != '.' {
            return Err(GitError::InvalidRepoName(format!("Invalid character '{}' in repository name", c)));
        }
    }

    // Explicitly reject any form of path traversal strings just to be safe,
    // though the character restriction above already handles slashes.
    if name.contains("..") {
        return Err(GitError::InvalidRepoName("Repository name cannot contain '..' (path traversal protection)".to_string()));
    }

    Ok(())
}
