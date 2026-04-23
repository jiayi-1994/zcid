import { useEffect, useState, useCallback, useMemo } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { Button, Space, Tag, Popconfirm, Message, Dropdown, Menu, Tooltip } from '@arco-design/web-react';
import { IconPlus, IconDelete, IconPlayArrow, IconMore, IconCopy, IconEdit, IconHistory, IconSearch } from '@arco-design/web-react/icon';
import { fetchPipelines, fetchPipeline, deletePipeline, copyPipeline, type PipelineSummary, type PipelineConfig } from '../../../services/pipeline';
import { triggerPipelineRun } from '../../../services/pipelineRun';
import { RunPipelineModal } from '../../../components/pipeline/RunPipelineModal';

const LOAD_ALL_SIZE = 200;

const statusConfig: Record<string, { color: string; bg: string; label: string; badgeClass: string }> = {
  draft: { color: '#6B7280', bg: '#F3F4F6', label: '草稿', badgeClass: 'pipeline-status-badge--pending' },
  active: { color: '#00C853', bg: '#E8F5E9', label: '已启用', badgeClass: 'pipeline-status-badge--success' },
  disabled: { color: '#FF9500', bg: '#FFF8E1', label: '已停用', badgeClass: 'pipeline-status-badge--cancelled' },
};

const triggerLabels: Record<string, string> = {
  manual: '手动触发',
  webhook: 'Webhook',
  scheduled: '定时触发',
};

