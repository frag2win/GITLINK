/**
 * RepoCard — Repository card for list views
 * Displays repo name, description, stats, and last activity.
 */

import { Link } from 'react-router-dom';
import type { Repo } from '@/types';
import { formatBytes, timeAgo } from '@/utils/format';

interface RepoCardProps {
  repo: Repo;
}

export default function RepoCard({ repo }: RepoCardProps) {
  return (
    <Link
      to={`/repos/${repo.name}`}
      className="card group block transition-shadow hover:shadow-md"
    >
      {/* Header row */}
      <div className="flex items-start justify-between">
        <div className="flex items-center gap-2">
          {/* Repo icon */}
          <svg
            className="h-5 w-5 shrink-0 text-surface-400 group-hover:text-brand-500"
            viewBox="0 0 16 16"
            fill="currentColor"
          >
            <path d="M2 2.5A2.5 2.5 0 0 1 4.5 0h8.75a.75.75 0 0 1 .75.75v12.5a.75.75 0 0 1-.75.75h-2.5a.75.75 0 0 1 0-1.5h1.75v-2h-8a1 1 0 0 0-.714 1.7.75.75 0 1 1-1.072 1.05A2.495 2.495 0 0 1 2 11.5Zm10.5-1h-8a1 1 0 0 0-1 1v6.708A2.486 2.486 0 0 1 4.5 9h8ZM5 12.25a.25.25 0 0 1 .25-.25h3.5a.25.25 0 0 1 .25.25v3.25a.25.25 0 0 1-.4.2l-1.45-1.087a.249.249 0 0 0-.3 0L5.4 15.7a.25.25 0 0 1-.4-.2Z" />
          </svg>
          <h3 className="font-semibold text-surface-900 group-hover:text-brand-600 dark:text-surface-100">
            {repo.name}
          </h3>
          {repo.isPrivate && (
            <span className="rounded-full border border-surface-300 px-1.5 py-0.5 text-[10px] font-medium text-surface-500 dark:border-surface-600">
              Private
            </span>
          )}
        </div>
      </div>

      {/* Description */}
      {repo.description && (
        <p className="mt-2 line-clamp-2 text-sm text-surface-500">
          {repo.description}
        </p>
      )}

      {/* Stats row */}
      <div className="mt-3 flex items-center gap-4 text-xs text-surface-400">
        <span className="flex items-center gap-1">
          <svg className="h-3.5 w-3.5" viewBox="0 0 16 16" fill="currentColor">
            <path d="M11.75 2.5a.75.75 0 1 0 0 1.5.75.75 0 0 0 0-1.5zm-2.25.75a2.25 2.25 0 1 1 3 2.122V6A2.5 2.5 0 0 1 10 8.5H6a1 1 0 0 0-1 1v1.128a2.251 2.251 0 1 1-1.5 0V5.372a2.25 2.25 0 1 1 1.5 0v1.836A2.493 2.493 0 0 1 6 7h4a1 1 0 0 0 1-1v-.628A2.25 2.25 0 0 1 9.5 3.25zM4.25 12a.75.75 0 1 0 0 1.5.75.75 0 0 0 0-1.5zM3.5 3.25a.75.75 0 1 1 1.5 0 .75.75 0 0 1-1.5 0z" />
          </svg>
          {repo.branchCount} branch{repo.branchCount !== 1 ? 'es' : ''}
        </span>
        <span>{formatBytes(repo.sizeBytes)}</span>
        <span className="ml-auto">
          Updated {timeAgo(repo.updatedAt)}
        </span>
      </div>
    </Link>
  );
}
