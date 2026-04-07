import { Button, Message, Popconfirm, Space, Tooltip, Tag } from '@arco-design/web-react';
import { IconDelete, IconEdit, IconPlayArrow, IconPlus, IconRefresh, IconCopy } from '@arco-design/web-react/icon';
import { useCallback, useEffect, useState } from 'react';
import { AppLayout } from '../../../components/layout/AppLayout';
import {
  fetchConnections,
  createConnection,
  updateConnection,
  deleteConnection,
  testConnection,
  getWebhookSecret,
  type GitConnection,
} from '../../../services/integration';
import { ConnectionFormModal } from './ConnectionFormModal';

const PROVIDER_CONFIG: Record<string, { icon: string; bg: string; label: string }> = {
  github: { icon: '🐙', bg: '#F6F8FA', label: 'GitHub' },
  gitlab: { icon: '🦊', bg: '#FFF4E6', label: 'GitLab' },
};

const STATUS_CONFIG: Record<string, { label: string; cssClass: string }> = {
  connected: { label: 'Connected', cssClass: 'env-health-status--healthy' },
  disconnected: { label: 'Disconnected', cssClass: 'env-health-status--down' },
  token_expired: { label: 'Token Expired', cssClass: 'env-health-status--degraded' },
};

export function IntegrationsPage() {
  const [connections, setConnections] = useState<GitConnection[]>([]);
  const [loading, setLoading] = useState(false);
  const [createVisible, setCreateVisible] = useState(false);
  const [editItem, setEditItem] = useState<GitConnection | null>(null);
  const [testingId, setTestingId] = useState<string | null>(null);

  const loadData = useCallback(async () => {
    setLoading(true);
    try {
      const data = await fetchConnections();
      setConnections(data.items || []);
    } catch {
      Message.error('加载集成列表失败');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => { loadData(); }, [loadData]);

  const handleCreate = async (data: {
    name: string;
    providerType: string;
    serverUrl: string;
    accessToken: string;
    description: string;
  }) => {
    await createConnection(data);
    await loadData();
  };

  const handleEdit = async (data: {
    name: string;
    accessToken: string;
    description: string;
  }) => {
    if (!editItem) return;
    await updateConnection(editItem.id, {
      name: data.name || undefined,
      accessToken: data.accessToken || undefined,
      description: data.description ?? undefined,
    });
    setEditItem(null);
    await loadData();
  };

  const handleDelete = async (id: string) => {
    try {
      await deleteConnection(id);
      Message.success('连接已删除');
      await loadData();
    } catch {
      Message.error('删除失败');
    }
  };

  const handleTest = async (id: string) => {
    setTestingId(id);
    try {
      const result = await testConnection(id);
      if (result.success) {
        Message.success('连接测试成功');
      } else {
        Message.warning(`连接测试失败: ${result.message}`);
      }
      await loadData();
    } catch {
      Message.error('测试请求失败');
    } finally {
      setTestingId(null);
    }
  };

  const handleCopyWebhookSecret = async (id: string) => {
    try {
      const secret = await getWebhookSecret(id);
      await navigator.clipboard.writeText(secret);
      Message.success('Webhook Secret 已复制到剪贴板');
    } catch {
      Message.error('获取 Webhook Secret 失败');
    }
  };

  const connectedCount = connections.filter((c) => c.status === 'connected').length;

  return (
    <AppLayout>
      <div className="page-container">
        {/* Header */}
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: 24 }}>
          <div>
            <h3 className="page-title" style={{ fontSize: 22, marginBottom: 4 }}>Integration Management</h3>
            <p style={{ margin: 0, fontSize: 13, color: 'var(--muted-foreground)' }}>
              管理 Git 仓库连接、Registry 和 Webhook 配置
            </p>
          </div>
          <Space size={8}>
            <Button
              icon={<IconRefresh />}
              onClick={loadData}
              style={{ borderRadius: 8, height: 40 }}
            >
              刷新
            </Button>
            <Button
              type="primary"
              icon={<IconPlus />}
              onClick={() => setCreateVisible(true)}
              style={{ borderRadius: 8, height: 40, fontWeight: 600 }}
            >
              + Add Connection
            </Button>
          </Space>
        </div>

        {/* Metrics */}
        <div className="metrics-grid" style={{ gridTemplateColumns: 'repeat(3, 1fr)', marginBottom: 24 }}>
          <div className="metric-card">
            <span className="metric-card-label">TOTAL INTEGRATIONS</span>
            <span className="metric-card-value">{connections.length}</span>
          </div>
          <div className="metric-card">
            <span className="metric-card-label">CONNECTED</span>
            <span className="metric-card-value" style={{ color: 'var(--success)' }}>{connectedCount}</span>
          </div>
          <div className="metric-card">
            <span className="metric-card-label">NEEDS ATTENTION</span>
            <span className="metric-card-value" style={{ color: connections.length - connectedCount > 0 ? 'var(--warning)' : 'var(--muted-foreground)' }}>
              {connections.length - connectedCount}
            </span>
          </div>
        </div>

        {/* Integration Cards */}
        {loading ? (
          <div style={{ padding: '60px 0', textAlign: 'center', color: 'var(--muted-foreground)' }}>
            加载中...
          </div>
        ) : connections.length === 0 ? (
          <div style={{
            padding: '60px 0', textAlign: 'center',
            background: 'var(--card)', borderRadius: 12, border: '1px solid var(--border)',
          }}>
            <div style={{ fontSize: 48, marginBottom: 12 }}>🔗</div>
            <div style={{ fontSize: 16, fontWeight: 600, color: 'var(--foreground)', marginBottom: 4 }}>
              暂无集成连接
            </div>
            <div style={{ fontSize: 13, color: 'var(--muted-foreground)', marginBottom: 16 }}>
              添加 Git 仓库连接，开始管理源代码
            </div>
            <Button
              type="primary"
              icon={<IconPlus />}
              onClick={() => setCreateVisible(true)}
              style={{ borderRadius: 8 }}
            >
              添加连接
            </Button>
          </div>
        ) : (
          <div className="integration-grid">
            {connections.map((conn) => {
              const provider = PROVIDER_CONFIG[conn.providerType] || { icon: '🔌', bg: '#F3F4F6', label: conn.providerType };
              const status = STATUS_CONFIG[conn.status] || { label: conn.status, cssClass: 'env-health-status--degraded' };
              return (
                <div key={conn.id} className="integration-card">
                  <div className="integration-card-header">
                    <div
                      className="integration-card-icon"
                      style={{ background: provider.bg, fontSize: 22 }}
                    >
                      {provider.icon}
                    </div>
                    <div style={{ flex: 1, minWidth: 0 }}>
                      <div className="integration-card-name">{conn.name}</div>
                      <div className="integration-card-desc">
                        {provider.label} · {conn.serverUrl}
                      </div>
                    </div>
                    <span className={`env-health-status ${status.cssClass}`}>
                      {status.label}
                    </span>
                  </div>

                  <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
                    <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: 13 }}>
                      <span style={{ color: 'var(--muted-foreground)' }}>Token</span>
                      <span style={{ fontFamily: 'var(--font-mono)', color: 'var(--zcid-text-4)', fontSize: 12 }}>
                        {conn.tokenMask}
                      </span>
                    </div>
                    {conn.description && (
                      <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: 13 }}>
                        <span style={{ color: 'var(--muted-foreground)' }}>Description</span>
                        <span style={{ color: 'var(--foreground)' }}>{conn.description}</span>
                      </div>
                    )}
                    <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: 13 }}>
                      <span style={{ color: 'var(--muted-foreground)' }}>Created</span>
                      <span style={{ color: 'var(--foreground)', fontSize: 12 }}>{conn.createdAt}</span>
                    </div>
                  </div>

                  <div className="integration-card-footer">
                    <Tag size="small" style={{
                      borderRadius: 20, background: provider.bg, border: 'none', fontWeight: 500,
                    }}>
                      {provider.label}
                    </Tag>
                    <Space size={4}>
                      <Tooltip content="测试连接">
                        <Button
                          type="text" size="small"
                          icon={<IconPlayArrow />}
                          loading={testingId === conn.id}
                          onClick={() => handleTest(conn.id)}
                          style={{ color: 'var(--success)' }}
                        />
                      </Tooltip>
                      <Tooltip content="复制 Webhook Secret">
                        <Button
                          type="text" size="small"
                          icon={<IconCopy />}
                          onClick={() => handleCopyWebhookSecret(conn.id)}
                        />
                      </Tooltip>
                      <Tooltip content="编辑">
                        <Button
                          type="text" size="small"
                          icon={<IconEdit />}
                          onClick={() => setEditItem(conn)}
                          style={{ color: 'var(--primary)' }}
                        />
                      </Tooltip>
                      <Popconfirm title="确认删除此连接？" onOk={() => handleDelete(conn.id)}>
                        <Tooltip content="删除">
                          <Button type="text" size="small" status="danger" icon={<IconDelete />} />
                        </Tooltip>
                      </Popconfirm>
                    </Space>
                  </div>
                </div>
              );
            })}
          </div>
        )}

        <ConnectionFormModal
          visible={createVisible}
          onClose={() => setCreateVisible(false)}
          onSubmit={handleCreate}
        />
        {editItem && (
          <ConnectionFormModal
            visible={!!editItem}
            onClose={() => setEditItem(null)}
            onSubmit={handleEdit}
            editMode
            initialValues={{ name: editItem.name, description: editItem.description }}
          />
        )}
      </div>
    </AppLayout>
  );
}
