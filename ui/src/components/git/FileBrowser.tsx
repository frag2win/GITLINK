/**
 * FileBrowser — File / folder tree browser
 * Displays the repository file tree at a given path and ref.
 * Clicking a directory navigates deeper; clicking a file opens the viewer.
 */

import type { FileEntry } from '@/types';
import { formatBytes, timeAgo } from '@/utils/format';

interface FileBrowserProps {
  /** File entries at the current directory level. */
  entries: FileEntry[];
  /** Current path within the repo (empty string = root). */
  currentPath: string;
  /** Called when a directory is clicked. */
  onNavigate: (path: string) => void;
  /** Called when a file is clicked. */
  onFileSelect: (entry: FileEntry) => void;
}

export default function FileBrowser({
  entries,
  currentPath,
  onNavigate,
  onFileSelect,
}: FileBrowserProps) {
  // Sort: directories first, then files, alphabetically within each group
  const sorted = [...entries].sort((a, b) => {
    if (a.type !== b.type) return a.type === 'directory' ? -1 : 1;
    return a.name.localeCompare(b.name);
  });

  return (
    <div className="card overflow-hidden p-0">
      {/* Breadcrumb path */}
      {currentPath && (
        <div className="border-b border-surface-200 px-4 py-2 dark:border-surface-700">
          <nav className="flex items-center gap-1 text-sm">
            <button
              onClick={() => onNavigate('')}
              className="text-brand-600 hover:underline"
            >
              root
            </button>
            {currentPath.split('/').map((segment, i, arr) => {
              const path = arr.slice(0, i + 1).join('/');
              return (
                <span key={path} className="flex items-center gap-1">
                  <span className="text-surface-400">/</span>
                  {i === arr.length - 1 ? (
                    <span className="font-medium">{segment}</span>
                  ) : (
                    <button
                      onClick={() => onNavigate(path)}
                      className="text-brand-600 hover:underline"
                    >
                      {segment}
                    </button>
                  )}
                </span>
              );
            })}
          </nav>
        </div>
      )}

      {/* File list */}
      <table className="w-full text-sm">
        <thead>
          <tr className="border-b border-surface-200 bg-surface-50 text-left text-xs font-medium uppercase tracking-wider text-surface-500 dark:border-surface-700 dark:bg-surface-800">
            <th className="px-4 py-2">Name</th>
            <th className="hidden px-4 py-2 sm:table-cell">Last Commit</th>
            <th className="px-4 py-2 text-right">Size</th>
          </tr>
        </thead>
        <tbody>
          {/* Parent directory link */}
          {currentPath && (
            <tr
              className="cursor-pointer border-b border-surface-100 hover:bg-surface-50 dark:border-surface-800 dark:hover:bg-surface-800/50"
              onClick={() => {
                const parent = currentPath.split('/').slice(0, -1).join('/');
                onNavigate(parent);
              }}
            >
              <td className="px-4 py-2 text-surface-500" colSpan={3}>
                📁 ..
              </td>
            </tr>
          )}

          {sorted.map((entry) => (
            <tr
              key={entry.path}
              className="cursor-pointer border-b border-surface-100 transition-colors hover:bg-surface-50 dark:border-surface-800 dark:hover:bg-surface-800/50"
              onClick={() =>
                entry.type === 'directory'
                  ? onNavigate(entry.path)
                  : onFileSelect(entry)
              }
            >
              <td className="px-4 py-2">
                <span className="flex items-center gap-2">
                  <span className="text-base">
                    {entry.type === 'directory' ? '📁' : '📄'}
                  </span>
                  <span className={entry.type === 'directory' ? 'font-medium text-brand-600' : ''}>
                    {entry.name}
                  </span>
                </span>
              </td>
              <td className="hidden px-4 py-2 text-xs text-surface-500 sm:table-cell">
                {entry.lastCommit && (
                  <span title={entry.lastCommit.message}>
                    {entry.lastCommit.message.substring(0, 50)}
                    {entry.lastCommit.message.length > 50 ? '…' : ''}
                    <span className="ml-2 text-surface-400">
                      {timeAgo(entry.lastCommit.authorDate)}
                    </span>
                  </span>
                )}
              </td>
              <td className="px-4 py-2 text-right text-xs text-surface-400">
                {entry.type === 'file' && entry.sizeBytes != null
                  ? formatBytes(entry.sizeBytes)
                  : '—'}
              </td>
            </tr>
          ))}

          {sorted.length === 0 && (
            <tr>
              <td className="px-4 py-8 text-center text-surface-400" colSpan={3}>
                This directory is empty.
              </td>
            </tr>
          )}
        </tbody>
      </table>
    </div>
  );
}
