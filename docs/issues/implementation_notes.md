# Implementation Issues, Errors, and Skipped Parts

This document records any bugs encountered, workarounds used, and parts of the architecture or implementation that were skipped or modified during development.

## 1. Rust Toolchain on Windows (cargo check failure)
- **Issue:** Running `cargo check` locally in the `services/git-server` directory fails with `error: error calling dlltool 'dlltool.exe': program not found`.
- **Reason:** The local Windows environment is missing specific build components (like `dlltool.exe`) for compiling some Rust dependencies.
- **Resolution/Status:** Skipped locally. The project relies on containerization (`distroless`/Linux environments via Docker) for the actual deployment of the Rust `git-server`. The local Windows failure is ignored to focus on the business logic.


## 2. Missing Error Variant in Scaffold
- **Issue:** The `services/git-server/src/git/sanitize.rs` required throwing an error for invalid repository names, but the `GitError` enum in `services/git-server/src/error.rs` did not include a specific variant for this.
- **Resolution/Status:** Added `InvalidRepoName(String)` to `GitError` manually during the implementation of `sanitize.rs` to allow for clean, typed error handling instead of falling back to a generic `Other` variant.

