/**
 * LocalRepo — File API
 * Browse repository file trees and read file content.
 */

import { get } from './client';
import type { FileEntry } from '@/types';

/** Parameters for browsing files in a repository tree. */
export interface BrowseFilesParams {
  /** Branch or ref to browse (defaults to default branch). */
  ref?: string;
  /** Directory path within the repository (defaults to root "/"). */
  path?: string;
}

/** Response for file content requests. */
export interface FileContentResponse {
  /** File name. */
  name: string;
  /** Full path relative to repo root. */
  path: string;
  /** MIME type of the file (e.g., "text/plain", "application/octet-stream"). */
  mimeType: string;
  /** File size in bytes. */
  sizeBytes: number;
  /** File content (base64 encoded for binary files, plain text for text files). */
  content: string;
  /** Whether the content is base64 encoded. */
  isBinary: boolean;
  /** The ref (branch/tag/hash) this content was read from. */
  ref: string;
}

/** Browse files and directories at a given path in a repository. */
export function browseFiles(
  repoName: string,
  params?: BrowseFilesParams,
): Promise<FileEntry[]> {
  const query: Record<string, string> = {};
  if (params?.ref) query.ref = params.ref;
  if (params?.path) query.path = params.path;
  return get<FileEntry[]>(
    `/repos/${encodeURIComponent(repoName)}/files`,
    query,
  );
}

/** Get the content of a specific file from the repository. */
export function getFileContent(
  repoName: string,
  filePath: string,
  ref?: string,
): Promise<FileContentResponse> {
  const query: Record<string, string> = {};
  if (ref) query.ref = ref;
  return get<FileContentResponse>(
    `/repos/${encodeURIComponent(repoName)}/files/${filePath}`,
    query,
  );
}
