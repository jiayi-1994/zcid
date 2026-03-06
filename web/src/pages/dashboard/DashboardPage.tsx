import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Grid, Skeleton, Tag, Message } from '@arco-design/web-react';
import { IconApps, IconCheckCircle, IconCloseCircle, IconThunderbolt } from '@arco-design/web-react/icon';
import { AppLayout } from '../../components/layout/AppLayout';
import { OnboardingCard, ONBOARDING_DISMISSED_KEY } from '../../components/onboarding/OnboardingCard';
import { fetchDashboardData, type DashboardProject, type DashboardStats } from '../../services/dashboard';

const healthColors: Record<string, string> = {
  healthy: 'green',
  degraded: 'orangered',
  failed: 'red',
  unknown: 'gray',
};

const healthLabels: Record<string, string> = {
  healthy: '健康',
  degraded: '降级',
  failed: '异常',
  unknown: '未知',
};

const runStatusColors: Record<string, string> = {
  succeeded: 'green',
  failed: 'red',
  cancelled: 'orange',
  running: 'arcoblue',
  pending: 'gray',
  queued: 'blue',
};

const runStatusLabels: Record<string, string> = {
  succeeded: '成功',
  failed: '失败',
  cancelled: '已取消',
  running: '运行中',
  pending: '待执行',
  queued: '排队中',
};

function StatCard({ icon, iconClass, label, value }: {
  icon: React.ReactNode;
  iconClass: string;
  label: string;
  value: number | string;
}) {
  return (
    <div className="stat-card">
      <div className={`stat-card-icon ${iconClass}`}>{icon}</div>
      <div className="stat-card-label">{label}</div>
      <div className="stat-card-value">{value}</div>
    </div>
  );
}

export function DashboardPage() {
  const navigate = useNavigate();
  const [projects, setProjects] = useState<DashboardProject[]>([]);
  const [stats, setStats] = useState<DashboardStats | null>(null);
  const [loading, setLoading] = useState(true);
  const [showOnboarding, setShowOnboarding] = useState(() =>
    typeof window !== 'undefined' ? !localStorage.getItem(ONBOARDING_DISMISSED_KEY) : false
  );

  useEffect(() => {
    let mounted = true;
    const load = async () => {
      try {
        const data = await fetchDashboardData();
        if (mounted) {
          setProjects(data.projects);
          setStats(data.stats);
        }
      } catch {
        if (mounted) Message.error('加载仪表盘数据失败');
      } finally {
        if (mounted) setLoading(false);
      }
    };
    load();
    return () => { mounted = false; };
  }, []);

  return (
    <AppLayout>
      <div className="page-container">
        <div className="page-header">
          <div>
            <h2 className="page-title">Dashboard</h2>
            <p className="page-subtitle">项目概览与构建状态一览</p>
          </div>
        </div>

        {showOnboarding && (
          <OnboardingCard onDismiss={() => setShowOnboarding(false)} />
        )}

        {stats && (
          <Grid.Row gutter={16} style={{ marginBottom: 'var(--zcid-space-section)' }}>
            <Grid.Col xs={24} sm={12} md={6}>
              <StatCard
                icon={<IconApps />}
                iconClass="stat-card-icon--primary"
                label="项目总数"
                value={stats.totalProjects}
              />
            </Grid.Col>
            <Grid.Col xs={24} sm={12} md={6}>
              <StatCard
                icon={<IconThunderbolt />}
                iconClass="stat-card-icon--warning"
                label="流水线总数"
                value={stats.totalPipelines}
              />
            </Grid.Col>
            <Grid.Col xs={24} sm={12} md={6}>
              <StatCard
                icon={<IconCheckCircle />}
                iconClass="stat-card-icon--success"
                label="最近运行成功"
                value={stats.recentRunsSuccess}
              />
            </Grid.Col>
            <Grid.Col xs={24} sm={12} md={6}>
              <StatCard
                icon={<IconCloseCircle />}
                iconClass="stat-card-icon--error"
                label="最近运行失败"
                value={stats.recentRunsFail}
              />
            </Grid.Col>
          </Grid.Row>
        )}

        <h3 className="section-title">项目概览</h3>
        <Grid.Row gutter={16}>
          {loading
            ? Array.from({ length: 4 }).map((_, i) => (
                <Grid.Col key={i} xs={24} sm={12} md={8} lg={6}>
                  <div className="stat-card">
                    <Skeleton text={{ rows: 3 }} animation />
                  </div>
                </Grid.Col>
              ))
            : projects.map((p) => (
                <Grid.Col key={p.id} xs={24} sm={12} md={8} lg={6}>
                  <div
                    className="project-card"
                    onClick={() => navigate(`/projects/${p.id}/environments`)}
                  >
                    <div className="project-card-name">{p.name}</div>
                    <div className="project-card-desc">
                      {p.description || '暂无描述'}
                    </div>
                    <div className="project-card-footer">
                      {p.lastRunStatus && (
                        <Tag size="small" color={runStatusColors[p.lastRunStatus] ?? 'default'}>
                          {runStatusLabels[p.lastRunStatus] ?? p.lastRunStatus}
                        </Tag>
                      )}
                      <Tag size="small" color={healthColors[p.envHealthSummary] ?? 'default'}>
                        {healthLabels[p.envHealthSummary]}
                      </Tag>
                    </div>
                  </div>
                </Grid.Col>
              ))}
        </Grid.Row>
      </div>
    </AppLayout>
  );
}
