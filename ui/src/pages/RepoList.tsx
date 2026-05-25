/**
 * RepoList — Repository list page
 * Shows all repositories with search, sort, and a create button.
 */

import { useState } from 'react';
import { useApi } from '@/hooks/useApi';
import { listRepos, type ListReposParams } from '@/api/repos';
import RepoCard from '@/components/repo/RepoCard';
import CreateRepoModal from '@/components/repo/CreateRepoModal';
import LoadingSpinner from '@/components/common/LoadingSpinner';
import EmptyState from '@/components/common/EmptyState';
import type { Repo, PaginatedResponse } from '@/types';

export default function RepoList() {
  const [showCreate, setShowCreate] = useState(false);
  const [search, setSearch] = useState('');
  const [sortBy, setSortBy] = useState<ListReposParams['sort']>('updated');

  const { data, loading, refetch } = useApi<PaginatedResponse<Repo>>(
    () => listRepos({ search, sort: sortBy, order: 'desc', pageSize: 50 }),
    [search, sortBy],
  );

  const repos = data?.items ?? [];

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h1 className="text-2xl font-bold text-surface-900 dark:text-surface-50">
            Repositories
          </h1>
          <p className="mt-1 text-sm text-surface-500">
            {data ? `${data.total} repositories` : 'Loading…'}
          </p>
        </div>
        <button
          onClick={() => setShowCreate(true)}
          className="btn-primary"
        >
          + New Repository
        </button>
      </div>

      {/* Filters */}
      <div className="flex flex-col gap-3 sm:flex-row">
        <input
          type="text"
          className="input max-w-sm"
          placeholder="Search repositories…"
          value={search}
          onChange={(e) => setSearch(e.target.value)}
        />
        <select
          className="input max-w-[200px]"
          value={sortBy}
          onChange={(e) => setSortBy(e.target.value as ListReposParams['sort'])}
        >
          <option value="updated">Recently updated</option>
          <option value="created">Recently created</option>
          <option value="name">Name</option>
          <option value="size">Size</option>
        </select>
      </div>

      {/* Content */}
      {loading ? (
        <div className="py-16">
          <LoadingSpinner label="Loading repositories…" size="lg" />
        </div>
      ) : repos.length === 0 ? (
        <EmptyState
          title="No repositories found"
          description={
            search
              ? `No repositories match "${search}". Try a different search term.`
              : 'Create your first repository to get started with LocalRepo.'
          }
          icon={
            <svg className="h-16 w-16" viewBox="0 0 16 16" fill="currentColor">
              <path d="M2 2.5A2.5 2.5 0 0 1 4.5 0h8.75a.75.75 0 0 1 .75.75v12.5a.75.75 0 0 1-.75.75h-2.5a.75.75 0 0 1 0-1.5h1.75v-2h-8a1 1 0 0 0-.714 1.7.75.75 0 1 1-1.072 1.05A2.495 2.495 0 0 1 2 11.5Zm10.5-1h-8a1 1 0 0 0-1 1v6.708A2.486 2.486 0 0 1 4.5 9h8ZM5 12.25a.25.25 0 0 1 .25-.25h3.5a.25.25 0 0 1 .25.25v3.25a.25.25 0 0 1-.4.2l-1.45-1.087a.249.249 0 0 0-.3 0L5.4 15.7a.25.25 0 0 1-.4-.2Z" />
            </svg>
          }
          action={
            !search ? (
              <button onClick={() => setShowCreate(true)} className="btn-primary">
                + Create Repository
              </button>
            ) : undefined
          }
        />
      ) : (
        <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-3">
          {repos.map((repo) => (
            <RepoCard key={repo.name} repo={repo} />
          ))}
        </div>
      )}

      {/* Create modal */}
      <CreateRepoModal
        isOpen={showCreate}
        onClose={() => setShowCreate(false)}
        onCreated={refetch}
      />
    </div>
  );
}
