import { useCallback, useEffect, useState } from 'react';
import { Message } from '@arco-design/web-react';
import { useParams } from 'react-router-dom';
import { useAuthStore } from '../../../stores/auth';
import { fetchEnvironments, createEnvironment, deleteEnvironment, type EnvironmentItem } from '../../../services/project';
import { extractErrorMessage } from '../../../services/http';
import { PageHeader } from '../../../components/ui/PageHeader';
import { Card } from '../../../components/ui/Card';
import { Metric } from '../../../components/ui/Metric';
import { Btn } from '../../../components/ui/Btn';
import { Badge } from '../../../components/ui/Badge';
import { ZModal } from '../../../components/ui/ZModal';
import { Field } from '../../../components/ui/Field';
import { IPlus, IServer } from '../../../components/ui/icons';

type HealthState = 'healthy' | 'warning' | 'degraded' | 'unknown' | 'stale';

const HEALTH_TONE: Record<HealthState, 'green' | 'amber' | 'red' | 'grey'> = {
  healthy: 'green',
  warning: 'amber',
  degraded: 'red',
  unknown: 'grey',
  stale: 'amber',
};

const HEALTH_LABEL: Record<HealthState, string> = {
  healthy: 'Healthy',
  warning: 'Warning',
  degraded: 'Degraded',
  unknown: 'Unknown',
  stale: 'Stale',
};

