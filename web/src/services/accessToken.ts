import { http } from './http';

export interface AccessTokenRecord {
  id: string;
  type: 'personal' | 'project';
  name: string;
  tokenPrefix: string;
  scopes: string[];
  userId?: string;
  projectId?: string;
  createdBy: string;
  expiresAt: string;
  lastUsedAt?: string;
  revokedAt?: string;
  createdAt: string;
}

export interface CreateAccessTokenPayload {
  name: string;
  type: 'personal' | 'project';
  scopes: string[];
  expiresAt: string;
  projectId?: string;
}

export interface CreateAccessTokenResult {
  token: AccessTokenRecord;
  rawToken: string;
}

export async function fetchAccessTokens() {
  const resp = await http.get<{ code: number; data: { items: AccessTokenRecord[]; total: number } }>('/admin/access-tokens');
  return resp.data.data;
}

export async function createAccessToken(payload: CreateAccessTokenPayload) {
  const resp = await http.post<{ code: number; data: CreateAccessTokenResult }>('/admin/access-tokens', payload);
  return resp.data.data;
}

export async function revokeAccessToken(tokenId: string) {
  await http.post(`/admin/access-tokens/${tokenId}/revoke`);
}
