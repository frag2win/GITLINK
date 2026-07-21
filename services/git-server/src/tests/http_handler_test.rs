#[cfg(test)]
mod tests {
    use crate::config::Config;
    use crate::handlers::http_handler::{
        handle_info_refs, handle_receive_pack, handle_upload_pack,
    };
    use crate::socket::protocol::{InfoRefsRequest, ReceivePackRequest, UploadPackRequest};

    /// Build a minimal Config pointing at a temp directory.
    /// No real repos need to exist — these tests should all be rejected
    /// by validate_repo_name() *before* the filesystem is ever touched.
    fn make_config(repos_path: &str) -> Config {
        Config {
            repos_path: repos_path.to_string(),
            ipc_network: "unix".to_string(),
            ipc_address: "/tmp/test.sock".to_string(),
            log_level: "error".to_string(),
        }
    }

    /// Returns true if the response carries a ValidationError code,
    /// which is what all three handlers set when validate_repo_name rejects input.
    fn is_validation_error(resp: &crate::socket::protocol::GitCommandResponse) -> bool {
        resp.error
            .as_ref()
            .map(|e| e.code == "ValidationError")
            .unwrap_or(false)
    }

    // ── Traversal / injection inputs that MUST be rejected ───────────────────

    const INVALID_REPO_NAMES: &[&str] = &[
        "",                        // empty
        "../etc",                  // classic dot-dot traversal
        "../../etc/passwd",        // deeper traversal
        "/etc/passwd",             // absolute path (Path::join footgun)
        "repo/subdir",             // slash — would escape repos_path via join
        "repo\\subdir",            // backslash variant
        "repo name",               // space (shell injection vector)
        "repo;evil",               // semicolon (shell injection)
        "repo|evil",               // pipe
        "repo`evil`",              // backtick
        "repo$evil",               // dollar sign
        "a".repeat(101).as_str(),  // exceeds max length
    ];

    // ── Valid inputs that must NOT be rejected by the name validator ──────────
    // (The actual Git operation will fail because no real repo exists, but
    //  the error code must be "GitError", not "ValidationError".)

    const VALID_REPO_NAMES: &[&str] = &[
        "myrepo",
        "my-repo",
        "my_repo.git",
        "repo1",
        "My.Repo-v2_final",
    ];

    // ─────────────────────────────────────────────────────────────────────────
    // handle_info_refs
    // ─────────────────────────────────────────────────────────────────────────

    #[tokio::test]
    async fn test_info_refs_rejects_traversal_names() {
        let config = make_config("/nonexistent/repos");
        for &name in INVALID_REPO_NAMES {
            let req = InfoRefsRequest {
                repo_name: name.to_string(),
                service: "git-upload-pack".to_string(),
            };
            let resp = handle_info_refs(req, &config).await;
            assert!(
                is_validation_error(&resp),
                "handle_info_refs: expected ValidationError for repo_name={:?}, got: {:?}",
                name,
                resp.error
            );
        }
    }

    #[tokio::test]
    async fn test_info_refs_passes_valid_names_to_git_layer() {
        let config = make_config("/nonexistent/repos");
        for &name in VALID_REPO_NAMES {
            let req = InfoRefsRequest {
                repo_name: name.to_string(),
                service: "git-upload-pack".to_string(),
            };
            let resp = handle_info_refs(req, &config).await;
            assert!(
                !is_validation_error(&resp),
                "handle_info_refs: unexpectedly rejected valid repo_name={:?}",
                name
            );
        }
    }

    // ─────────────────────────────────────────────────────────────────────────
    // handle_upload_pack
    // ─────────────────────────────────────────────────────────────────────────

    #[tokio::test]
    async fn test_upload_pack_rejects_traversal_names() {
        let config = make_config("/nonexistent/repos");
        for &name in INVALID_REPO_NAMES {
            let req = UploadPackRequest {
                repo_name: name.to_string(),
                body: vec![],
            };
            let resp = handle_upload_pack(req, &config).await;
            assert!(
                is_validation_error(&resp),
                "handle_upload_pack: expected ValidationError for repo_name={:?}, got: {:?}",
                name,
                resp.error
            );
        }
    }

    #[tokio::test]
    async fn test_upload_pack_passes_valid_names_to_git_layer() {
        let config = make_config("/nonexistent/repos");
        for &name in VALID_REPO_NAMES {
            let req = UploadPackRequest {
                repo_name: name.to_string(),
                body: vec![],
            };
            let resp = handle_upload_pack(req, &config).await;
            assert!(
                !is_validation_error(&resp),
                "handle_upload_pack: unexpectedly rejected valid repo_name={:?}",
                name
            );
        }
    }

    // ─────────────────────────────────────────────────────────────────────────
    // handle_receive_pack
    // ─────────────────────────────────────────────────────────────────────────

    #[tokio::test]
    async fn test_receive_pack_rejects_traversal_names() {
        let config = make_config("/nonexistent/repos");
        for &name in INVALID_REPO_NAMES {
            let req = ReceivePackRequest {
                repo_name: name.to_string(),
                body: vec![],
            };
            let resp = handle_receive_pack(req, &config).await;
            assert!(
                is_validation_error(&resp),
                "handle_receive_pack: expected ValidationError for repo_name={:?}, got: {:?}",
                name,
                resp.error
            );
        }
    }

    #[tokio::test]
    async fn test_receive_pack_passes_valid_names_to_git_layer() {
        let config = make_config("/nonexistent/repos");
        for &name in VALID_REPO_NAMES {
            let req = ReceivePackRequest {
                repo_name: name.to_string(),
                body: vec![],
            };
            let resp = handle_receive_pack(req, &config).await;
            assert!(
                !is_validation_error(&resp),
                "handle_receive_pack: unexpectedly rejected valid repo_name={:?}",
                name
            );
        }
    }
}
