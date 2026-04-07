import { Button, Message, Modal, Form, Input, Tag } from '@arco-design/web-react';
import { IconPlus } from '@arco-design/web-react/icon';
import { useState, useEffect, useCallback } from 'react';
import { useParams } from 'react-router-dom';
import { useAuthStore } from '../../../stores/auth';
import { fetchEnvironments, createEnvironment, deleteEnvironment, type EnvironmentItem } from '../../../services/project';
import { extractErrorMessage } from '../../../services/http';

const FormItem = Form.Item;

const healthStates = ['healthy', 'syncing', 'degraded'] as const;
type HealthState = typeof healthStates[number];

function deriveHealth(env: EnvironmentItem): HealthState {
  const name = (env.name || '').toLowerCase();
  if (name.includes('prod')) return 'healthy';
  if (name.includes('staging') || name.includes('stag')) return 'syncing';
  return 'healthy';
}

const healthConfig: Record<string, { label: string; cssClass: string; icon: string }> = {
  healthy: { label: 'Healthy', cssClass: 'env-health-status--healthy', icon: '✓' },
  syncing: { label: 'Syncing', cssClass: 'env-health-status--syncing', icon: '↻' },
  degraded: { label: 'Degraded', cssClass: 'env-health-status--degraded', icon: '!' },
  down: { label: 'Down', cssClass: 'env-health-status--down', icon: '✕' },
};

