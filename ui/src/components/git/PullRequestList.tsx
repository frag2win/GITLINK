import { useState } from 'react';
import { useApi } from '@/hooks/useApi';
import { listPullRequests } from '@/api/pulls';
import type { PullRequest, Branch } from '@/types';
import LoadingSpinner from '@/components/common/LoadingSpinner';
import PullRequestCreate from './PullRequestCreate';
import PullRequestDetail from './PullRequestDetail';

interface PullRequestListProps {
  repoName: string;
  branches: Branch[];
}

export default function PullRequestList({ repoName, branches }: PullRequestListProps) {
  const [view, setView] = useState<'list' | 'create' | 'detail'>('list');
  const [selectedPR, setSelectedPR] = useState<PullRequest | null>(null);

  const { data: prs, loading, refetch } = useApi<PullRequest[]>(
    () => listPullRequests(repoName),
    [repoName, view]
  );

  if (view === 'create') {
    return (
      <PullRequestCreate
        repoName={repoName}
        branches={branches}
        onCancel={() => setView('list')}
        onCreated={() => {
          setView('list');
          refetch();
        }}
      />
    );
  }

  if (view === 'detail' && selectedPR) {
    return (
      <PullRequestDetail
        repoName={repoName}
        pr={selectedPR}
        onBack={() => {
          setSelectedPR(null);
          setView('list');
          refetch();
        }}
      />
    );
  }

  if (loading) {
    return <LoadingSpinner label="Loading Pull Requests..." />;
  }

  return (
    <div className="card space-y-4">
      <div className="flex justify-between items-center">
        <h3 className="text-lg font-semibold">Pull Requests</h3>
        <button className="btn-primary" onClick={() => setView('create')}>
          New Pull Request
        </button>
      </div>

      {!prs || prs.length === 0 ? (
        <p className="text-sm text-surface-500">No pull requests found.</p>
      ) : (
        <ul className="divide-y divide-surface-100 dark:divide-surface-800">
          {prs.map((pr) => (
            <li
              key={pr.ID}
              className="py-3 flex justify-between items-center cursor-pointer hover:bg-surface-50 dark:hover:bg-surface-800 px-2 rounded"
              onClick={() => {
                setSelectedPR(pr);
                setView('detail');
              }}
            >
              <div>
                <span className="font-semibold">{pr.title}</span>
                <span className="ml-2 text-xs text-surface-400">#{pr.ID}</span>
                <div className="text-sm text-surface-500 mt-1">
                  {pr.headBranch} into {pr.baseBranch}
                </div>
              </div>
              <span
                className={`text-xs px-2 py-1 rounded-full font-medium ${
                  pr.status === 'open'
                    ? 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-300'
                    : 'bg-purple-100 text-purple-800 dark:bg-purple-900 dark:text-purple-300'
                }`}
              >
                {pr.status.toUpperCase()}
              </span>
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}
