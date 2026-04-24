import { useCallback, useEffect, useMemo, useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { Message } from '@arco-design/web-react';
import { fetchPipelines, fetchPipeline, deletePipeline, copyPipeline, type PipelineSummary, type PipelineConfig } from '../../../services/pipeline';
import { triggerPipelineRun } from '../../../services/pipelineRun';
import { RunPipelineModal } from '../../../components/pipeline/RunPipelineModal';
import { PageHeader } from '../../../components/ui/PageHeader';
import { Metric } from '../../../components/ui/Metric';
import { Btn } from '../../../components/ui/Btn';
import { StatusBadge } from '../../../components/ui/StatusBadge';
import { IPlus, IPlay, ILayers, ISearch, IEdit, ICopy, ITrash, IZap } from '../../../components/ui/icons';

const LOAD_ALL_SIZE = 200;

const TRIGGER_LABELS: Record<string, string> = {
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

  const filteredItems = useMemo(() => allItems.filter((item) => {
    if (statusFilter && item.status !== statusFilter) return false;
    if (search && !item.name.toLowerCase().includes(search.toLowerCase())) return false;
    return true;
  }), [allItems, statusFilter, search]);

  const metrics = useMemo(() => ({
    total: allItems.length,
    active: allItems.filter((i) => i.status === 'active').length,
    draft: allItems.filter((i) => i.status === 'draft').length,
    disabled: allItems.filter((i) => i.status === 'disabled').length,
  }), [allItems]);

  const handleDelete = useCallback(async (pipelineId: string) => {
    if (!projectId) return;
    try { await deletePipeline(projectId, pipelineId); Message.success('删除成功'); loadData(); }
    catch { Message.error('删除失败'); }
  }, [projectId, loadData]);

  const handleCopy = useCallback(async (pipelineId: string) => {
    if (!projectId) return;
    try { await copyPipeline(projectId, pipelineId); Message.success('复制成功'); loadData(); }
    catch { Message.error('复制失败'); }
  }, [projectId, loadData]);

  const handleOpenRunModal = useCallback(async (record: PipelineSummary) => {
    setRunModalPipeline(record);
    if (projectId) {
      try { const full = await fetchPipeline(projectId, record.id); setRunModalConfig(full.config); }
      catch { setRunModalConfig(null); }
    }
  }, [projectId]);

  const handleRunSubmit = useCallback(async (params: { params?: Record<string, string>; gitBranch?: string; gitCommit?: string }) => {
    if (!projectId || !runModalPipeline) return;
    try { await triggerPipelineRun(projectId, runModalPipeline.id, params); Message.success('已触发运行'); setRunModalPipeline(null); }
    catch { Message.error('触发运行失败'); }
  }, [projectId, runModalPipeline]);

  const STATUS_FILTERS = [
    { value: '', label: '全部' },
    { value: 'active', label: 'Active' },
    { value: 'draft', label: 'Draft' },
    { value: 'disabled', label: 'Disabled' },
  ];

  return (
    <>
      <PageHeader
        crumb="Pipelines"
        title="Automated Pipelines"
        sub="Real-time status of your CI/CD automation. 自动化构建与部署流水线。"
        actions={
          <Btn size="sm" variant="primary" icon={<IPlus size={13} />} onClick={() => navigate(`/projects/${projectId}/pipelines/new`)}>
            Create Pipeline
          </Btn>
        }
      />
      <div style={{ padding: '24px 24px 48px', display: 'flex', flexDirection: 'column', gap: 18 }}>
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(4,1fr)', gap: 14 }}>
          <Metric label="TOTAL" value={metrics.total} icon={<ILayers size={14} />} iconBg="var(--z-100)" iconColor="var(--z-700)" onClick={() => setStatusFilter('')} />
          <Metric label="ACTIVE" value={metrics.active} icon={<IZap size={14} />} iconBg="var(--green-soft)" iconColor="var(--green-ink)" trendTone="green" onClick={() => setStatusFilter('active')} />
          <Metric label="DRAFT" value={metrics.draft} icon={<IEdit size={14} />} iconBg="var(--blue-soft)" iconColor="var(--blue-ink)" onClick={() => setStatusFilter('draft')} />
          <Metric label="DISABLED" value={metrics.disabled} icon={<ILayers size={14} />} iconBg="var(--amber-soft)" iconColor="var(--amber-ink)" onClick={() => setStatusFilter('disabled')} />
        </div>

        {/* filter bar */}
        <div style={{ display: 'flex', gap: 8, alignItems: 'center' }}>
          <div className="input-wrap" style={{ flex: 1, maxWidth: 320 }}>
            <ISearch size={13} />
            <input className="input input--with-icon" placeholder="搜索流水线..." value={search} onChange={(e) => setSearch(e.target.value)} />
          </div>
          <div style={{ display: 'flex', gap: 4 }}>
            {STATUS_FILTERS.map((f) => (
              <Btn
                key={f.value}
                size="sm"
                variant={statusFilter === f.value ? 'primary' : 'ghost'}
                onClick={() => setStatusFilter(f.value)}
              >
                {f.label}
              </Btn>
            ))}
          </div>
        </div>

        {loading ? (
          <div style={{ padding: '40px 0', textAlign: 'center', color: 'var(--z-400)' }}>加载中...</div>
        ) : filteredItems.length === 0 ? (
          <div style={{ padding: '48px 0', textAlign: 'center', color: 'var(--z-500)' }}>
            <div style={{ fontSize: 14, fontWeight: 500, marginBottom: 4 }}>暂无流水线</div>
            <div style={{ fontSize: 12.5, marginBottom: 14 }}>创建第一条流水线，开始自动化构建</div>
            <Btn variant="primary" icon={<IPlus size={13} />} onClick={() => navigate(`/projects/${projectId}/pipelines/new`)}>创建流水线</Btn>
          </div>
        ) : (
          <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
            {filteredItems.map((item) => (
              <div
                key={item.id}
                className="card"
                style={{ padding: '12px 16px', display: 'flex', alignItems: 'center', gap: 14, cursor: 'pointer' }}
                onClick={() => navigate(`/projects/${projectId}/pipelines/${item.id}`)}
              >
                <div style={{ flex: 1, minWidth: 0 }}>
                  <div style={{ fontSize: 13.5, fontWeight: 600, marginBottom: 2 }}>{item.name}</div>
                  <div className="sub" style={{ fontSize: 11.5 }}>
                    {TRIGGER_LABELS[item.triggerType] || item.triggerType}
                    {item.description ? ` · ${item.description}` : ''}
                    {' · '}
                    {new Date(item.updatedAt).toLocaleString()}
                  </div>
                </div>
                <StatusBadge status={item.status} />
                <div style={{ display: 'flex', gap: 4 }} onClick={(e) => e.stopPropagation()}>
                  <Btn size="xs" variant="primary" iconOnly icon={<IPlay size={11} />} title="运行" onClick={() => handleOpenRunModal(item)} />
                  <Btn size="xs" variant="ghost" iconOnly icon={<ICopy size={12} />} title="复制" onClick={() => handleCopy(item.id)} />
                  <Btn size="xs" variant="ghost" iconOnly icon={<ITrash size={12} />} title="删除" onClick={() => handleDelete(item.id)} />
                </div>
              </div>
            ))}
          </div>
        )}
      </div>

      <RunPipelineModal
        visible={!!runModalPipeline}
        pipeline={runModalPipeline}
        pipelineConfig={runModalConfig}
        onClose={() => { setRunModalPipeline(null); setRunModalConfig(null); }}
        onSubmit={handleRunSubmit}
      />
    </>
  );
}
