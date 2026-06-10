/**
 * RepoDetail — Repository detail page with tabs
 * Provides tabs for Files, Commits, Branches, and Settings.
 */

import { useState, useCallback } from 'react';
import { useParams } from 'react-router-dom';
import { useApi } from '@/hooks/useApi';
import { getRepo } from '@/api/repos';
import { listBranches } from '@/api/branches';
import { listCommits } from '@/api/commits';
import { browseFiles } from '@/api/files';
import type { Repo, Branch, Commit, FileEntry, PaginatedResponse } from '@/types';

import RepoHeader from '@/components/repo/RepoHeader';
import BranchSelector from '@/components/git/BranchSelector';
import FileBrowser from '@/components/git/FileBrowser';
import CommitList from '@/components/git/CommitList';
import CommitDetail from '@/components/git/CommitDetail';
import PullRequestList from '@/components/git/PullRequestList';
import LoadingSpinner from '@/components/common/LoadingSpinner';

type Tab = 'files' | 'commits' | 'branches' | 'pulls' | 'settings';

export default function RepoDetail() {
  const { name } = useParams<{ name: string }>();
  const [activeTab, setActiveTab] = useState<Tab>('files');
  const [currentBranch, setCurrentBranch] = useState<string>('');
  const [currentPath, setCurrentPath] = useState('');
  const [selectedCommit, setSelectedCommit] = useState<Commit | null>(null);

  // Fetch repo metadata
  const { data: repo, loading: repoLoading } = useApi<Repo>(
    () => getRepo(name!),
    [name],
  );

  // Set default branch once repo loads
  if (repo && !currentBranch) {
    setCurrentBranch(repo.defaultBranch);
  }

  // Fetch branches
  const { data: branches } = useApi<Branch[]>(
    () => (name ? listBranches(name) : Promise.resolve([])),
    [name],
  );

  // Fetch files for current branch/path
  const { data: files, loading: filesLoading } = useApi<FileEntry[]>(
    () =>
      name && currentBranch
        ? browseFiles(name, { ref: currentBranch, path: currentPath || undefined })
        : Promise.resolve([]),
    [name, currentBranch, currentPath],
  );

  // Fetch commits for current branch
  const { data: commitsData, loading: commitsLoading } = useApi<PaginatedResponse<Commit>>(
    () =>
      name && currentBranch
        ? listCommits(name, { ref: currentBranch, pageSize: 30 })
        : Promise.resolve({ items: [], total: 0, page: 1, pageSize: 30, hasMore: false }),
    [name, currentBranch],
  );

  const handleBranchChange = useCallback((branchName: string) => {
    setCurrentBranch(branchName);
    setCurrentPath('');
    setSelectedCommit(null);
  }, []);

  const tabs: Array<{ key: Tab; label: string }> = [
    { key: 'files', label: 'Files' },
    { key: 'commits', label: 'Commits' },
    { key: 'branches', label: `Branches (${branches?.length ?? 0})` },
    { key: 'pulls', label: 'Pull Requests' },
    { key: 'settings', label: 'Settings' },
  ];

  if (repoLoading) {
    return (
      <div className="py-16">
        <LoadingSpinner label="Loading repository…" size="lg" />
      </div>
    );
  }

  if (!repo) {
    return (
      <div className="py-16 text-center">
        <h2 className="text-xl font-semibold text-surface-900 dark:text-surface-50">
          Repository not found
        </h2>
        <p className="mt-2 text-surface-500">
          The repository "{name}" does not exist or could not be loaded.
        </p>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <RepoHeader repo={repo} />

      {/* Branch selector + tabs */}
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <BranchSelector
          branches={branches ?? []}
          currentBranch={currentBranch}
          onSelect={handleBranchChange}
        />

        <nav className="flex gap-1">
          {tabs.map((tab) => (
            <button
              key={tab.key}
              onClick={() => { setActiveTab(tab.key); setSelectedCommit(null); }}
              className={`rounded-md px-3 py-1.5 text-sm font-medium transition-colors ${
                activeTab === tab.key
                  ? 'bg-brand-50 text-brand-700 dark:bg-brand-950 dark:text-brand-300'
                  : 'text-surface-500 hover:bg-surface-100 hover:text-surface-700 dark:hover:bg-surface-800'
              }`}
            >
              {tab.label}
            </button>
          ))}
        </nav>
      </div>

      {/* Tab content */}
      {activeTab === 'files' && (
        filesLoading ? (
          <LoadingSpinner label="Loading files…" />
        ) : (
          <FileBrowser
            entries={files ?? []}
            currentPath={currentPath}
            onNavigate={setCurrentPath}
            onFileSelect={(entry) => {
              // TODO: Open FileViewer for selected file
              console.log('Selected file:', entry.path);
            }}
          />
        )
      )}

      {activeTab === 'commits' && (
        selectedCommit ? (
          <CommitDetail
            commit={selectedCommit}
            onBack={() => setSelectedCommit(null)}
          />
        ) : (
          <CommitList
            commits={commitsData?.items ?? []}
            onSelect={setSelectedCommit}
            loading={commitsLoading}
          />
        )
      )}

      {activeTab === 'branches' && (
        <div className="card">
          <h3 className="mb-4 text-lg font-semibold">Branches</h3>
          {branches && branches.length > 0 ? (
            <ul className="divide-y divide-surface-100 dark:divide-surface-800">
              {branches.map((branch) => (
                <li key={branch.name} className="flex items-center justify-between py-3">
                  <div className="flex items-center gap-2">
                    <svg className="h-4 w-4 text-surface-400" viewBox="0 0 16 16" fill="currentColor">
                      <path d="M11.75 2.5a.75.75 0 1 0 0 1.5.75.75 0 0 0 0-1.5zm-2.25.75a2.25 2.25 0 1 1 3 2.122V6A2.5 2.5 0 0 1 10 8.5H6a1 1 0 0 0-1 1v1.128a2.251 2.251 0 1 1-1.5 0V5.372a2.25 2.25 0 1 1 1.5 0v1.836A2.493 2.493 0 0 1 6 7h4a1 1 0 0 0 1-1v-.628A2.25 2.25 0 0 1 9.5 3.25zM4.25 12a.75.75 0 1 0 0 1.5.75.75 0 0 0 0-1.5zM3.5 3.25a.75.75 0 1 1 1.5 0 .75.75 0 0 1-1.5 0z" />
                    </svg>
                    <span className="font-mono text-sm">{branch.name}</span>
                    {branch.isDefault && (
                      <span className="rounded-full bg-brand-50 px-2 py-0.5 text-[10px] font-medium text-brand-600 dark:bg-brand-950 dark:text-brand-300">
                        default
                      </span>
                    )}
                  </div>
                  <span className="text-xs text-surface-400">
                    Updated {new Date(branch.lastActivity).toLocaleDateString()}
                  </span>
                </li>
              ))}
            </ul>
          ) : (
            <p className="text-sm text-surface-400">No branches found.</p>
          )}
        </div>
      )}

      {activeTab === 'pulls' && name && (
        <PullRequestList repoName={name} branches={branches ?? []} />
      )}

      {activeTab === 'settings' && (
        <div className="card space-y-6">
          <h3 className="text-lg font-semibold">Repository Settings</h3>

          <div className="space-y-4">
            <div>
              <label className="mb-1 block text-sm font-medium">Repository Name</label>
              <input
                type="text"
                className="input max-w-md"
                value={repo.name}
                disabled
              />
            </div>
            <div>
              <label className="mb-1 block text-sm font-medium">Description</label>
              <textarea
                className="input max-w-md"
                rows={3}
                defaultValue={repo.description}
              />
            </div>
            <div>
              <label className="mb-1 block text-sm font-medium">Default Branch</label>
              <input
                type="text"
                className="input max-w-md"
                defaultValue={repo.defaultBranch}
              />
            </div>
          </div>

          <div className="flex gap-2 border-t border-surface-200 pt-4 dark:border-surface-700">
            <button className="btn-primary">Save Changes</button>
            <button className="btn-secondary text-red-600 hover:text-red-700">
              Delete Repository
            </button>
          </div>

          <div className="border-t border-surface-200 pt-6 dark:border-surface-700">
            <h4 className="mb-4 text-md font-semibold">Branch Protection</h4>
            <div className="space-y-4">
              {branches?.map(b => (
                <div key={b.name} className="flex items-center justify-between">
                  <span className="font-mono text-sm">{b.name}</span>
                  <button
                    onClick={async () => {
                      const reqPR = confirm(`Toggle Require PR for ${b.name}?`);
                      if (reqPR !== null) {
                        await import('@/api/branches').then(m => m.protectBranch(repo.name, b.name, reqPR));
                        alert('Branch protection updated.');
                      }
                    }}
                    className="btn-secondary text-xs"
                  >
                    Toggle Protection
                  </button>
                </div>
              ))}
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
