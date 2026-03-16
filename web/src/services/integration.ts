import { http } from './http';
import type { ApiResponse } from './types';

export interface GitConnection {
  id: string;
  name: string;
  providerType: string;
  serverUrl: string;
  tokenType: string;
  tokenMask: string;
  status: string;
  description: string;
  createdBy: string;
  createdAt: string;
  updatedAt: string;
}

export interface GitConnectionList {
  items: GitConnection[];
  total: number;
}

export interface TestConnectionResult {
  success: boolean;
  message: string;
}

export async function fetchConnections(): Promise<GitConnectionList> {
  const res = await http.get<ApiResponse<GitConnectionList>>('/admin/integrations');
  return res.data.data;
}

export async function createConnection(data: {
  name: string;
  providerType: string;
  serverUrl: string;
  accessToken: string;
  description?: string;
}): Promise<GitConnection> {
  const res = await http.post<ApiResponse<GitConnection>>('/admin/integrations', data);
  return res.data.data;
}

export async function updateConnection(
  connId: string,
  data: { name?: string; accessToken?: string; description?: string }
): Promise<void> {
  await http.put(`/admin/integrations/${connId}`, data);
}

export async function deleteConnection(connId: string): Promise<void> {
  await http.delete(`/admin/integrations/${connId}`);
}

export async function testConnection(connId: string): Promise<TestConnectionResult> {
  const res = await http.post<ApiResponse<TestConnectionResult>>(
    `/admin/integrations/${connId}/test`
  );
  return res.data.data;
}

export async function getWebhookSecret(connId: string): Promise<string> {
  const res = await http.get<ApiResponse<{ webhookSecret: string }>>(
    `/admin/integrations/${connId}/webhook-secret`
  );
  return res.data.data.webhookSecret;
}

export interface GitBranch {
  name: string;
  isDefault: boolean;
}

export async function listBranches(
  connId: string,
  repoFullName: string,
  refresh = false,
): Promise<GitBranch[]> {
  const res = await http.get<ApiResponse<{ items: GitBranch[] }>>(
    `/admin/integrations/${connId}/repos/${encodeURIComponent(repoFullName)}/branches`,
    { params: refresh ? { refresh: 'true' } : undefined },
  );
  return res.data.data.items ?? [];
}
