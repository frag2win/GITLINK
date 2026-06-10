/**
 * LocalRepo — Core TypeScript interfaces
 * Shared type definitions used across the frontend application.
 */

/** Represents a Git repository managed by LocalRepo. */
export interface Repo {
  /** Unique repository name (URL-safe slug). */
  name: string;
  /** Human-readable description of the repository. */
  description: string;
  /** Default branch name (e.g., "main"). */
  defaultBranch: string;
  /** Whether the repository is private or publicly accessible on the network. */
  isPrivate: boolean;
  /** Disk size in bytes. */
  sizeBytes: number;
  /** ISO-8601 timestamp of creation. */
  createdAt: string;
  /** ISO-8601 timestamp of last push/update. */
  updatedAt: string;
  /** Clone URL for the repository. */
  cloneUrl: string;
  /** Number of branches. */
  branchCount: number;
  /** Number of contributors. */
  contributorCount: number;
  /** Latest commit on the default branch. */
  lastCommit?: Commit;
}

/** Represents a Git branch within a repository. */
export interface Branch {
  /** Branch name (e.g., "main", "feature/auth"). */
  name: string;
  /** SHA-1 hash of the branch HEAD commit. */
  headCommitHash: string;
  /** Whether this is the default branch. */
  isDefault: boolean;
  /** ISO-8601 timestamp of the latest commit on this branch. */
  lastActivity: string;
}

/** Represents a single Git commit. */
export interface Commit {
  /** Full SHA-1 hash of the commit. */
  hash: string;
  /** Short (7-char) hash for display. */
  shortHash: string;
  /** Commit message (first line / subject). */
  message: string;
  /** Full commit message body (may be empty). */
  body: string;
  /** Author display name. */
  authorName: string;
  /** Author email. */
  authorEmail: string;
  /** ISO-8601 timestamp of when the commit was authored. */
  authorDate: string;
  /** Parent commit hash(es). Merge commits have multiple parents. */
  parents: string[];
  /** Files changed in this commit. */
  filesChanged?: number;
  /** Lines added. */
  additions?: number;
  /** Lines deleted. */
  deletions?: number;
}

/** Represents a contributor with access to a repository. */
export interface Contributor {
  /** Unique contributor ID. */
  id: string;
  /** Display name. */
  name: string;
  /** Email address. */
  email: string;
  /** Role within the repository. */
  role: 'owner' | 'admin' | 'write' | 'read';
  /** SSH public key fingerprint for authentication. */
  sshKeyFingerprint?: string;
  /** ISO-8601 timestamp of when the contributor was added. */
  addedAt: string;
}

/** Audit log entry for tracking repository events. */
export interface AuditLogEntry {
  /** Unique entry ID. */
  id: string;
  /** Type of event. */
  action: 'push' | 'clone' | 'branch_create' | 'branch_delete' | 'contributor_add' | 'contributor_remove' | 'repo_create' | 'repo_delete' | 'settings_change';
  /** Human-readable description of the event. */
  description: string;
  /** Who performed the action. */
  actorName: string;
  /** ISO-8601 timestamp. */
  timestamp: string;
  /** Repository name (if applicable). */
  repoName?: string;
  /** Additional metadata for the event. */
  metadata?: Record<string, string>;
}

/** Represents a file or directory entry in the repository tree. */
export interface FileEntry {
  /** File or directory name. */
  name: string;
  /** Full path relative to the repository root. */
  path: string;
  /** Whether this entry is a directory or a file. */
  type: 'file' | 'directory';
  /** File size in bytes (files only). */
  sizeBytes?: number;
  /** Last commit that touched this file. */
  lastCommit?: Pick<Commit, 'hash' | 'shortHash' | 'message' | 'authorDate'>;
}

/** Tracks the real-time connection status of the local node. */
export interface ConnectionStatus {
  /** Whether the local API server is reachable. */
  isOnline: boolean;
  /** Current sync state. */
  syncState: 'idle' | 'syncing' | 'error';
  /** Number of connected peers. */
  connectedPeers: number;
  /** ISO-8601 timestamp of last successful sync. */
  lastSyncAt?: string;
  /** Human-readable error message (when syncState is 'error'). */
  errorMessage?: string;
  /** Local node's peer ID. */
  peerId?: string;
}

/** Paginated API response wrapper. */
export interface PaginatedResponse<T> {
  items: T[];
  total: number;
  page: number;
  pageSize: number;
  hasMore: boolean;
}

/** Standard API error response. */
export interface ApiError {
  code: string;
  message: string;
  details?: Record<string, string>;
}

/** Represents a Pull Request in a repository. */
export interface PullRequest {
  ID: number;
  CreatedAt: string;
  UpdatedAt: string;
  DeletedAt: string | null;
  repository_id: number;
  title: string;
  description: string;
  baseBranch: string;
  headBranch: string;
  status: 'open' | 'merged' | 'closed';
}
