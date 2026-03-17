import { useEffect, useState, useCallback, useMemo } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { Table, Button, Space, Tag, Popconfirm, Message, Dropdown, Menu, Tooltip } from '@arco-design/web-react';
import { IconPlus, IconDelete, IconPlayArrow, IconMore, IconCopy, IconEdit, IconHistory } from '@arco-design/web-react/icon';
import { fetchPipelines, fetchPipeline, deletePipeline, copyPipeline, type PipelineSummary, type PipelineConfig } from '../../../services/pipeline';
import { triggerPipelineRun } from '../../../services/pipelineRun';
import { RunPipelineModal } from '../../../components/pipeline/RunPipelineModal';
import { ListFilters } from '../../../components/common/ListFilters';
import { useQueryFilters } from '../../../hooks/useQueryFilters';

const PAGE_SIZE = 20;
const LOAD_ALL_SIZE = 200;

const statusColors: Record<string, string> = {
  draft: 'gray',
  active: 'green',
  disabled: 'orangered',
};

const statusLabels: Record<string, string> = {
  draft: '草稿',
  active: '已启用',
  disabled: '已停用',
};

const triggerLabels: Record<string, string> = {
  manual: '手动',
  webhook: 'Webhook',
  scheduled: '定时',
};

const statusFilterOptions = [
  { label: '全部', value: '' },
  { label: '草稿', value: 'draft' },
  { label: '已启用', value: 'active' },
  { label: '已停用', value: 'disabled' },
];

const PIPELINE_FILTER_DEFAULTS = { status: '', search: '' };

