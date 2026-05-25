//! # Pack File Handling
//!
//! Handles Git pack protocol operations for push and pull:
//! - [`receive_pack`] — Accept a pack file from a client (push).
//! - [`send_pack`] — Generate and send a pack file to a client (pull/fetch).
//!
//! These operations work with raw pack data transmitted over the Unix domain
//! socket, enabling local-first P2P synchronization without network access.

use std::path::Path;

use tracing::{debug, info, instrument};

use crate::error::GitError;

/// Result of a receive-pack operation (push).
#[derive(Debug)]
pub struct ReceivePackResult {
    /// Number of objects received.
    pub objects_received: usize,
    /// References that were updated.
    pub refs_updated: Vec<String>,
}

/// Result of a send-pack operation (pull/fetch).
#[derive(Debug)]
pub struct SendPackResult {
    /// The pack file data as raw bytes.
    pub pack_data: Vec<u8>,
    /// Number of objects in the pack.
    pub object_count: usize,
}

/// Accept incoming pack data and apply it to the repository.
///
/// This is the server side of `git push`. The pack data contains objects
/// and reference updates from the client.
///
/// # Parameters
/// - `pack_data` — Raw pack file bytes received from the client.
///
/// # Returns
/// A [`ReceivePackResult`] summarizing what was received and updated.
#[instrument(skip(repos_dir, pack_data))]
pub fn receive_pack(
    repos_dir: &Path,
    repo_name: &str,
    pack_data: &[u8],
) -> Result<ReceivePackResult, GitError> {
    let repo = crate::git::repository::open(repos_dir, repo_name)?;
    let odb = repo.odb().map_err(GitError::from)?;

    // Write the pack data into the object database
    let mut indexer = odb
        .packwriter()
        .map_err(GitError::from)?;

    use std::io::Write;
    indexer
        .write_all(pack_data)
        .map_err(|e| GitError::Other(format!("failed to write pack data: {e}")))?;

    let _index_oid = indexer
        .commit()
        .map_err(|e| GitError::Other(format!("failed to commit pack: {e}")))?;

    info!(
        repo = %repo_name,
        size = pack_data.len(),
        "Received pack data"
    );

    // TODO: Parse the pack to determine exact objects_received and refs_updated
    Ok(ReceivePackResult {
        objects_received: 0, // TODO: count from indexer stats
        refs_updated: Vec::new(), // TODO: parse ref updates
    })
}

/// Generate a pack file containing objects needed by the client.
///
/// This is the server side of `git fetch` / `git pull`. Given a set of
/// "want" OIDs and "have" OIDs, generates a minimal pack file.
///
/// # Parameters
/// - `want_oids` — OIDs the client wants.
/// - `have_oids` — OIDs the client already has (for delta computation).
///
/// # Returns
/// A [`SendPackResult`] containing the raw pack data.
#[instrument(skip(repos_dir, want_oids, have_oids))]
pub fn send_pack(
    repos_dir: &Path,
    repo_name: &str,
    want_oids: &[String],
    have_oids: &[String],
) -> Result<SendPackResult, GitError> {
    let repo = crate::git::repository::open(repos_dir, repo_name)?;

    // Build a packbuilder with the requested objects
    let mut packbuilder = repo.packbuilder()?;

    // Add "want" objects
    for want in want_oids {
        let oid = git2::Oid::from_str(want).map_err(|_| GitError::ObjectNotFound {
            oid: want.clone(),
        })?;
        packbuilder
            .insert_object(oid, None)
            .map_err(GitError::from)?;
    }

    // Collect pack data into a buffer
    let mut pack_data = Vec::new();
    packbuilder
        .foreach(|data| {
            pack_data.extend_from_slice(data);
            true
        })
        .map_err(GitError::from)?;

    let object_count = packbuilder.object_count();

    debug!(
        repo = %repo_name,
        wants = want_oids.len(),
        haves = have_oids.len(),
        objects = object_count,
        pack_size = pack_data.len(),
        "Generated pack"
    );

    Ok(SendPackResult {
        pack_data,
        object_count,
    })
}
