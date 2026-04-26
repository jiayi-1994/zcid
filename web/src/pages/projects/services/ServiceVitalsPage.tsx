import { useEffect, useState } from 'react';
import { Message } from '@arco-design/web-react';
import { Link, useParams } from 'react-router-dom';
import { fetchServiceVitals, type ServiceVitals } from '../../../services/project';
import { extractErrorMessage } from '../../../services/http';
import { PageHeader } from '../../../components/ui/PageHeader';
import { Card } from '../../../components/ui/Card';
import { Metric } from '../../../components/ui/Metric';
import { Badge } from '../../../components/ui/Badge';
import { Btn } from '../../../components/ui/Btn';
import { IServer, IZap, IRocket, IAlert, IChevR } from '../../../components/ui/icons';

const HEALTH_TONE: Record<string, 'green' | 'amber' | 'red' | 'grey'> = {
  healthy: 'green',
  warning: 'amber',
  degraded: 'red',
  stale: 'amber',
  unknown: 'grey',
};

const HEALTH_LABEL: Record<string, string> = {
  healthy: 'Healthy',
  warning: 'Warning',
  degraded: 'Degraded',
  stale: 'Stale',
  unknown: 'Unknown',
};

function fmt(value?: string) {
  if (!value) return '-';
  const date = new Date(value);
  return Number.isNaN(date.getTime()) ? value : date.toLocaleString();
}

function statusTone(status: string): 'green' | 'amber' | 'red' | 'grey' {
  if (status === 'succeeded' || status === 'healthy' || status === 'active') return 'green';
  if (status === 'failed' || status === 'degraded') return 'red';
  if (status === 'running' || status === 'queued' || status === 'pending' || status === 'syncing' || status === 'warning') return 'amber';
  return 'grey';
}

