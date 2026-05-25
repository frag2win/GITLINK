/**
 * LocalRepo — Contributor API
 * Manage contributors (access control) for repositories.
 */

import { get, post, del } from './client';
import type { Contributor } from '@/types';

/** Parameters for adding a contributor to a repository. */
export interface AddContributorParams {
  /** Display name of the contributor. */
  name: string;
  /** Email address. */
  email: string;
  /** Permission role to grant. */
  role: 'admin' | 'write' | 'read';
  /** SSH public key (optional, can be added later). */
  sshPublicKey?: string;
}

/** List all contributors for a repository. */
export function listContributors(repoName: string): Promise<Contributor[]> {
  return get<Contributor[]>(`/repos/${encodeURIComponent(repoName)}/contributors`);
}

/** Add a new contributor to a repository. */
export function addContributor(
  repoName: string,
  params: AddContributorParams,
): Promise<Contributor> {
  return post<Contributor>(
    `/repos/${encodeURIComponent(repoName)}/contributors`,
    params,
  );
}

/** Remove a contributor from a repository. */
export function removeContributor(repoName: string, contributorId: string): Promise<void> {
  return del(
    `/repos/${encodeURIComponent(repoName)}/contributors/${encodeURIComponent(contributorId)}`,
  );
}