export default function PipelineListPage() {
  const { id: projectId } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [allItems, setAllItems] = useState<PipelineSummary[]>([]);
  const [loading, setLoading] = useState(false);
  const [search, setSearch] = useState('');
  const [statusFilter, setStatusFilter] = useState('');
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
      if (statusFilter && item.status !== statusFilter) return false;
      if (search) {
        const q = search.toLowerCase();
        if (!item.name.toLowerCase().includes(q)) return false;
      }
      return true;
    });
  }, [allItems, statusFilter, search]);

  const metrics = useMemo(() => {
    const total = allItems.length;
    const active = allItems.filter((i) => i.status === 'active').length;
    const draft = allItems.filter((i) => i.status === 'draft').length;
    const disabled = allItems.filter((i) => i.status === 'disabled').length;
    return { total, active, draft, disabled };
  }, [allItems]);

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

  return (
    <div className="page-container">
      <div className="page-header">
        <div>
          <div className="breadcrumb">Pipelines</div>
          <h1 className="page-title">Automated Pipelines</h1>
          <p className="page-subtitle">
            Real-time status of your deployment infrastructure and service health across all environments.
          </p>
        </div>
        <Button
          type="primary"
          icon={<IconPlus />}
          onClick={() => navigate(`/projects/${projectId}/pipelines/new`)}
          size="large"
        >
          Create Pipeline
        </Button>
      </div>

      <div className="metrics-grid" style={{ gridTemplateColumns: 'repeat(4, 1fr)' }}>
        <div className="metric-card" onClick={() => setStatusFilter('')} style={{ cursor: 'pointer' }}>
          <span className="metric-card-label">Total Pipelines</span>
          <span className="metric-card-value">{metrics.total}</span>
          <span className="metric-card-sub">所有流水线</span>
        </div>
        <div className="metric-card" onClick={() => setStatusFilter('active')} style={{ cursor: 'pointer' }}>
          <span className="metric-card-label">Active</span>
          <span className="metric-card-value">{metrics.active}</span>
          <span className="metric-card-sub">已启用 · running smoothly</span>
        </div>
        <div className="metric-card" onClick={() => setStatusFilter('draft')} style={{ cursor: 'pointer' }}>
          <span className="metric-card-label">Draft</span>
          <span className="metric-card-value">{metrics.draft}</span>
          <span className="metric-card-sub">待发布</span>
        </div>
        <div className="metric-card" onClick={() => setStatusFilter('disabled')} style={{ cursor: 'pointer' }}>
          <span className="metric-card-label">Disabled</span>
          <span className="metric-card-value">{metrics.disabled}</span>
          <span className="metric-card-sub">已停用</span>
        </div>
      </div>

      <div style={{
        display: 'flex', gap: 'var(--space-3)', marginBottom: 'var(--space-5)',
        alignItems: 'center',
        padding: 'var(--space-3) var(--space-4)',
        background: 'var(--surface-container-lowest)',
        borderRadius: 'var(--radius-lg)',
      }}>
        <div style={{ position: 'relative', flex: 1 }}>
          <IconSearch style={{ position: 'absolute', left: 12, top: '50%', transform: 'translateY(-50%)', color: 'var(--on-surface-variant)', fontSize: 16 }} />
          <input
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder="搜索流水线..."
            style={{
              width: '100%', height: 38,
              border: '1px solid var(--ghost-border-strong)',
              borderRadius: 'var(--radius-sm)',
              paddingLeft: 40, paddingRight: 12, fontSize: 13, outline: 'none',
              background: 'var(--surface-container-low)',
              color: 'var(--on-surface)',
              transition: 'all 200ms',
            }}
            onFocus={(e) => { e.target.style.borderColor = 'var(--primary)'; e.target.style.boxShadow = 'var(--shadow-glow-primary)'; }}
            onBlur={(e) => { e.target.style.borderColor = 'var(--ghost-border-strong)'; e.target.style.boxShadow = 'none'; }}
          />
        </div>
        <Space size={8}>
          {['', 'active', 'draft', 'disabled'].map((s) => (
            <Tag
              key={s}
              style={{
                cursor: 'pointer', borderRadius: 'var(--radius-full)', padding: '5px 14px',
                background: statusFilter === s ? 'var(--primary-gradient)' : 'var(--surface-container-low)',
                color: statusFilter === s ? '#fff' : 'var(--on-surface)',
                border: 'none', fontWeight: 600, fontSize: 12,
                transition: 'all 150ms',
              }}
              onClick={() => setStatusFilter(s)}
            >
              {s === '' ? '全部' : statusConfig[s]?.label || s}
            </Tag>
          ))}
        </Space>
      </div>

      {/* Pipeline List */}
      {loading ? (
        <div style={{ padding: '60px 0', textAlign: 'center', color: 'var(--muted-foreground)' }}>
          加载中...
        </div>
      ) : filteredItems.length === 0 ? (
        <div className="zcid-card empty-state">
          <div className="empty-state-title">暂无流水线</div>
          <div className="empty-state-desc">
            创建你的第一条流水线，开始自动化构建
          </div>
          <Button
            type="primary"
            icon={<IconPlus />}
            onClick={() => navigate(`/projects/${projectId}/pipelines/new`)}
          >
            创建流水线
          </Button>
        </div>
      ) : (
        <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
          {filteredItems.map((item) => {
            const cfg = statusConfig[item.status] || statusConfig.draft;
            return (
              <div
                key={item.id}
                className="pipeline-status-card"
                onClick={() => navigate(`/projects/${projectId}/pipelines/${item.id}`)}
              >
                <div className={`pipeline-status-icon pipeline-status-icon--${item.status === 'active' ? 'success' : item.status === 'disabled' ? 'pending' : 'pending'}`}>
                  {item.status === 'active' ? '✓' : item.status === 'disabled' ? '⏸' : '📝'}
                </div>
                <div className="pipeline-status-info">
                  <div className="pipeline-status-name">{item.name}</div>
                  <div className="pipeline-status-meta">
                    {triggerLabels[item.triggerType] || item.triggerType}
                    {' · '}
                    {item.description || '无描述'}
                    {' · '}
                    {new Date(item.updatedAt).toLocaleString()}
                  </div>
                </div>
                <span className={`pipeline-status-badge ${cfg.badgeClass}`}>
                  <span style={{ width: 6, height: 6, borderRadius: '50%', background: cfg.color }} />
                  {cfg.label}
                </span>
                <div style={{ display: 'flex', gap: 4, alignItems: 'center' }} onClick={(e) => e.stopPropagation()}>
                  <Tooltip content="运行流水线">
                    <Button
                      size="small"
                      type="primary"
                      icon={<IconPlayArrow />}
                      onClick={() => handleOpenRunModal(item)}
                      style={{ borderRadius: 6 }}
                    />
                  </Tooltip>
                  <Tooltip content="运行历史">
                    <Button
                      size="small"
                      type="outline"
                      icon={<IconHistory />}
                      onClick={() => navigate(`/projects/${projectId}/pipelines/${item.id}/runs`)}
                      style={{ borderRadius: 6 }}
                    />
                  </Tooltip>
                  <Dropdown
                    droplist={
                      <Menu onClickMenuItem={(key) => {
                        if (key === 'edit') navigate(`/projects/${projectId}/pipelines/${item.id}`);
                        else if (key === 'copy') handleCopy(item.id);
                      }}>
                        <Menu.Item key="edit"><IconEdit style={{ marginRight: 8 }} />编辑</Menu.Item>
                        <Menu.Item key="copy"><IconCopy style={{ marginRight: 8 }} />复制</Menu.Item>
                      </Menu>
                    }
                    position="br"
                  >
                    <Button size="small" type="text" icon={<IconMore />} style={{ borderRadius: 6 }} />
                  </Dropdown>
                  <Popconfirm title="确定删除此流水线？" onOk={() => handleDelete(item.id)}>
                    <Tooltip content="删除">
                      <Button size="small" type="text" status="danger" icon={<IconDelete />} style={{ borderRadius: 6 }} />
                    </Tooltip>
                  </Popconfirm>
                </div>
              </div>
            );
          })}
        </div>
      )}
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
