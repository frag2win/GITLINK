//! # Merge Engine
//!
//! Implements the Git merge process broken into testable steps:
//! - [`AnalyzeMerge`]
//! - [`DetectConflicts`]
//! - [`FastForwardMerge`]
//! - [`ThreeWayMerge`]
//! - [`CreateMergeCommit`]
//! - [`UpdateReferences`]

use std::path::Path;
use git2::{Repository, Oid, Commit, Index, ObjectType, Signature};
use tracing::{debug, instrument};

use crate::error::GitError;

#[derive(Debug)]
pub struct MergeConflict {
    pub path: String,
    pub conflict_type: String,
    pub base_oid: String,
    pub ours_oid: String,
    pub theirs_oid: String,
    pub is_binary: bool,
}

#[derive(Debug)]
pub enum MergeResult {
    Success { new_head: String },
    Conflicts(Vec<MergeConflict>),
}

pub struct MergeEngine<'a> {
    repo: &'a Repository,
}

impl<'a> MergeEngine<'a> {
    pub fn new(repo: &'a Repository) -> Self {
        Self { repo }
    }

    /// Entry point for a merge request.
    #[instrument(skip(self))]
    pub fn execute(
        &self,
        base_branch: &str,
        head_branch: &str,
        author_name: &str,
        author_email: &str,
        message: &str,
    ) -> Result<MergeResult, GitError> {
        let ours_ref = self.repo.find_reference(base_branch)
            .map_err(|_| GitError::RefNotFound { name: base_branch.to_string() })?;
        let theirs_ref = self.repo.find_reference(head_branch)
            .map_err(|_| GitError::RefNotFound { name: head_branch.to_string() })?;

        let ours_commit = ours_ref.peel_to_commit()?;
        let theirs_commit = theirs_ref.peel_to_commit()?;

        let theirs_annotated = self.repo.find_annotated_commit(theirs_commit.id())?;

        let (analysis, _preference) = self.repo.merge_analysis(&[&theirs_annotated])?;

        if analysis.is_up_to_date() {
            return Err(GitError::Other("Already up to date".into()));
        }

        if analysis.is_fast_forward() {
            return self.fast_forward_merge(base_branch, &theirs_commit);
        } else if analysis.is_normal() {
            return self.three_way_merge(base_branch, &ours_commit, &theirs_commit, author_name, author_email, message);
        }

        Err(GitError::Other("Unsupported merge type".into()))
    }

    fn fast_forward_merge(&self, refname: &str, target_commit: &Commit) -> Result<MergeResult, GitError> {
        self.update_reference(refname, target_commit.id(), "fast-forward")?;
        Ok(MergeResult::Success {
            new_head: target_commit.id().to_string(),
        })
    }

    fn three_way_merge(
        &self,
        refname: &str,
        ours_commit: &Commit,
        theirs_commit: &Commit,
        author_name: &str,
        author_email: &str,
        message: &str,
    ) -> Result<MergeResult, GitError> {
        let mut index = self.repo.merge_commits(ours_commit, theirs_commit, None)?;

        if index.has_conflicts() {
            let conflicts = self.detect_conflicts(&mut index)?;
            return Ok(MergeResult::Conflicts(conflicts));
        }

        let new_head = self.create_merge_commit(&mut index, ours_commit, theirs_commit, author_name, author_email, message)?;
        self.update_reference(refname, new_head, "merge")?;

        Ok(MergeResult::Success {
            new_head: new_head.to_string(),
        })
    }

    fn detect_conflicts(&self, index: &mut Index) -> Result<Vec<MergeConflict>, GitError> {
        let mut conflicts = Vec::new();
        
        let iter = index.conflicts()?;
        for conflict in iter {
            let conflict = conflict?;
            
            // A conflict has up to 3 entries: ancestor, ours, theirs.
            let path = conflict.our.as_ref().or(conflict.their.as_ref()).or(conflict.ancestor.as_ref())
                .map(|e| String::from_utf8_lossy(&e.path).to_string())
                .unwrap_or_else(|| "unknown".to_string());

            let base_oid = conflict.ancestor.map(|e| e.id.to_string()).unwrap_or_default();
            let ours_oid = conflict.our.map(|e| e.id.to_string()).unwrap_or_default();
            let theirs_oid = conflict.their.map(|e| e.id.to_string()).unwrap_or_default();

            // Simple conflict typing
            let conflict_type = if ours_oid.is_empty() {
                "deleted_by_us".into()
            } else if theirs_oid.is_empty() {
                "deleted_by_them".into()
            } else {
                "content_conflict".into()
            };

            // Not performing deep binary detection here, just defaulting for now.
            // Would read blob to check for \0 bytes.
            let is_binary = false;

            conflicts.push(MergeConflict {
                path,
                conflict_type,
                base_oid,
                ours_oid,
                theirs_oid,
                is_binary,
            });
        }
        
        Ok(conflicts)
    }

    fn create_merge_commit(
        &self,
        index: &mut Index,
        ours: &Commit,
        theirs: &Commit,
        author_name: &str,
        author_email: &str,
        message: &str,
    ) -> Result<Oid, GitError> {
        let tree_oid = index.write_tree_to(self.repo)?;
        let tree = self.repo.find_tree(tree_oid)?;

        let sig = Signature::now(author_name, author_email)?;
        let commit_oid = self.repo.commit(
            None,
            &sig,
            &sig,
            message,
            &tree,
            &[ours, theirs],
        )?;

        Ok(commit_oid)
    }

    fn update_reference(&self, refname: &str, target: Oid, log_message: &str) -> Result<(), GitError> {
        let mut reference = self.repo.find_reference(refname)?;
        reference.set_target(target, log_message)?;
        Ok(())
    }
}
