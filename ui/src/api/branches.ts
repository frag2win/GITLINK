/**
 * LocalRepo — Branch API
 * Operations for managing branches within a repository.
 */

import { get, post, del } from './client';
import type { Branch } from '@/types';

/** Parameters for creating a new branch. */
export interface CreateBranchParams {
  /** Name for the new branch. */
  name: string;
  /** Source branch or commit hash to branch from. */
  startPoint?: string;
}

/** List all branches for a given repository. */
export function listBranches(repoName: string): Promise<Branch[]> {
  return get<Branch[]>(`/repos/${encodeURIComponent(repoName)}/branches`);
}

/** Create a new branch in the repository. */
export function createBranch(repoName: string, params: CreateBranchParams): Promise<Branch> {
  return post<Branch>(`/repos/${encodeURIComponent(repoName)}/branches`, params);
}

/** Delete a branch from the repository. Cannot delete the default branch. */
export function deleteBranch(repoName: string, branchName: string): Promise<void> {
  return del(`/repos/${encodeURIComponent(repoName)}/branches/${encodeURIComponent(branchName)}`);
}
