/**
 * useApi — Custom hook for API calls with loading/error state
 * Provides a consistent pattern for data fetching across the application.
 */

import { useState, useEffect, useCallback, useRef } from 'react';
import { ApiClientError } from '@/api/client';

interface UseApiState<T> {
  /** The fetched data (undefined until first successful load). */
  data: T | undefined;
  /** Whether a request is currently in flight. */
  loading: boolean;
  /** Error from the latest request, if any. */
  error: ApiClientError | Error | null;
}

interface UseApiReturn<T> extends UseApiState<T> {
  /** Manually re-fetch the data. */
  refetch: () => void;
}

/**
 * Hook that executes an async API call and manages loading / error / data state.
 *
 * @param fetcher - Async function that returns the data.
 * @param deps - Dependency array; the fetcher re-runs when deps change.
 */
export function useApi<T>(
  fetcher: () => Promise<T>,
  deps: unknown[] = [],
): UseApiReturn<T> {
  const [state, setState] = useState<UseApiState<T>>({
    data: undefined,
    loading: true,
    error: null,
  });

  // Track whether the component is still mounted to avoid state updates after unmount.
  const mountedRef = useRef(true);
  useEffect(() => {
    return () => {
      mountedRef.current = false;
    };
  }, []);

  const execute = useCallback(async () => {
    setState((prev) => ({ ...prev, loading: true, error: null }));
    try {
      const data = await fetcher();
      if (mountedRef.current) {
        setState({ data, loading: false, error: null });
      }
    } catch (err) {
      if (mountedRef.current) {
        setState((prev) => ({
          ...prev,
          loading: false,
          error: err instanceof Error ? err : new Error(String(err)),
        }));
      }
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, deps);

  useEffect(() => {
    execute();
  }, [execute]);

  return {
    ...state,
    refetch: execute,
  };
}
