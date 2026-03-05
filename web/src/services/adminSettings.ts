import { http } from './http';

export interface SystemSettings {
  k8sApiUrl: string;
  defaultRegistry: string;
  argocdUrl: string;
}

export interface HealthStatus {
  status: string;
  checks: Record<string, string>;
}

export interface IntegrationStatus {
  name: string;
  status: string;
  detail: string;
}

export async function fetchSettings() {
  const resp = await http.get<{ code: number; data: SystemSettings }>('/admin/settings');
  return resp.data.data;
}

export async function updateSettings(data: SystemSettings) {
  const resp = await http.put<{ code: number; data: SystemSettings }>('/admin/settings', data);
  return resp.data.data;
}

export async function fetchHealth() {
  const resp = await http.get<{ status: string; checks: Record<string, string> }>('/admin/health');
  return resp.data;
}

export async function fetchIntegrationsStatus() {
  const resp = await http.get<{ code: number; data: { integrations: IntegrationStatus[] } }>(
    '/admin/integrations/status',
  );
  return resp.data.data.integrations;
}
