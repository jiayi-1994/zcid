import { http } from './http';
import type { ApiResponse } from './types';

export interface VariableItem {
  id: string;
  scope: string;
  projectId?: string;
  key: string;
  value: string;
  varType: string;
  description: string;
  createdBy: string;
  createdAt: string;
  updatedAt: string;
}

export interface VariableList {
  items: VariableItem[];
  total: number;
}

export async function fetchProjectVariables(projectId: string): Promise<VariableList> {
  const res = await http.get<ApiResponse<VariableList>>(`/projects/${projectId}/variables`);
  return res.data.data;
}

export async function fetchMergedVariables(projectId: string): Promise<VariableList> {
  const res = await http.get<ApiResponse<VariableList>>(`/projects/${projectId}/variables/merged`);
  return res.data.data;
}

export async function createProjectVariable(
  projectId: string,
  data: { key: string; value: string; varType?: string; description?: string }
): Promise<VariableItem> {
  const res = await http.post<ApiResponse<VariableItem>>(`/projects/${projectId}/variables`, data);
  return res.data.data;
}

export async function updateProjectVariable(
  projectId: string,
  varId: string,
  data: { value?: string; description?: string }
): Promise<void> {
  await http.put(`/projects/${projectId}/variables/${varId}`, data);
}

export async function deleteProjectVariable(projectId: string, varId: string): Promise<void> {
  await http.delete(`/projects/${projectId}/variables/${varId}`);
}

export async function fetchGlobalVariables(): Promise<VariableList> {
  const res = await http.get<ApiResponse<VariableList>>('/admin/variables');
  return res.data.data;
}

export async function createGlobalVariable(
  data: { key: string; value: string; varType?: string; description?: string }
): Promise<VariableItem> {
  const res = await http.post<ApiResponse<VariableItem>>('/admin/variables', data);
  return res.data.data;
}

export async function updateGlobalVariable(
  varId: string,
  data: { value?: string; description?: string }
): Promise<void> {
  await http.put(`/admin/variables/${varId}`, data);
}

export async function deleteGlobalVariable(varId: string): Promise<void> {
  await http.delete(`/admin/variables/${varId}`);
}
