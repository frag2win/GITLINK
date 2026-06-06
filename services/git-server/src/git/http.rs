use std::process::{Command, Stdio};
use std::io::Write;
use crate::error::GitError;

pub fn info_refs(repo_path: &str, service: &str) -> Result<Vec<u8>, GitError> {
    // service is "git-upload-pack" or "git-receive-pack"
    let output = Command::new(service)
        .arg("--stateless-rpc")
        .arg("--advertise-refs")
        .arg(repo_path)
        .output()
        .map_err(|e| GitError::IoError(e))?;

    if !output.status.success() {
        return Err(GitError::GitCommandFailed(
            String::from_utf8_lossy(&output.stderr).to_string()
        ));
    }

    Ok(output.stdout)
}

pub fn upload_pack(repo_path: &str, input: &[u8]) -> Result<Vec<u8>, GitError> {
    let mut child = Command::new("git-upload-pack")
        .arg("--stateless-rpc")
        .arg(repo_path)
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .map_err(|e| GitError::IoError(e))?;

    // Write the git client's request body to git-upload-pack stdin
    if let Some(mut stdin) = child.stdin.take() {
        stdin.write_all(input).map_err(|e| GitError::IoError(e))?;
    }

    let output = child.wait_with_output()
        .map_err(|e| GitError::IoError(e))?;

    Ok(output.stdout)
}

pub fn receive_pack(repo_path: &str, input: &[u8]) -> Result<Vec<u8>, GitError> {
    let mut child = Command::new("git-receive-pack")
        .arg("--stateless-rpc")
        .arg(repo_path)
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .stderr(Stdio::piped())
        .spawn()
        .map_err(|e| GitError::IoError(e))?;

    if let Some(mut stdin) = child.stdin.take() {
        stdin.write_all(input).map_err(|e| GitError::IoError(e))?;
    }

    let output = child.wait_with_output()
        .map_err(|e| GitError::IoError(e))?;

    Ok(output.stdout)
}
