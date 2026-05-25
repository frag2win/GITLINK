/**
 * Header — Top navigation bar
 * Displays the LocalRepo logo, main navigation links,
 * and a real-time connection status indicator.
 */

import { Link, useLocation } from 'react-router-dom';
import StatusBadge from '@/components/common/StatusBadge';
import { useConnection } from '@/hooks/useConnection';

/** Navigation link definition. */
interface NavLink {
  label: string;
  to: string;
}

const NAV_LINKS: NavLink[] = [
  { label: 'Dashboard', to: '/' },
  { label: 'Repositories', to: '/repos' },
  { label: 'Settings', to: '/settings' },
];

export default function Header() {
  const location = useLocation();
  const { status } = useConnection();

  return (
    <header className="sticky top-0 z-30 border-b border-surface-200 bg-white/80 backdrop-blur dark:border-surface-700 dark:bg-surface-900/80">
      <div className="flex h-14 items-center justify-between px-4 lg:px-6">
        {/* Logo */}
        <Link to="/" className="flex items-center gap-2 font-semibold text-brand-600">
          <svg
            className="h-7 w-7"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth={2}
            strokeLinecap="round"
            strokeLinejoin="round"
          >
            <circle cx="12" cy="12" r="3" />
            <path d="M12 3v6m0 6v6" />
            <path d="M5.63 5.63l4.24 4.24m4.24 4.24l4.24 4.24" />
            <path d="M3 12h6m6 0h6" />
          </svg>
          <span className="text-lg">LocalRepo</span>
        </Link>

        {/* Navigation */}
        <nav className="hidden items-center gap-1 md:flex">
          {NAV_LINKS.map((link) => {
            const isActive =
              link.to === '/'
                ? location.pathname === '/'
                : location.pathname.startsWith(link.to);
            return (
              <Link
                key={link.to}
                to={link.to}
                className={`rounded-md px-3 py-1.5 text-sm font-medium transition-colors ${
                  isActive
                    ? 'bg-brand-50 text-brand-700 dark:bg-brand-950 dark:text-brand-300'
                    : 'text-surface-600 hover:bg-surface-100 hover:text-surface-900 dark:text-surface-400 dark:hover:bg-surface-800 dark:hover:text-surface-100'
                }`}
              >
                {link.label}
              </Link>
            );
          })}
        </nav>

        {/* Connection status */}
        <div className="flex items-center gap-3">
          <StatusBadge
            status={status.isOnline ? (status.syncState === 'syncing' ? 'syncing' : 'online') : 'offline'}
          />
          <span className="hidden text-xs text-surface-500 sm:inline">
            {status.connectedPeers} peer{status.connectedPeers !== 1 ? 's' : ''}
          </span>
        </div>
      </div>
    </header>
  );
}
