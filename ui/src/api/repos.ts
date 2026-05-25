/**
 * LocalRepo — Repository API
 * CRUD operations for Git repositories.
 */

import { get, post, del } from './client';
import type { Repo, PaginatedResponse } from '@/types';

/** Parameters for creating a new repository. */
export interface CreateRepoParams {
  name: string;
  description?: string;
  defaultBranch?: string;
  isPrivate?: boolean;
}

/** Optional filters when listing repositories. */
export interface ListReposParams {
  page?: number;
  pageSize?: number;
  search?: string;
  sort?: 'name' | 'updated' | 'created' | 'size';
  order?: 'asc' | 'desc';
}

/** Fetch a paginated list of repositories. */
export function listRepos(params?: ListReposParams): Promise<PaginatedResponse<Repo>> {
  const query: Record<string, string> = {};
  if (params?.page) query.page = String(params.page);
  if (params?.pageSize) query.pageSize = String(params.pageSize);
  if (params?.search) query.search = params.search;
  if (params?.sort) query.sort = params.sort;
  if (params?.order) query.order = params.order;
  return get<PaginatedResponse<Repo>>('/repos', query);
}

/** Fetch a single repository by name. */
export function getRepo(name: string): Promise<Repo> {
  return get<Repo>(`/repos/${encodeURIComponent(name)}`);
}

/** Create a new repository. */
export function createRepo(params: CreateRepoParams): Promise<Repo> {
  return post<Repo>('/repos', params);
}

/** Delete a repository by name. This action is irreversible. */
export function deleteRepo(name: string): Promise<void> {
  return del(`/repos/${encodeURIComponent(name)}`);
}
