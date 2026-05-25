/**
 * FileViewer — File content viewer
 * Displays file content with line numbers and basic syntax highlighting support.
 * Shows a toolbar with file name, size, and raw download link.
 */

import { formatBytes } from '@/utils/format';

interface FileViewerProps {
  /** File name for display. */
  fileName: string;
  /** Full path within the repository. */
  filePath: string;
  /** File content as a string. */
  content: string;
  /** Whether the file is binary (show download prompt instead). */
  isBinary: boolean;
  /** File size in bytes. */
  sizeBytes: number;
  /** Callback to navigate back to the file browser. */
  onBack: () => void;
}

/**
 * Attempt to infer a language from the file extension for future
 * syntax highlighting integration.
 */
function inferLanguage(fileName: string): string {
  const ext = fileName.split('.').pop()?.toLowerCase();
  const languageMap: Record<string, string> = {
    ts: 'typescript',
    tsx: 'tsx',
    js: 'javascript',
    jsx: 'jsx',
    py: 'python',
    go: 'go',
    rs: 'rust',
    rb: 'ruby',
    java: 'java',
    css: 'css',
    html: 'html',
    json: 'json',
    yaml: 'yaml',
    yml: 'yaml',
    md: 'markdown',
    sh: 'bash',
    sql: 'sql',
    toml: 'toml',
    dockerfile: 'dockerfile',
  };
  return ext ? languageMap[ext] ?? 'plaintext' : 'plaintext';
}

export default function FileViewer({
  fileName,
  filePath,
  content,
  isBinary,
  sizeBytes,
  onBack,
}: FileViewerProps) {
  const language = inferLanguage(fileName);
  const lines = content.split('\n');

  return (
    <div className="card overflow-hidden p-0">
      {/* Toolbar */}
      <div className="flex items-center justify-between border-b border-surface-200 px-4 py-2 dark:border-surface-700">
        <div className="flex items-center gap-2">
          <button
            onClick={onBack}
            className="text-sm text-brand-600 hover:underline"
          >
            ← Back
          </button>
          <span className="text-surface-300">/</span>
          <span className="font-mono text-sm font-medium">{filePath}</span>
        </div>
        <div className="flex items-center gap-3 text-xs text-surface-400">
          <span>{formatBytes(sizeBytes)}</span>
          <span className="rounded bg-surface-100 px-1.5 py-0.5 font-mono dark:bg-surface-700">
            {language}
          </span>
        </div>
      </div>

      {/* Content */}
      {isBinary ? (
        <div className="flex flex-col items-center justify-center py-16 text-center">
          <p className="text-surface-500">
            Binary file ({formatBytes(sizeBytes)}) — cannot be displayed.
          </p>
          <button className="btn-primary mt-4">Download File</button>
        </div>
      ) : (
        <div className="overflow-x-auto">
          <table className="w-full font-mono text-sm">
            <tbody>
              {lines.map((line, i) => (
                <tr key={i} className="hover:bg-surface-50 dark:hover:bg-surface-800/50">
                  <td className="select-none border-r border-surface-200 px-3 py-0.5 text-right text-xs text-surface-400 dark:border-surface-700">
                    {i + 1}
                  </td>
                  <td className="whitespace-pre px-4 py-0.5">
                    {line || '\u00A0'}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}
