import { fetchProjects } from './project';
import { fetchPipelines } from './pipeline';
import { fetchPipelineRuns } from './pipelineRun';
import { fetchDeployments } from './deployment';

export interface DashboardProject {
  id: string;
  name: string;
  description: string;
  lastRunStatus?: string;
  envHealthSummary: 'healthy' | 'degraded' | 'failed' | 'unknown';
}

export interface DashboardStats {
  totalProjects: number;
  totalPipelines: number;
  recentRunsSuccess: number;
  recentRunsFail: number;
}

export interface DashboardData {
  projects: DashboardProject[];
  stats: DashboardStats;
}

function getEnvHealthFromDeployments(
  items: { healthStatus?: string }[]
): 'healthy' | 'degraded' | 'failed' | 'unknown' {
  if (!items.length) return 'unknown';
  const healthSet = new Set(
    items
      .map((d) => (d.healthStatus ?? '').toLowerCase())
      .filter(Boolean)
  );
  if (healthSet.has('healthy') && healthSet.size === 1) return 'healthy';
  if (healthSet.has('degraded') || healthSet.has('progressing')) return 'degraded';
  if (healthSet.has('suspended') || healthSet.has('missing') || healthSet.has('unknown')) return 'failed';
  if (healthSet.has('failed')) return 'failed';
  return healthSet.size ? 'degraded' : 'unknown';
}

export async function fetchDashboardData(): Promise<DashboardData> {
  const projectRes = await fetchProjects(1, 100);
  const projects = projectRes.items ?? [];
  const stats: DashboardStats = {
    totalProjects: projectRes.total,
    totalPipelines: 0,
    recentRunsSuccess: 0,
    recentRunsFail: 0,
  };

  const dashboardProjects: DashboardProject[] = [];
  const runStatuses: string[] = [];

  await Promise.all(
    projects.slice(0, 20).map(async (p) => {
      try {
        const [pipelinesRes, deploymentsRes] = await Promise.all([
          fetchPipelines(p.id, 1, 10),
          fetchDeployments(p.id, 1, 50),
        ]);

        stats.totalPipelines += pipelinesRes.total;

        let lastRunStatus: string | undefined;
        const pipelines = pipelinesRes.items ?? [];
        const firstPipeline = pipelines.find((pl) => pl.status === 'active') ?? pipelines[0];
        if (firstPipeline) {
          try {
            const runsRes = await fetchPipelineRuns(p.id, firstPipeline.id, 1, 1);
            const latest = runsRes.items?.[0];
            if (latest) {
              lastRunStatus = latest.status;
              runStatuses.push(latest.status);
            }
          } catch {
            // ignore
          }
        }

        const envHealth = getEnvHealthFromDeployments(deploymentsRes.items ?? []);
        dashboardProjects.push({
          id: p.id,
          name: p.name,
          description: p.description ?? '',
          lastRunStatus,
          envHealthSummary: envHealth,
        });
      } catch {
        dashboardProjects.push({
          id: p.id,
          name: p.name,
          description: p.description ?? '',
          envHealthSummary: 'unknown',
        });
      }
    })
  );

  for (const s of runStatuses) {
    if (s === 'succeeded') stats.recentRunsSuccess += 1;
    else if (s === 'failed' || s === 'cancelled') stats.recentRunsFail += 1;
  }

  return {
    projects: dashboardProjects,
    stats,
  };
}
