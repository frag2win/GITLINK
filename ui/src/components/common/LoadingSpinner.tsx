/**
 * LoadingSpinner — Animated loading indicator
 * Displays a spinning circle to indicate pending async operations.
 */

interface LoadingSpinnerProps {
  /** Size of the spinner. */
  size?: 'sm' | 'md' | 'lg';
  /** Optional label shown beneath the spinner. */
  label?: string;
}

const SIZE_CLASSES = {
  sm: 'h-4 w-4 border-2',
  md: 'h-8 w-8 border-2',
  lg: 'h-12 w-12 border-3',
} as const;

export default function LoadingSpinner({ size = 'md', label }: LoadingSpinnerProps) {
  return (
    <div className="flex flex-col items-center justify-center gap-2">
      <div
        className={`animate-spin rounded-full border-brand-200 border-t-brand-600 ${SIZE_CLASSES[size]}`}
        role="status"
        aria-label="Loading"
      />
      {label && (
        <span className="text-sm text-surface-500">{label}</span>
      )}
    </div>
  );
}
