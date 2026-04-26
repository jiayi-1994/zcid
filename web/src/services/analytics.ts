import { http } from './http';
import type { ApiResponse } from './types';

export interface AnalyticsSummary {
  totalRuns: number;
  successRate: number;
  medianDurationMs: number;
  p95DurationMs: number;
}

export interface DailyStat {
  date: string;
  total: number;
  succeeded: number;
  failed: number;
  medianDurationMs: number;
  successRate: number;
}

export interface TopFailingStep {
  stepName: string;
  taskRunName: string;
  failureCount: number;
  totalCount: number;
  failureRate: number;
}

export interface TopPipeline {
  pipelineId: string;
  pipelineName: string;
  runCount: number;
  successRate: number;
}

export interface AnalyticsResponse {
  range: '7d' | '30d' | '90d';
  summary: AnalyticsSummary;
  dailyStats: DailyStat[];
  topFailingSteps: TopFailingStep[];
  topPipelines: TopPipeline[];
}

export async function fetchAnalytics(projectId: string, range: '7d' | '30d' | '90d') {
  const res = await http.get<ApiResponse<AnalyticsResponse>>(`/projects/${projectId}/analytics`, { params: { range } });
  return res.data.data;
}
