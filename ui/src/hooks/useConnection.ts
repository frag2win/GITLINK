/**
 * useConnection — Hook for monitoring connection status
 * Polls the API health endpoint to track whether the local node
 * is online, offline, or actively syncing with peers.
 */

import { useState, useEffect, useCallback, useRef } from 'react';
import type { ConnectionStatus } from '@/types';

/** Default connection status when no health data is available yet. */
const DEFAULT_STATUS: ConnectionStatus = {
  isOnline: false,
  syncState: 'idle',
  connectedPeers: 0,
};

/** Interval between health checks in milliseconds. */
const POLL_INTERVAL_MS = 5000;

interface UseConnectionReturn {
  /** Current connection status. */
  status: ConnectionStatus;
  /** Force an immediate health check. */
  refresh: () => void;
}

/**
 * Polls `/api/health` to determine the connection status of the local node.
 * Updates automatically every 5 seconds.
 */
export function useConnection(): UseConnectionReturn {
  const [status, setStatus] = useState<ConnectionStatus>(DEFAULT_STATUS);
  const intervalRef = useRef<ReturnType<typeof setInterval> | null>(null);

  const checkHealth = useCallback(async () => {
    try {
      const res = await fetch('/api/health', {
        method: 'GET',
        headers: { Accept: 'application/json' },
      });

      if (res.ok) {
        const data: ConnectionStatus = await res.json();
        setStatus({ ...data, isOnline: true });
      } else {
        setStatus((prev) => ({ ...prev, isOnline: false, syncState: 'error' }));
      }
    } catch {
      setStatus((prev) => ({
        ...prev,
        isOnline: false,
        syncState: 'error',
        connectedPeers: 0,
      }));
    }
  }, []);

  useEffect(() => {
    // Run immediately on mount
    checkHealth();

    // Set up polling interval
    intervalRef.current = setInterval(checkHealth, POLL_INTERVAL_MS);

    return () => {
      if (intervalRef.current) {
        clearInterval(intervalRef.current);
      }
    };
  }, [checkHealth]);

  return { status, refresh: checkHealth };
}
