import { useCallback, useEffect, useMemo, useState } from 'react';
import { Message } from '@arco-design/web-react';
import { useParams } from 'react-router-dom';
import { fetchAnalytics, type AnalyticsResponse, type DailyStat } from '../../../services/analytics';
import { PageHeader } from '../../../components/ui/PageHeader';
import { Card } from '../../../components/ui/Card';
import { Metric } from '../../../components/ui/Metric';
import { Badge } from '../../../components/ui/Badge';
import { Segmented } from '../../../components/ui/Segmented';
import { IClock, IGrid, ITarget, IZap } from '../../../components/ui/icons';

type RangeValue = '7d' | '30d' | '90d';

function pct(value: number) {
  return `${Math.round(value * 1000) / 10}%`;
}

function duration(ms: number) {
  if (!ms) return '-';
  const seconds = Math.round(ms / 1000);
  if (seconds < 60) return `${seconds}s`;
  const minutes = Math.floor(seconds / 60);
  return minutes < 60 ? `${minutes}m ${seconds % 60}s` : `${Math.floor(minutes / 60)}h ${minutes % 60}m`;
}

function DailyTrendChart({ data }: { data: DailyStat[] }) {
  const maxTotal = Math.max(1, ...data.map((d) => d.total));
  const points = data.map((d, i) => {
    const x = data.length <= 1 ? 50 : (i / (data.length - 1)) * 100;
    const y = 100 - d.successRate * 100;
    return `${x},${y}`;
  }).join(' ');

  if (data.length === 0) {
    return <div style={{ padding: '48px 0', textAlign: 'center', color: 'var(--z-400)' }}>这个时间范围内暂无运行数据。</div>;
  }

  return (
    <div style={{ display: 'grid', gridTemplateRows: '160px auto', gap: 10 }}>
      <div style={{ position: 'relative', height: 160, borderBottom: '1px solid var(--z-150)', borderLeft: '1px solid var(--z-150)', padding: '8px 8px 0' }}>
        <div style={{ position: 'absolute', inset: '8px 8px 0', display: 'flex', alignItems: 'end', gap: 6 }}>
          {data.map((d) => (
            <div key={d.date} title={`${d.date}: ${d.total} runs, ${pct(d.successRate)} success`} style={{ flex: 1, minWidth: 10, display: 'flex', alignItems: 'end', gap: 2 }}>
              <div style={{ width: '100%', height: `${Math.max(3, (d.total / maxTotal) * 120)}px`, borderRadius: '6px 6px 0 0', background: 'linear-gradient(180deg, var(--blue), var(--blue-soft))', opacity: 0.78 }} />
            </div>
          ))}
        </div>
        <svg viewBox="0 0 100 100" preserveAspectRatio="none" style={{ position: 'absolute', inset: '8px 8px 0', width: 'calc(100% - 16px)', height: 'calc(100% - 8px)', overflow: 'visible' }}>
          <polyline points={points} fill="none" stroke="var(--green)" strokeWidth="2.5" vectorEffect="non-scaling-stroke" />
        </svg>
      </div>
      <div style={{ display: 'flex', justifyContent: 'space-between', gap: 8, color: 'var(--z-400)', fontSize: 11 }}>
        <span>{data[0]?.date}</span>
        <span>蓝色柱 = 运行数 · 绿色线 = 成功率</span>
        <span>{data[data.length - 1]?.date}</span>
      </div>
    </div>
  );
}

