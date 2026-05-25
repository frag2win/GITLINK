/**
 * Settings — Application settings page
 * Manages contributor access, SSH keys, and connection configuration.
 */

import { useState } from 'react';
import { useConnection } from '@/hooks/useConnection';
import StatusBadge from '@/components/common/StatusBadge';

/** Tabs within the settings page. */
type SettingsTab = 'general' | 'ssh-keys' | 'connection';

export default function Settings() {
  const [activeTab, setActiveTab] = useState<SettingsTab>('general');
  const { status } = useConnection();

  const tabs: Array<{ key: SettingsTab; label: string }> = [
    { key: 'general', label: 'General' },
    { key: 'ssh-keys', label: 'SSH Keys' },
    { key: 'connection', label: 'Connection' },
  ];

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-surface-900 dark:text-surface-50">
          Settings
        </h1>
        <p className="mt-1 text-sm text-surface-500">
          Configure your LocalRepo instance.
        </p>
      </div>

      {/* Tabs */}
      <nav className="flex gap-1 border-b border-surface-200 pb-1 dark:border-surface-700">
        {tabs.map((tab) => (
          <button
            key={tab.key}
            onClick={() => setActiveTab(tab.key)}
            className={`rounded-t-md px-4 py-2 text-sm font-medium transition-colors ${
              activeTab === tab.key
                ? 'border-b-2 border-brand-600 text-brand-700 dark:text-brand-300'
                : 'text-surface-500 hover:text-surface-700 dark:hover:text-surface-300'
            }`}
          >
            {tab.label}
          </button>
        ))}
      </nav>

      {/* General settings */}
      {activeTab === 'general' && (
        <div className="card space-y-6">
          <h2 className="text-lg font-semibold">General Settings</h2>

          <div className="space-y-4">
            <div>
              <label htmlFor="display-name" className="mb-1 block text-sm font-medium">
                Display Name
              </label>
              <input
                id="display-name"
                type="text"
                className="input max-w-md"
                placeholder="Your name"
                defaultValue=""
              />
              <p className="mt-1 text-xs text-surface-400">
                Shown to peers when they connect to your node.
              </p>
            </div>

            <div>
              <label htmlFor="email" className="mb-1 block text-sm font-medium">
                Email Address
              </label>
              <input
                id="email"
                type="email"
                className="input max-w-md"
                placeholder="you@example.com"
                defaultValue=""
              />
            </div>

            <div>
              <label htmlFor="data-dir" className="mb-1 block text-sm font-medium">
                Data Directory
              </label>
              <input
                id="data-dir"
                type="text"
                className="input max-w-lg font-mono text-sm"
                defaultValue="~/.localrepo/data"
                disabled
              />
              <p className="mt-1 text-xs text-surface-400">
                Where repositories and configuration are stored on disk.
              </p>
            </div>
          </div>

          <div className="border-t border-surface-200 pt-4 dark:border-surface-700">
            <button className="btn-primary">Save Settings</button>
          </div>
        </div>
      )}

      {/* SSH Keys */}
      {activeTab === 'ssh-keys' && (
        <div className="card space-y-6">
          <div className="flex items-center justify-between">
            <h2 className="text-lg font-semibold">SSH Keys</h2>
            <button className="btn-primary text-sm">+ Add SSH Key</button>
          </div>

          <p className="text-sm text-surface-500">
            SSH keys are used to authenticate Git push/pull operations.
            Add your public key below to grant access.
          </p>

          {/* Placeholder key list */}
          <div className="rounded-lg border border-dashed border-surface-300 p-8 text-center dark:border-surface-600">
            <p className="text-sm text-surface-400">No SSH keys configured.</p>
            <p className="mt-1 text-xs text-surface-400">
              Add your first SSH public key to enable authenticated Git operations.
            </p>
          </div>
        </div>
      )}

      {/* Connection */}
      {activeTab === 'connection' && (
        <div className="card space-y-6">
          <h2 className="text-lg font-semibold">Connection Configuration</h2>

          {/* Current status */}
          <div className="rounded-lg bg-surface-50 p-4 dark:bg-surface-800">
            <h3 className="mb-2 text-sm font-medium">Current Status</h3>
            <div className="flex items-center gap-4">
              <StatusBadge
                status={status.isOnline ? (status.syncState === 'syncing' ? 'syncing' : 'online') : 'offline'}
                size="md"
              />
              <span className="text-sm text-surface-500">
                {status.connectedPeers} peer{status.connectedPeers !== 1 ? 's' : ''} connected
              </span>
              {status.peerId && (
                <code className="ml-auto rounded bg-surface-200 px-2 py-1 font-mono text-xs dark:bg-surface-700">
                  {status.peerId}
                </code>
              )}
            </div>
          </div>

          {/* Connection settings */}
          <div className="space-y-4">
            <div>
              <label htmlFor="listen-addr" className="mb-1 block text-sm font-medium">
                Listen Address
              </label>
              <input
                id="listen-addr"
                type="text"
                className="input max-w-md font-mono text-sm"
                defaultValue="0.0.0.0:9418"
              />
              <p className="mt-1 text-xs text-surface-400">
                Address and port to listen for incoming peer connections.
              </p>
            </div>

            <div>
              <label htmlFor="api-port" className="mb-1 block text-sm font-medium">
                API Server Port
              </label>
              <input
                id="api-port"
                type="number"
                className="input max-w-[200px]"
                defaultValue={3000}
              />
            </div>

            <div>
              <label htmlFor="bootstrap-peers" className="mb-1 block text-sm font-medium">
                Bootstrap Peers
              </label>
              <textarea
                id="bootstrap-peers"
                className="input max-w-lg font-mono text-sm"
                rows={4}
                placeholder="Enter peer addresses, one per line…"
                defaultValue=""
              />
              <p className="mt-1 text-xs text-surface-400">
                Initial peers to connect to when the node starts. One address per line.
              </p>
            </div>
          </div>

          <div className="border-t border-surface-200 pt-4 dark:border-surface-700">
            <button className="btn-primary">Save Connection Settings</button>
          </div>
        </div>
      )}
    </div>
  );
}
