import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Card, Grid, Skeleton, Space, Statistic, Tag, Typography, Message } from '@arco-design/web-react';
import { IconApps, IconCheckCircle, IconCloseCircle } from '@arco-design/web-react/icon';
import { AppLayout } from '../../components/layout/AppLayout';
import { OnboardingCard } from '../../components/onboarding/OnboardingCard';
import { ONBOARDING_DISMISSED_KEY } from '../../components/onboarding/OnboardingCard';
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
          <h2 className="page-title">Dashboard</h2>
        </div>

        {showOnboarding && (
          <OnboardingCard onDismiss={() => setShowOnboarding(false)} />
        )}

        {stats && (
          <Grid.Row gutter={16} style={{ marginBottom: 'var(--zcid-space-section)' }}>
            <Grid.Col xs={24} sm={12} md={8} lg={6}>
              <Card className="zcid-card">
                <Statistic
                  title="项目总数"
                  value={stats.totalProjects}
                  prefix={<IconApps />}
                />
              </Card>
            </Grid.Col>
            <Grid.Col xs={24} sm={12} md={8} lg={6}>
              <Card className="zcid-card">
                <Statistic
                  title="流水线总数"
                  value={stats.totalPipelines}
                />
              </Card>
            </Grid.Col>
            <Grid.Col xs={24} sm={12} md={8} lg={6}>
              <Card className="zcid-card">
                <Statistic
                  title="最近运行成功"
                  value={stats.recentRunsSuccess}
                  prefix={<IconCheckCircle style={{ color: 'var(--zcid-color-success)' }} />}
                />
              </Card>
            </Grid.Col>
            <Grid.Col xs={24} sm={12} md={8} lg={6}>
              <Card className="zcid-card">
                <Statistic
                  title="最近运行失败"
                  value={stats.recentRunsFail}
                  prefix={<IconCloseCircle style={{ color: 'var(--zcid-color-danger)' }} />}
                />
              </Card>
            </Grid.Col>
          </Grid.Row>
        )}

        <Typography.Title heading={5} style={{ marginTop: 0, marginBottom: 'var(--zcid-space-section)' }}>
          项目概览
        </Typography.Title>
        <Grid.Row gutter={16}>
          {loading
            ? Array.from({ length: 4 }).map((_, i) => (
                <Grid.Col key={i} xs={24} sm={12} md={8} lg={6}>
                  <Card className="zcid-card">
                    <Skeleton text={{ rows: 4 }} animation />
                  </Card>
                </Grid.Col>
              ))
            : projects.map((p) => (
                <Grid.Col key={p.id} xs={24} sm={12} md={8} lg={6}>
                  <Card
                    className="zcid-card zcid-card-interactive"
                    onClick={() => navigate(`/projects/${p.id}/environments`)}
                  >
                    <Space direction="vertical" size="small" style={{ width: '100%' }}>
                      <Typography.Text bold>{p.name}</Typography.Text>
                      <Typography.Text type="secondary" ellipsis={{ showTooltip: true }}>
                        {p.description || '暂无描述'}
                      </Typography.Text>
                      <Space>
                        {p.lastRunStatus && (
                          <Tag color={runStatusColors[p.lastRunStatus] ?? 'default'}>
                            {runStatusLabels[p.lastRunStatus] ?? p.lastRunStatus}
                          </Tag>
                        )}
                        <Tag color={healthColors[p.envHealthSummary] ?? 'default'}>
                          {healthLabels[p.envHealthSummary]}
                        </Tag>
                      </Space>
                    </Space>
                  </Card>
                </Grid.Col>
              ))}
        </Grid.Row>
      </div>
    </AppLayout>
  );
}
