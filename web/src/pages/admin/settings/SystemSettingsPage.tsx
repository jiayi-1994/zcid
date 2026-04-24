import { useCallback, useEffect, useState } from 'react';
import { Message } from '@arco-design/web-react';
import { AppLayout } from '../../../components/layout/AppLayout';
import {
  fetchHealth, fetchIntegrationsStatus, fetchSettings, updateSettings,
  type HealthStatus, type IntegrationStatus, type SystemSettings,
} from '../../../services/adminSettings';
import { PageHeader } from '../../../components/ui/PageHeader';
import { Card } from '../../../components/ui/Card';
import { Btn } from '../../../components/ui/Btn';
import { Badge } from '../../../components/ui/Badge';
import { StatusBadge } from '../../../components/ui/StatusBadge';
import { Field } from '../../../components/ui/Field';
import { IRefresh } from '../../../components/ui/icons';

export function SystemSettingsPage() {
  const [settings, setSettings] = useState<SystemSettings | null>(null);
  const [health, setHealth] = useState<HealthStatus | null>(null);
  const [integrations, setIntegrations] = useState<IntegrationStatus[]>([]);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [form, setForm] = useState({ k8sApiUrl: '', defaultRegistry: '', argocdUrl: '' });

  const load = useCallback(async () => {
    setLoading(true);
    try {
      const [s, h, i] = await Promise.all([fetchSettings(), fetchHealth(), fetchIntegrationsStatus()]);
      setSettings(s); setHealth(h); setIntegrations(i);
      setForm({ k8sApiUrl: s.k8sApiUrl ?? '', defaultRegistry: s.defaultRegistry ?? '', argocdUrl: s.argocdUrl ?? '' });
    } catch { Message.error('加载系统设置失败'); }
    finally { setLoading(false); }
  }, []);

  useEffect(() => { load(); }, [load]);

  const handleSave = async () => {
    setSaving(true);
    try { const updated = await updateSettings(form as SystemSettings); setSettings(updated); Message.success('系统设置已保存'); }
    catch { Message.error('保存失败'); }
    finally { setSaving(false); }
  };

  const healthTone = (s: string): 'green' | 'amber' | 'red' =>
    s === 'ok' || s === 'healthy' ? 'green' : s === 'degraded' ? 'amber' : 'red';

  if (loading) {
    return <AppLayout><div style={{ padding: '40px 24px', color: 'var(--z-400)' }}>加载中...</div></AppLayout>;
  }

  return (
    <AppLayout>
      <PageHeader crumb="System › Settings" title="System Settings" sub="Platform configuration & health. 平台级配置、健康状态与集成监控。" />
      <div style={{ padding: '24px 24px 48px', display: 'flex', flexDirection: 'column', gap: 18, maxWidth: 900 }}>
        <Card title="平台配置" extra={<Btn size="sm" variant="primary" onClick={handleSave} disabled={saving}>{saving ? '保存中...' : '保存配置'}</Btn>}>
          <div style={{ display: 'flex', flexDirection: 'column', gap: 14 }}>
            <Field label="K8s API Server 地址" required>
              <input className="input" value={form.k8sApiUrl} onChange={(e) => setForm({ ...form, k8sApiUrl: e.target.value })} placeholder="https://kubernetes.default.svc" />
            </Field>
            <Field label="默认镜像仓库" required>
              <input className="input" value={form.defaultRegistry} onChange={(e) => setForm({ ...form, defaultRegistry: e.target.value })} placeholder="harbor.example.com" />
            </Field>
            <Field label="ArgoCD 地址" required>
              <input className="input" value={form.argocdUrl} onChange={(e) => setForm({ ...form, argocdUrl: e.target.value })} placeholder="https://argocd.example.com" />
            </Field>
          </div>
        </Card>

        {health && (
          <Card
            title={
              <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
                <h2>健康状态</h2>
                <Badge tone={healthTone(health.status)} dot>
                  {health.status === 'ok' || health.status === 'healthy' ? 'ok · 全部健康' : health.status}
                </Badge>
              </div>
            }
            extra={<Btn size="sm" icon={<IRefresh size={13} />} onClick={load}>刷新</Btn>}
          >
            <div style={{ display: 'grid', gridTemplateColumns: 'repeat(2,1fr)', rowGap: 10, columnGap: 20, fontSize: 12.5 }}>
              {Object.entries(health.checks).map(([k, v]) => (
                <div key={k} style={{ display: 'flex', justifyContent: 'space-between', padding: '8px 0', borderBottom: '1px dashed var(--z-150)' }}>
                  <span>{k}</span>
                  <Badge tone={healthTone(v)} dot>{v}</Badge>
                </div>
              ))}
            </div>
          </Card>
        )}

        <Card title="集成状态">
          <div style={{ display: 'flex', flexDirection: 'column', gap: 10 }}>
            {integrations.map((it) => (
              <div key={it.name} style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', padding: '8px 0', borderBottom: '1px dashed var(--z-150)' }}>
                <div>
                  <div style={{ fontSize: 12.5, fontWeight: 500 }}>{it.name}</div>
                  {it.detail && <div className="sub mono" style={{ fontSize: 11 }}>{it.detail}</div>}
                </div>
                <StatusBadge status={it.status === 'ok' || it.status === 'healthy' ? 'connected' : it.status === 'degraded' ? 'token_expired' : 'disconnected'} />
              </div>
            ))}
            {integrations.length === 0 && <div style={{ color: 'var(--z-400)', fontSize: 12 }}>暂无集成数据</div>}
          </div>
        </Card>
      </div>
    </AppLayout>
  );
}

export default SystemSettingsPage;
