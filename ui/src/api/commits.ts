/**
 * LocalRepo — Commit API
 * Read-only operations for browsing commit history.
 */

import { get } from './client';
import type { Commit, PaginatedResponse } from '@/types';

/** Optional filters when listing commits. */
export interface ListCommitsParams {
  /** Branch or ref to list commits from. */
  ref?: string;
  /** Pagination page number. */
  page?: number;
  /** Number of commits per page. */
  pageSize?: number;
  /** Filter commits by author name or email. */
  author?: string;
  /** Only show commits after this ISO-8601 date. */
  since?: string;
  /** Only show commits before this ISO-8601 date. */
  until?: string;
}

/** Fetch paginated commit history for a repository. */
export function listCommits(
  repoName: string,
  params?: ListCommitsParams,
): Promise<PaginatedResponse<Commit>> {
  const query: Record<string, string> = {};
  if (params?.ref) query.ref = params.ref;
  if (params?.page) query.page = String(params.page);
  if (params?.pageSize) query.pageSize = String(params.pageSize);
  if (params?.author) query.author = params.author;
  if (params?.since) query.since = params.since;
  if (params?.until) query.until = params.until;
  return get<PaginatedResponse<Commit>>(
    `/repos/${encodeURIComponent(repoName)}/commits`,
    query,
  );
}

/** Fetch a single commit by its hash. */
export function getCommit(repoName: string, hash: string): Promise<Commit> {
  return get<Commit>(
    `/repos/${encodeURIComponent(repoName)}/commits/${encodeURIComponent(hash)}`,
  );
}
