import React, { useState } from 'react';
import { mergePullRequest } from '@/api/pulls';
import type { PullRequest } from '@/types';

interface Props {
  repoName: string;
  pr: PullRequest;
  onBack: () => void;
}

export default function PullRequestDetail({ repoName, pr, onBack }: Props) {
  const [loading, setLoading] = useState(false);

  const handleMerge = async () => {
    setLoading(true);
    try {
      await mergePullRequest(repoName, pr.ID);
      alert('Pull Request merged successfully!');
      onBack();
    } catch (err: any) {
      alert('Failed to merge PR: ' + err.message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="card space-y-6">
      <button className="btn-secondary text-sm mb-4" onClick={onBack}>
        &larr; Back to Pull Requests
      </button>

      <div>
        <h2 className="text-xl font-bold">{pr.title} <span className="text-surface-400 font-normal">#{pr.ID}</span></h2>
        <div className="flex gap-2 items-center mt-2">
          <span
            className={`text-xs px-2 py-1 rounded-full font-medium ${
              pr.status === 'open'
                ? 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-300'
                : 'bg-purple-100 text-purple-800 dark:bg-purple-900 dark:text-purple-300'
            }`}
          >
            {pr.status.toUpperCase()}
          </span>
          <span className="text-sm text-surface-500">
            {pr.headBranch} into {pr.baseBranch}
          </span>
        </div>
      </div>

      <div className="prose dark:prose-invert max-w-none text-sm p-4 bg-surface-50 dark:bg-surface-800 rounded-md">
        {pr.description || <em className="text-surface-400">No description provided.</em>}
      </div>

      {pr.status === 'open' && (
        <div className="pt-4 border-t border-surface-200 dark:border-surface-700">
          <button 
            className="btn-primary bg-green-600 hover:bg-green-700 text-white" 
            onClick={handleMerge}
            disabled={loading}
          >
            {loading ? 'Merging...' : 'Merge Pull Request'}
          </button>
        </div>
      )}
    </div>
  );
}
