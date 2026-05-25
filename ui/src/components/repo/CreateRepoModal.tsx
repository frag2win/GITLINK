/**
 * CreateRepoModal — Modal form for creating a new repository
 * Collects name, description, visibility, and default branch settings.
 */

import { useState, type FormEvent } from 'react';
import Modal from '@/components/common/Modal';
import { createRepo, type CreateRepoParams } from '@/api/repos';

interface CreateRepoModalProps {
  isOpen: boolean;
  onClose: () => void;
  /** Called after a repository is successfully created. */
  onCreated?: () => void;
}

export default function CreateRepoModal({ isOpen, onClose, onCreated }: CreateRepoModalProps) {
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [defaultBranch, setDefaultBranch] = useState('main');
  const [isPrivate, setIsPrivate] = useState(false);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const resetForm = () => {
    setName('');
    setDescription('');
    setDefaultBranch('main');
    setIsPrivate(false);
    setError(null);
  };

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError(null);

    const params: CreateRepoParams = {
      name: name.trim(),
      description: description.trim() || undefined,
      defaultBranch,
      isPrivate,
    };

    try {
      await createRepo(params);
      resetForm();
      onCreated?.();
      onClose();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create repository');
    } finally {
      setLoading(false);
    }
  };

  return (
    <Modal
      isOpen={isOpen}
      onClose={onClose}
      title="Create New Repository"
      footer={
        <>
          <button
            type="button"
            className="btn-secondary"
            onClick={() => { resetForm(); onClose(); }}
            disabled={loading}
          >
            Cancel
          </button>
          <button
            type="submit"
            form="create-repo-form"
            className="btn-primary"
            disabled={loading || !name.trim()}
          >
            {loading ? 'Creating…' : 'Create Repository'}
          </button>
        </>
      }
    >
      <form id="create-repo-form" onSubmit={handleSubmit} className="space-y-4">
        {error && (
          <div className="rounded-md bg-red-50 p-3 text-sm text-red-700 dark:bg-red-950 dark:text-red-300">
            {error}
          </div>
        )}

        {/* Repository name */}
        <div>
          <label htmlFor="repo-name" className="mb-1 block text-sm font-medium">
            Repository Name <span className="text-red-500">*</span>
          </label>
          <input
            id="repo-name"
            type="text"
            className="input"
            placeholder="my-awesome-project"
            value={name}
            onChange={(e) => setName(e.target.value)}
            pattern="[a-zA-Z0-9._-]+"
            required
            autoFocus
          />
          <p className="mt-1 text-xs text-surface-400">
            Use letters, numbers, hyphens, dots, and underscores.
          </p>
        </div>

        {/* Description */}
        <div>
          <label htmlFor="repo-desc" className="mb-1 block text-sm font-medium">
            Description
          </label>
          <textarea
            id="repo-desc"
            className="input min-h-[80px] resize-y"
            placeholder="A brief description of your repository"
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            rows={3}
          />
        </div>

        {/* Default branch */}
        <div>
          <label htmlFor="default-branch" className="mb-1 block text-sm font-medium">
            Default Branch
          </label>
          <input
            id="default-branch"
            type="text"
            className="input"
            value={defaultBranch}
            onChange={(e) => setDefaultBranch(e.target.value)}
          />
        </div>

        {/* Visibility */}
        <div className="flex items-center gap-2">
          <input
            id="is-private"
            type="checkbox"
            className="h-4 w-4 rounded border-surface-300 text-brand-600 focus:ring-brand-500"
            checked={isPrivate}
            onChange={(e) => setIsPrivate(e.target.checked)}
          />
          <label htmlFor="is-private" className="text-sm font-medium">
            Private repository
          </label>
        </div>
      </form>
    </Modal>
  );
}
