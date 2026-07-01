import * as api from './client';

export interface User {
  id: number;
  username: string;
  email: string;
  peer_id?: string;
}

export interface AuthResponse {
  token: string;
  user: User;
}

export interface SSHKey {
  ID: number;
  Name: string;
  PublicKey: string;
  Fingerprint: string;
  CreatedAt: string;
}

export const authApi = {
  login: async (username: string, password: string): Promise<AuthResponse> => {
    return await api.post<AuthResponse>('/auth/login', { username, password });
  },

  register: async (username: string, email: string, password: string): Promise<{message: string}> => {
    return await api.post<{message: string}>('/auth/register', { username, email, password });
  },

  getMe: async (): Promise<User> => {
    return await api.get<User>('/auth/me');
  },

  listSSHKeys: async (): Promise<SSHKey[]> => {
    return await api.get<SSHKey[]>('/user/keys');
  },

  addSSHKey: async (name: string, publicKey: string): Promise<SSHKey> => {
    return await api.post<SSHKey>('/user/keys', { name, publicKey });
  },

  deleteSSHKey: async (id: number): Promise<void> => {
    await api.del(`/user/keys/${id}`);
  }
};
