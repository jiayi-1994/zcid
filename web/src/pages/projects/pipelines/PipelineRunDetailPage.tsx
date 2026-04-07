import { useEffect, useState, useCallback } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import {
  Skeleton,
  Message,
  Typography,
  Button,
  Space,
  Tag,
  Popconfirm,
  Grid,
} from '@arco-design/web-react';
import { IconArrowLeft, IconClose, IconBranch, IconCode, IconClockCircle, IconUser } from '@arco-design/web-react/icon';
import {
  fetchPipelineRun,
  cancelPipelineRun,
  fetchRunArtifacts,
  fetchArchivedLogs,
  type PipelineRun,
  type Artifact,
  type LogEntry,
  type ArchivedLogsResponse,
} from '../../../services/pipelineRun';

const { Text } = Typography;
const { Row, Col } = Grid;

const statusConfig: Record<string, { color: string; label: string; badgeClass: string; dotClass: string; stageClass: string }> = {
  pending:   { color: '#6B7280', label: '待执行',   badgeClass: 'pipeline-status-badge--pending',   dotClass: 'stage-progress-dot--pending',   stageClass: 'stage-progress-item--pending' },
  queued:    { color: '#0066FF', label: '排队中',   badgeClass: 'pipeline-status-badge--running',   dotClass: 'stage-progress-dot--active',    stageClass: 'stage-progress-item--active' },
  running:   { color: '#0066FF', label: '运行中',   badgeClass: 'pipeline-status-badge--running',   dotClass: 'stage-progress-dot--active',    stageClass: 'stage-progress-item--active' },
  succeeded: { color: '#00C853', label: '成功',     badgeClass: 'pipeline-status-badge--success',   dotClass: 'stage-progress-dot--completed', stageClass: 'stage-progress-item--completed' },
  failed:    { color: '#FF3B30', label: '失败',     badgeClass: 'pipeline-status-badge--failed',    dotClass: 'stage-progress-dot--completed', stageClass: 'stage-progress-item--completed' },
  cancelled: { color: '#FF9500', label: '已取消',   badgeClass: 'pipeline-status-badge--cancelled', dotClass: 'stage-progress-dot--pending',   stageClass: 'stage-progress-item--pending' },
};

function formatDuration(start?: string, end?: string): string {
  if (!start) return '-';
  const s = new Date(start).getTime();
  const e = end ? new Date(end).getTime() : Date.now();
  const diff = Math.floor((e - s) / 1000);
  if (diff < 60) return `${diff}s`;
  const min = Math.floor(diff / 60);
  const sec = diff % 60;
  return `${min}m ${sec}s`;
}