export default function PipelineListPage() {
  const { id: projectId } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [filters, setFilters] = useQueryFilters(PIPELINE_FILTER_DEFAULTS);
  const [allItems, setAllItems] = useState<PipelineSummary[]>([]);
  const [loading, setLoading] = useState(false);
  const [page, setPage] = useState(1);
  const [runModalPipeline, setRunModalPipeline] = useState<PipelineSummary | null>(null);
  const [runModalConfig, setRunModalConfig] = useState<PipelineConfig | null>(null);

  const loadData = useCallback(async () => {
    if (!projectId) return;
    setLoading(true);
    try {
      const result = await fetchPipelines(projectId, 1, LOAD_ALL_SIZE);
      setAllItems(result.items);
    } catch {
      Message.error('加载流水线列表失败');
    } finally {
      setLoading(false);
    }
  }, [projectId]);

  useEffect(() => { loadData(); }, [loadData]);

  const filteredItems = useMemo(() => {
    return allItems.filter((item) => {
      if (filters.status && item.status !== filters.status) return false;
      if (filters.search) {
        const q = filters.search.toLowerCase();
        if (!item.name.toLowerCase().includes(q)) return false;
      }
      return true;
    });
  }, [allItems, filters.status, filters.search]);

  useEffect(() => { setPage(1); }, [filters.status, filters.search]);

  const paginatedItems = useMemo(() => {
    const start = (page - 1) * PAGE_SIZE;
    return filteredItems.slice(start, start + PAGE_SIZE);
  }, [filteredItems, page]);

  const handleDelete = useCallback(async (pipelineId: string) => {
    if (!projectId) return;
    try {
      await deletePipeline(projectId, pipelineId);
      Message.success('删除成功');
      loadData();
    } catch {
      Message.error('删除失败');
    }
  }, [projectId, loadData]);

  const handleCopy = useCallback(async (pipelineId: string) => {
    if (!projectId) return;
    try {
      await copyPipeline(projectId, pipelineId);
      Message.success('复制成功');
      loadData();
    } catch {
      Message.error('复制失败');
    }
  }, [projectId, loadData]);

  const handleOpenRunModal = useCallback(async (record: PipelineSummary) => {
    setRunModalPipeline(record);
    if (projectId) {
      try {
        const full = await fetchPipeline(projectId, record.id);
        setRunModalConfig(full.config);
      } catch {
        setRunModalConfig(null);
      }
    }
  }, [projectId]);

  const handleRunSubmit = useCallback(async (params: { params?: Record<string, string>; gitBranch?: string; gitCommit?: string }) => {
    if (!projectId || !runModalPipeline) return;
    try {
      await triggerPipelineRun(projectId, runModalPipeline.id, params);
      Message.success('已触发运行');
      setRunModalPipeline(null);
    } catch {
      Message.error('触发运行失败');
    }
  }, [projectId, runModalPipeline]);

  const columns = useMemo(() => [
    {
      title: '名称',
      dataIndex: 'name',
      render: (name: string, record: PipelineSummary) => (
        <Button type="text" style={{ padding: 0, fontWeight: 500 }} onClick={() => navigate(`/projects/${projectId}/pipelines/${record.id}`)}>
          {name}
        </Button>
      ),
    },
    { title: '描述', dataIndex: 'description', render: (v: string) => <span style={{ color: 'var(--color-text-3)' }}>{v || '-'}</span> },
    {
      title: '状态',
      dataIndex: 'status',
      width: 100,
      render: (status: string) => <Tag color={statusColors[status] || 'default'} size="small">{statusLabels[status] || status}</Tag>,
    },
    {
      title: '触发方式',
      dataIndex: 'triggerType',
      width: 100,
      render: (type: string) => <span style={{ color: 'var(--color-text-2)' }}>{triggerLabels[type] || type}</span>,
    },
    {
      title: '更新时间',
      dataIndex: 'updatedAt',
      width: 180,
      render: (time: string) => <span style={{ color: 'var(--color-text-3)', fontSize: 13 }}>{new Date(time).toLocaleString()}</span>,
    },
    {
      title: '操作',
      width: 200,
      render: (_: unknown, record: PipelineSummary) => (
        <Space size={4}>
          <Tooltip content="运行流水线">
            <Button
              size="small"
              type="primary"
              icon={<IconPlayArrow />}
              onClick={() => handleOpenRunModal(record)}
              style={{ borderRadius: 4 }}
            >
              运行
            </Button>
          </Tooltip>
          <Tooltip content="运行历史">
            <Button
              size="small"
              type="outline"
              icon={<IconHistory />}
              onClick={() => navigate(`/projects/${projectId}/pipelines/${record.id}/runs`)}
              style={{ borderRadius: 4 }}
            />
          </Tooltip>
          <Dropdown
            droplist={
              <Menu onClickMenuItem={(key) => {
                if (key === 'edit') navigate(`/projects/${projectId}/pipelines/${record.id}`);
                else if (key === 'copy') handleCopy(record.id);
              }}>
                <Menu.Item key="edit"><IconEdit style={{ marginRight: 8 }} />编辑</Menu.Item>
                <Menu.Item key="copy"><IconCopy style={{ marginRight: 8 }} />复制</Menu.Item>
              </Menu>
            }
            position="br"
          >
            <Button size="small" type="text" icon={<IconMore />} style={{ borderRadius: 4 }} />
          </Dropdown>
          <Popconfirm title="确定删除此流水线？" onOk={() => handleDelete(record.id)}>
            <Tooltip content="删除">
              <Button size="small" type="text" status="danger" icon={<IconDelete />} style={{ borderRadius: 4 }} />
            </Tooltip>
          </Popconfirm>
        </Space>
      ),
    },
  ], [projectId, navigate, handleDelete, handleCopy, handleOpenRunModal]);

  return (
    <div className="page-container">
      <div className="page-header">
        <h3 className="page-title">流水线</h3>
        <Space size={12}>
          <ListFilters
            filters={[
              { key: 'search', type: 'search', placeholder: '按名称搜索' },
              { key: 'status', type: 'select', placeholder: '按状态筛选', options: statusFilterOptions },
            ]}
            values={filters}
            onChange={setFilters}
          />
          <Button type="primary" icon={<IconPlus />} onClick={() => navigate(`/projects/${projectId}/pipelines/new`)} style={{ borderRadius: 6 }}>
            创建流水线
          </Button>
        </Space>
      </div>
      <Table
        rowKey="id"
        columns={columns}
        data={paginatedItems}
        loading={loading}
        border={false}
        stripe
        hover
        style={{ borderRadius: 8 }}
        pagination={{
          current: page,
          pageSize: PAGE_SIZE,
          total: filteredItems.length,
          onChange: setPage,
          showTotal: true,
          style: { marginTop: 16 },
        }}
      />
      <RunPipelineModal
        visible={!!runModalPipeline}
        pipeline={runModalPipeline}
        pipelineConfig={runModalConfig}
        onClose={() => { setRunModalPipeline(null); setRunModalConfig(null); }}
        onSubmit={handleRunSubmit}
      />
    </div>
  );
}
