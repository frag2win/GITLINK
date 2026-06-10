import React, { useState } from 'react';
import { createPullRequest } from '@/api/pulls';
import type { Branch } from '@/types';

interface Props {
  repoName: string;
  branches: Branch[];
  onCancel: () => void;
  onCreated: () => void;
}

export default function PullRequestCreate({ repoName, branches, onCancel, onCreated }: Props) {
  const [title, setTitle] = useState('');
  const [description, setDescription] = useState('');
  const [baseBranch, setBaseBranch] = useState(branches.find(b => b.isDefault)?.name || '');
  const [headBranch, setHeadBranch] = useState('');
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!title || !baseBranch || !headBranch) return;

    setLoading(true);
    try {
      await createPullRequest(repoName, { title, description, baseBranch, headBranch });
      onCreated();
    } catch (err: any) {
      alert('Failed to create PR: ' + err.message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <form onSubmit={handleSubmit} className="card space-y-4">
      <h3 className="text-lg font-semibold">New Pull Request</h3>
      
      <div className="flex gap-4">
        <div className="flex-1">
          <label className="block text-sm font-medium mb-1">Base Branch</label>
          <select 
            className="input" 
            value={baseBranch} 
            onChange={(e) => setBaseBranch(e.target.value)}
            required
          >
            <option value="">Select base...</option>
            {branches.map(b => (
              <option key={b.name} value={b.name}>{b.name}</option>
            ))}
          </select>
        </div>
        <div className="flex-1">
          <label className="block text-sm font-medium mb-1">Head Branch (compare)</label>
          <select 
            className="input" 
            value={headBranch} 
            onChange={(e) => setHeadBranch(e.target.value)}
            required
          >
            <option value="">Select head...</option>
            {branches.map(b => (
              <option key={b.name} value={b.name}>{b.name}</option>
            ))}
          </select>
        </div>
      </div>

      <div>
        <label className="block text-sm font-medium mb-1">Title</label>
        <input 
          className="input" 
          type="text" 
          value={title} 
          onChange={(e) => setTitle(e.target.value)} 
          placeholder="e.g., Add new feature"
          required 
        />
      </div>

      <div>
        <label className="block text-sm font-medium mb-1">Description</label>
        <textarea 
          className="input" 
          rows={4} 
          value={description} 
          onChange={(e) => setDescription(e.target.value)} 
          placeholder="Describe your changes..."
        />
      </div>

      <div className="flex gap-2">
        <button type="submit" className="btn-primary" disabled={loading}>
          {loading ? 'Creating...' : 'Create Pull Request'}
        </button>
        <button type="button" className="btn-secondary" onClick={onCancel} disabled={loading}>
          Cancel
        </button>
      </div>
    </form>
  );
}