export default function PipelineRunDetailPage() {
  const { id: projectId, pipelineId, runId } = useParams<{ id: string; pipelineId: string; runId: string }>();
  const navigate = useNavigate();
  const [run, setRun] = useState<PipelineRun | null>(null);
  const [artifacts, setArtifacts] = useState<Artifact[]>([]);
  const [logs, setLogs] = useState<LogEntry[]>([]);
  const [logsTotal, setLogsTotal] = useState(0);
  const [loading, setLoading] = useState(false);

  const loadData = useCallback(async () => {
    if (!projectId || !pipelineId || !runId) return;
    setLoading(true);
    try {
      const [runData, artData, logsData] = await Promise.all([
        fetchPipelineRun(projectId, pipelineId, runId),
        fetchRunArtifacts(projectId, pipelineId, runId).catch(() => []),
        fetchArchivedLogs(projectId, runId, 1, 100).catch(
          (): ArchivedLogsResponse => ({ items: [], total: 0, page: 1, pageSize: 50 })
        ),
      ]);
      setRun(runData);
      setArtifacts(artData);
      setLogs(logsData.items ?? []);
      setLogsTotal(logsData.total ?? 0);
    } catch {
      Message.error('加载运行详情失败');
    } finally {
      setLoading(false);
    }
  }, [projectId, pipelineId, runId]);

  useEffect(() => { loadData(); }, [loadData]);

  const handleCancel = async () => {
    if (!projectId || !pipelineId || !runId) return;
    try {
      await cancelPipelineRun(projectId, pipelineId, runId);
      Message.success('已取消');
      loadData();
    } catch {
      Message.error('取消失败');
    }
  };

  const canCancel = run && ['pending', 'queued', 'running'].includes(run.status);

  if (loading && !run) {
    return (
      <div className="page-container">
        <Skeleton text={{ rows: 6 }} animation />
      </div>
    );
  }

  if (!run) {
    return (
      <div className="page-container">
        <Message type="error">运行记录不存在</Message>
      </div>
    );
  }

  const cfg = statusConfig[run.status] || statusConfig.pending;

  const mockStages = [
    { name: 'Source Checkout', status: run.status === 'succeeded' || run.status === 'failed' ? 'completed' : run.status === 'running' ? 'completed' : 'pending' },
    { name: 'Build Artifacts', status: run.status === 'succeeded' || run.status === 'failed' ? 'completed' : run.status === 'running' ? 'active' : 'pending' },
    { name: 'Image Push', status: run.status === 'succeeded' ? 'completed' : 'pending' },
    { name: 'K8s Deployment', status: run.status === 'succeeded' ? 'completed' : 'pending' },
  ];

  return (
    <div className="page-container" style={{ maxWidth: 1000 }}>
      {/* Header */}
      <div style={{
        display: 'flex', justifyContent: 'space-between', alignItems: 'center',
        marginBottom: 24,
      }}>
        <Space size={16} align="center">
          <Button
            type="text"
            icon={<IconArrowLeft />}
            onClick={() => navigate(`/projects/${projectId}/pipelines/${pipelineId}/runs`)}
            style={{ borderRadius: 8, color: 'var(--muted-foreground)' }}
          />
          <div>
            <h3 style={{ margin: 0, fontSize: 22, fontWeight: 700, color: 'var(--foreground)' }}>
              Build #{run.runNumber}
            </h3>
            <span style={{ fontSize: 13, color: 'var(--muted-foreground)' }}>
              {run.triggeredBy?.replace('admin-bootstrap-', 'admin#') ?? 'unknown'}
              {' · '}
              {formatDuration(run.startedAt, run.finishedAt)}
            </span>
          </div>
          <span className={`pipeline-status-badge ${cfg.badgeClass}`}>
            <span style={{ width: 6, height: 6, borderRadius: '50%', background: cfg.color }} />
            {cfg.label}
          </span>
        </Space>
        {canCancel && (
          <Popconfirm title="确定取消此运行？" onOk={handleCancel}>
            <Button type="outline" status="danger" icon={<IconClose />} style={{ borderRadius: 8 }}>
              取消运行
            </Button>
          </Popconfirm>
        )}
      </div>

      {/* Stage Progress */}
      <div className="stage-progress-bar">
        {mockStages.map((stage) => (
          <div key={stage.name} className={`stage-progress-item stage-progress-item--${stage.status}`}>
            <div className={`stage-progress-dot stage-progress-dot--${stage.status}`} />
            <div className="stage-progress-label">{stage.name}</div>
            <div className="stage-progress-status">
              {stage.status === 'completed' ? 'Completed' : stage.status === 'active' ? 'Active' : 'Pending'}
            </div>
          </div>
        ))}
      </div>

      {/* Info Cards */}
      <Row gutter={12} style={{ marginBottom: 20 }}>
        <Col span={6}>
          <div className="metric-card" style={{ padding: 16 }}>
            <span className="metric-card-label">TRIGGER</span>
            <span style={{ fontSize: 14, fontWeight: 600, color: 'var(--foreground)' }}>
              {run.triggerType === 'manual' ? '手动' : run.triggerType}
            </span>
          </div>
        </Col>
        <Col span={6}>
          <div className="metric-card" style={{ padding: 16 }}>
            <span className="metric-card-label">
              <IconBranch style={{ fontSize: 12, marginRight: 4 }} />BRANCH
            </span>
            <span style={{ fontSize: 14, fontWeight: 600, color: 'var(--foreground)' }}>
              {run.gitBranch ?? '-'}
            </span>
          </div>
        </Col>
        <Col span={6}>
          <div className="metric-card" style={{ padding: 16 }}>
            <span className="metric-card-label">
              <IconUser style={{ fontSize: 12, marginRight: 4 }} />TRIGGERED BY
            </span>
            <span style={{ fontSize: 14, fontWeight: 600, color: 'var(--foreground)' }}>
              {run.triggeredBy?.replace('admin-bootstrap-', 'admin#') ?? '-'}
            </span>
          </div>
        </Col>
        <Col span={6}>
          <div className="metric-card" style={{ padding: 16 }}>
            <span className="metric-card-label">
              <IconCode style={{ fontSize: 12, marginRight: 4 }} />COMMIT
            </span>
            <span style={{ fontSize: 14, fontWeight: 600, color: 'var(--foreground)', fontFamily: 'var(--font-mono)' }}>
              {run.gitCommit ? run.gitCommit.substring(0, 8) : '-'}
            </span>
          </div>
        </Col>
      </Row>

      {/* Detail Info */}
      <div className="config-panel" style={{ marginBottom: 20 }}>
        <div className="config-panel-header">Run Details</div>
        <div className="config-panel-body">
          <Row gutter={[24, 16]}>
            <Col span={12}>
              <Text type="secondary" style={{ fontSize: 12, display: 'block', marginBottom: 4 }}>
                <IconClockCircle style={{ fontSize: 12, marginRight: 4 }} />开始时间
              </Text>
              <Text style={{ fontSize: 14, fontWeight: 500 }}>
                {run.startedAt ? new Date(run.startedAt).toLocaleString() : '尚未开始'}
              </Text>
            </Col>
            <Col span={12}>
              <Text type="secondary" style={{ fontSize: 12, display: 'block', marginBottom: 4 }}>
                <IconClockCircle style={{ fontSize: 12, marginRight: 4 }} />结束时间
              </Text>
              <Text style={{ fontSize: 14, fontWeight: 500 }}>
                {run.finishedAt ? new Date(run.finishedAt).toLocaleString() : '进行中'}
              </Text>
            </Col>
            <Col span={12}>
              <Text type="secondary" style={{ fontSize: 12, display: 'block', marginBottom: 4 }}>Tekton Name</Text>
              <Text style={{ fontSize: 14, fontWeight: 500, fontFamily: 'var(--font-mono)' }}>
                {run.tektonName ?? '-'}
              </Text>
            </Col>
            <Col span={12}>
              <Text type="secondary" style={{ fontSize: 12, display: 'block', marginBottom: 4 }}>Namespace</Text>
              <Text style={{ fontSize: 14, fontWeight: 500, fontFamily: 'var(--font-mono)' }}>
                {run.namespace ?? '-'}
              </Text>
            </Col>
            {run.params && Object.keys(run.params).length > 0 && (
              <Col span={24}>
                <Text type="secondary" style={{ fontSize: 12, display: 'block', marginBottom: 8 }}>运行参数</Text>
                <div style={{ display: 'flex', gap: 8, flexWrap: 'wrap' }}>
                  {Object.entries(run.params).map(([k, v]) => (
                    <Tag key={k} style={{ borderRadius: 6, background: 'var(--primary-light)', color: 'var(--primary)', border: 'none' }}>
                      {k}={v}
                    </Tag>
                  ))}
                </div>
              </Col>
            )}
            {run.errorMessage && (
              <Col span={24}>
                <div style={{
                  padding: '12px 16px', background: '#FFF1F0', borderRadius: 8,
                  color: 'var(--destructive)', fontSize: 13, border: '1px solid #FFD6D6',
                }}>
                  {run.errorMessage}
                </div>
              </Col>
            )}
          </Row>
        </div>
      </div>

      {/* Build Logs - Dark Terminal */}
      <div className="build-log-terminal" style={{ marginBottom: 20 }}>
        <div className="build-log-header">
          <span className="build-log-title">Build Output</span>
          <Space size={8}>
            <div style={{ width: 10, height: 10, borderRadius: '50%', background: '#FF5F57' }} />
            <div style={{ width: 10, height: 10, borderRadius: '50%', background: '#FFBD2E' }} />
            <div style={{ width: 10, height: 10, borderRadius: '50%', background: '#28C840' }} />
          </Space>
        </div>
        <div className="build-log-body">
          {logsTotal > 0 ? (
            logs.map((l, i) => {
              const levelClass = l.level === 'error' ? 'build-log-content--error'
                : l.level === 'info' ? 'build-log-content--info'
                : '';
              return (
                <div key={i} className="build-log-line">
                  <span className="build-log-linenum">{i + 1}</span>
                  <span className={`build-log-content ${levelClass}`}>
                    [{l.level.toUpperCase()}] {l.content}
                  </span>
                </div>
              );
            })
          ) : (
            <div style={{ padding: '40px 0', textAlign: 'center', color: '#64748B' }}>
              <div style={{ fontSize: 32, marginBottom: 8 }}>{'>'}_</div>
              <div>Waiting for build output...</div>
              <div style={{ fontSize: 12, marginTop: 4, color: '#475569' }}>暂无归档日志</div>
            </div>
          )}
        </div>
      </div>

      {/* Artifacts */}
      {artifacts.length > 0 && (
        <div className="config-panel">
          <div className="config-panel-header">Build Artifacts</div>
          <div className="config-panel-body">
            <Space wrap size={12}>
              {artifacts.map((a) => (
                <a key={a.name} href={a.url} target="_blank" rel="noopener noreferrer">
                  <Tag style={{
                    borderRadius: 8, cursor: 'pointer', padding: '6px 14px',
                    background: 'var(--primary-light)', color: 'var(--primary)',
                    border: '1px solid var(--primary)', fontWeight: 500,
                  }}>
                    {a.name}
                    {a.size != null ? ` (${(a.size / 1024).toFixed(1)} KB)` : ''}
                  </Tag>
                </a>
              ))}
            </Space>
          </div>
        </div>
      )}
    </div>
  );
}
