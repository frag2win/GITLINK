/**
 * BranchSelector — Branch dropdown selector
 * Allows switching between branches with a search filter.
 */

import { useState, useRef, useEffect } from 'react';
import type { Branch } from '@/types';

interface BranchSelectorProps {
  /** Available branches. */
  branches: Branch[];
  /** Currently selected branch name. */
  currentBranch: string;
  /** Called when a branch is selected. */
  onSelect: (branchName: string) => void;
}

export default function BranchSelector({
  branches,
  currentBranch,
  onSelect,
}: BranchSelectorProps) {
  const [isOpen, setIsOpen] = useState(false);
  const [search, setSearch] = useState('');
  const dropdownRef = useRef<HTMLDivElement>(null);

  // Close dropdown on outside click
  useEffect(() => {
    function handleClickOutside(e: MouseEvent) {
      if (dropdownRef.current && !dropdownRef.current.contains(e.target as Node)) {
        setIsOpen(false);
      }
    }
    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  const filtered = branches.filter((b) =>
    b.name.toLowerCase().includes(search.toLowerCase()),
  );

  return (
    <div className="relative" ref={dropdownRef}>
      {/* Trigger button */}
      <button
        onClick={() => setIsOpen(!isOpen)}
        className="btn-secondary flex items-center gap-2"
      >
        {/* Branch icon */}
        <svg className="h-4 w-4" viewBox="0 0 16 16" fill="currentColor">
          <path d="M11.75 2.5a.75.75 0 1 0 0 1.5.75.75 0 0 0 0-1.5zm-2.25.75a2.25 2.25 0 1 1 3 2.122V6A2.5 2.5 0 0 1 10 8.5H6a1 1 0 0 0-1 1v1.128a2.251 2.251 0 1 1-1.5 0V5.372a2.25 2.25 0 1 1 1.5 0v1.836A2.493 2.493 0 0 1 6 7h4a1 1 0 0 0 1-1v-.628A2.25 2.25 0 0 1 9.5 3.25zM4.25 12a.75.75 0 1 0 0 1.5.75.75 0 0 0 0-1.5zM3.5 3.25a.75.75 0 1 1 1.5 0 .75.75 0 0 1-1.5 0z" />
        </svg>
        <span className="max-w-[120px] truncate font-mono text-sm">{currentBranch}</span>
        <svg className="h-3 w-3" viewBox="0 0 12 12" fill="currentColor">
          <path d="M6 8.825c-.2 0-.4-.1-.5-.2l-3.3-3.3c-.3-.3-.3-.8 0-1.1.3-.3.8-.3 1.1 0L6 6.925l2.7-2.7c.3-.3.8-.3 1.1 0 .3.3.3.8 0 1.1l-3.3 3.3c-.1.1-.3.2-.5.2z" />
        </svg>
      </button>

      {/* Dropdown */}
      {isOpen && (
        <div className="absolute left-0 z-20 mt-1 w-64 rounded-lg border border-surface-200 bg-white shadow-lg dark:border-surface-700 dark:bg-surface-800">
          {/* Search input */}
          <div className="border-b border-surface-200 p-2 dark:border-surface-700">
            <input
              type="text"
              className="input py-1.5 text-sm"
              placeholder="Filter branches…"
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              autoFocus
            />
          </div>

          {/* Branch list */}
          <ul className="max-h-60 overflow-y-auto py-1">
            {filtered.map((branch) => (
              <li key={branch.name}>
                <button
                  className={`flex w-full items-center gap-2 px-3 py-1.5 text-left text-sm transition-colors hover:bg-surface-50 dark:hover:bg-surface-700 ${
                    branch.name === currentBranch
                      ? 'bg-brand-50 font-medium text-brand-700 dark:bg-brand-950 dark:text-brand-300'
                      : 'text-surface-700 dark:text-surface-300'
                  }`}
                  onClick={() => {
                    onSelect(branch.name);
                    setIsOpen(false);
                    setSearch('');
                  }}
                >
                  {branch.name === currentBranch && (
                    <span className="text-brand-500">✓</span>
                  )}
                  <span className="truncate font-mono">{branch.name}</span>
                  {branch.isDefault && (
                    <span className="ml-auto shrink-0 rounded-full bg-surface-100 px-1.5 py-0.5 text-[10px] font-medium text-surface-500 dark:bg-surface-700">
                      default
                    </span>
                  )}
                </button>
              </li>
            ))}
            {filtered.length === 0 && (
              <li className="px-3 py-4 text-center text-sm text-surface-400">
                No matching branches.
              </li>
            )}
          </ul>
        </div>
      )}
    </div>
  );
}
