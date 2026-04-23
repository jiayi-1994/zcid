import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Skeleton, Message, Button } from '@arco-design/web-react';
import {
  IconApps, IconCheckCircle, IconCloseCircle, IconThunderbolt,
  IconPlus, IconArrowRight, IconPlayArrow, IconBook, IconLink, IconRefresh,
} from '@arco-design/web-react/icon';
import { AppLayout } from '../../components/layout/AppLayout';
import { fetchDashboardData, type DashboardProject, type DashboardStats } from '../../services/dashboard';
import { useAuthStore } from '../../stores/auth';

const runStatusCfg: Record<string, { cls: string; label: string }> = {
  succeeded: { cls: 'pipeline-status-badge--success', label: 'Success' },
  failed:    { cls: 'pipeline-status-badge--failed', label: 'Failed' },
  running:   { cls: 'pipeline-status-badge--running', label: 'Running' },
  queued:    { cls: 'pipeline-status-badge--running', label: 'Queued' },
  pending:   { cls: 'pipeline-status-badge--pending', label: 'Pending' },
  cancelled: { cls: 'pipeline-status-badge--cancelled', label: 'Cancelled' },
};

function MetricTile({ icon, label, value, trend, tone }: {
  icon: React.ReactNode;
  label: string;
  value: number | string;
  trend?: string;
  tone?: 'primary' | 'success' | 'error';
}) {
  const toneCls = tone === 'success' ? 'stat-card-icon--success'
    : tone === 'error' ? 'stat-card-icon--error'
    : 'stat-card-icon--primary';
  return (
    <div className="metric-card">
      <div className={`stat-card-icon ${toneCls}`}>{icon}</div>
      <div className="metric-card-label">{label}</div>
      <div className="metric-card-value">{value}</div>
      {trend && <div className="metric-card-sub">{trend}</div>}
    </div>
  );
}

function QuickAction({ icon, title, desc, onClick }: {
  icon: React.ReactNode; title: string; desc: string; onClick: () => void;
}) {
  return (
    <div className="onboarding-step" onClick={onClick}>
      <div className="onboarding-step-icon">{icon}</div>
      <div className="onboarding-step-title">{title}</div>
      <div className="onboarding-step-desc">{desc}</div>
    </div>
  );
}

function ProjectRow({ project, onClick }: { project: DashboardProject; onClick: () => void }) {
  const cfg = project.lastRunStatus ? runStatusCfg[project.lastRunStatus] : null;
  const statusIconCls = project.lastRunStatus === 'succeeded' ? 'pipeline-status-icon--success'
    : project.lastRunStatus === 'failed' ? 'pipeline-status-icon--failed'
    : project.lastRunStatus === 'running' || project.lastRunStatus === 'queued' ? 'pipeline-status-icon--running'
    : 'pipeline-status-icon--pending';

  return (
    <div className="pipeline-status-card" onClick={onClick}>
      <div className={`pipeline-status-icon ${statusIconCls}`}>
        {project.name.charAt(0).toUpperCase()}
      </div>
      <div className="pipeline-status-info">
        <div className="pipeline-status-name">{project.name}</div>
        <div className="pipeline-status-meta">{project.description || '暂无描述'}</div>
      </div>
      {cfg && (
        <span className={`pipeline-status-badge ${cfg.cls}`}>{cfg.label}</span>
      )}
      <IconArrowRight style={{ fontSize: 14, color: 'var(--on-surface-variant)', flexShrink: 0 }} />
    </div>
  );
}

