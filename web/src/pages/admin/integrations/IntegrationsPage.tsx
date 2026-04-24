import { useCallback, useEffect, useState } from 'react';
import { Message } from '@arco-design/web-react';
import { AppLayout } from '../../../components/layout/AppLayout';
import {
  fetchConnections, createConnection, updateConnection, deleteConnection, testConnection, getWebhookSecret,
  type GitConnection,
} from '../../../services/integration';
import { ConnectionFormModal } from './ConnectionFormModal';
import { PageHeader } from '../../../components/ui/PageHeader';
import { Metric } from '../../../components/ui/Metric';
import { Btn } from '../../../components/ui/Btn';
import { StatusBadge } from '../../../components/ui/StatusBadge';
import { IPlug, ICheck, IAlert, IRefresh, IPlus, IPlay, ICopy, IEdit, ITrash } from '../../../components/ui/icons';

const PROVIDER_ICON: Record<string, string> = { github: '🐙', gitlab: '🦊' };

export function IntegrationsPage() {
  const [connections, setConnections] = useState<GitConnection[]>([]);
  const [loading, setLoading] = useState(false);
  const [createVisible, setCreateVisible] = useState(false);
  const [editItem, setEditItem] = useState<GitConnection | null>(null);
  const [testingId, setTestingId] = useState<string | null>(null);

  const loadData = useCallback(async () => {
    setLoading(true);
    try { const data = await fetchConnections(); setConnections(data.items || []); }
    catch { Message.error('加载集成列表失败'); }
    finally { setLoading(false); }
  }, []);

  useEffect(() => { loadData(); }, [loadData]);

  const handleCreate = async (data: { name: string; providerType: string; serverUrl: string; accessToken: string; description: string }) => {
    await createConnection(data); await loadData();
  };
  const handleEdit = async (data: { name: string; accessToken: string; description: string }) => {
    if (!editItem) return;
    await updateConnection(editItem.id, { name: data.name || undefined, accessToken: data.accessToken || undefined, description: data.description ?? undefined });
    setEditItem(null); await loadData();
  };
  const handleDelete = async (id: string) => {
    try { await deleteConnection(id); Message.success('连接已删除'); await loadData(); }
    catch { Message.error('删除失败'); }
  };
  const handleTest = async (id: string) => {
    setTestingId(id);
    try {
      const result = await testConnection(id);
      result.success ? Message.success('连接测试成功') : Message.warning(`连接测试失败: ${result.message}`);
      await loadData();
    } catch { Message.error('测试请求失败'); }
    finally { setTestingId(null); }
  };
  const handleCopyWebhook = async (id: string) => {
    try { const s = await getWebhookSecret(id); await navigator.clipboard.writeText(s); Message.success('Webhook Secret 已复制'); }
    catch { Message.error('获取 Webhook Secret 失败'); }
  };

  const connectedCount = connections.filter((c) => c.status === 'connected').length;
  const needsAttention = connections.length - connectedCount;

  return (
    <AppLayout>
      <PageHeader
        crumb="Settings › Integrations"
        title="Integration Management"
        sub="Connect and manage your external CI/CD toolchain. 连接并管理 Git 源、镜像仓库、通知渠道。"
        actions={
          <>
            <Btn size="sm" icon={<IRefresh size={13} />} onClick={loadData}>Refresh</Btn>
            <Btn size="sm" variant="primary" icon={<IPlus size={13} />} onClick={() => setCreateVisible(true)}>Connect New Service</Btn>
          </>
        }
      />
      <div style={{ padding: 24, display: 'flex', flexDirection: 'column', gap: 18 }}>
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3,1fr)', gap: 14 }}>
          <Metric label="TOTAL INTEGRATIONS" value={connections.length} icon={<IPlug size={14} />} iconBg="var(--z-100)" iconColor="var(--z-700)" />
          <Metric label="CONNECTED" value={connectedCount} icon={<ICheck size={14} />} iconBg="var(--green-soft)" iconColor="var(--green-ink)" trend="all healthy" trendTone="green" />
          <Metric label="NEEDS ATTENTION" value={needsAttention} icon={<IAlert size={14} />} iconBg="var(--amber-soft)" iconColor="var(--amber-ink)" trend={needsAttention > 0 ? 'review tokens' : undefined} trendTone="amber" />
        </div>

        {loading ? (
          <div style={{ padding: '40px 0', textAlign: 'center', color: 'var(--z-400)' }}>加载中...</div>
        ) : connections.length === 0 ? (
          <div style={{ padding: '48px 0', textAlign: 'center', color: 'var(--z-500)' }}>
            <div style={{ fontSize: 14, fontWeight: 500, marginBottom: 4 }}>暂无集成连接</div>
            <div style={{ fontSize: 12.5, marginBottom: 14 }}>添加 Git 仓库连接，开始管理源代码</div>
            <Btn variant="primary" icon={<IPlus size={13} />} onClick={() => setCreateVisible(true)}>添加连接</Btn>
          </div>
        ) : (
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(320px, 1fr))', gap: 14 }}>
            {connections.map((conn) => (
              <div key={conn.id} className="card" style={{ padding: 0 }}>
                <div style={{ padding: '14px 14px 10px', display: 'flex', gap: 11, alignItems: 'flex-start' }}>
                  <div style={{ width: 36, height: 36, borderRadius: 8, background: 'var(--z-100)', display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: 18, flex: 'none' }}>
                    {PROVIDER_ICON[conn.providerType] ?? '🔌'}
                  </div>
                  <div style={{ flex: 1, minWidth: 0 }}>
                    <div style={{ fontSize: 13.5, fontWeight: 600 }}>{conn.name}</div>
                    <div className="sub" style={{ fontSize: 11.5 }}>{conn.providerType} · {conn.serverUrl}</div>
                  </div>
                </div>
                <div style={{ padding: '6px 14px 12px', display: 'flex', flexDirection: 'column', gap: 6, fontSize: 11.5 }}>
                  <div style={{ display: 'flex', justifyContent: 'space-between', gap: 10 }}>
                    <span style={{ color: 'var(--z-500)' }}>Token</span>
                    <span className="mono" style={{ color: 'var(--z-700)' }}>{conn.tokenMask}</span>
                  </div>
                  {conn.description && (
                    <div style={{ display: 'flex', justifyContent: 'space-between', gap: 10 }}>
                      <span style={{ color: 'var(--z-500)' }}>Description</span>
                      <span style={{ color: 'var(--z-800)', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap', maxWidth: 180 }}>{conn.description}</span>
                    </div>
                  )}
                  <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                    <span style={{ color: 'var(--z-500)' }}>Created</span>
                    <span className="mono" style={{ color: 'var(--z-700)' }}>{conn.createdAt}</span>
                  </div>
                </div>
                <div style={{ padding: '10px 14px', borderTop: '1px solid var(--z-150)', background: 'var(--z-25)', display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
                  <StatusBadge status={conn.status} />
                  <div style={{ display: 'inline-flex', gap: 4 }}>
                    <Btn size="xs" variant="ghost" iconOnly icon={<IPlay size={11} />} title="Test" onClick={() => handleTest(conn.id)} disabled={testingId === conn.id} />
                    <Btn size="xs" variant="ghost" iconOnly icon={<ICopy size={12} />} title="Copy webhook secret" onClick={() => handleCopyWebhook(conn.id)} />
                    <Btn size="xs" variant="ghost" iconOnly icon={<IEdit size={12} />} onClick={() => setEditItem(conn)} />
                    <Btn size="xs" variant="ghost" iconOnly icon={<ITrash size={12} />} onClick={() => handleDelete(conn.id)} />
                  </div>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>

      <ConnectionFormModal visible={createVisible} onClose={() => setCreateVisible(false)} onSubmit={handleCreate} />
      {editItem && (
        <ConnectionFormModal visible={!!editItem} onClose={() => setEditItem(null)} onSubmit={handleEdit} editMode initialValues={{ name: editItem.name, description: editItem.description }} />
      )}
    </AppLayout>
  );
}
