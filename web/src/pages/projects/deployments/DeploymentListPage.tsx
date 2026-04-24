import { useCallback, useEffect, useState } from 'react';
import { Message } from '@arco-design/web-react';
import { useParams, useNavigate } from 'react-router-dom';
import { fetchDeployments, triggerDeploy, type DeploymentSummary, type DeploymentList } from '../../../services/deployment';
import { fetchEnvironments, type EnvironmentItem } from '../../../services/project';
import { extractErrorMessage } from '../../../services/http';
import { PageHeader } from '../../../components/ui/PageHeader';
import { Card } from '../../../components/ui/Card';
import { Btn } from '../../../components/ui/Btn';
import { StatusBadge } from '../../../components/ui/StatusBadge';
import { ZModal } from '../../../components/ui/ZModal';
import { ZSelect } from '../../../components/ui/ZSelect';
import { Field } from '../../../components/ui/Field';
import { IPlus, IChevL, IChevR } from '../../../components/ui/icons';

export function DeploymentListPage() {
  const { id: projectId } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [data, setData] = useState<DeploymentList>({ items: [], total: 0, page: 1, pageSize: 20 });
  const [loading, setLoading] = useState(false);
  const [page, setPage] = useState(1);
  const [modalVisible, setModalVisible] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [envs, setEnvs] = useState<EnvironmentItem[]>([]);
  const [form, setForm] = useState({ environmentId: '', image: '', pipelineRunId: '' });

  const loadData = useCallback(async (p: number) => {
    if (!projectId) return;
    setLoading(true);
    try {
      const result = await fetchDeployments(projectId, p, 20);
      setData(result);
    } catch {
      Message.error('加载部署列表失败');
    } finally {
      setLoading(false);
    }
  }, [projectId]);

  const loadEnvs = useCallback(async () => {
    if (!projectId) return;
    try {
      const r = await fetchEnvironments(projectId, 1, 100);
      setEnvs(r.items ?? []);
    } catch { /* silent */ }
  }, [projectId]);

  useEffect(() => { loadData(page); }, [page, loadData]);

  const openModal = () => {
    loadEnvs();
    setForm({ environmentId: '', image: '', pipelineRunId: '' });
    setModalVisible(true);
  };

  const handleSubmit = async () => {
    if (!form.environmentId || !form.image) { Message.error('请选择环境并填写镜像地址'); return; }
    if (!projectId) return;
    setSubmitting(true);
    try {
      await triggerDeploy(projectId, {
        environmentId: form.environmentId,
        image: form.image,
        pipelineRunId: form.pipelineRunId || undefined,
      });
      Message.success('部署已触发');
      setModalVisible(false);
      loadData(page);
    } catch (err: unknown) {
      Message.error(extractErrorMessage(err, '触发失败'));
    } finally {
      setSubmitting(false);
    }
  };

  const totalPages = Math.ceil(data.total / 20);

  return (
    <>
      <PageHeader
        crumb="Project · Delivery"
        title="部署管理"
        sub="触发与追踪 ArgoCD 部署同步状态。"
        actions={
          <Btn size="sm" variant="primary" icon={<IPlus size={13} />} onClick={openModal}>
            触发部署
          </Btn>
        }
      />
      <div style={{ padding: 24, display: 'flex', flexDirection: 'column', gap: 14 }}>
        <Card padding={false}>
          {loading ? (
            <div style={{ padding: '40px 0', textAlign: 'center', color: 'var(--z-400)' }}>加载中...</div>
          ) : (
            <table className="table">
              <thead>
                <tr>
                  <th>镜像</th><th>环境</th><th>状态</th><th>同步</th><th>健康</th><th>部署人</th><th>时间</th><th>操作</th>
                </tr>
              </thead>
              <tbody>
                {data.items.map((d: DeploymentSummary) => (
                  <tr key={d.id}>
                    <td><span className="code" style={{ fontSize: 11.5 }}>{d.image}</span></td>
                    <td><span className="sub">{d.environmentId || '-'}</span></td>
                    <td><StatusBadge status={d.status} /></td>
                    <td><span className="sub">{d.syncStatus ?? '-'}</span></td>
                    <td><span className="sub">{d.healthStatus ?? '-'}</span></td>
                    <td><span className="mono sub" style={{ fontSize: 11 }}>{d.deployedBy}</span></td>
                    <td><span className="sub mono" style={{ fontSize: 11 }}>{new Date(d.createdAt).toLocaleString()}</span></td>
                    <td>
                      <Btn size="xs" variant="ghost" onClick={() => navigate(`/projects/${projectId}/deployments/${d.id}`)}>
                        详情
                      </Btn>
                    </td>
                  </tr>
                ))}
                {data.items.length === 0 && !loading && (
                  <tr><td colSpan={8} style={{ textAlign: 'center', padding: '40px 0', color: 'var(--z-400)' }}>暂无部署记录</td></tr>
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

      {modalVisible && (
        <ZModal
          title="触发部署"
          onClose={() => setModalVisible(false)}
          footer={
            <div style={{ display: 'flex', justifyContent: 'flex-end', gap: 8 }}>
              <Btn onClick={() => setModalVisible(false)}>取消</Btn>
              <Btn variant="primary" onClick={handleSubmit} disabled={submitting}>
                {submitting ? '触发中...' : '触发'}
              </Btn>
            </div>
          }
        >
          <div style={{ display: 'flex', flexDirection: 'column', gap: 14 }}>
            <Field label="环境" required>
              <ZSelect
                width={300}
                value={form.environmentId}
                options={[{ value: '', label: '选择环境' }, ...envs.map((e) => ({ value: e.id, label: `${e.name} (${e.namespace})` }))]}
                onChange={(v) => setForm({ ...form, environmentId: v })}
              />
            </Field>
            <Field label="镜像" required>
              <input className="input" value={form.image} onChange={(e) => setForm({ ...form, image: e.target.value })} placeholder="nginx:latest 或 registry.io/app:v1" />
            </Field>
            <Field label="Pipeline Run ID（可选）">
              <input className="input" value={form.pipelineRunId} onChange={(e) => setForm({ ...form, pipelineRunId: e.target.value })} placeholder="关联的 Pipeline Run ID" />
            </Field>
          </div>
        </ZModal>
      )}
    </>
  );
}
