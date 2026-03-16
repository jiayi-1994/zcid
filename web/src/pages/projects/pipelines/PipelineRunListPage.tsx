import { useEffect, useState, useCallback, useMemo } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { Table, Button, Space, Tag, Message, Popconfirm, Tooltip, Divider } from '@arco-design/web-react';
import { IconPlayArrow, IconArrowLeft, IconEye, IconClose, IconBranch } from '@arco-design/web-react/icon';
import {
  fetchPipelineRuns,
  cancelPipelineRun,
  triggerPipelineRun,
  type PipelineRunSummary,
  type PipelineRunList,
} from '../../../services/pipelineRun';
import { fetchPipeline, type PipelineSummary, type PipelineConfig } from '../../../services/pipeline';
import { RunPipelineModal } from '../../../components/pipeline/RunPipelineModal';
import { ListFilters } from '../../../components/common/ListFilters';
import { useQueryFilters } from '../../../hooks/useQueryFilters';

const statusConfig: Record<string, { color: string; bg: string }> = {
  pending:   { color: '#86909C', bg: '#F2F3F5' },
  queued:    { color: '#165DFF', bg: '#E8F3FF' },
  running:   { color: '#0FC6C2', bg: '#E8FFFB' },
  succeeded: { color: '#00B42A', bg: '#E8FFEA' },
  failed:    { color: '#F53F3F', bg: '#FFECE8' },
  cancelled: { color: '#FF7D00', bg: '#FFF7E8' },
};

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

const triggerLabels: Record<string, string> = {
  manual: '手动',
  webhook: 'Webhook',
  scheduled: '定时',
};

const runStatusFilterOptions = [
  { label: '全部', value: '' },
  { label: '待执行', value: 'pending' },
  { label: '排队中', value: 'queued' },
  { label: '运行中', value: 'running' },
  { label: '成功', value: 'succeeded' },
  { label: '失败', value: 'failed' },
  { label: '已取消', value: 'cancelled' },
];

const runTriggerFilterOptions = [
  { label: '全部', value: '' },
  { label: '手动', value: 'manual' },
  { label: 'Webhook', value: 'webhook' },
  { label: '定时', value: 'scheduled' },
];

const RUN_FILTER_DEFAULTS = { status: '', triggerType: '', commit: '' };

function formatDuration(startedAt?: string, finishedAt?: string): string {
  if (!startedAt || !finishedAt) return '-';
  const start = new Date(startedAt).getTime();
  const end = new Date(finishedAt).getTime();
  const sec = Math.floor((end - start) / 1000);
  if (sec < 60) return `${sec}s`;
  const min = Math.floor(sec / 60);
  const s = sec % 60;
  if (min < 60) return `${min}m ${s}s`;
  const h = Math.floor(min / 60);
  const m = min % 60;
  return `${h}h ${m}m`;
}

