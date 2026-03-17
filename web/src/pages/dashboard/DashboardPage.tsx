import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Grid, Skeleton, Tag, Message, Button } from '@arco-design/web-react';
import { IconApps, IconCheckCircle, IconCloseCircle, IconThunderbolt, IconPlus, IconArrowRight, IconPlayArrow, IconBook, IconLink } from '@arco-design/web-react/icon';
import { AppLayout } from '../../components/layout/AppLayout';
import { fetchDashboardData, type DashboardProject, type DashboardStats } from '../../services/dashboard';
import { useAuthStore } from '../../stores/auth';

const { Row, Col } = Grid;

const runStatusLabels: Record<string, string> = {
  succeeded: '成功', failed: '失败', cancelled: '已取消',
  running: '运行中', pending: '待执行', queued: '排队中',
};

function MetricCard({ icon, label, value, trend }: {
  icon: React.ReactNode; label: string; value: number | string; trend?: string;
}) {
  return (
    <div style={{
      background: 'var(--card)', border: '1px solid var(--border)', borderRadius: 10,
      padding: '16px 20px', display: 'flex', alignItems: 'center', gap: 16,
    }}>
      <div style={{
        width: 40, height: 40, borderRadius: 8, background: 'var(--muted)',
        display: 'flex', alignItems: 'center', justifyContent: 'center',
        fontSize: 18, color: 'var(--foreground)', flexShrink: 0,
      }}>
        {icon}
      </div>
      <div>
        <div style={{ fontSize: 12, color: 'var(--muted-foreground)', marginBottom: 2 }}>{label}</div>
        <div style={{ display: 'flex', alignItems: 'baseline', gap: 6 }}>
          <span style={{ fontSize: 22, fontWeight: 700, color: 'var(--foreground)', letterSpacing: -0.5 }}>{value}</span>
          {trend && <span style={{ fontSize: 11, color: 'var(--muted-foreground)' }}>{trend}</span>}
        </div>
      </div>
    </div>
  );
}

function QuickAction({ icon, title, desc, onClick }: {
  icon: React.ReactNode; title: string; desc: string; onClick: () => void;
}) {
  return (
    <div
      onClick={onClick}
      style={{
        display: 'flex', alignItems: 'center', gap: 12, padding: '12px 14px',
        borderRadius: 8, cursor: 'pointer', transition: 'background 120ms',
        border: '1px solid var(--border)',
      }}
      onMouseEnter={(e) => { (e.currentTarget as HTMLElement).style.background = 'var(--muted)'; }}
      onMouseLeave={(e) => { (e.currentTarget as HTMLElement).style.background = 'transparent'; }}
    >
      <div style={{
        width: 36, height: 36, borderRadius: 8, background: 'var(--muted)',
        display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: 16,
        flexShrink: 0, color: 'var(--foreground)',
      }}>
        {icon}
      </div>
      <div style={{ flex: 1, minWidth: 0 }}>
        <div style={{ fontSize: 13, fontWeight: 600, color: 'var(--foreground)' }}>{title}</div>
        <div style={{ fontSize: 12, color: 'var(--muted-foreground)', marginTop: 1 }}>{desc}</div>
      </div>
      <IconArrowRight style={{ fontSize: 14, color: 'var(--muted-foreground)', flexShrink: 0 }} />
    </div>
  );
}

