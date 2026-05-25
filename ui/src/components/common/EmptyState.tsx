/**
 * EmptyState — Placeholder for empty content areas
 * Shown when a list has no items, search returns no results, etc.
 */

import { type ReactNode } from 'react';

interface EmptyStateProps {
  /** Headline text (e.g., "No repositories yet"). */
  title: string;
  /** Descriptive subtext. */
  description?: string;
  /** Optional icon element rendered above the title. */
  icon?: ReactNode;
  /** Optional action element (e.g., a button) rendered below the description. */
  action?: ReactNode;
}

export default function EmptyState({ title, description, icon, action }: EmptyStateProps) {
  return (
    <div className="flex flex-col items-center justify-center rounded-lg border-2 border-dashed border-surface-200 px-6 py-16 text-center dark:border-surface-700">
      {icon && (
        <div className="mb-4 text-surface-300 dark:text-surface-600">
          {icon}
        </div>
      )}
      <h3 className="text-lg font-semibold text-surface-900 dark:text-surface-100">
        {title}
      </h3>
      {description && (
        <p className="mt-1 max-w-sm text-sm text-surface-500">
          {description}
        </p>
      )}
      {action && <div className="mt-4">{action}</div>}
    </div>
  );
}
