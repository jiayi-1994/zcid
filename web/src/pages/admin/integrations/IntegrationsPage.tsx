import { Button, Message, Popconfirm, Space, Tooltip } from '@arco-design/web-react';
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

const PROVIDER_CONFIG: Record<string, { icon: string; label: string }> = {
  github: { icon: '🐙', label: 'GitHub' },
  gitlab: { icon: '🦊', label: 'GitLab' },
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
        <div className="page-header">
          <div>
            <div className="breadcrumb">Settings › Integrations</div>
            <h1 className="page-title">Integration Management</h1>
            <p className="page-subtitle">
              Connect and manage your external CI/CD toolchain. Monitor connectivity status and streamline deployment workflows.
            </p>
          </div>
          <Space size={8}>
            <Button icon={<IconRefresh />} onClick={loadData}>
              刷新
            </Button>
            <Button
              type="primary"
              icon={<IconPlus />}
              onClick={() => setCreateVisible(true)}
              size="large"
            >
              Connect New Service
            </Button>
          </Space>
        </div>

        {/* Metrics */}
        <div className="metrics-grid" style={{ gridTemplateColumns: 'repeat(3, 1fr)' }}>
          <div className="metric-card">
            <span className="metric-card-label">Total Integrations</span>
            <span className="metric-card-value">{connections.length}</span>
            <span className="metric-card-sub">configured providers</span>
          </div>
          <div className="metric-card">
            <span className="metric-card-label">Connected</span>
            <span className="metric-card-value">{connectedCount}</span>
            <span className="metric-card-sub">healthy & streaming</span>
          </div>
          <div className="metric-card">
            <span className="metric-card-label">Needs Attention</span>
            <span className="metric-card-value">{connections.length - connectedCount}</span>
            <span className="metric-card-sub">
              {connections.length - connectedCount > 0 ? '需检查连接' : 'all green'}
            </span>
          </div>
        </div>

        {/* Integration Cards */}
        {loading ? (
          <div style={{ padding: '60px 0', textAlign: 'center', color: 'var(--muted-foreground)' }}>
            加载中...
          </div>
        ) : connections.length === 0 ? (
          <div className="zcid-card empty-state">
            <div className="empty-state-title">暂无集成连接</div>
            <div className="empty-state-desc">添加 Git 仓库连接，开始管理源代码</div>
            <Button type="primary" icon={<IconPlus />} onClick={() => setCreateVisible(true)}>
              添加连接
            </Button>
          </div>
        ) : (
          <div className="integration-grid">
            {connections.map((conn) => {
              const provider = PROVIDER_CONFIG[conn.providerType] || { icon: '🔌', label: conn.providerType };
              const statusDotCls = conn.status === 'connected' ? 'integration-status-dot--connected'
                : conn.status === 'token_expired' ? 'integration-status-dot--pending'
                : 'integration-status-dot--failed';
              const statusLabel = STATUS_CONFIG[conn.status]?.label || conn.status;
              return (
                <div key={conn.id} className="integration-card">
                  <div className="integration-card-header">
                    <div className="integration-card-icon">
                      {provider.icon}
                    </div>
                    <div style={{ flex: 1, minWidth: 0 }}>
                      <div className="integration-card-name">{conn.name}</div>
                      <div className="integration-card-desc">
                        {provider.label} · {conn.serverUrl}
                      </div>
                    </div>
                  </div>

                  <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
                    <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: 13 }}>
                      <span style={{ color: 'var(--on-surface-variant)' }}>Token</span>
                      <span style={{ fontFamily: 'var(--font-mono)', color: 'var(--on-surface-variant)', fontSize: 12 }}>
                        {conn.tokenMask}
                      </span>
                    </div>
                    {conn.description && (
                      <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: 13 }}>
                        <span style={{ color: 'var(--on-surface-variant)' }}>Description</span>
                        <span style={{ color: 'var(--on-surface)' }}>{conn.description}</span>
                      </div>
                    )}
                    <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: 13 }}>
                      <span style={{ color: 'var(--on-surface-variant)' }}>Created</span>
                      <span style={{ color: 'var(--on-surface)', fontSize: 12, fontFamily: 'var(--font-mono)' }}>{conn.createdAt}</span>
                    </div>
                  </div>

                  <div className="integration-card-footer">
                    <span className={`integration-status-dot ${statusDotCls}`}>
                      {statusLabel}
                    </span>
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
