/**
 * CommitDetail — Single commit detail view
 * Shows full commit metadata, parent refs, and a diff summary.
 */

import type { Commit } from '@/types';
import { formatDate, truncateHash } from '@/utils/format';

interface CommitDetailProps {
  /** The commit to display. */
  commit: Commit;
  /** Callback to navigate back to the commit list. */
  onBack: () => void;
}

export default function CommitDetail({ commit, onBack }: CommitDetailProps) {
  return (
    <div className="space-y-4">
      {/* Back button */}
      <button
        onClick={onBack}
        className="text-sm text-brand-600 hover:underline"
      >
        ← Back to commits
      </button>

      {/* Commit header card */}
      <div className="card">
        <h2 className="text-xl font-semibold text-surface-900 dark:text-surface-50">
          {commit.message}
        </h2>

        {commit.body && (
          <pre className="mt-3 whitespace-pre-wrap rounded bg-surface-50 p-3 font-mono text-sm text-surface-600 dark:bg-surface-800 dark:text-surface-400">
            {commit.body}
          </pre>
        )}

        {/* Metadata grid */}
        <div className="mt-4 grid gap-3 border-t border-surface-200 pt-4 text-sm dark:border-surface-700 sm:grid-cols-2">
          <div>
            <span className="text-xs font-medium uppercase tracking-wider text-surface-400">
              Author
            </span>
            <p className="mt-0.5 text-surface-700 dark:text-surface-300">
              {commit.authorName}
              <span className="ml-1 text-surface-400">
                &lt;{commit.authorEmail}&gt;
              </span>
            </p>
          </div>
          <div>
            <span className="text-xs font-medium uppercase tracking-wider text-surface-400">
              Date
            </span>
            <p className="mt-0.5 text-surface-700 dark:text-surface-300">
              {formatDate(commit.authorDate)}
            </p>
          </div>
          <div>
            <span className="text-xs font-medium uppercase tracking-wider text-surface-400">
              Commit Hash
            </span>
            <p className="mt-0.5">
              <code className="font-mono text-xs text-surface-600 dark:text-surface-400">
                {commit.hash}
              </code>
            </p>
          </div>
          <div>
            <span className="text-xs font-medium uppercase tracking-wider text-surface-400">
              Parent{commit.parents.length > 1 ? 's' : ''}
            </span>
            <div className="mt-0.5 flex flex-wrap gap-1">
              {commit.parents.length > 0 ? (
                commit.parents.map((parent) => (
                  <code
                    key={parent}
                    className="rounded bg-surface-100 px-1.5 py-0.5 font-mono text-xs text-surface-600 dark:bg-surface-700 dark:text-surface-400"
                  >
                    {truncateHash(parent)}
                  </code>
                ))
              ) : (
                <span className="text-surface-400">Initial commit</span>
              )}
            </div>
          </div>
        </div>
      </div>

      {/* Diff stats summary */}
      {(commit.filesChanged != null || commit.additions != null || commit.deletions != null) && (
        <div className="card">
          <h3 className="mb-2 text-sm font-semibold text-surface-700 dark:text-surface-300">
            Changes
          </h3>
          <div className="flex items-center gap-4 text-sm">
            {commit.filesChanged != null && (
              <span className="text-surface-500">
                {commit.filesChanged} file{commit.filesChanged !== 1 ? 's' : ''} changed
              </span>
            )}
            {commit.additions != null && (
              <span className="font-medium text-green-600">
                +{commit.additions} additions
              </span>
            )}
            {commit.deletions != null && (
              <span className="font-medium text-red-600">
                -{commit.deletions} deletions
              </span>
            )}
          </div>
          {/* Placeholder for future inline diff view */}
          <div className="mt-4 rounded border border-dashed border-surface-300 p-8 text-center text-sm text-surface-400 dark:border-surface-600">
            Inline diff view coming soon.
          </div>
        </div>
      )}
    </div>
  );
}
