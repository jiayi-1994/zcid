import { useCallback, useEffect, useMemo, useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { Message } from '@arco-design/web-react';
import {
  fetchPipelineRuns, cancelPipelineRun, triggerPipelineRun,
  type PipelineRunSummary, type PipelineRunList,
} from '../../../services/pipelineRun';
import { fetchPipeline, type PipelineSummary, type PipelineConfig } from '../../../services/pipeline';
import { RunPipelineModal } from '../../../components/pipeline/RunPipelineModal';
import { ListFilters } from '../../../components/common/ListFilters';
import { useQueryFilters } from '../../../hooks/useQueryFilters';
import { PageHeader } from '../../../components/ui/PageHeader';
import { Card } from '../../../components/ui/Card';
import { Btn } from '../../../components/ui/Btn';
import { StatusBadge } from '../../../components/ui/StatusBadge';
import { IArrL, IPlay, IChevL, IChevR } from '../../../components/ui/icons';

const RUN_FILTER_DEFAULTS = { status: '', triggerType: '', commit: '' };

const TRIGGER_LABELS: Record<string, string> = {
  manual: '手动', webhook: 'Webhook', scheduled: '定时',
};

function formatDuration(startedAt?: string, finishedAt?: string): string {
  if (!startedAt || !finishedAt) return '-';
  const sec = Math.floor((new Date(finishedAt).getTime() - new Date(startedAt).getTime()) / 1000);
  if (sec < 60) return `${sec}s`;
  const min = Math.floor(sec / 60);
  if (min < 60) return `${min}m ${sec % 60}s`;
  return `${Math.floor(min / 60)}h ${min % 60}m`;
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

  const filteredItems = useMemo(() => (data.items ?? []).filter((item) => {
    if (filters.status && item.status !== filters.status) return false;
    if (filters.triggerType && item.triggerType !== filters.triggerType) return false;
    if (filters.commit) {
      const commit = (item as PipelineRunSummary & { gitCommit?: string }).gitCommit ?? '';
      if (!commit.toLowerCase().includes(filters.commit.toLowerCase())) return false;
    }
    return true;
  }), [data.items, filters]);

  const loadPipeline = useCallback(async () => {
    if (!projectId || !pipelineId) return;
    try {
      const p = await fetchPipeline(projectId, pipelineId);
      setPipeline({ id: p.id, projectId: p.projectId, name: p.name, description: p.description, status: p.status, triggerType: p.triggerType, concurrencyPolicy: p.concurrencyPolicy, createdBy: p.createdBy, createdAt: p.createdAt, updatedAt: p.updatedAt });
      setPipelineConfig(p.config);
    } catch { Message.error('加载流水线失败'); }
  }, [projectId, pipelineId]);

  const loadData = useCallback(async (p: number) => {
    if (!projectId || !pipelineId) return;
    setLoading(true);
    try {
      const result = await fetchPipelineRuns(projectId, pipelineId, p, 20);
      setData(result);
    } catch { Message.error('加载运行历史失败'); }
    finally { setLoading(false); }
  }, [projectId, pipelineId]);

  useEffect(() => { loadPipeline(); }, [loadPipeline]);
  useEffect(() => { loadData(page); }, [page, loadData]);

  const handleRunSubmit = async (params: { params?: Record<string, string>; gitBranch?: string; gitCommit?: string }) => {
    if (!projectId || !pipelineId) return;
    try {
      await triggerPipelineRun(projectId, pipelineId, params);
      Message.success('已触发运行');
      setRunModalVisible(false);
      loadData(page);
    } catch { Message.error('触发运行失败'); }
  };

  const handleCancel = async (runId: string) => {
    if (!projectId || !pipelineId) return;
    try { await cancelPipelineRun(projectId, pipelineId, runId); Message.success('已取消'); loadData(page); }
    catch { Message.error('取消失败'); }
  };

  const totalPages = Math.ceil(data.total / 20);

  return (
    <>
      <PageHeader
        crumb={`Pipelines · ${pipeline?.name ?? '...'}`}
        title="运行历史"
        sub={pipeline?.name ? `${pipeline.name} 的执行记录` : ''}
        actions={
          <>
            <Btn size="sm" icon={<IArrL size={13} />} onClick={() => navigate(`/projects/${projectId}/pipelines`)}>返回</Btn>
            <Btn size="sm" variant="primary" icon={<IPlay size={13} />} onClick={() => setRunModalVisible(true)}>触发运行</Btn>
          </>
        }
      />
      <div style={{ padding: 24, display: 'flex', flexDirection: 'column', gap: 14 }}>
        <div style={{ display: 'flex', gap: 8, alignItems: 'center', flexWrap: 'wrap' }}>
          <ListFilters
            filters={[
              { key: 'commit', type: 'search', placeholder: 'Commit SHA...' },
              { key: 'status', type: 'select', placeholder: '全部状态', options: [
                { label: '待执行', value: 'pending' },
                { label: '排队中', value: 'queued' },
                { label: '运行中', value: 'running' },
                { label: '成功', value: 'succeeded' },
                { label: '失败', value: 'failed' },
                { label: '已取消', value: 'cancelled' },
              ]},
              { key: 'triggerType', type: 'select', placeholder: '全部触发方式', options: [
                { label: '手动', value: 'manual' },
                { label: 'Webhook', value: 'webhook' },
                { label: '定时', value: 'scheduled' },
              ]},
            ]}
            values={filters}
            onChange={setFilters}
          />
          <span className="sub" style={{ fontSize: 11.5, marginLeft: 'auto' }}>{data.total} 条记录</span>
        </div>

        <Card padding={false}>
          {loading ? (
            <div style={{ padding: '40px 0', textAlign: 'center', color: 'var(--z-400)' }}>加载中...</div>
          ) : (
            <table className="table">
              <thead>
                <tr>
                  <th>#</th><th>状态</th><th>触发方式</th><th>触发人</th><th>分支</th><th>耗时</th><th>时间</th><th style={{ textAlign: 'right' }}>操作</th>
                </tr>
              </thead>
              <tbody>
                {filteredItems.map((run) => (
                  <tr key={run.id}>
                    <td><span className="mono sub" style={{ fontSize: 11.5 }}>#{run.runNumber}</span></td>
                    <td><StatusBadge status={run.status} /></td>
                    <td><span className="sub">{TRIGGER_LABELS[run.triggerType] || run.triggerType}</span></td>
                    <td><span className="mono sub" style={{ fontSize: 11.5 }}>{run.triggeredBy ?? '-'}</span></td>
                    <td><span className="code" style={{ fontSize: 11 }}>{run.gitBranch ?? '-'}</span></td>
                    <td><span className="mono sub" style={{ fontSize: 11.5 }}>{formatDuration(run.startedAt, run.finishedAt)}</span></td>
                    <td><span className="sub mono" style={{ fontSize: 11 }}>{new Date(run.createdAt).toLocaleString()}</span></td>
                    <td style={{ textAlign: 'right' }}>
                      <div style={{ display: 'inline-flex', gap: 4 }}>
                        <Btn size="xs" variant="ghost" onClick={() => navigate(`/projects/${projectId}/pipelines/${pipelineId}/runs/${run.id}`)}>详情</Btn>
                        {['pending', 'queued', 'running'].includes(run.status) && (
                          <Btn size="xs" variant="ghost" onClick={() => handleCancel(run.id)}>取消</Btn>
                        )}
                      </div>
                    </td>
                  </tr>
                ))}
                {filteredItems.length === 0 && !loading && (
                  <tr><td colSpan={8} style={{ textAlign: 'center', padding: '40px 0', color: 'var(--z-400)' }}>暂无运行记录</td></tr>
                )}
              </tbody>
            </table>
          )}
        </Card>

        {totalPages > 1 && (
          <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: 11.5, color: 'var(--z-500)' }}>
            <span>共 {data.total} 条 · 第 {page} / {totalPages} 页</span>
            <div style={{ display: 'flex', gap: 4 }}>
              <Btn size="xs" variant="ghost" iconOnly icon={<IChevL size={12} />} disabled={page <= 1} onClick={() => setPage((p) => p - 1)} />
              <Btn size="xs" variant="outline">{page}</Btn>
              <Btn size="xs" variant="ghost" iconOnly icon={<IChevR size={12} />} disabled={page >= totalPages} onClick={() => setPage((p) => p + 1)} />
            </div>
          </div>
        )}
      </div>

      <RunPipelineModal
        visible={runModalVisible}
        pipeline={pipeline}
        pipelineConfig={pipelineConfig}
        onClose={() => setRunModalVisible(false)}
        onSubmit={handleRunSubmit}
      />
    </>
  );
}
