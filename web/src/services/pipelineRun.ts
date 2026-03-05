import { http } from './http';
import type { ApiResponse } from './types';

export interface PipelineRun {
  id: string;
  pipelineId: string;
  projectId: string;
  runNumber: number;
  status: string;
  triggerType: string;
  triggeredBy?: string;
  gitBranch?: string;
  gitCommit?: string;
  gitAuthor?: string;
  gitMessage?: string;
  params?: Record<string, string>;
  tektonName?: string;
  namespace?: string;
  startedAt?: string;
  finishedAt?: string;
  errorMessage?: string;
  artifacts?: Artifact[];
  createdAt: string;
  updatedAt: string;
}

export interface PipelineRunSummary {
  id: string;
  pipelineId: string;
  runNumber: number;
  status: string;
  triggerType: string;
  triggeredBy?: string;
  gitBranch?: string;
  gitCommit?: string;
  startedAt?: string;
  finishedAt?: string;
  createdAt: string;
}

export interface PipelineRunList {
  items: PipelineRunSummary[];
  total: number;
  page: number;
  pageSize: number;
}

export interface Artifact {
  type: string;
  name: string;
  url: string;
  size?: number;
}

export interface LogEntry {
  seq: number;
  stepId: string;
  content: string;
  level: string;
  timestamp: string;
}

export interface ArchivedLogsResponse {
  items: LogEntry[];
  total: number;
  page: number;
  pageSize: number;
}

export async function fetchPipelineRuns(
  projectId: string,
  pipelineId: string,
  page = 1,
  pageSize = 20
): Promise<PipelineRunList> {
  const res = await http.get<ApiResponse<PipelineRunList>>(
    `/projects/${projectId}/pipelines/${pipelineId}/runs`,
    { params: { page, pageSize } }
  );
  return res.data.data;
}

export async function fetchPipelineRun(
  projectId: string,
  pipelineId: string,
  runId: string
): Promise<PipelineRun> {
  const res = await http.get<ApiResponse<PipelineRun>>(
    `/projects/${projectId}/pipelines/${pipelineId}/runs/${runId}`
  );
  return res.data.data;
}

export async function triggerPipelineRun(
  projectId: string,
  pipelineId: string,
  params: { params?: Record<string, string>; gitBranch?: string; gitCommit?: string }
): Promise<PipelineRun> {
  const res = await http.post<ApiResponse<PipelineRun>>(
    `/projects/${projectId}/pipelines/${pipelineId}/runs`,
    params
  );
  return res.data.data;
}

export async function cancelPipelineRun(
  projectId: string,
  pipelineId: string,
  runId: string
): Promise<void> {
  await http.post(`/projects/${projectId}/pipelines/${pipelineId}/runs/${runId}/cancel`);
}

export async function fetchRunArtifacts(
  projectId: string,
  pipelineId: string,
  runId: string
): Promise<Artifact[]> {
  const res = await http.get<ApiResponse<{ artifacts: Artifact[] }>>(
    `/projects/${projectId}/pipelines/${pipelineId}/runs/${runId}/artifacts`
  );
  return res.data.data.artifacts ?? [];
}

export async function fetchArchivedLogs(
  projectId: string,
  runId: string,
  page = 1,
  pageSize = 50
): Promise<ArchivedLogsResponse> {
  const res = await http.get<ApiResponse<ArchivedLogsResponse>>(
    `/projects/${projectId}/pipeline-runs/${runId}/logs`,
    { params: { page, pageSize } }
  );
  return res.data.data;
}