export default function AnalyticsPage() {
  const { id: projectId } = useParams<{ id: string }>();
  const [range, setRange] = useState<RangeValue>('7d');
  const [data, setData] = useState<AnalyticsResponse | null>(null);
  const [loading, setLoading] = useState(false);

  const load = useCallback(async () => {
    if (!projectId) return;
    setLoading(true);
    try {
      setData(await fetchAnalytics(projectId, range));
    } catch {
      Message.error('加载分析数据失败');
      setData(null);
    } finally {
      setLoading(false);
    }
  }, [projectId, range]);

  useEffect(() => { void load(); }, [load]);

  const summary = data?.summary ?? { totalRuns: 0, successRate: 0, medianDurationMs: 0, p95DurationMs: 0 };
  const hasData = summary.totalRuns > 0;
  const rangeLabel = useMemo(() => ({ '7d': '7 天', '30d': '30 天', '90d': '90 天' }[range]), [range]);

  return (
    <>
      <PageHeader
        crumb="Project · Analytics"
        title="流水线分析"
        sub={`过去 ${rangeLabel} 的运行成功率、耗时趋势和失败热点。`}
        actions={<Segmented value={range} options={[{ value: '7d', label: '7d' }, { value: '30d', label: '30d' }, { value: '90d', label: '90d' }]} onChange={(v) => setRange(v as RangeValue)} />}
      />
      <div style={{ padding: 24, display: 'flex', flexDirection: 'column', gap: 16 }}>
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(4, minmax(0, 1fr))', gap: 14 }}>
          <Metric label="TOTAL RUNS" value={loading ? '...' : summary.totalRuns} icon={<IGrid size={14} />} iconBg="var(--z-100)" iconColor="var(--z-700)" />
          <Metric label="SUCCESS RATE" value={loading ? '...' : pct(summary.successRate)} icon={<ITarget size={14} />} iconBg="var(--green-soft)" iconColor="var(--green-ink)" />
          <Metric label="MEDIAN DURATION" value={loading ? '...' : duration(summary.medianDurationMs)} icon={<IClock size={14} />} iconBg="var(--amber-soft)" iconColor="var(--amber-ink)" />
          <Metric label="P95 DURATION" value={loading ? '...' : duration(summary.p95DurationMs)} icon={<IZap size={14} />} iconBg="var(--blue-soft)" iconColor="var(--blue-ink)" />
        </div>

        {!loading && !hasData ? (
          <Card><div style={{ padding: '48px 0', textAlign: 'center', color: 'var(--z-400)' }}>暂无运行数据。触发几次流水线后，这里会显示趋势与失败热点。</div></Card>
        ) : (
          <>
            <Card title="Daily Trend"><DailyTrendChart data={data?.dailyStats ?? []} /></Card>
            <div style={{ display: 'grid', gridTemplateColumns: '1.2fr 1fr', gap: 16 }}>
              <Card title="Top Failing Steps" padding={false}>
                <table className="table"><thead><tr><th>Step</th><th>Stage</th><th>Failures</th><th>Rate</th></tr></thead><tbody>
                  {(data?.topFailingSteps ?? []).map((row) => (
                    <tr key={`${row.taskRunName}-${row.stepName}`}><td>{row.stepName}</td><td><span className="mono sub">{row.taskRunName}</span></td><td>{row.failureCount} / {row.totalCount}</td><td><Badge tone={row.failureRate > 0.5 ? 'red' : 'amber'}>{pct(row.failureRate)}</Badge></td></tr>
                  ))}
                  {(data?.topFailingSteps ?? []).length === 0 && <tr><td colSpan={4} style={{ padding: '32px 0', textAlign: 'center', color: 'var(--z-400)' }}>暂无失败步骤</td></tr>}
                </tbody></table>
              </Card>
              <Card title="Most Triggered Pipelines" padding={false}>
                <table className="table"><thead><tr><th>Pipeline</th><th>Runs</th><th>Success</th></tr></thead><tbody>
                  {(data?.topPipelines ?? []).map((row) => (
                    <tr key={row.pipelineId}><td>{row.pipelineName}</td><td>{row.runCount}</td><td><Badge tone={row.successRate >= 0.8 ? 'green' : 'amber'}>{pct(row.successRate)}</Badge></td></tr>
                  ))}
                  {(data?.topPipelines ?? []).length === 0 && <tr><td colSpan={3} style={{ padding: '32px 0', textAlign: 'center', color: 'var(--z-400)' }}>暂无流水线统计</td></tr>}
                </tbody></table>
              </Card>
            </div>
          </>
        )}
      </div>
    </>
  );
}
