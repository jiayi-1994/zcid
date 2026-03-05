import {
  Badge,
  Button,
  Card,
  Descriptions,
  Form,
  Input,
  Message,
  Space,
  Skeleton,
  Typography,
} from '@arco-design/web-react';
import { IconRefresh } from '@arco-design/web-react/icon';
import { useCallback, useEffect, useState } from 'react';
import { AppLayout } from '../../../components/layout/AppLayout';
import {
  fetchHealth,
  fetchIntegrationsStatus,
  fetchSettings,
  updateSettings,
  type HealthStatus,
  type IntegrationStatus,
  type SystemSettings,
} from '../../../services/adminSettings';

export function SystemSettingsPage() {
  const [settings, setSettings] = useState<SystemSettings | null>(null);
  const [health, setHealth] = useState<HealthStatus | null>(null);
  const [integrations, setIntegrations] = useState<IntegrationStatus[]>([]);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [form] = Form.useForm();

  const load = useCallback(async () => {
    setLoading(true);
    try {
      const [s, h, i] = await Promise.all([
        fetchSettings(),
        fetchHealth(),
        fetchIntegrationsStatus(),
      ]);
      setSettings(s);
      setHealth(h);
      setIntegrations(i);
      form.setFieldsValue(s);
    } catch {
      Message.error('加载系统设置失败');
    } finally {
      setLoading(false);
    }
  }, [form]);

  useEffect(() => { load(); }, [load]);

  const handleSave = async () => {
    try {
      const values = await form.validate();
      setSaving(true);
      const updated = await updateSettings(values as SystemSettings);
      setSettings(updated);
      Message.success('系统设置已保存');
    } catch {
      /* validation */
    } finally {
      setSaving(false);
    }
  };

  const statusColor = (s: string) => {
    if (s === 'ok' || s === 'healthy') return 'green';
    if (s === 'degraded' || s === 'fail') return 'red';
    return 'orange';
  };

  if (loading) return <AppLayout><div className="page-container"><Skeleton text={{ rows: 6 }} animation /></div></AppLayout>;

  return (
    <AppLayout>
      <div className="page-container" style={{ maxWidth: 900 }}>
        <div className="page-header">
          <h3 className="page-title">系统设置</h3>
        </div>

        <Card title="平台配置" style={{ marginBottom: 24 }}>
          <Form form={form} layout="vertical" style={{ maxWidth: 600 }}>
            <Form.Item label="K8s API Server 地址" field="k8sApiUrl">
              <Input placeholder="https://kubernetes.default.svc" />
            </Form.Item>
            <Form.Item label="默认镜像仓库" field="defaultRegistry">
              <Input placeholder="harbor.example.com" />
            </Form.Item>
            <Form.Item label="ArgoCD 地址" field="argocdUrl">
              <Input placeholder="https://argocd.example.com" />
            </Form.Item>
            <Form.Item>
              <Button type="primary" loading={saving} onClick={handleSave}>
                保存配置
              </Button>
            </Form.Item>
          </Form>
        </Card>

        <Card
          title="健康状态"
          style={{ marginBottom: 24 }}
          extra={<Button icon={<IconRefresh />} size="small" onClick={load}>刷新</Button>}
        >
          {health && (
            <>
              <Space style={{ marginBottom: 12 }}>
                <Typography.Text bold>总体状态:</Typography.Text>
                <Badge color={statusColor(health.status)} text={health.status} />
              </Space>
              <Descriptions
                column={1}
                data={Object.entries(health.checks).map(([k, v]) => ({
                  label: k,
                  value: <Badge color={statusColor(v)} text={v} />,
                }))}
                border
                style={{ marginTop: 8 }}
              />
            </>
          )}
        </Card>

        <Card title="集成状态">
          {integrations.map((item) => (
            <div key={item.name} style={{ display: 'flex', alignItems: 'center', padding: '8px 0', borderBottom: '1px solid var(--color-border)' }}>
              <Typography.Text bold style={{ width: 120 }}>{item.name}</Typography.Text>
              <Badge color={statusColor(item.status)} text={item.status} />
              {item.detail && (
                <Typography.Text type="secondary" style={{ marginLeft: 12, fontSize: 12 }}>
                  {item.detail}
                </Typography.Text>
              )}
            </div>
          ))}
        </Card>
      </div>
    </AppLayout>
  );
}

export default SystemSettingsPage;
