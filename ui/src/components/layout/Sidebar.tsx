/**
 * Sidebar — Side navigation panel
 * Displays a list of repositories for quick navigation
 * and quick-action buttons.
 */

import { Link, useLocation } from 'react-router-dom';
import { useApi } from '@/hooks/useApi';
import { listRepos } from '@/api/repos';
import type { Repo, PaginatedResponse } from '@/types';

export default function Sidebar() {
  const location = useLocation();
  const { data, loading } = useApi<PaginatedResponse<Repo>>(
    () => listRepos({ pageSize: 20, sort: 'updated', order: 'desc' }),
    [],
  );

  const repos = data?.items ?? [];

  return (
    <aside className="hidden w-60 shrink-0 border-r border-surface-200 bg-surface-50 dark:border-surface-700 dark:bg-surface-900 lg:block">
      <div className="flex h-full flex-col">
        {/* Quick Actions */}
        <div className="border-b border-surface-200 p-4 dark:border-surface-700">
          <h3 className="mb-2 text-xs font-semibold uppercase tracking-wider text-surface-500">
            Quick Actions
          </h3>
          <Link
            to="/repos"
            className="btn-primary w-full text-center text-xs"
          >
            + New Repository
          </Link>
        </div>

        {/* Repository list */}
        <div className="flex-1 overflow-y-auto p-4">
          <h3 className="mb-2 text-xs font-semibold uppercase tracking-wider text-surface-500">
            Repositories
          </h3>

          {loading && (
            <div className="space-y-2">
              {[1, 2, 3].map((i) => (
                <div
                  key={i}
                  className="h-8 animate-pulse rounded bg-surface-200 dark:bg-surface-700"
                />
              ))}
            </div>
          )}

          {!loading && repos.length === 0 && (
            <p className="text-xs text-surface-400">No repositories yet.</p>
          )}

          <ul className="space-y-0.5">
            {repos.map((repo) => {
              const isActive = location.pathname === `/repos/${repo.name}`;
              return (
                <li key={repo.name}>
                  <Link
                    to={`/repos/${repo.name}`}
                    className={`flex items-center gap-2 rounded-md px-2 py-1.5 text-sm transition-colors ${
                      isActive
                        ? 'bg-brand-50 font-medium text-brand-700 dark:bg-brand-950 dark:text-brand-300'
                        : 'text-surface-700 hover:bg-surface-100 dark:text-surface-300 dark:hover:bg-surface-800'
                    }`}
                  >
                    {/* Repo icon */}
                    <svg
                      className="h-4 w-4 shrink-0 text-surface-400"
                      viewBox="0 0 16 16"
                      fill="currentColor"
                    >
                      <path d="M2 2.5A2.5 2.5 0 0 1 4.5 0h8.75a.75.75 0 0 1 .75.75v12.5a.75.75 0 0 1-.75.75h-2.5a.75.75 0 0 1 0-1.5h1.75v-2h-8a1 1 0 0 0-.714 1.7.75.75 0 1 1-1.072 1.05A2.495 2.495 0 0 1 2 11.5Zm10.5-1h-8a1 1 0 0 0-1 1v6.708A2.486 2.486 0 0 1 4.5 9h8ZM5 12.25a.25.25 0 0 1 .25-.25h3.5a.25.25 0 0 1 .25.25v3.25a.25.25 0 0 1-.4.2l-1.45-1.087a.249.249 0 0 0-.3 0L5.4 15.7a.25.25 0 0 1-.4-.2Z" />
                    </svg>
                    <span className="truncate">{repo.name}</span>
                    {repo.isPrivate && (
                      <span className="ml-auto text-[10px] text-surface-400">🔒</span>
                    )}
                  </Link>
                </li>
              );
            })}
          </ul>
        </div>
      </div>
    </aside>
  );
}
