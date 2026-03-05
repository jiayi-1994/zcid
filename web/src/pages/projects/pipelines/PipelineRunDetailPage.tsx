import { useEffect, useState, useCallback } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import {
  Skeleton,
  Message,
  Typography,
  Button,
  Space,
  Descriptions,
  Tag,
  Card,
  Link,
  Popconfirm,
} from '@arco-design/web-react';
import { IconArrowLeft, IconClose } from '@arco-design/web-react/icon';
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

const { Paragraph } = Typography;

const statusColors: Record<string, string> = {
  pending: 'gray',
  queued: 'blue',
  running: 'arcoblue',
  succeeded: 'green',
  failed: 'red',
  cancelled: 'orange',
};

const statusLabels: Record<string, string> = {
  pending: '待执行',
  queued: '排队中',
  running: '运行中',
  succeeded: '成功',
  failed: '失败',
  cancelled: '已取消',
};

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

  useEffect(() => {
    loadData();
  }, [loadData]);

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
        <Skeleton text={{ rows: 4 }} animation />
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
    <div className="page-container">
      <div className="page-header">
        <Space>
          <Button
            type="text"
            icon={<IconArrowLeft />}
            onClick={() => navigate(`/projects/${projectId}/pipelines/${pipelineId}/runs`)}
          >
            返回
          </Button>
          <h3 className="page-title">
            运行 #{run.runNumber}
          </h3>
          <Tag color={statusColors[run.status] || 'default'}>{statusLabels[run.status] || run.status}</Tag>
        </Space>
        {canCancel && (
          <Popconfirm title="确定取消此运行？" onOk={handleCancel}>
            <Button type="primary" status="warning" icon={<IconClose />}>
              取消
            </Button>
          </Popconfirm>
        )}
      </div>

      <Card title="基本信息" style={{ marginBottom: 16 }}>
        <Descriptions
          column={1}
          data={[
            { label: '运行编号', value: run.runNumber },
            { label: '状态', value: <Tag color={statusColors[run.status]}>{statusLabels[run.status]}</Tag> },
            { label: '触发方式', value: run.triggerType },
            { label: '触发人', value: run.triggeredBy ?? '-' },
            { label: '分支', value: run.gitBranch ?? '-' },
            { label: 'Commit', value: run.gitCommit ?? '-' },
            { label: '开始时间', value: run.startedAt ? new Date(run.startedAt).toLocaleString() : '-' },
            { label: '结束时间', value: run.finishedAt ? new Date(run.finishedAt).toLocaleString() : '-' },
            {
              label: '参数',
              value: run.params && Object.keys(run.params).length > 0
                ? JSON.stringify(run.params)
                : '-',
            },
            run.errorMessage ? { label: '错误信息', value: run.errorMessage } : null,
          ].filter(Boolean) as { label: string; value: React.ReactNode }[]}
        />
      </Card>

      <Card title="日志" style={{ marginBottom: 16 }}>
        <Paragraph type="secondary">日志查看器（将使用 xterm.js 实现）</Paragraph>
        {logsTotal > 0 ? (
          <pre className="log-viewer">
            {logs.map((l) => `[${l.level}] ${l.content}`).join('\n')}
          </pre>
        ) : (
          <Paragraph type="secondary">暂无归档日志</Paragraph>
        )}
      </Card>

      {artifacts.length > 0 && (
        <Card title="构建产物">
          <Space wrap>
            {artifacts.map((a) => (
              <Link key={a.name} href={a.url} target="_blank" rel="noopener noreferrer">
                {a.name}
                {a.size != null ? ` (${(a.size / 1024).toFixed(1)} KB)` : ''}
              </Link>
            ))}
          </Space>
        </Card>
      )}
    </div>
  );
}
