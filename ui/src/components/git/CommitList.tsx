/**
 * CommitList — Commit history list
 * Displays a chronological list of commits with author, message, and hash.
 */

import type { Commit } from '@/types';
import { truncateHash, timeAgo } from '@/utils/format';

interface CommitListProps {
  /** Array of commits to display. */
  commits: Commit[];
  /** Called when a commit is clicked for detail view. */
  onSelect?: (commit: Commit) => void;
  /** Whether more commits are being loaded. */
  loading?: boolean;
}

export default function CommitList({ commits, onSelect, loading }: CommitListProps) {
  if (commits.length === 0 && !loading) {
    return (
      <div className="card py-12 text-center text-sm text-surface-500">
        No commits found.
      </div>
    );
  }

  return (
    <div className="card overflow-hidden p-0">
      <ul className="divide-y divide-surface-100 dark:divide-surface-800">
        {commits.map((commit) => (
          <li
            key={commit.hash}
            className="flex cursor-pointer items-start gap-4 px-4 py-3 transition-colors hover:bg-surface-50 dark:hover:bg-surface-800/50"
            onClick={() => onSelect?.(commit)}
          >
            {/* Commit dot */}
            <div className="mt-1.5 flex flex-col items-center">
              <div className="h-3 w-3 rounded-full border-2 border-brand-500 bg-white dark:bg-surface-800" />
            </div>

            {/* Commit info */}
            <div className="min-w-0 flex-1">
              <p className="truncate font-medium text-surface-900 dark:text-surface-100">
                {commit.message}
              </p>
              <div className="mt-1 flex flex-wrap items-center gap-x-3 gap-y-1 text-xs text-surface-500">
                <span className="font-medium text-surface-700 dark:text-surface-300">
                  {commit.authorName}
                </span>
                <span>committed {timeAgo(commit.authorDate)}</span>
                {commit.filesChanged != null && (
                  <span>
                    {commit.filesChanged} file{commit.filesChanged !== 1 ? 's' : ''} changed
                    {commit.additions != null && (
                      <span className="ml-1 text-green-600">+{commit.additions}</span>
                    )}
                    {commit.deletions != null && (
                      <span className="ml-1 text-red-600">-{commit.deletions}</span>
                    )}
                  </span>
                )}
              </div>
            </div>

            {/* Hash badge */}
            <code className="shrink-0 rounded bg-surface-100 px-2 py-1 font-mono text-xs text-surface-600 dark:bg-surface-700 dark:text-surface-400">
              {commit.shortHash ?? truncateHash(commit.hash)}
            </code>
          </li>
        ))}
      </ul>

      {loading && (
        <div className="border-t border-surface-100 px-4 py-3 text-center text-sm text-surface-400 dark:border-surface-800">
          Loading more commits…
        </div>
      )}
    </div>
  );
}
