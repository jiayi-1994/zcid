import { useEffect, useState, useCallback } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import {
  Skeleton,
  Message,
  Typography,
  Button,
  Space,
  Tag,
  Card,
  Link,
  Popconfirm,
  Grid,
  Divider,
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

const { Paragraph, Text } = Typography;
const { Row, Col } = Grid;

const statusConfig: Record<string, { color: string; label: string; bg: string }> = {
  pending:   { color: '#86909C', label: '待执行', bg: '#F2F3F5' },
  queued:    { color: '#165DFF', label: '排队中', bg: '#E8F3FF' },
  running:   { color: '#0FC6C2', label: '运行中', bg: '#E8FFFB' },
  succeeded: { color: '#00B42A', label: '成功',   bg: '#E8FFEA' },
  failed:    { color: '#F53F3F', label: '失败',   bg: '#FFECE8' },
  cancelled: { color: '#FF7D00', label: '已取消', bg: '#FFF7E8' },
};

function StatusBadge({ status }: { status: string }) {
  const cfg = statusConfig[status] || { color: '#86909C', label: status, bg: '#F2F3F5' };
  return (
    <span style={{
      display: 'inline-flex', alignItems: 'center', gap: 6,
      padding: '4px 12px', borderRadius: 20,
      background: cfg.bg, color: cfg.color,
      fontWeight: 600, fontSize: 13,
    }}>
      <span style={{ width: 8, height: 8, borderRadius: '50%', background: cfg.color, display: 'inline-block' }} />
      {cfg.label}
    </span>
  );
}

function InfoItem({ label, value, icon }: { label: string; value: React.ReactNode; icon?: React.ReactNode }) {
  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 4 }}>
      <Text type="secondary" style={{ fontSize: 12 }}>{icon && <span style={{ marginRight: 4 }}>{icon}</span>}{label}</Text>
      <Text style={{ fontSize: 14, fontWeight: 500 }}>{value || '-'}</Text>
    </div>
  );
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

  return (
    <div className="page-container" style={{ maxWidth: 960 }}>
      {/* Header */}
      <div style={{
        display: 'flex', justifyContent: 'space-between', alignItems: 'center',
        marginBottom: 24, padding: '16px 20px',
        background: 'var(--color-bg-2)', borderRadius: 12,
        border: '1px solid var(--color-border)',
      }}>
        <Space size={16} align="center">
          <Button
            type="text"
            icon={<IconArrowLeft />}
            onClick={() => navigate(`/projects/${projectId}/pipelines/${pipelineId}/runs`)}
            style={{ borderRadius: 8 }}
          >
            返回
          </Button>
          <Divider type="vertical" style={{ height: 24 }} />
          <div>
            <div style={{ fontSize: 18, fontWeight: 600 }}>运行 #{run.runNumber}</div>
          </div>
          <StatusBadge status={run.status} />
        </Space>
        {canCancel && (
          <Popconfirm title="确定取消此运行？" onOk={handleCancel}>
            <Button type="outline" status="danger" icon={<IconClose />} style={{ borderRadius: 6 }}>
              取消运行
            </Button>
          </Popconfirm>
        )}
      </div>

      {/* Summary Cards */}
      <Row gutter={16} style={{ marginBottom: 20 }}>
        <Col span={6}>
          <Card style={{ borderRadius: 10, textAlign: 'center' }} bodyStyle={{ padding: '16px 12px' }}>
            <InfoItem label="触发方式" value={run.triggerType === 'manual' ? '手动' : run.triggerType} />
          </Card>
        </Col>
        <Col span={6}>
          <Card style={{ borderRadius: 10, textAlign: 'center' }} bodyStyle={{ padding: '16px 12px' }}>
            <InfoItem label="分支" value={run.gitBranch ?? '-'} icon={<IconBranch style={{ fontSize: 12 }} />} />
          </Card>
        </Col>
        <Col span={6}>
          <Card style={{ borderRadius: 10, textAlign: 'center' }} bodyStyle={{ padding: '16px 12px' }}>
            <InfoItem label="触发人" value={run.triggeredBy?.replace('admin-bootstrap-', 'admin#') ?? '-'} icon={<IconUser style={{ fontSize: 12 }} />} />
          </Card>
        </Col>
        <Col span={6}>
          <Card style={{ borderRadius: 10, textAlign: 'center' }} bodyStyle={{ padding: '16px 12px' }}>
            <InfoItem label="Commit" value={run.gitCommit ? run.gitCommit.substring(0, 8) : '-'} icon={<IconCode style={{ fontSize: 12 }} />} />
          </Card>
        </Col>
      </Row>

      {/* Detail Card */}
      <Card title="详细信息" style={{ marginBottom: 20, borderRadius: 10 }} headerStyle={{ borderBottom: '1px solid var(--color-border)' }}>
        <Row gutter={[24, 20]}>
          <Col span={12}>
            <InfoItem label="开始时间" value={run.startedAt ? new Date(run.startedAt).toLocaleString() : '尚未开始'} icon={<IconClockCircle style={{ fontSize: 12 }} />} />
          </Col>
          <Col span={12}>
            <InfoItem label="结束时间" value={run.finishedAt ? new Date(run.finishedAt).toLocaleString() : '进行中'} icon={<IconClockCircle style={{ fontSize: 12 }} />} />
          </Col>
          <Col span={12}>
            <InfoItem label="Tekton 名称" value={run.tektonName ?? '-'} />
          </Col>
          <Col span={12}>
            <InfoItem label="命名空间" value={run.namespace ?? '-'} />
          </Col>
          {run.params && Object.keys(run.params).length > 0 && (
            <Col span={24}>
              <Text type="secondary" style={{ fontSize: 12, marginBottom: 8, display: 'block' }}>运行参数</Text>
              <div style={{ display: 'flex', gap: 8, flexWrap: 'wrap' }}>
                {Object.entries(run.params).map(([k, v]) => (
                  <Tag key={k} color="arcoblue" style={{ borderRadius: 4 }}>{k}={v}</Tag>
                ))}
              </div>
            </Col>
          )}
          {run.errorMessage && (
            <Col span={24}>
              <Text type="secondary" style={{ fontSize: 12, marginBottom: 4, display: 'block' }}>错误信息</Text>
              <div style={{
                padding: '8px 12px', background: '#FFECE8', borderRadius: 6,
                color: '#F53F3F', fontSize: 13,
              }}>
                {run.errorMessage}
              </div>
            </Col>
          )}
        </Row>
      </Card>

      {/* Logs */}
      <Card
        title="构建日志"
        style={{ marginBottom: 20, borderRadius: 10 }}
        headerStyle={{ borderBottom: '1px solid var(--color-border)' }}
      >
        {logsTotal > 0 ? (
          <pre style={{
            background: '#1D2129', color: '#C9CDD4', padding: 16, borderRadius: 8,
            fontSize: 13, lineHeight: 1.6, overflow: 'auto', maxHeight: 400,
            fontFamily: '"Fira Code", "Consolas", monospace',
          }}>
            {logs.map((l) => `[${l.level}] ${l.content}`).join('\n')}
          </pre>
        ) : (
          <div style={{
            padding: '32px 0', textAlign: 'center',
            color: 'var(--color-text-3)',
          }}>
            <IconCode style={{ fontSize: 32, marginBottom: 8, display: 'block', margin: '0 auto 8px' }} />
            <Paragraph type="secondary">日志查看器（将使用 xterm.js 实现）</Paragraph>
            <Text type="secondary" style={{ fontSize: 12 }}>暂无归档日志</Text>
          </div>
        )}
      </Card>

      {/* Artifacts */}
      {artifacts.length > 0 && (
        <Card title="构建产物" style={{ borderRadius: 10 }} headerStyle={{ borderBottom: '1px solid var(--color-border)' }}>
          <Space wrap size={12}>
            {artifacts.map((a) => (
              <Link key={a.name} href={a.url} target="_blank" rel="noopener noreferrer">
                <Tag color="blue" style={{ borderRadius: 4, cursor: 'pointer' }}>
                  {a.name}
                  {a.size != null ? ` (${(a.size / 1024).toFixed(1)} KB)` : ''}
                </Tag>
              </Link>
            ))}
          </Space>
        </Card>
      )}
    </div>
  );
}