export default function PipelineRunListPage() {
  const { id: projectId, pipelineId } = useParams<{ id: string; pipelineId: string }>();
  const navigate = useNavigate();
  const [filters, setFilters] = useQueryFilters(RUN_FILTER_DEFAULTS);
  const [pipeline, setPipeline] = useState<PipelineSummary | null>(null);
  const [pipelineConfig, setPipelineConfig] = useState<PipelineConfig | null>(null);
  const [data, setData] = useState<PipelineRunList>({ items: [], total: 0, page: 1, pageSize: 20 });
  const [loading, setLoading] = useState(false);
  const [page, setPage] = useState(1);
  const [runModalVisible, setRunModalVisible] = useState(false);

  const filteredItems = useMemo(() => {
    return (data.items ?? []).filter((item) => {
      if (filters.status && item.status !== filters.status) return false;
      if (filters.triggerType && item.triggerType !== filters.triggerType) return false;
      if (filters.commit) {
        const commit = (item as PipelineRunSummary & { gitCommit?: string }).gitCommit ?? '';
        if (!commit.toLowerCase().includes(filters.commit.toLowerCase())) return false;
      }
      return true;
    });
  }, [data.items, filters.status, filters.triggerType, filters.commit]);

  const loadPipeline = useCallback(async () => {
    if (!projectId || !pipelineId) return;
    try {
      const p = await fetchPipeline(projectId, pipelineId);
      setPipeline({
        id: p.id,
        projectId: p.projectId,
        name: p.name,
        description: p.description,
        status: p.status,
        triggerType: p.triggerType,
        concurrencyPolicy: p.concurrencyPolicy,
        createdBy: p.createdBy,
        createdAt: p.createdAt,
        updatedAt: p.updatedAt,
      });
      setPipelineConfig(p.config);
    } catch {
      Message.error('加载流水线失败');
    }
  }, [projectId, pipelineId]);

  const loadData = useCallback(async (p: number) => {
    if (!projectId || !pipelineId) return;
    setLoading(true);
    try {
      const result = await fetchPipelineRuns(projectId, pipelineId, p, 20);
      setData(result);
    } catch {
      Message.error('加载运行历史失败');
    } finally {
      setLoading(false);
    }
  }, [projectId, pipelineId]);

  useEffect(() => {
    loadPipeline();
  }, [loadPipeline]);

  useEffect(() => {
    loadData(page);
  }, [page, loadData]);

  const handleTrigger = async () => {
    setRunModalVisible(true);
  };

  const handleRunSubmit = async (params: { params?: Record<string, string>; gitBranch?: string; gitCommit?: string }) => {
    if (!projectId || !pipelineId) return;
    try {
      await triggerPipelineRun(projectId, pipelineId, params);
      Message.success('已触发运行');
      setRunModalVisible(false);
      loadData(page);
    } catch {
      Message.error('触发运行失败');
    }
  };

  const handleCancel = async (runId: string) => {
    if (!projectId || !pipelineId) return;
    try {
      await cancelPipelineRun(projectId, pipelineId, runId);
      Message.success('已取消');
      loadData(page);
    } catch {
      Message.error('取消失败');
    }
  };

  const columns = useMemo(
    () => [
      {
        title: '#',
        dataIndex: 'runNumber',
        width: 80,
      },
      {
        title: '状态',
        dataIndex: 'status',
        width: 110,
        render: (status: string) => {
          const cfg = statusConfig[status] || { color: '#86909C', bg: '#F2F3F5' };
          return (
            <span style={{
              display: 'inline-flex', alignItems: 'center', gap: 6,
              padding: '3px 10px', borderRadius: 12,
              background: cfg.bg, color: cfg.color,
              fontWeight: 600, fontSize: 12,
            }}>
              <span style={{ width: 6, height: 6, borderRadius: '50%', background: cfg.color }} />
              {statusLabels[status] || status}
            </span>
          );
        },
      },
      {
        title: '触发方式',
        dataIndex: 'triggerType',
        render: (type: string) => triggerLabels[type] || type,
      },
      {
        title: '触发人',
        dataIndex: 'triggeredBy',
        render: (v: string | undefined) => v ?? '-',
      },
      {
        title: '分支',
        dataIndex: 'gitBranch',
        render: (v: string | undefined) => v ? (
          <Space size={4}><IconBranch style={{ fontSize: 13, color: 'var(--color-text-3)' }} /><span>{v}</span></Space>
        ) : '-',
      },
      {
        title: '耗时',
        render: (_: unknown, record: PipelineRunSummary) =>
          formatDuration(record.startedAt, record.finishedAt),
      },
      {
        title: '时间',
        dataIndex: 'createdAt',
        render: (t: string) => new Date(t).toLocaleString(),
      },
      {
        title: '操作',
        width: 160,
        render: (_: unknown, record: PipelineRunSummary) => (
          <Space size={4}>
            <Tooltip content="查看详情">
              <Button
                size="small"
                type="outline"
                icon={<IconEye />}
                onClick={() => navigate(`/projects/${projectId}/pipelines/${pipelineId}/runs/${record.id}`)}
                style={{ borderRadius: 4 }}
              >
                详情
              </Button>
            </Tooltip>
            {(record.status === 'pending' || record.status === 'queued' || record.status === 'running') && (
              <Popconfirm title="确定取消此运行？" onOk={() => handleCancel(record.id)}>
                <Tooltip content="取消运行">
                  <Button size="small" type="outline" status="danger" icon={<IconClose />} style={{ borderRadius: 4 }}>
                    取消
                  </Button>
                </Tooltip>
              </Popconfirm>
            )}
          </Space>
        ),
      },
    ],
    [projectId, pipelineId, navigate]
  );

  return (
    <div className="page-container">
      <div className="page-header">
        <Space wrap>
          <Button type="text" icon={<IconArrowLeft />} onClick={() => navigate(`/projects/${projectId}/pipelines`)}>
            返回
          </Button>
          <h3 className="page-title">
            {pipeline?.name ?? '流水线'} - 运行历史
          </h3>
        </Space>
        <Space>
          <ListFilters
            filters={[
              { key: 'commit', type: 'search', placeholder: '按 Commit SHA 搜索' },
              { key: 'status', type: 'select', placeholder: '按状态筛选', options: runStatusFilterOptions },
              { key: 'triggerType', type: 'select', placeholder: '按触发方式筛选', options: runTriggerFilterOptions },
            ]}
            values={filters}
            onChange={setFilters}
          />
          <Button type="primary" icon={<IconPlayArrow />} onClick={handleTrigger}>
            触发运行
          </Button>
        </Space>
      </div>
      <Table
        rowKey="id"
        columns={columns}
        data={filteredItems}
        loading={loading}
        border={false}
        stripe
        hover
        style={{ borderRadius: 8 }}
        pagination={{
          current: page,
          pageSize: 20,
          total: data.total,
          onChange: setPage,
          showTotal: true,
          style: { marginTop: 16 },
        }}
      />
      <RunPipelineModal
        visible={runModalVisible}
        pipeline={pipeline}
        pipelineConfig={pipelineConfig}
        onClose={() => setRunModalVisible(false)}
        onSubmit={handleRunSubmit}
      />
    </div>
  );
}
