/**
 * LocalRepo — API Client
 * Centralized HTTP client for communicating with the Go api-server.
 * All API modules (repos, branches, commits, etc.) use this client.
 */

import type { ApiError } from '@/types';

/** Base URL for API requests. In dev mode Vite proxies /api to localhost:3000. */
const BASE_URL = '/api';

/** Default headers sent with every request. */
const DEFAULT_HEADERS: HeadersInit = {
  'Content-Type': 'application/json',
  'Accept': 'application/json',
};

/**
 * Custom error class that wraps API error responses.
 */
export class ApiClientError extends Error {
  public readonly status: number;
  public readonly code: string;
  public readonly details?: Record<string, string>;

  constructor(status: number, apiError: ApiError) {
    super(apiError.message);
    this.name = 'ApiClientError';
    this.status = status;
    this.code = apiError.code;
    this.details = apiError.details;
  }
}

/**
 * Core fetch wrapper with automatic error handling,
 * JSON parsing, and base URL resolution.
 */
async function request<T>(
  endpoint: string,
  options: RequestInit = {},
): Promise<T> {
  const url = `${BASE_URL}${endpoint}`;
  
  const token = localStorage.getItem('token');
  const authHeaders: HeadersInit = token ? { 'Authorization': `Bearer ${token}` } : {};

  const response = await fetch(url, {
    ...options,
    headers: {
      ...DEFAULT_HEADERS,
      ...authHeaders,
      ...options.headers,
    },
  });

  // Handle 204 No Content (successful delete, etc.)
  if (response.status === 204) {
    return undefined as T;
  }

  const body = await response.json();

  if (!response.ok) {
    const apiError: ApiError = body ?? {
      code: 'UNKNOWN_ERROR',
      message: `Request failed with status ${response.status}`,
    };
    throw new ApiClientError(response.status, apiError);
  }

  return body as T;
}

/** Perform a GET request. */
export function get<T>(endpoint: string, params?: Record<string, string>): Promise<T> {
  const query = params ? `?${new URLSearchParams(params).toString()}` : '';
  return request<T>(`${endpoint}${query}`, { method: 'GET' });
}

/** Perform a POST request with a JSON body. */
export function post<T>(endpoint: string, body?: unknown): Promise<T> {
  return request<T>(endpoint, {
    method: 'POST',
    body: body ? JSON.stringify(body) : undefined,
  });
}

/** Perform a PUT request with a JSON body. */
export function put<T>(endpoint: string, body?: unknown): Promise<T> {
  return request<T>(endpoint, {
    method: 'PUT',
    body: body ? JSON.stringify(body) : undefined,
  });
}

/** Perform a DELETE request. */
export function del<T = void>(endpoint: string): Promise<T> {
  return request<T>(endpoint, { method: 'DELETE' });
}

/** Perform a PATCH request with a JSON body. */
export function patch<T>(endpoint: string, body?: unknown): Promise<T> {
  return request<T>(endpoint, {
    method: 'PATCH',
    body: body ? JSON.stringify(body) : undefined,
  });
}