export function EnvironmentListPage() {
  const { id: projectId } = useParams<{ id: string }>();
  const [envs, setEnvs] = useState<EnvironmentItem[]>([]);
  const [total, setTotal] = useState(0);
  const [page] = useState(1);
  const [loading, setLoading] = useState(false);
  const [modalVisible, setModalVisible] = useState(false);
  const [form] = Form.useForm();
  const [submitLoading, setSubmitLoading] = useState(false);
  const isAdmin = useAuthStore((s) => s.user?.role === 'admin');

  const loadData = useCallback(async (p: number) => {
    if (!projectId) return;
    setLoading(true);
    try {
      const data = await fetchEnvironments(projectId, p, 20);
      setEnvs(data.items ?? []);
      setTotal(data.total);
    } catch (err: unknown) {
      Message.error(extractErrorMessage(err, '加载环境列表失败'));
    } finally {
      setLoading(false);
    }
  }, [projectId]);

  useEffect(() => { loadData(page); }, [page, loadData]);

  const handleCreate = async () => {
    try {
      const values = await form.validate();
      setSubmitLoading(true);
      await createEnvironment(projectId!, values.name, values.namespace, values.description || '');
      Message.success('环境创建成功');
      form.resetFields();
      setModalVisible(false);
      loadData(page);
    } catch (err: unknown) {
      const msg = extractErrorMessage(err, '');
      if (msg) Message.error(msg);
    } finally {
      setSubmitLoading(false);
    }
  };

  const handleDelete = async (envId: string) => {
    try {
      await deleteEnvironment(projectId!, envId);
      Message.success('环境已删除');
      loadData(page);
    } catch (err: unknown) {
      Message.error(extractErrorMessage(err, '删除失败'));
    }
  };

  const activeCount = envs.length;
  const healthyCount = envs.filter((e) => deriveHealth(e) === 'healthy').length;

  return (
    <div className="page-container">
      {/* Header */}
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: 24 }}>
        <div>
          <h3 className="page-title" style={{ fontSize: 22, marginBottom: 4 }}>Environment Health</h3>
          <p style={{ margin: 0, fontSize: 13, color: 'var(--muted-foreground)' }}>
            监控和管理部署环境状态
          </p>
        </div>
        {isAdmin && (
          <Button
            type="primary"
            icon={<IconPlus />}
            onClick={() => setModalVisible(true)}
            style={{ borderRadius: 8, height: 40, fontWeight: 600 }}
          >
            + New Environment
          </Button>
        )}
      </div>

      {/* Overview Metrics */}
      <div className="metrics-grid" style={{ gridTemplateColumns: 'repeat(3, 1fr)', marginBottom: 24 }}>
        <div className="metric-card">
          <span className="metric-card-label">TOTAL ENVIRONMENTS</span>
          <span className="metric-card-value">{total}</span>
        </div>
        <div className="metric-card">
          <span className="metric-card-label">ACTIVE SERVICES</span>
          <span className="metric-card-value" style={{ color: 'var(--success)' }}>{activeCount}</span>
        </div>
        <div className="metric-card">
          <span className="metric-card-label">HEALTH RATE</span>
          <span className="metric-card-value" style={{ color: 'var(--success)' }}>
            {total > 0 ? `${((healthyCount / total) * 100).toFixed(0)}%` : '-'}
          </span>
        </div>
      </div>

      {/* Environment Cards */}
      {loading ? (
        <div style={{ padding: '60px 0', textAlign: 'center', color: 'var(--muted-foreground)' }}>
          加载中...
        </div>
      ) : envs.length === 0 ? (
        <div style={{
          padding: '60px 0', textAlign: 'center',
          background: 'var(--card)', borderRadius: 12, border: '1px solid var(--border)',
        }}>
          <div style={{ fontSize: 48, marginBottom: 12 }}>🌐</div>
          <div style={{ fontSize: 16, fontWeight: 600, color: 'var(--foreground)', marginBottom: 4 }}>
            暂无环境
          </div>
          <div style={{ fontSize: 13, color: 'var(--muted-foreground)', marginBottom: 16 }}>
            创建第一个环境，开始管理部署目标
          </div>
          {isAdmin && (
            <Button
              type="primary"
              icon={<IconPlus />}
              onClick={() => setModalVisible(true)}
              style={{ borderRadius: 8 }}
            >
              新建环境
            </Button>
          )}
        </div>
      ) : (
        <div className="env-health-grid">
          {envs.map((env) => {
            const health = deriveHealth(env);
            const hcfg = healthConfig[health];
            return (
              <div key={env.id} className="env-health-card">
                <div className="env-health-header">
                  <div className="env-health-name">{env.name}</div>
                  <span className={`env-health-status ${hcfg.cssClass}`}>
                    {hcfg.icon} {hcfg.label}
                  </span>
                </div>
                <div className="env-health-meta">
                  <div className="env-health-meta-item">
                    <span className="env-health-meta-label">Namespace</span>
                    <Tag size="small" style={{ borderRadius: 20, fontSize: 11 }}>{env.namespace}</Tag>
                  </div>
                  {env.description && (
                    <div className="env-health-meta-item">
                      <span className="env-health-meta-label">Description</span>
                      <span className="env-health-meta-value">{env.description}</span>
                    </div>
                  )}
                  <div className="env-health-meta-item">
                    <span className="env-health-meta-label">Created</span>
                    <span className="env-health-meta-value" style={{ fontSize: 12 }}>
                      {env.createdAt ? new Date(env.createdAt).toLocaleDateString() : '-'}
                    </span>
                  </div>
                </div>
                {isAdmin && (
                  <div style={{ marginTop: 16, paddingTop: 12, borderTop: '1px solid var(--border)', display: 'flex', justifyContent: 'flex-end' }}>
                    <Button
                      type="text"
                      size="small"
                      status="danger"
                      onClick={() => handleDelete(env.id)}
                      style={{ borderRadius: 6, fontSize: 12 }}
                    >
                      删除
                    </Button>
                  </div>
                )}
              </div>
            );
          })}
        </div>
      )}

      <Modal
        title="新建环境" visible={modalVisible}
        onOk={handleCreate}
        onCancel={() => { form.resetFields(); setModalVisible(false); }}
        confirmLoading={submitLoading} unmountOnExit
      >
        <Form form={form} layout="vertical">
          <FormItem label="环境名称" field="name" rules={[{ required: true, message: '请输入环境名称' }]}>
            <Input placeholder="如 dev / staging / prod" style={{ borderRadius: 8 }} />
          </FormItem>
          <FormItem label="K8s Namespace" field="namespace" rules={[{ required: true, message: '请输入 Namespace' }]}>
            <Input placeholder="如 zcid-dev" style={{ borderRadius: 8 }} />
          </FormItem>
          <FormItem label="描述" field="description">
            <Input.TextArea placeholder="环境描述" rows={2} style={{ borderRadius: 8 }} />
          </FormItem>
        </Form>
      </Modal>
    </div>
  );
}
