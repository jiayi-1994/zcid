import { http } from './http';
import type { ApiResponse } from './types';

export interface Project {
  id: string;
  name: string;
  description: string;
  ownerId: string;
  status: string;
  createdAt: string;
  updatedAt: string;
}

export interface ProjectList {
  items: Project[];
  total: number;
  page: number;
  pageSize: number;
}

export interface EnvironmentItem {
  id: string;
  projectId: string;
  name: string;
  namespace: string;
  description: string;
  status: string;
  health?: EnvironmentHealth;
  createdAt: string;
  updatedAt: string;
}

export interface EnvironmentHealth {
  status: 'healthy' | 'warning' | 'degraded' | 'unknown' | 'stale';
  reason: string;
  lastSignalAt?: string;
  stale: boolean;
}

export interface EnvironmentList {
  items: EnvironmentItem[];
  total: number;
  page: number;
  pageSize: number;
}

export interface ServiceItem {
  id: string;
  projectId: string;
  name: string;
  description: string;
  repoUrl: string;
  serviceType?: string;
  language?: string;
  owner?: string;
  tags?: string[];
  pipelineIds?: string[];
  environmentIds?: string[];
  status: string;
  createdAt: string;
  updatedAt: string;
}

export interface ServiceVitals {
  service: ServiceItem;
  summary: {
    status: 'healthy' | 'warning' | 'degraded' | 'unknown' | 'stale';
    reason: string;
    lastSignalAt?: string;
    hasDeliveryData: boolean;
    hasDeploymentData: boolean;
    activeWarningCount: number;
  };
  linkedPipelines: Array<{ id: string; name: string; status: string; repoUrl: string; createdAt: string; updatedAt: string }>;
  recentRuns: Array<{ id: string; pipelineId: string; pipelineName: string; runNumber: number; status: string; startedAt?: string; finishedAt?: string; errorMessage?: string; createdAt: string }>;
  latestDeployments: Array<{ id: string; environmentId: string; environmentName: string; pipelineRunId?: string; image: string; status: string; syncStatus?: string; healthStatus?: string; errorMessage?: string; startedAt?: string; finishedAt?: string; createdAt: string }>;
  activeSignals: Array<{ id: string; targetType: string; targetId: string; source: string; status: string; rawStatus: string; severity: string; reason: string; message: string; observedAt: string; staleAfter?: string }>;
  warnings: Array<{ stepName: string; taskRunName: string; status: string; durationMs?: number; exitCode?: number; runId: string; pipelineId: string; pipelineName: string; runNumber: number; runPath: string; createdAt: string }>;
  emptyStates: string[];
  refreshedAt: string;
}

export interface ServiceList {
  items: ServiceItem[];
  total: number;
  page: number;
  pageSize: number;
}

export interface MemberItem {
  userId: string;
  username: string;
  role: string;
  joinedAt: string;
}

export interface MemberList {
  items: MemberItem[];
  total: number;
}

export async function fetchProjects(page = 1, pageSize = 20): Promise<ProjectList> {
  const res = await http.get<ApiResponse<ProjectList>>('/projects', { params: { page, pageSize } });
  return res.data.data;
}

export async function fetchProject(id: string): Promise<Project> {
  const res = await http.get<ApiResponse<Project>>(`/projects/${id}`);
  return res.data.data;
}

export async function createProject(name: string, description: string): Promise<Project> {
  const res = await http.post<ApiResponse<Project>>('/projects', { name, description });
  return res.data.data;
}

export async function updateProject(id: string, data: { name?: string; description?: string }): Promise<Project> {
  const res = await http.put<ApiResponse<Project>>(`/projects/${id}`, data);
  return res.data.data;
}

export async function deleteProject(id: string): Promise<void> {
  await http.delete(`/projects/${id}`);
}

export async function fetchEnvironments(projectId: string, page = 1, pageSize = 20): Promise<EnvironmentList> {
  const res = await http.get<ApiResponse<EnvironmentList>>(`/projects/${projectId}/environments`, { params: { page, pageSize } });
  return res.data.data;
}

export async function createEnvironment(projectId: string, name: string, namespace: string, description: string): Promise<EnvironmentItem> {
  const res = await http.post<ApiResponse<EnvironmentItem>>(`/projects/${projectId}/environments`, { name, namespace, description });
  return res.data.data;
}

export async function deleteEnvironment(projectId: string, envId: string): Promise<void> {
  await http.delete(`/projects/${projectId}/environments/${envId}`);
}

export async function fetchServices(projectId: string, page = 1, pageSize = 20): Promise<ServiceList> {
  const res = await http.get<ApiResponse<ServiceList>>(`/projects/${projectId}/services`, { params: { page, pageSize } });
  return res.data.data;
}

export async function fetchServiceVitals(projectId: string, serviceId: string): Promise<ServiceVitals> {
  const res = await http.get<ApiResponse<ServiceVitals>>(`/projects/${projectId}/services/${serviceId}/vitals`);
  return res.data.data;
}

export async function createService(projectId: string, name: string, description: string, repoUrl: string): Promise<ServiceItem> {
  const res = await http.post<ApiResponse<ServiceItem>>(`/projects/${projectId}/services`, { name, description, repoUrl });
  return res.data.data;
}

export async function deleteService(projectId: string, svcId: string): Promise<void> {
  await http.delete(`/projects/${projectId}/services/${svcId}`);
}

export async function fetchMembers(projectId: string): Promise<MemberList> {
  const res = await http.get<ApiResponse<MemberList>>(`/projects/${projectId}/members`);
  return res.data.data;
}

export async function addMember(projectId: string, userId: string, role: string): Promise<void> {
  await http.post(`/projects/${projectId}/members`, { userId, role });
}

export async function updateMemberRole(projectId: string, userId: string, role: string): Promise<void> {
  await http.put(`/projects/${projectId}/members/${userId}`, { role });
}

export async function removeMember(projectId: string, userId: string): Promise<void> {
  await http.delete(`/projects/${projectId}/members/${userId}`);
}
