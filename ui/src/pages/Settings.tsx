/**
 * Settings — Application settings page
 * Manages contributor access, SSH keys, and connection configuration.
 */

import { useState, useEffect } from 'react';
import { useConnection } from '@/hooks/useConnection';
import StatusBadge from '@/components/common/StatusBadge';
import { authApi, SSHKey } from '@/api/auth';
import { useAuth } from '@/contexts/AuthContext';
import { Trash2, Key, AlertCircle } from 'lucide-react';

/** Tabs within the settings page. */
type SettingsTab = 'general' | 'ssh-keys' | 'connection';

function SSHKeysSettings() {
  const [keys, setKeys] = useState<SSHKey[]>([]);
  const [isAdding, setIsAdding] = useState(false);
  const [newKeyName, setNewKeyName] = useState('');
  const [newKeyContent, setNewKeyContent] = useState('');
  const [error, setError] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    loadKeys();
  }, []);

  const loadKeys = async () => {
    try {
      const data = await authApi.listSSHKeys();
      setKeys(data);
    } catch (err) {
      console.error(err);
    } finally {
      setIsLoading(false);
    }
  };

  const handleAdd = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    try {
      await authApi.addSSHKey(newKeyName, newKeyContent);
      setNewKeyName('');
      setNewKeyContent('');
      setIsAdding(false);
      await loadKeys();
    } catch (err: any) {
      setError(err.message || 'Failed to add SSH key');
    }
  };

  const handleDelete = async (id: number) => {
    if (!window.confirm('Are you sure you want to delete this key?')) return;
    try {
      await authApi.deleteSSHKey(id);
      await loadKeys();
    } catch (err) {
      console.error(err);
    }
  };

  return (
    <div className="card space-y-6">
      <div className="flex items-center justify-between">
        <h2 className="text-lg font-semibold">SSH Keys</h2>
        {!isAdding && (
          <button onClick={() => setIsAdding(true)} className="btn-primary text-sm">
            + Add SSH Key
          </button>
        )}
      </div>

      <p className="text-sm text-surface-500">
        SSH keys are used to authenticate Git push/pull operations.
        Add your public key below to grant access.
      </p>

      {error && (
        <div className="rounded-md bg-red-50 p-4 border border-red-100 flex items-center">
          <AlertCircle className="h-5 w-5 text-red-400 mr-2" />
          <span className="text-sm text-red-800">{error}</span>
        </div>
      )}

      {isAdding && (
        <form onSubmit={handleAdd} className="bg-surface-50 p-4 rounded-lg border border-surface-200 space-y-4">
          <div>
            <label className="block text-sm font-medium mb-1">Key Name</label>
            <input 
              required
              type="text" 
              className="input w-full max-w-md" 
              placeholder="e.g. My MacBook"
              value={newKeyName}
              onChange={(e) => setNewKeyName(e.target.value)}
            />
          </div>
          <div>
            <label className="block text-sm font-medium mb-1">Key Content (Public Key)</label>
            <textarea 
              required
              className="input w-full font-mono text-sm" 
              rows={4}
              placeholder="ssh-ed25519 AAAA..."
              value={newKeyContent}
              onChange={(e) => setNewKeyContent(e.target.value)}
            />
          </div>
          <div className="flex gap-2">
            <button type="submit" className="btn-primary text-sm">Save Key</button>
            <button type="button" onClick={() => setIsAdding(false)} className="btn-secondary text-sm">Cancel</button>
          </div>
        </form>
      )}

      {isLoading ? (
        <div className="py-8 text-center text-surface-500 text-sm">Loading keys...</div>
      ) : keys.length === 0 ? (
        <div className="rounded-lg border border-dashed border-surface-300 p-8 text-center dark:border-surface-600">
          <p className="text-sm text-surface-400">No SSH keys configured.</p>
          <p className="mt-1 text-xs text-surface-400">
            Add your first SSH public key to enable authenticated Git operations.
          </p>
        </div>
      ) : (
        <div className="space-y-3">
          {keys.map((key) => (
            <div key={key.ID} className="flex items-start justify-between p-4 rounded-lg border border-surface-200 bg-white dark:bg-surface-800 dark:border-surface-700">
              <div className="flex gap-3">
                <div className="mt-1">
                  <Key className="w-5 h-5 text-surface-400" />
                </div>
                <div>
                  <h3 className="font-medium text-surface-900 dark:text-surface-50">{key.Name}</h3>
                  <p className="font-mono text-xs text-surface-500 mt-1">{key.Fingerprint}</p>
                  <p className="text-xs text-surface-400 mt-1">Added on {new Date(key.CreatedAt).toLocaleDateString()}</p>
                </div>
              </div>
              <button 
                onClick={() => handleDelete(key.ID)}
                className="text-red-500 hover:text-red-600 p-2 rounded hover:bg-red-50 dark:hover:bg-red-900/20 transition-colors"
                title="Delete key"
              >
                <Trash2 className="w-4 h-4" />
              </button>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}

export default function Settings() {
  const [activeTab, setActiveTab] = useState<SettingsTab>('general');
  const { status } = useConnection();
  const { user, logout } = useAuth();

  const tabs: Array<{ key: SettingsTab; label: string }> = [
    { key: 'general', label: 'General' },
    { key: 'ssh-keys', label: 'SSH Keys' },
    { key: 'connection', label: 'Connection' },
  ];

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <div>
          <h1 className="text-2xl font-bold text-surface-900 dark:text-surface-50">
            Settings
          </h1>
          <p className="mt-1 text-sm text-surface-500">
            Configure your LocalRepo instance.
          </p>
        </div>
        <div>
          <button onClick={logout} className="btn-secondary text-sm text-red-600 hover:bg-red-50 border-red-200">
            Sign out
          </button>
        </div>
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
                Username
              </label>
              <input
                id="display-name"
                type="text"
                className="input max-w-md bg-surface-50"
                value={user?.username || ''}
                disabled
              />
              <p className="mt-1 text-xs text-surface-400">
                Your authenticated username.
              </p>
            </div>

            <div>
              <label htmlFor="email" className="mb-1 block text-sm font-medium">
                Email Address
              </label>
              <input
                id="email"
                type="email"
                className="input max-w-md bg-surface-50"
                value={user?.email || ''}
                disabled
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
        </div>
      )}

      {/* SSH Keys */}
      {activeTab === 'ssh-keys' && <SSHKeysSettings />}

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
