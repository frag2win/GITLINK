/**
 * Dashboard — Main landing page
 * Shows recent activity, repository statistics, and connection status overview.
 */

import { useConnection } from '@/hooks/useConnection';
import { useApi } from '@/hooks/useApi';
import { listRepos } from '@/api/repos';
import StatusBadge from '@/components/common/StatusBadge';
import LoadingSpinner from '@/components/common/LoadingSpinner';
import type { Repo, PaginatedResponse } from '@/types';
import { timeAgo, formatBytes } from '@/utils/format';
import { Link } from 'react-router-dom';

export default function Dashboard() {
  const { status } = useConnection();
  const { data, loading } = useApi<PaginatedResponse<Repo>>(
    () => listRepos({ pageSize: 5, sort: 'updated', order: 'desc' }),
    [],
  );

  const repos = data?.items ?? [];
  const totalRepos = data?.total ?? 0;

  return (
    <div className="space-y-6">
      {/* Page title */}
      <div>
        <h1 className="text-2xl font-bold text-surface-900 dark:text-surface-50">
          Dashboard
        </h1>
        <p className="mt-1 text-sm text-surface-500">
          Welcome to LocalRepo — your local-first, peer-to-peer Git hosting platform.
        </p>
      </div>

      {/* Status cards */}
      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        {/* Connection */}
        <div className="card">
          <h3 className="text-xs font-medium uppercase tracking-wider text-surface-400">
            Connection
          </h3>
          <div className="mt-2 flex items-center gap-2">
            <StatusBadge
              status={status.isOnline ? (status.syncState === 'syncing' ? 'syncing' : 'online') : 'offline'}
              size="md"
            />
          </div>
          <p className="mt-2 text-xs text-surface-500">
            {status.connectedPeers} connected peer{status.connectedPeers !== 1 ? 's' : ''}
          </p>
        </div>

        {/* Repositories */}
        <div className="card">
          <h3 className="text-xs font-medium uppercase tracking-wider text-surface-400">
            Repositories
          </h3>
          <p className="mt-2 text-3xl font-bold text-surface-900 dark:text-surface-50">
            {loading ? '—' : totalRepos}
          </p>
          <p className="mt-1 text-xs text-surface-500">Total repositories</p>
        </div>

        {/* Last sync */}
        <div className="card">
          <h3 className="text-xs font-medium uppercase tracking-wider text-surface-400">
            Last Sync
          </h3>
          <p className="mt-2 text-lg font-semibold text-surface-900 dark:text-surface-50">
            {status.lastSyncAt ? timeAgo(status.lastSyncAt) : 'Never'}
          </p>
          <p className="mt-1 text-xs text-surface-500">
            {status.syncState === 'error' ? status.errorMessage ?? 'Sync error' : 'Sync status OK'}
          </p>
        </div>

        {/* Node ID */}
        <div className="card">
          <h3 className="text-xs font-medium uppercase tracking-wider text-surface-400">
            Node ID
          </h3>
          <p className="mt-2 truncate font-mono text-sm text-surface-700 dark:text-surface-300">
            {status.peerId ?? 'Not connected'}
          </p>
          <p className="mt-1 text-xs text-surface-500">Your peer identity</p>
        </div>
      </div>

      {/* Recent repositories */}
      <div className="card">
        <div className="mb-4 flex items-center justify-between">
          <h2 className="text-lg font-semibold">Recently Updated</h2>
          <Link to="/repos" className="text-sm text-brand-600 hover:underline">
            View all →
          </Link>
        </div>

        {loading ? (
          <div className="py-8">
            <LoadingSpinner label="Loading repositories…" />
          </div>
        ) : repos.length === 0 ? (
          <p className="py-8 text-center text-sm text-surface-400">
            No repositories yet.{' '}
            <Link to="/repos" className="text-brand-600 hover:underline">
              Create your first repository
            </Link>
          </p>
        ) : (
          <ul className="divide-y divide-surface-100 dark:divide-surface-800">
            {repos.map((repo) => (
              <li key={repo.name}>
                <Link
                  to={`/repos/${repo.name}`}
                  className="flex items-center justify-between py-3 transition-colors hover:bg-surface-50 dark:hover:bg-surface-800/50"
                >
                  <div>
                    <span className="font-medium text-surface-900 dark:text-surface-100">
                      {repo.name}
                    </span>
                    {repo.description && (
                      <p className="mt-0.5 text-xs text-surface-500">
                        {repo.description}
                      </p>
                    )}
                  </div>
                  <div className="text-right text-xs text-surface-400">
                    <div>{formatBytes(repo.sizeBytes)}</div>
                    <div>Updated {timeAgo(repo.updatedAt)}</div>
                  </div>
                </Link>
              </li>
            ))}
          </ul>
        )}
      </div>
    </div>
  );
}