function ProjectRow({ project, onClick }: { project: DashboardProject; onClick: () => void }) {
  const statusCfg: Record<string, { dot: string; label: string }> = {
    succeeded: { dot: '#22C55E', label: '成功' },
    failed:    { dot: '#EF4444', label: '失败' },
    running:   { dot: '#3B82F6', label: '运行中' },
    queued:    { dot: '#3B82F6', label: '排队中' },
    pending:   { dot: '#A1A1AA', label: '待执行' },
    cancelled: { dot: '#F59E0B', label: '已取消' },
  };
  const cfg = project.lastRunStatus ? statusCfg[project.lastRunStatus] || null : null;

  return (
    <div
      onClick={onClick}
      style={{
        display: 'flex', alignItems: 'center', gap: 12,
        padding: '10px 14px', borderRadius: 8, cursor: 'pointer',
        transition: 'background 120ms',
      }}
      onMouseEnter={(e) => { (e.currentTarget as HTMLElement).style.background = 'var(--muted)'; }}
      onMouseLeave={(e) => { (e.currentTarget as HTMLElement).style.background = 'transparent'; }}
    >
      <div style={{
        width: 32, height: 32, borderRadius: 8, flexShrink: 0,
        background: 'var(--primary)', color: 'var(--primary-foreground)',
        display: 'flex', alignItems: 'center', justifyContent: 'center',
        fontSize: 13, fontWeight: 600,
      }}>
        {project.name.charAt(0).toUpperCase()}
      </div>
      <div style={{ flex: 1, minWidth: 0 }}>
        <div style={{ fontSize: 13, fontWeight: 500, color: 'var(--foreground)', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
          {project.name}
        </div>
        <div style={{ fontSize: 12, color: 'var(--muted-foreground)', marginTop: 1, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
          {project.description || '暂无描述'}
        </div>
      </div>
      {cfg && (
        <span style={{
          display: 'inline-flex', alignItems: 'center', gap: 4,
          padding: '2px 8px', borderRadius: 999, fontSize: 11, fontWeight: 500,
          background: 'var(--muted)', color: 'var(--muted-foreground)',
        }}>
          <span style={{ width: 5, height: 5, borderRadius: '50%', background: cfg.dot }} />
          {cfg.label}
        </span>
      )}
      <IconArrowRight style={{ fontSize: 12, color: '#D4D4D8', flexShrink: 0 }} />
    </div>
  );
}

export function DashboardPage() {
  const navigate = useNavigate();
  const user = useAuthStore((s) => s.user);
  const [projects, setProjects] = useState<DashboardProject[]>([]);
  const [stats, setStats] = useState<DashboardStats | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let mounted = true;
    fetchDashboardData()
      .then((data) => { if (mounted) { setProjects(data.projects); setStats(data.stats); } })
      .catch(() => { if (mounted) Message.error('加载仪表盘数据失败'); })
      .finally(() => { if (mounted) setLoading(false); });
    return () => { mounted = false; };
  }, []);

  const greeting = (() => {
    const h = new Date().getHours();
    if (h < 12) return '早上好';
    if (h < 18) return '下午好';
    return '晚上好';
  })();

  return (
    <AppLayout>
      <div className="page-container" style={{ maxWidth: 1100 }}>
        {/* Greeting header */}
        <div style={{ marginBottom: 24 }}>
          <h2 style={{ margin: 0, fontSize: 20, fontWeight: 600, color: 'var(--foreground)', letterSpacing: -0.3 }} aria-label="Dashboard">
            {greeting}，{user?.username ?? 'User'} 👋
          </h2>
          <p style={{ margin: '4px 0 0', fontSize: 13, color: 'var(--muted-foreground)' }}>
            这是你的 CI/CD 工作台概览
          </p>
        </div>

        {/* Metrics */}
        {loading ? (
          <Row gutter={12} style={{ marginBottom: 24 }}>
            {[1,2,3,4].map(i => (
              <Col key={i} span={6}>
                <div style={{ background: 'var(--card)', border: '1px solid var(--border)', borderRadius: 10, padding: 20 }}>
                  <Skeleton text={{ rows: 2 }} animation />
                </div>
              </Col>
            ))}
          </Row>
        ) : stats && (
          <Row gutter={12} style={{ marginBottom: 24 }}>
            <Col span={6}><MetricCard icon={<IconApps />} label="项目" value={stats.totalProjects} /></Col>
            <Col span={6}><MetricCard icon={<IconThunderbolt />} label="流水线" value={stats.totalPipelines} /></Col>
            <Col span={6}><MetricCard icon={<IconCheckCircle />} label="近期成功" value={stats.recentRunsSuccess} trend="次运行" /></Col>
            <Col span={6}><MetricCard icon={<IconCloseCircle />} label="近期失败" value={stats.recentRunsFail} trend="次运行" /></Col>
          </Row>
        )}

        {/* Two-column layout */}
        <Row gutter={16}>
          {/* Left: Projects */}
          <Col span={16}>
            <div style={{
              background: 'var(--card)', border: '1px solid var(--border)', borderRadius: 10,
              overflow: 'hidden',
            }}>
              <div style={{
                display: 'flex', alignItems: 'center', justifyContent: 'space-between',
                padding: '14px 18px', borderBottom: '1px solid var(--border)',
              }}>
                <span style={{ fontSize: 14, fontWeight: 600, color: 'var(--foreground)' }}>项目</span>
                <Button size="mini" type="text" onClick={() => navigate('/projects')} style={{ fontSize: 12, color: 'var(--muted-foreground)' }}>
                  查看全部 <IconArrowRight style={{ fontSize: 10 }} />
                </Button>
              </div>
              <div style={{ padding: '6px 8px' }}>
                {loading ? (
                  <div style={{ padding: 20 }}><Skeleton text={{ rows: 4 }} animation /></div>
                ) : projects.length === 0 ? (
                  <div style={{ textAlign: 'center', padding: '32px 16px', color: 'var(--muted-foreground)' }}>
                    <div style={{ fontSize: 13, marginBottom: 12 }}>还没有项目</div>
                    <Button size="small" type="primary" icon={<IconPlus />} onClick={() => navigate('/projects')} style={{ borderRadius: 6 }}>
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
          </Col>

          {/* Right: Quick Actions */}
          <Col span={8}>
            <div style={{
              background: 'var(--card)', border: '1px solid var(--border)', borderRadius: 10,
              overflow: 'hidden',
            }}>
              <div style={{
                padding: '14px 18px', borderBottom: '1px solid var(--border)',
              }}>
                <span style={{ fontSize: 14, fontWeight: 600, color: 'var(--foreground)' }}>快速操作</span>
              </div>
              <div style={{ padding: '8px', display: 'flex', flexDirection: 'column', gap: 6 }}>
                <QuickAction
                  icon={<IconPlus />}
                  title="新建项目"
                  desc="创建一个新的 CI/CD 项目"
                  onClick={() => navigate('/projects')}
                />
                <QuickAction
                  icon={<IconPlayArrow />}
                  title="创建流水线"
                  desc="配置自动化构建和部署"
                  onClick={() => navigate('/projects')}
                />
                <QuickAction
                  icon={<IconLink />}
                  title="集成管理"
                  desc="连接 Git 仓库和镜像仓库"
                  onClick={() => navigate('/admin/integrations')}
                />
                <QuickAction
                  icon={<IconBook />}
                  title="查看文档"
                  desc="了解平台功能和最佳实践"
                  onClick={() => window.open('https://github.com/jiayi-1994/zcid', '_blank')}
                />
              </div>
            </div>
          </Col>
        </Row>
      </div>
    </AppLayout>
  );
}
