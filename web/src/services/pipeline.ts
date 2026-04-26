import { http } from './http';
import type { ApiResponse } from './types';

export interface PipelineConfig {
  schemaVersion: string;
  stages: StageConfig[];
  params?: ParamConfig[];
  metadata?: Record<string, string>;
}

export interface StageConfig {
  id: string;
  name: string;
  steps: StepConfig[];
}

export interface StepConfig {
  id: string;
  name: string;
  type: string;
  image?: string;
  command?: string[];
  args?: string[];
  env?: Record<string, string>;
  config?: Record<string, unknown>;
}

export interface ParamConfig {
  name: string;
  type: string;
  options?: string[];
  defaultValue?: string;
  description?: string;
  required: boolean;
}

export interface Pipeline {
  id: string;
  projectId: string;
  name: string;
  description: string;
  status: string;
  config: PipelineConfig;
  triggerType: string;
  concurrencyPolicy: string;
  createdBy: string;
  createdAt: string;
  updatedAt: string;
}

export interface PipelineSummary {
  id: string;
  projectId: string;
  name: string;
  description: string;
  status: string;
  triggerType: string;
  concurrencyPolicy: string;
  createdBy: string;
  createdAt: string;
  updatedAt: string;
}

export interface PipelineList {
  items: PipelineSummary[];
  total: number;
  page: number;
  pageSize: number;
}

export interface PipelineTemplate {
  id: string;
  name: string;
  description: string;
  category: string;
  params: ParamConfig[];
  config?: PipelineConfig;
}

export async function fetchPipelines(projectId: string, page = 1, pageSize = 20): Promise<PipelineList> {
  const res = await http.get<ApiResponse<PipelineList>>(`/projects/${projectId}/pipelines`, { params: { page, pageSize } });
  return res.data.data;
}

export async function fetchPipeline(projectId: string, pipelineId: string): Promise<Pipeline> {
  const res = await http.get<ApiResponse<Pipeline>>(`/projects/${projectId}/pipelines/${pipelineId}`);
  return res.data.data;
}

export async function createPipeline(projectId: string, data: {
  name: string;
  description?: string;
  config?: PipelineConfig;
  templateId?: string;
  templateParams?: Record<string, string>;
}): Promise<Pipeline> {
  const res = await http.post<ApiResponse<Pipeline>>(`/projects/${projectId}/pipelines`, data);
  return res.data.data;
}

export async function createPipelineFromTemplate(projectId: string, data: {
  name: string;
  templateId: string;
  params?: Record<string, string>;
}): Promise<Pipeline> {
  const res = await http.post<ApiResponse<Pipeline>>(`/projects/${projectId}/pipelines/from-template`, data);
  return res.data.data;
}

export async function updatePipeline(projectId: string, pipelineId: string, data: {
  name?: string;
  description?: string;
  status?: string;
  config?: PipelineConfig;
  triggerType?: string;
  concurrencyPolicy?: string;
}): Promise<Pipeline> {
  const res = await http.put<ApiResponse<Pipeline>>(`/projects/${projectId}/pipelines/${pipelineId}`, data);
  return res.data.data;
}

export async function deletePipeline(projectId: string, pipelineId: string): Promise<void> {
  await http.delete(`/projects/${projectId}/pipelines/${pipelineId}`);
}

export async function copyPipeline(projectId: string, pipelineId: string): Promise<Pipeline> {
  const res = await http.post<ApiResponse<Pipeline>>(`/projects/${projectId}/pipelines/${pipelineId}/copy`);
  return res.data.data;
}

export async function fetchTemplates(): Promise<{ items: PipelineTemplate[]; total: number }> {
  const res = await http.get<ApiResponse<{ items: PipelineTemplate[]; total: number }>>('/pipeline-templates');
  return res.data.data;
}

export async function fetchTemplate(templateId: string): Promise<PipelineTemplate> {
  const res = await http.get<ApiResponse<PipelineTemplate>>(`/pipeline-templates/${templateId}`);
  return res.data.data;
}
