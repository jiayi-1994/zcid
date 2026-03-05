import { useEffect, useState, useCallback, useMemo } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { Table, Button, Space, Tag, Popconfirm, Message } from '@arco-design/web-react';
import { IconPlus, IconCopy, IconDelete, IconEdit, IconPlayArrow } from '@arco-design/web-react/icon';
import { fetchPipelines, deletePipeline, copyPipeline, type PipelineSummary, type PipelineList } from '../../../services/pipeline';
import { RunPipelineModal } from '../../../components/pipeline/RunPipelineModal';
import { ListFilters } from '../../../components/common/ListFilters';
import { useQueryFilters } from '../../../hooks/useQueryFilters';

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
  const [data, setData] = useState<PipelineList>({ items: [], total: 0, page: 1, pageSize: 20 });
  const [loading, setLoading] = useState(false);
  const [page, setPage] = useState(1);
  const [runModalPipeline, setRunModalPipeline] = useState<PipelineSummary | null>(null);

  const loadData = useCallback(async (p: number) => {
    if (!projectId) return;
    setLoading(true);
    try {
      const result = await fetchPipelines(projectId, p, 20);
      setData(result);
    } catch {
      Message.error('加载流水线列表失败');
    } finally {
      setLoading(false);
    }
  }, [projectId]);

  useEffect(() => { loadData(page); }, [page, loadData]);

  const filteredItems = useMemo(() => {
    return data.items.filter((item) => {
      if (filters.status && item.status !== filters.status) return false;
      if (filters.search) {
        const q = filters.search.toLowerCase();
        if (!item.name.toLowerCase().includes(q)) return false;
      }
      return true;
    });
  }, [data.items, filters.status, filters.search]);

  const handleDelete = async (pipelineId: string) => {
    if (!projectId) return;
    try {
      await deletePipeline(projectId, pipelineId);
      Message.success('删除成功');
      loadData(page);
    } catch {
      Message.error('删除失败');
    }
  };

  const handleCopy = async (pipelineId: string) => {
    if (!projectId) return;
    try {
      await copyPipeline(projectId, pipelineId);
      Message.success('复制成功');
      loadData(page);
    } catch {
      Message.error('复制失败');
    }
  };

  const columns = useMemo(() => [
    {
      title: '名称',
      dataIndex: 'name',
      render: (name: string, record: PipelineSummary) => (
        <Button type="text" onClick={() => navigate(`/projects/${projectId}/pipelines/${record.id}`)}>
          {name}
        </Button>
      ),
    },
    { title: '描述', dataIndex: 'description' },
    {
      title: '状态',
      dataIndex: 'status',
      render: (status: string) => <Tag color={statusColors[status] || 'default'}>{statusLabels[status] || status}</Tag>,
    },
    {
      title: '触发方式',
      dataIndex: 'triggerType',
      render: (type: string) => triggerLabels[type] || type,
    },
    {
      title: '更新时间',
      dataIndex: 'updatedAt',
      render: (time: string) => new Date(time).toLocaleString(),
    },
    {
      title: '操作',
      render: (_: unknown, record: PipelineSummary) => (
        <Space>
          <Button
            size="small"
            type="text"
            icon={<IconPlayArrow />}
            onClick={() => setRunModalPipeline(record)}
          >
            运行
          </Button>
          <Button
            size="small"
            type="text"
            onClick={() => navigate(`/projects/${projectId}/pipelines/${record.id}/runs`)}
          >
            运行历史
          </Button>
          <Button
            size="small"
            type="text"
            icon={<IconEdit />}
            onClick={() => navigate(`/projects/${projectId}/pipelines/${record.id}`)}
          />
          <Button size="small" type="text" icon={<IconCopy />} onClick={() => handleCopy(record.id)} />
          <Popconfirm title="确定删除此流水线？" onOk={() => handleDelete(record.id)}>
            <Button size="small" type="text" status="danger" icon={<IconDelete />} />
          </Popconfirm>
        </Space>
      ),
    },
  ], [projectId, navigate, handleDelete, handleCopy]);

  return (
    <div className="page-container">
      <div className="page-header">
        <h3 className="page-title">流水线</h3>
        <Space>
          <ListFilters
            filters={[
              { key: 'search', type: 'search', placeholder: '按名称搜索' },
              { key: 'status', type: 'select', placeholder: '按状态筛选', options: statusFilterOptions },
            ]}
            values={filters}
            onChange={setFilters}
          />
          <Button type="primary" icon={<IconPlus />} onClick={() => navigate(`/projects/${projectId}/pipelines/new`)}>
            创建流水线
          </Button>
        </Space>
      </div>
      <Table
        rowKey="id"
        columns={columns}
        data={filteredItems}
        loading={loading}
        pagination={{
          current: page,
          pageSize: 20,
          total: data.total,
          onChange: setPage,
          showTotal: true,
        }}
      />
      <RunPipelineModal
        visible={!!runModalPipeline}
        pipeline={runModalPipeline}
        onClose={() => setRunModalPipeline(null)}
      />
    </div>
  );
}
