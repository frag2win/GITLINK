/**
 * LocalRepo — Pull Request API
 * Operations for managing pull requests.
 */

import { get, post } from './client';
import type { PullRequest } from '@/types';

export interface CreatePullRequestParams {
  title: string;
  description: string;
  baseBranch: string;
  headBranch: string;
}

export interface MergePullRequestResponse {
  status: string;
  hash: string;
}

/** List all pull requests for a given repository. */
export function listPullRequests(repoName: string): Promise<PullRequest[]> {
  return get<PullRequest[]>(`/repos/${encodeURIComponent(repoName)}/pulls`);
}

/** Create a new pull request. */
export function createPullRequest(repoName: string, params: CreatePullRequestParams): Promise<PullRequest> {
  return post<PullRequest>(`/repos/${encodeURIComponent(repoName)}/pulls`, params);
}

/** Merge a pull request. */
export function mergePullRequest(repoName: string, prId: number | string): Promise<MergePullRequestResponse> {
  return post<MergePullRequestResponse>(`/repos/${encodeURIComponent(repoName)}/pulls/${encodeURIComponent(prId)}/merge`, {});
}