function ServiceVitalsPage() {
  const { id: projectId, serviceId } = useParams<{ id: string; serviceId: string }>();
  const [data, setData] = useState<ServiceVitals | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!projectId || !serviceId) return;
    setLoading(true);
    setError(null);
    fetchServiceVitals(projectId, serviceId)
      .then((payload) => {
        setData(payload);
        setError(null);
      })
      .catch((err: unknown) => {
        const message = extractErrorMessage(err, '加载服务健康视图失败');
        setData(null);
        setError(message);
        Message.error(message);
      })
      .finally(() => setLoading(false));
  }, [projectId, serviceId]);

  if (loading) {
    return (
      <>
        <PageHeader crumb="Project · Service Vitals" title="Service Vitals" sub="加载服务健康和交付证据..." />
        <div style={{ padding: 24, color: 'var(--z-500)' }}>加载中...</div>
      </>
    );
  }

  if (error || !data) {
    return (
      <>
        <PageHeader crumb="Project · Service Vitals" title="Service Vitals" sub="加载服务健康和交付证据..." />
        <div style={{ padding: 24 }}>
          <Card>
            <div role="alert" style={{ color: 'var(--red-ink)', fontWeight: 600 }}>加载失败：{error ?? '服务健康数据不可用'}</div>
          </Card>
        </div>
      </>
    );
  }

  const health = data.summary.status || 'unknown';
  const service = data.service;

  return (
    <>
      <PageHeader
        crumb="Project · Service Vitals"
        title={service.name}
        sub="服务健康、交付证据和信号矩阵。"
        actions={<Btn size="sm" onClick={() => history.back()}>返回</Btn>}
      />
      <div style={{ padding: '24px 24px 48px', display: 'flex', flexDirection: 'column', gap: 16 }}>
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(4, minmax(0, 1fr))', gap: 14 }}>
          <Metric label="HEALTH" value={HEALTH_LABEL[health] ?? health} icon={<IServer size={14} />} iconBg="var(--z-100)" iconColor="var(--z-700)" trend={data.summary.reason} trendTone={health === 'healthy' ? 'green' : health === 'degraded' ? 'red' : 'amber'} />
          <Metric label="PIPELINES" value={data.linkedPipelines.length} icon={<IZap size={14} />} iconBg="var(--blue-soft)" iconColor="var(--blue-ink)" />
          <Metric label="DEPLOYMENTS" value={data.latestDeployments.length} icon={<IRocket size={14} />} iconBg="var(--green-soft)" iconColor="var(--green-ink)" />
          <Metric label="WARNINGS" value={data.summary.activeWarningCount} icon={<IAlert size={14} />} iconBg="var(--amber-soft)" iconColor="var(--amber-ink)" />
        </div>

        <Card>
          <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', gap: 16 }}>
            <div>
              <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 6 }}>
                <Badge tone={HEALTH_TONE[health] ?? 'grey'} dot>{HEALTH_LABEL[health] ?? health}</Badge>
                {data.summary.lastSignalAt && <span className="sub">last signal: {fmt(data.summary.lastSignalAt)}</span>}
              </div>
              <div style={{ color: 'var(--z-700)', fontSize: 13 }}>{data.summary.reason}</div>
            </div>
            <div style={{ display: 'flex', gap: 6, flexWrap: 'wrap', justifyContent: 'flex-end' }}>
              {service.serviceType && <span className="badge badge--blue">{service.serviceType}</span>}
              {service.language && <span className="badge badge--cyan">{service.language}</span>}
              {service.owner && <span className="badge badge--grey">owner: {service.owner}</span>}
              {(service.tags ?? []).map((tag) => <span key={tag} className="badge badge--grey">{tag}</span>)}
            </div>
          </div>
        </Card>

        {data.emptyStates.length > 0 && (
          <Card>
            <div style={{ fontWeight: 600, marginBottom: 8 }}>Missing evidence</div>
            <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
              {data.emptyStates.map((msg) => <div key={msg} className="sub">• {msg}</div>)}
            </div>
          </Card>
        )}

        {data.warnings.length > 0 && (
          <Card padding={false}>
            <div style={{ padding: '14px 16px', borderBottom: '1px solid var(--z-150)', fontWeight: 600 }}>Active warnings</div>
            <table className="table">
              <thead><tr><th>Step</th><th>Pipeline</th><th>Status</th><th>Run</th></tr></thead>
              <tbody>
                {data.warnings.map((warning) => (
                  <tr key={`${warning.runId}-${warning.taskRunName}-${warning.stepName}`}>
                    <td><span style={{ fontWeight: 500 }}>{warning.taskRunName} / {warning.stepName}</span></td>
                    <td>{warning.pipelineName}</td>
                    <td><Badge tone={statusTone(warning.status)} dot>{warning.status}</Badge></td>
                    <td><Link to={warning.runPath}>#{warning.runNumber} <IChevR size={11} /></Link></td>
                  </tr>
                ))}
              </tbody>
            </table>
          </Card>
        )}

        <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 14 }}>
          <Card padding={false}>
            <div style={{ padding: '14px 16px', borderBottom: '1px solid var(--z-150)', fontWeight: 600 }}>Linked pipelines</div>
            <table className="table">
              <thead><tr><th>Name</th><th>Status</th><th>Repo</th></tr></thead>
              <tbody>
                {data.linkedPipelines.map((p) => (
                  <tr key={p.id}><td>{p.name}</td><td><Badge tone={statusTone(p.status)} dot>{p.status}</Badge></td><td><span className="sub">{p.repoUrl || '-'}</span></td></tr>
                ))}
                {data.linkedPipelines.length === 0 && <tr><td colSpan={3} className="sub" style={{ textAlign: 'center', padding: 24 }}>No linked pipelines</td></tr>}
              </tbody>
            </table>
          </Card>
          <Card padding={false}>
            <div style={{ padding: '14px 16px', borderBottom: '1px solid var(--z-150)', fontWeight: 600 }}>Latest deployments</div>
            <table className="table">
              <thead><tr><th>Environment</th><th>Status</th><th>Image</th></tr></thead>
              <tbody>
                {data.latestDeployments.map((d) => (
                  <tr key={d.id}><td>{d.environmentName}</td><td><Badge tone={statusTone(d.status)} dot>{d.status}</Badge></td><td><span className="sub">{d.image}</span></td></tr>
                ))}
                {data.latestDeployments.length === 0 && <tr><td colSpan={3} className="sub" style={{ textAlign: 'center', padding: 24 }}>No deployments</td></tr>}
              </tbody>
            </table>
          </Card>
        </div>

        <Card padding={false}>
          <div style={{ padding: '14px 16px', borderBottom: '1px solid var(--z-150)', fontWeight: 600 }}>Recent runs</div>
          <table className="table">
            <thead><tr><th>Run</th><th>Pipeline</th><th>Status</th><th>Created</th></tr></thead>
            <tbody>
              {data.recentRuns.map((run) => (
                <tr key={run.id}><td>#{run.runNumber}</td><td>{run.pipelineName}</td><td><Badge tone={statusTone(run.status)} dot>{run.status}</Badge></td><td className="sub">{fmt(run.createdAt)}</td></tr>
              ))}
              {data.recentRuns.length === 0 && <tr><td colSpan={4} className="sub" style={{ textAlign: 'center', padding: 24 }}>No recent runs</td></tr>}
            </tbody>
          </table>
        </Card>

        <Card padding={false}>
          <div style={{ padding: '14px 16px', borderBottom: '1px solid var(--z-150)', fontWeight: 600 }}>Signal matrix</div>
          <table className="table">
            <thead><tr><th>Target</th><th>Source</th><th>Status</th><th>Reason</th><th>Observed</th></tr></thead>
            <tbody>
              {data.activeSignals.map((sig) => (
                <tr key={sig.id}>
                  <td>{sig.targetType}:{sig.targetId}</td>
                  <td>{sig.source}</td>
                  <td><Badge tone={HEALTH_TONE[sig.status] ?? 'grey'} dot>{sig.status}</Badge></td>
                  <td><span className="sub">{sig.message || sig.reason || '-'}</span></td>
                  <td className="sub">{fmt(sig.observedAt)}</td>
                </tr>
              ))}
              {data.activeSignals.length === 0 && <tr><td colSpan={5} className="sub" style={{ textAlign: 'center', padding: 24 }}>No health signals</td></tr>}
            </tbody>
          </table>
        </Card>
      </div>
    </>
  );
}

export default ServiceVitalsPage;
