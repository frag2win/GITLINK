/**
 * LocalRepo — Utility / formatting functions
 * Pure helper functions used across the UI for display formatting.
 */

/**
 * Format an ISO-8601 date string into a human-readable local format.
 * @example formatDate('2024-06-15T10:30:00Z') → "Jun 15, 2024"
 */
export function formatDate(isoDate: string): string {
  const date = new Date(isoDate);
  return date.toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
  });
}

/**
 * Format byte counts into human-readable sizes.
 * @example formatBytes(1536) → "1.5 KB"
 */
export function formatBytes(bytes: number): string {
  if (bytes === 0) return '0 B';

  const units = ['B', 'KB', 'MB', 'GB', 'TB'];
  const k = 1024;
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  const value = bytes / Math.pow(k, i);

  return `${value.toFixed(i === 0 ? 0 : 1)} ${units[i]}`;
}

/**
 * Truncate a full Git hash to a short display form.
 * @example truncateHash('abc123def456...') → "abc123d"
 */
export function truncateHash(hash: string, length = 7): string {
  return hash.substring(0, length);
}

/**
 * Compute a human-readable "time ago" string from an ISO-8601 date.
 * @example timeAgo('2024-06-15T10:00:00Z') → "3 hours ago"
 */
export function timeAgo(isoDate: string): string {
  const date = new Date(isoDate);
  const now = new Date();
  const seconds = Math.floor((now.getTime() - date.getTime()) / 1000);

  if (seconds < 0) return 'just now';

  const intervals: Array<{ label: string; seconds: number }> = [
    { label: 'year', seconds: 31536000 },
    { label: 'month', seconds: 2592000 },
    { label: 'week', seconds: 604800 },
    { label: 'day', seconds: 86400 },
    { label: 'hour', seconds: 3600 },
    { label: 'minute', seconds: 60 },
  ];

  for (const interval of intervals) {
    const count = Math.floor(seconds / interval.seconds);
    if (count >= 1) {
      return `${count} ${interval.label}${count > 1 ? 's' : ''} ago`;
    }
  }

  return 'just now';
}
