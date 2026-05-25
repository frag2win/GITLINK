/**
 * StatusBadge — Online / Offline / Syncing indicator
 * Small pill badge that shows the current connection state
 * with a colored dot and label.
 */

interface StatusBadgeProps {
  status: 'online' | 'offline' | 'syncing';
  /** Optional label override. */
  label?: string;
  /** Size variant. */
  size?: 'sm' | 'md';
}

const STATUS_CONFIG = {
  online: {
    label: 'Online',
    dotColor: 'bg-status-online',
    bgColor: 'bg-green-50 dark:bg-green-950',
    textColor: 'text-green-700 dark:text-green-300',
  },
  offline: {
    label: 'Offline',
    dotColor: 'bg-status-offline',
    bgColor: 'bg-red-50 dark:bg-red-950',
    textColor: 'text-red-700 dark:text-red-300',
  },
  syncing: {
    label: 'Syncing',
    dotColor: 'bg-status-syncing animate-pulse',
    bgColor: 'bg-amber-50 dark:bg-amber-950',
    textColor: 'text-amber-700 dark:text-amber-300',
  },
} as const;

export default function StatusBadge({ status, label, size = 'sm' }: StatusBadgeProps) {
  const config = STATUS_CONFIG[status];
  const sizeClasses = size === 'sm' ? 'px-2 py-0.5 text-xs' : 'px-3 py-1 text-sm';

  return (
    <span
      className={`inline-flex items-center gap-1.5 rounded-full font-medium ${config.bgColor} ${config.textColor} ${sizeClasses}`}
    >
      <span className={`h-2 w-2 rounded-full ${config.dotColor}`} />
      {label ?? config.label}
    </span>
  );
}
