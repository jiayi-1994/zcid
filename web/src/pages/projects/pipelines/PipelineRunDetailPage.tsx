import { useCallback, useEffect, useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { Message } from '@arco-design/web-react';
import {
  fetchPipelineRun, cancelPipelineRun, fetchRunArtifacts, fetchArchivedLogs, fetchStepExecutions,
  type PipelineRun, type Artifact, type LogEntry, type ArchivedLogsResponse, type StepExecution,
} from '../../../services/pipelineRun';
import { PageHeader } from '../../../components/ui/PageHeader';
import { Card } from '../../../components/ui/Card';
import { Metric } from '../../../components/ui/Metric';
import { Btn } from '../../../components/ui/Btn';
import { StatusBadge } from '../../../components/ui/StatusBadge';
import { StepsWaterfall } from '../../../components/pipeline/StepsWaterfall';
import { IArrL, IPlay, IClock, IUser, ICode } from '../../../components/ui/icons';

function formatDuration(start?: string, end?: string): string {
  if (!start) return '-';
  const diff = Math.floor(((end ? new Date(end).getTime() : Date.now()) - new Date(start).getTime()) / 1000);
  if (diff < 60) return `${diff}s`;
  const min = Math.floor(diff / 60);
  return min < 60 ? `${min}m ${diff % 60}s` : `${Math.floor(min / 60)}h ${min % 60}m`;
}


export default function PipelineRunDetailPage() {
  const { id: projectId, pipelineId, runId } = useParams<{ id: string; pipelineId: string; runId: string }>();
  const navigate = useNavigate();
  const [run, setRun] = useState<PipelineRun | null>(null);
  const [artifacts, setArtifacts] = useState<Artifact[]>([]);
  const [logs, setLogs] = useState<LogEntry[]>([]);
  const [logsTotal, setLogsTotal] = useState(0);
  const [stepExecutions, setStepExecutions] = useState<StepExecution[]>([]);
  const [stepError, setStepError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  const loadData = useCallback(async () => {
    if (!projectId || !pipelineId || !runId) return;
    setLoading(true);
    try {
      const [runData, artData, stepData, logsData] = await Promise.all([
        fetchPipelineRun(projectId, pipelineId, runId),
        fetchRunArtifacts(projectId, pipelineId, runId).catch(() => []),
        fetchStepExecutions(projectId, pipelineId, runId).catch(() => {
          setStepError('Step execution data is temporarily unavailable.');
          return [] as StepExecution[];
        }),
        fetchArchivedLogs(projectId, runId, 1, 100).catch(
          (): ArchivedLogsResponse => ({ items: [], total: 0, page: 1, pageSize: 50 })
        ),
      ]);
      setRun(runData);
      setArtifacts(artData);
      setStepExecutions(stepData);
      setStepError(null);
      setLogs(logsData.items ?? []);
      setLogsTotal(logsData.total ?? 0);
    } catch {
      Message.error('加载运行详情失败');
    } finally {
      setLoading(false);
    }
  }, [projectId, pipelineId, runId]);

  useEffect(() => { loadData(); }, [loadData]);

  useEffect(() => {
    if (!projectId || !pipelineId || !runId || !run || !['pending', 'queued', 'running'].includes(run.status)) return;
    const timer = window.setInterval(() => {
      Promise.all([
        fetchPipelineRun(projectId, pipelineId, runId),
        fetchStepExecutions(projectId, pipelineId, runId),
      ])
        .then(([nextRun, items]) => { setRun(nextRun); setStepExecutions(items); setStepError(null); })
        .catch(() => setStepError('Step execution data is temporarily unavailable.'));
    }, 3000);
    return () => window.clearInterval(timer);
  }, [projectId, pipelineId, runId, run?.status]);

  const handleCancel = async () => {
    if (!projectId || !pipelineId || !runId) return;
    try { await cancelPipelineRun(projectId, pipelineId, runId); Message.success('已取消'); loadData(); }
    catch { Message.error('取消失败'); }
  };

  if (loading && !run) {
    return (
      <>
        <PageHeader crumb="Build Observation" title="运行详情" />
        <div style={{ padding: '40px 24px', color: 'var(--z-400)' }}>加载中...</div>
      </>
    );
  }

  if (!run) {
    return (
      <>
        <PageHeader crumb="Build Observation" title="运行详情" />
        <div style={{ padding: '40px 24px', color: 'var(--z-400)' }}>运行记录不存在</div>
      </>
    );
  }

  const canCancel = ['pending', 'queued', 'running'].includes(run.status);
  const duration = formatDuration(run.startedAt, run.finishedAt);


  return (
    <>
      <PageHeader
        crumb="Build Observation"
        title={`Build #${run.runNumber}`}
        sub={`${run.triggeredBy?.replace('admin-bootstrap-', 'admin#') ?? 'unknown'} · ${duration}`}
        actions={
          <>
            <Btn size="sm" icon={<IArrL size={13} />} onClick={() => navigate(`/projects/${projectId}/pipelines/${pipelineId}/runs`)}>返回</Btn>
            {canCancel && <Btn size="sm" onClick={handleCancel}>取消运行</Btn>}
          </>
        }
      />
      <div style={{ padding: '24px 24px 48px', display: 'flex', flexDirection: 'column', gap: 18, maxWidth: 960 }}>
        {/* Info metrics */}
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(4,1fr)', gap: 14 }}>
          <Metric label="TRIGGER" value={run.triggerType === 'manual' ? '手动' : run.triggerType} icon={<IPlay size={14} />} iconBg="var(--z-100)" iconColor="var(--z-700)" />
          <Metric label="BRANCH" value={run.gitBranch ?? '-'} icon={<ICode size={14} />} iconBg="var(--blue-soft)" iconColor="var(--blue-ink)" />
          <Metric label="TRIGGERED BY" value={run.triggeredBy?.replace('admin-bootstrap-', 'admin#') ?? '-'} icon={<IUser size={14} />} iconBg="var(--z-100)" iconColor="var(--z-700)" />
          <Metric label="DURATION" value={duration} icon={<IClock size={14} />} iconBg="var(--amber-soft)" iconColor="var(--amber-ink)" />
        </div>

        {/* Details */}
        <Card title="Run Details" extra={<StatusBadge status={run.status} />}>
          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', rowGap: 14, columnGap: 24 }}>
            {[
              { label: 'Tekton Name', value: <span className="mono" style={{ fontSize: 11.5 }}>{run.tektonName ?? '-'}</span> },
              { label: 'Namespace', value: <span className="mono" style={{ fontSize: 11.5 }}>{run.namespace ?? '-'}</span> },
              { label: '开始时间', value: <span className="mono sub" style={{ fontSize: 11.5 }}>{run.startedAt ? new Date(run.startedAt).toLocaleString() : '尚未开始'}</span> },
              { label: '结束时间', value: <span className="mono sub" style={{ fontSize: 11.5 }}>{run.finishedAt ? new Date(run.finishedAt).toLocaleString() : '进行中'}</span> },
              { label: 'Commit', value: <span className="code" style={{ fontSize: 11 }}>{run.gitCommit ? run.gitCommit.substring(0, 8) : '-'}</span> },
            ].map(({ label, value }) => (
              <div key={label}>
                <div style={{ fontSize: 11, color: 'var(--z-400)', fontWeight: 500, textTransform: 'uppercase', letterSpacing: '0.05em', marginBottom: 3 }}>{label}</div>
                <div>{value}</div>
              </div>
            ))}
            {run.params && Object.keys(run.params).length > 0 && (
              <div style={{ gridColumn: '1 / -1' }}>
                <div style={{ fontSize: 11, color: 'var(--z-400)', fontWeight: 500, textTransform: 'uppercase', letterSpacing: '0.05em', marginBottom: 6 }}>运行参数</div>
                <div style={{ display: 'flex', gap: 6, flexWrap: 'wrap' }}>
                  {Object.entries(run.params).map(([k, v]) => (
                    <span key={k} className="code" style={{ fontSize: 11 }}>{k}={v}</span>
                  ))}
                </div>
              </div>
            )}
            {run.errorMessage && (
              <div style={{ gridColumn: '1 / -1', padding: '10px 12px', background: 'var(--red-soft)', borderRadius: 6, color: 'var(--red-ink)', fontSize: 12, fontFamily: 'var(--font-mono)' }}>
                {run.errorMessage}
              </div>
            )}
          </div>
        </Card>

        {/* Artifacts */}
        {artifacts.length > 0 && (
          <Card title="Build Artifacts">
            <div style={{ display: 'flex', gap: 8, flexWrap: 'wrap' }}>
              {artifacts.map((a) => (
                <a key={a.name} href={a.url} target="_blank" rel="noopener noreferrer" style={{ textDecoration: 'none' }}>
                  <span className="code" style={{ fontSize: 12, cursor: 'pointer' }}>
                    {a.name}{a.size != null ? ` (${(a.size / 1024).toFixed(1)} KB)` : ''}
                  </span>
                </a>
              ))}
            </div>
          </Card>
        )}

        {/* Step execution waterfall */}
        <Card title="Step Execution Waterfall">
          <StepsWaterfall items={stepExecutions} error={stepError} />
        </Card>

        {/* Build log terminal */}
        <Card title="Build Output" padding={false}>
          <div style={{
            background: '#0d1117', borderRadius: '0 0 8px 8px',
            minHeight: 240, maxHeight: 480, overflow: 'auto',
            fontFamily: 'var(--font-mono)', fontSize: 12, lineHeight: 1.7,
          }}>
            {logsTotal > 0 ? logs.map((l, i) => (
              <div key={i} style={{ display: 'flex', gap: 16, padding: '1px 16px' }}>
                <span style={{ color: '#484f58', userSelect: 'none', minWidth: 32, textAlign: 'right' }}>{i + 1}</span>
                <span style={{ color: l.level === 'error' ? '#f85149' : l.level === 'info' ? '#79c0ff' : '#e6edf3' }}>
                  [{l.level.toUpperCase()}] {l.content}
                </span>
              </div>
            )) : (
              <div style={{ padding: '40px 0', textAlign: 'center', color: '#484f58' }}>
                <div style={{ fontSize: 28, marginBottom: 8 }}>{'>'}_</div>
                <div style={{ color: '#8b949e' }}>Waiting for build output...</div>
              </div>
            )}
          </div>
        </Card>

      </div>
    </>
  );
}
