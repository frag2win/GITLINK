use crate::error::GitError;

/// Validate and sanitize a repository name to prevent path traversal and ensure it only
/// contains safe characters.
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

/// Validate Git branch names using standard rules (no spaces, no .., no control chars, no ~, ^, :, ?, *, [, leading -, etc.).
pub fn validate_branch_name(name: &str) -> Result<(), GitError> {
    if name.is_empty() {
        return Err(GitError::Other("Branch name cannot be empty".to_string()));
    }
    if name.len() > 250 {
        return Err(GitError::Other("Branch name too long".to_string()));
    }
    if name.starts_with('-') {
        return Err(GitError::Other("Branch name cannot start with '-'".to_string()));
    }
    if name.ends_with('.') || name.ends_with('/') {
        return Err(GitError::Other("Branch name cannot end with '.' or '/'".to_string()));
    }

    // Common forbidden characters in git refs:
    for c in name.chars() {
        if c.is_ascii_control() || c == ' ' || c == '~' || c == '^' || c == ':' || c == '?' || c == '*' || c == '[' || c == '\\' {
            return Err(GitError::Other(format!("Invalid character '{}' in branch name", c)));
        }
    }

    if name.contains("..") || name.contains("@{") || name.contains("//") {
        return Err(GitError::Other("Branch name contains forbidden sequences".to_string()));
    }

    Ok(())
}

/// Validate Git tag names (similar to branch names).
pub fn validate_tag_name(name: &str) -> Result<(), GitError> {
    if name.is_empty() {
        return Err(GitError::Other("Tag name cannot be empty".to_string()));
    }
    if name.len() > 250 {
        return Err(GitError::Other("Tag name too long".to_string()));
    }
    if name.starts_with('-') {
        return Err(GitError::Other("Tag name cannot start with '-'".to_string()));
    }
    if name.ends_with('.') || name.ends_with('/') {
        return Err(GitError::Other("Tag name cannot end with '.' or '/'".to_string()));
    }

    for c in name.chars() {
        if c.is_ascii_control() || c == ' ' || c == '~' || c == '^' || c == ':' || c == '?' || c == '*' || c == '[' || c == '\\' {
            return Err(GitError::Other(format!("Invalid character '{}' in tag name", c)));
        }
    }

    if name.contains("..") || name.contains("@{") || name.contains("//") {
        return Err(GitError::Other("Tag name contains forbidden sequences".to_string()));
    }

    Ok(())
}

/// Validate relative file path inside repository, preventing traversal.
pub fn validate_file_path(path: &str) -> Result<(), GitError> {
    if path.is_empty() {
        return Ok(());
    }

    // Disallow absolute paths
    if path.starts_with('/') || path.starts_with('\\') {
        return Err(GitError::Other("Absolute paths are not allowed".to_string()));
    }

    let normalized = path.replace('\\', "/");
    for segment in normalized.split('/') {
        if segment == ".." {
            return Err(GitError::Other("Path traversal sequence '..' is forbidden".to_string()));
        }
    }

    if path.contains('\0') {
        return Err(GitError::Other("Path contains null byte".to_string()));
    }

    Ok(())
}

/// Validate revision string (e.g. SHA-1 hash, branch name, ref string).
pub fn validate_revision(rev: &str) -> Result<(), GitError> {
    if rev.is_empty() {
        return Err(GitError::Other("Revision cannot be empty".to_string()));
    }

    // Revision shouldn't look like option flags or command execution arguments
    if rev.starts_with('-') {
        return Err(GitError::Other("Revision cannot start with '-'".to_string()));
    }

    for c in rev.chars() {
        if c.is_ascii_control() || c == ' ' || c == ';' || c == '&' || c == '|' || c == '$' || c == '`' || c == '"' || c == '\'' || c == '\\' {
            return Err(GitError::Other(format!("Invalid character '{}' in revision string", c)));
        }
    }

    Ok(())
}