export function EnvironmentListPage() {
  const { id: projectId } = useParams<{ id: string }>();
  const [envs, setEnvs] = useState<EnvironmentItem[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(false);
  const [modalVisible, setModalVisible] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [form, setForm] = useState({ name: '', namespace: '', description: '' });
  const isAdmin = useAuthStore((s) => s.user?.role === 'admin');

  const loadData = useCallback(async () => {
    if (!projectId) return;
    setLoading(true);
    try {
      const data = await fetchEnvironments(projectId, 1, 20);
      setEnvs(data.items ?? []);
      setTotal(data.total);
    } catch (err: unknown) {
      Message.error(extractErrorMessage(err, '加载环境列表失败'));
    } finally {
      setLoading(false);
    }
  }, [projectId]);

  useEffect(() => { loadData(); }, [loadData]);

  const handleCreate = async () => {
    if (!form.name || !form.namespace) { Message.error('请填写环境名称和 Namespace'); return; }
    setSubmitting(true);
    try {
      await createEnvironment(projectId!, form.name, form.namespace, form.description);
      Message.success('环境创建成功');
      setForm({ name: '', namespace: '', description: '' });
      setModalVisible(false);
      loadData();
    } catch (err: unknown) {
      Message.error(extractErrorMessage(err, '创建失败'));
    } finally {
      setSubmitting(false);
    }
  };

  const handleDelete = async (envId: string) => {
    try {
      await deleteEnvironment(projectId!, envId);
      Message.success('环境已删除');
      loadData();
    } catch (err: unknown) {
      Message.error(extractErrorMessage(err, '删除失败'));
    }
  };

  const envHealth = (env: EnvironmentItem): HealthState => env.health?.status ?? 'unknown';
  const healthyCount = envs.filter((e) => envHealth(e) === 'healthy').length;

  return (
    <>
      <PageHeader
        crumb="Project · Environments"
        title="Environment Health"
        sub="Deployment targets & rollback. 监控和管理部署环境状态。"
        actions={isAdmin && (
          <Btn size="sm" variant="primary" icon={<IPlus size={13} />} onClick={() => setModalVisible(true)}>
            New Environment
          </Btn>
        )}
      />
      <div style={{ padding: '24px 24px 48px', display: 'flex', flexDirection: 'column', gap: 18 }}>
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3,1fr)', gap: 14 }}>
          <Metric label="TOTAL ENVIRONMENTS" value={total} icon={<IServer size={14} />} iconBg="var(--z-100)" iconColor="var(--z-700)" />
          <Metric label="HEALTHY" value={healthyCount} icon={<IServer size={14} />} iconBg="var(--green-soft)" iconColor="var(--green-ink)" trend="all targets ok" trendTone="green" />
          <Metric
            label="HEALTH RATE"
            value={total > 0 ? `${Math.round((healthyCount / total) * 100)}%` : '—'}
            icon={<IServer size={14} />}
            iconBg="var(--blue-soft)"
            iconColor="var(--blue-ink)"
          />
        </div>

        {loading ? (
          <div style={{ padding: '40px 0', textAlign: 'center', color: 'var(--z-400)' }}>加载中...</div>
        ) : envs.length === 0 ? (
          <div style={{ padding: '48px 0', textAlign: 'center', color: 'var(--z-500)' }}>
            <div style={{ fontSize: 14, fontWeight: 500, marginBottom: 4 }}>暂无环境</div>
            <div style={{ fontSize: 12.5, marginBottom: 14 }}>创建第一个环境，开始管理部署目标</div>
            {isAdmin && (
              <Btn variant="primary" icon={<IPlus size={13} />} onClick={() => setModalVisible(true)}>新建环境</Btn>
            )}
          </div>
        ) : (
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(280px, 1fr))', gap: 14 }}>
            {envs.map((env) => {
              const health = envHealth(env);
              return (
                <Card key={env.id} padding={false}>
                  <div style={{ padding: '14px 16px 12px' }}>
                      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: 10 }}>
                        <div style={{ fontSize: 13.5, fontWeight: 600 }}>{env.name}</div>
                        <Badge tone={HEALTH_TONE[health]} dot>{HEALTH_LABEL[health]}</Badge>
                      </div>
                      <div style={{ fontSize: 11.5, color: 'var(--z-500)', marginBottom: 8 }}>
                        {env.health?.reason || 'No environment health signals yet'}
                      </div>
                    <div style={{ display: 'flex', flexDirection: 'column', gap: 6, fontSize: 12 }}>
                      <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                        <span style={{ color: 'var(--z-500)' }}>Namespace</span>
                        <span className="mono" style={{ color: 'var(--z-700)', fontSize: 11 }}>{env.namespace}</span>
                      </div>
                      {env.description && (
                        <div style={{ display: 'flex', justifyContent: 'space-between', gap: 10 }}>
                          <span style={{ color: 'var(--z-500)' }}>Description</span>
                          <span style={{ color: 'var(--z-800)', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap', maxWidth: 160 }}>{env.description}</span>
                        </div>
                      )}
                      <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                        <span style={{ color: 'var(--z-500)' }}>Created</span>
                        <span className="mono sub" style={{ fontSize: 11 }}>
                          {env.createdAt ? new Date(env.createdAt).toLocaleDateString() : '-'}
                        </span>
                      </div>
                    </div>
                  </div>
                  {isAdmin && (
                    <div style={{ padding: '8px 16px 12px', display: 'flex', justifyContent: 'flex-end' }}>
                      <Btn size="xs" variant="ghost" onClick={() => handleDelete(env.id)}>删除</Btn>
                    </div>
                  )}
                </Card>
              );
            })}
          </div>
        )}
      </div>

      {modalVisible && (
        <ZModal
          title="新建环境"
          onClose={() => setModalVisible(false)}
          footer={
            <div style={{ display: 'flex', justifyContent: 'flex-end', gap: 8 }}>
              <Btn onClick={() => setModalVisible(false)}>取消</Btn>
              <Btn variant="primary" onClick={handleCreate} disabled={submitting}>
                {submitting ? '创建中...' : '创建'}
              </Btn>
            </div>
          }
        >
          <div style={{ display: 'flex', flexDirection: 'column', gap: 14 }}>
            <Field label="环境名称" required>
              <input className="input" value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })} placeholder="dev / staging / prod" />
            </Field>
            <Field label="K8s Namespace" required>
              <input className="input" value={form.namespace} onChange={(e) => setForm({ ...form, namespace: e.target.value })} placeholder="zcid-dev" />
            </Field>
            <Field label="描述">
              <textarea className="input" rows={2} value={form.description} onChange={(e) => setForm({ ...form, description: e.target.value })} placeholder="环境描述（可选）" style={{ resize: 'vertical' }} />
            </Field>
          </div>
        </ZModal>
      )}
    </>
  );
}
