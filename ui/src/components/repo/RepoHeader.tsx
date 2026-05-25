/**
 * RepoHeader — Repository detail page header
 * Shows repository name, description, clone URL, and key stats.
 */

import type { Repo } from '@/types';
import { formatBytes, timeAgo } from '@/utils/format';
import { useState } from 'react';

interface RepoHeaderProps {
  repo: Repo;
}

export default function RepoHeader({ repo }: RepoHeaderProps) {
  const [copied, setCopied] = useState(false);

  const handleCopyCloneUrl = async () => {
    try {
      await navigator.clipboard.writeText(repo.cloneUrl);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch {
      // Clipboard API may not be available in all contexts
    }
  };

  return (
    <div className="card">
      {/* Title row */}
      <div className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
        <div>
          <div className="flex items-center gap-2">
            <h1 className="text-2xl font-bold text-surface-900 dark:text-surface-50">
              {repo.name}
            </h1>
            {repo.isPrivate && (
              <span className="rounded-full border border-surface-300 px-2 py-0.5 text-xs font-medium text-surface-500 dark:border-surface-600">
                Private
              </span>
            )}
          </div>
          {repo.description && (
            <p className="mt-1 text-sm text-surface-500">{repo.description}</p>
          )}
        </div>

        {/* Clone URL */}
        <div className="flex items-center gap-2">
          <code className="rounded-md bg-surface-100 px-3 py-1.5 font-mono text-xs text-surface-700 dark:bg-surface-700 dark:text-surface-300">
            {repo.cloneUrl}
          </code>
          <button
            onClick={handleCopyCloneUrl}
            className="btn-secondary px-2 py-1.5 text-xs"
            title="Copy clone URL"
          >
            {copied ? '✓' : 'Copy'}
          </button>
        </div>
      </div>

      {/* Stats */}
      <div className="mt-4 flex flex-wrap items-center gap-6 border-t border-surface-200 pt-4 text-sm text-surface-500 dark:border-surface-700">
        <div className="flex items-center gap-1.5">
          <span className="font-medium text-surface-700 dark:text-surface-300">
            {repo.branchCount}
          </span>
          branch{repo.branchCount !== 1 ? 'es' : ''}
        </div>
        <div className="flex items-center gap-1.5">
          <span className="font-medium text-surface-700 dark:text-surface-300">
            {repo.contributorCount}
          </span>
          contributor{repo.contributorCount !== 1 ? 's' : ''}
        </div>
        <div>{formatBytes(repo.sizeBytes)}</div>
        <div className="ml-auto text-xs">
          Created {timeAgo(repo.createdAt)} · Updated {timeAgo(repo.updatedAt)}
        </div>
      </div>
    </div>
  );
}