export function DashboardPage() {
  const navigate = useNavigate();
  const user = useAuthStore((s) => s.user);
  const [projects, setProjects] = useState<DashboardProject[]>([]);
  const [stats, setStats] = useState<DashboardStats | null>(null);
  const [loading, setLoading] = useState(true);
  const [reloadKey, setReloadKey] = useState(0);

  useEffect(() => {
    let mounted = true;
    setLoading(true);
    fetchDashboardData()
      .then((data) => { if (mounted) { setProjects(data.projects); setStats(data.stats); } })
      .catch(() => { if (mounted) Message.error('加载仪表盘数据失败'); })
      .finally(() => { if (mounted) setLoading(false); });
    return () => { mounted = false; };
  }, [reloadKey]);

  const greeting = (() => {
    const h = new Date().getHours();
    if (h < 12) return '早上好';
    if (h < 18) return '下午好';
    return '晚上好';
  })();

  const successRate = stats && (stats.recentRunsSuccess + stats.recentRunsFail) > 0
    ? Math.round((stats.recentRunsSuccess / (stats.recentRunsSuccess + stats.recentRunsFail)) * 100)
    : null;

  return (
    <AppLayout>
      <div className="page-container">
        <div className="page-header">
          <div>
            <div className="breadcrumb">Cloud-Native Overview</div>
            <h1 className="page-title">
              {greeting}，{user?.username ?? 'User'}
            </h1>
            <p className="page-subtitle">
              Global infrastructure health and deployment telemetry. 当前 CI/CD 工作台概览。
            </p>
          </div>
          <div style={{ display: 'flex', gap: 8 }}>
            <Button
              icon={<IconRefresh />}
              onClick={() => setReloadKey((k) => k + 1)}
            >
              Refresh
            </Button>
            <Button type="primary" icon={<IconPlus />} onClick={() => navigate('/projects')}>
              New Pipeline
            </Button>
          </div>
        </div>

        {loading ? (
          <div className="metrics-grid">
            {[1, 2, 3, 4].map(i => (
              <div key={i} className="metric-card">
                <Skeleton text={{ rows: 2 }} animation />
              </div>
            ))}
          </div>
        ) : stats && (
          <div className="metrics-grid">
            <MetricTile icon={<IconApps />} label="Total Projects" value={stats.totalProjects} tone="primary" />
            <MetricTile icon={<IconThunderbolt />} label="Total Pipelines" value={stats.totalPipelines} tone="primary" />
            <MetricTile
              icon={<IconCheckCircle />}
              label="Recent Success"
              value={stats.recentRunsSuccess}
              trend={successRate !== null ? `${successRate}% success rate` : undefined}
              tone="success"
            />
            <MetricTile
              icon={<IconCloseCircle />}
              label="Recent Failures"
              value={stats.recentRunsFail}
              trend={stats.recentRunsFail > 0 ? 'Needs attention' : 'All green'}
              tone="error"
            />
          </div>
        )}

        <div className="dashboard-hero">
          <div className="dashboard-main-col">
            <div className="zcid-card" style={{ padding: 'var(--space-6)' }}>
              <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: 'var(--space-4)' }}>
                <h2 className="section-title" style={{ margin: 0 }}>Projects</h2>
                <Button size="mini" type="text" onClick={() => navigate('/projects')}>
                  查看全部 <IconArrowRight style={{ fontSize: 10, marginLeft: 4 }} />
                </Button>
              </div>
              <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
                {loading ? (
                  <Skeleton text={{ rows: 4 }} animation />
                ) : projects.length === 0 ? (
                  <div className="empty-state">
                    <div className="empty-state-title">还没有项目</div>
                    <div className="empty-state-desc">创建首个 CI/CD 项目开始工作</div>
                    <Button type="primary" icon={<IconPlus />} onClick={() => navigate('/projects')}>
                      创建项目
                    </Button>
                  </div>
                ) : (
                  projects.map((p) => (
                    <ProjectRow key={p.id} project={p} onClick={() => navigate(`/projects/${p.id}/pipelines`)} />
                  ))
                )}
              </div>
            </div>
          </div>

          <div className="dashboard-side-col">
            <div className="zcid-card" style={{ padding: 'var(--space-6)' }}>
              <h2 className="section-title">快速操作</h2>
              <div className="onboarding-steps" style={{ gridTemplateColumns: '1fr' }}>
                <QuickAction icon={<IconPlus />} title="新建项目" desc="创建一个新的 CI/CD 项目"
                  onClick={() => navigate('/projects')} />
                <QuickAction icon={<IconPlayArrow />} title="创建流水线" desc="配置自动化构建与部署"
                  onClick={() => navigate('/projects')} />
                <QuickAction icon={<IconLink />} title="集成管理" desc="连接 Git 仓库与镜像仓库"
                  onClick={() => navigate('/admin/integrations')} />
                <QuickAction icon={<IconBook />} title="查看文档" desc="了解平台功能与最佳实践"
                  onClick={() => window.open('https://github.com/jiayi-1994/zcid', '_blank')} />
              </div>
            </div>
          </div>
        </div>
      </div>
    </AppLayout>
  );
}
