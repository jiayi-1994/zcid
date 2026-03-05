import { http } from './http';
import type { ApiResponse } from './types';

export interface Deployment {
  id: string;
  projectId: string;
  environmentId: string;
  pipelineRunId?: string;
  image: string;
  status: string;
  argoAppName?: string;
  syncStatus?: string;
  healthStatus?: string;
  errorMessage?: string;
  deployedBy: string;
  startedAt?: string;
  finishedAt?: string;
  createdAt: string;
  updatedAt: string;
}

export interface DeploymentSummary {
  id: string;
  environmentId: string;
  image: string;
  status: string;
  syncStatus?: string;
  healthStatus?: string;
  deployedBy: string;
  createdAt: string;
}

export interface DeploymentList {
  items: DeploymentSummary[];
  total: number;
  page: number;
  pageSize: number;
}

export async function fetchDeployments(
  projectId: string,
  page = 1,
  pageSize = 20
): Promise<DeploymentList> {
  const res = await http.get<ApiResponse<DeploymentList>>(
    `/projects/${projectId}/deployments`,
    { params: { page, pageSize } }
  );
  return res.data.data;
}

export async function fetchDeployment(
  projectId: string,
  deployId: string
): Promise<Deployment> {
  const res = await http.get<ApiResponse<Deployment>>(
    `/projects/${projectId}/deployments/${deployId}`
  );
  return res.data.data;
}

export async function triggerDeploy(
  projectId: string,
  body: { environmentId: string; image: string; pipelineRunId?: string }
): Promise<Deployment> {
  const res = await http.post<ApiResponse<Deployment>>(
    `/projects/${projectId}/deployments`,
    body
  );
  return res.data.data;
}

export async function fetchDeployStatus(
  projectId: string,
  deployId: string
): Promise<Deployment> {
  const res = await http.get<ApiResponse<Deployment>>(
    `/projects/${projectId}/deployments/${deployId}/status`
  );
  return res.data.data;
}

export async function resyncDeploy(
  projectId: string,
  deployId: string
): Promise<Deployment> {
  const res = await http.post<ApiResponse<Deployment>>(
    `/projects/${projectId}/deployments/${deployId}/resync`
  );
  return res.data.data;
}

export async function rollbackDeploy(
  projectId: string,
  deployId: string
): Promise<Deployment> {
  const res = await http.post<ApiResponse<Deployment>>(
    `/projects/${projectId}/deployments/${deployId}/rollback`
  );
  return res.data.data;
}

export async function fetchDeployHistory(
  projectId: string,
  envId: string,
  page = 1,
  pageSize = 20
): Promise<DeploymentList> {
  const res = await http.get<ApiResponse<DeploymentList>>(
    `/projects/${projectId}/deployments/environments/${envId}/deploy-history`,
    { params: { page, pageSize } }
  );
  return res.data.data;
}
