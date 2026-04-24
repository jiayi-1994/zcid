import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Message } from '@arco-design/web-react';
import { AppLayout } from '../../components/layout/AppLayout';
import { fetchDashboardData, type DashboardProject, type DashboardStats } from '../../services/dashboard';
import { useAuthStore } from '../../stores/auth';
import { PageHeader } from '../../components/ui/PageHeader';
import { Metric } from '../../components/ui/Metric';
import { Card } from '../../components/ui/Card';
import { StatusBadge } from '../../components/ui/StatusBadge';
import { Avatar } from '../../components/ui/Avatar';
import { Btn } from '../../components/ui/Btn';
import {
  IFolder, IZap, ICheck, IX, IRefresh, IPlus, IChevR, IArrR,
  IPlug, IBook,
} from '../../components/ui/icons';

function greet() {
  const h = new Date().getHours();
  if (h < 6) return '深夜好';
  if (h < 12) return '早上好';
  if (h < 18) return '下午好';
  return '晚上好';
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
      .then((d) => { if (mounted) { setProjects(d.projects); setStats(d.stats); } })
      .catch(() => { if (mounted) Message.error('加载仪表盘数据失败'); })
      .finally(() => { if (mounted) setLoading(false); });
    return () => { mounted = false; };
  }, [reloadKey]);

  const successRate = stats && (stats.recentRunsSuccess + stats.recentRunsFail) > 0
    ? Math.round((stats.recentRunsSuccess / (stats.recentRunsSuccess + stats.recentRunsFail)) * 100)
    : null;

  return (
    <AppLayout>
      <PageHeader
        crumb="Cloud-Native Overview"
        title={`${greet()}，${user?.username ?? 'User'}`}
        sub="Global infrastructure health and deployment telemetry. 当前 CI/CD 工作台概览。"
        actions={
          <>
            <Btn size="sm" icon={<IRefresh size={13} />} onClick={() => setReloadKey((k) => k + 1)}>Refresh</Btn>
            <Btn size="sm" variant="primary" icon={<IPlus size={13} />} onClick={() => navigate('/projects')}>New Pipeline</Btn>
          </>
        }
      />

      <div style={{ padding: 24, display: 'flex', flexDirection: 'column', gap: 20 }}>
        {/* Metrics */}
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(4, 1fr)', gap: 14 }}>
          <Metric
            label="TOTAL PROJECTS"
            value={loading ? '—' : (stats?.totalProjects ?? 0)}
            icon={<IFolder size={14} />}
            iconBg="color-mix(in oklch, var(--accent-1), white 85%)"
            iconColor="var(--accent-ink)"
            trend={loading ? undefined : `${projects.length} active`}
          />
          <Metric
            label="TOTAL PIPELINES"
            value={loading ? '—' : (stats?.totalPipelines ?? 0)}
            icon={<IZap size={14} />}
            iconBg="var(--z-100)"
            iconColor="var(--z-700)"
          />
          <Metric
            label="RECENT SUCCESS"
            value={loading ? '—' : (stats?.recentRunsSuccess ?? 0)}
            icon={<ICheck size={14} />}
            trend={successRate !== null ? `${successRate}% success rate` : undefined}
            trendTone="green"
            iconBg="var(--green-soft)"
            iconColor="var(--green-ink)"
          />
          <Metric
            label="RECENT FAILURES"
            value={loading ? '—' : (stats?.recentRunsFail ?? 0)}
            icon={<IX size={14} />}
            trend={(stats?.recentRunsFail ?? 0) > 0 ? 'Needs attention' : undefined}
            trendTone="red"
            iconBg="var(--red-soft)"
            iconColor="var(--red-ink)"
          />
        </div>

        {/* Main grid */}
        <div style={{ display: 'grid', gridTemplateColumns: '2fr 1fr', gap: 20 }}>
          <Card
            padding={false}
            title={
              <div style={{ display: 'flex', alignItems: 'baseline', gap: 8 }}>
                <h2>Projects</h2>
                <span className="sub">最近活跃</span>
              </div>
            }
            extra={
              <a style={{ fontSize: 12, color: 'var(--accent-ink)', cursor: 'pointer' }} onClick={() => navigate('/projects')}>
                查看全部 →
              </a>
            }
          >
            <div style={{ padding: 8 }}>
              {loading ? (
                <div style={{ padding: '24px 0', textAlign: 'center', color: 'var(--z-400)' }}>加载中...</div>
              ) : projects.length === 0 ? (
                <div style={{ padding: '24px 0', textAlign: 'center', color: 'var(--z-400)' }}>还没有项目</div>
              ) : (
                projects.slice(0, 5).map((p) => (
                  <div key={p.id} className="lrow" onClick={() => navigate(`/projects/${p.id}/pipelines`)}>
                    <Avatar name={p.name} size="sm" round />
                    <div style={{ flex: 1, minWidth: 0 }}>
                      <div style={{ fontSize: 13, fontWeight: 500, color: 'var(--z-900)' }}>{p.name}</div>
                      <div className="sub" style={{ fontSize: 11.5, whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis' }}>
                        {p.description || '暂无描述'}
                      </div>
                    </div>
                    {p.lastRunStatus && <StatusBadge status={p.lastRunStatus} />}
                    <IChevR size={14} style={{ color: 'var(--z-400)' }} />
                  </div>
                ))
              )}
            </div>
          </Card>

          <Card title="快速操作">
            <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
              {[
                { i: <IFolder size={15} />, t: '新建项目', s: '配置 git、变量和环境', action: () => navigate('/projects') },
                { i: <IZap size={15} />, t: '创建流水线', s: '从模板或空白开始', action: () => navigate('/projects') },
                { i: <IPlug size={15} />, t: '集成管理', s: '连接 GitHub / GitLab', action: () => navigate('/admin/integrations') },
                { i: <IBook size={15} />, t: '查看文档', s: '产品指南 & API', action: () => window.open('https://github.com/jiayi-1994/zcid', '_blank') },
              ].map((x) => (
                <div key={x.t} className="lrow" style={{ padding: '9px 10px' }} onClick={x.action}>
                  <div style={{ width: 30, height: 30, borderRadius: 8, background: 'var(--z-100)', color: 'var(--z-700)', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>{x.i}</div>
                  <div style={{ flex: 1, minWidth: 0 }}>
                    <div style={{ fontSize: 12.5, fontWeight: 500 }}>{x.t}</div>
                    <div className="sub" style={{ fontSize: 11 }}>{x.s}</div>
                  </div>
                  <IArrR size={13} style={{ color: 'var(--z-400)' }} />
                </div>
              ))}
            </div>
          </Card>
        </div>
      </div>
    </AppLayout>
  );
}
