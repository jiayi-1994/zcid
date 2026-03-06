import { Badge, Button, Message, Popconfirm, Space, Table, Tag, Tooltip } from '@arco-design/web-react';
import { IconCopy, IconDelete, IconEdit, IconPlayArrow, IconPlus, IconRefresh } from '@arco-design/web-react/icon';
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

const STATUS_CONFIG: Record<string, { color: string; text: string }> = {
  connected: { color: 'green', text: '已连接' },
  disconnected: { color: 'red', text: '已断开' },
  token_expired: { color: 'orange', text: 'Token 过期' },
};

const PROVIDER_LABELS: Record<string, string> = {
  gitlab: 'GitLab',
  github: 'GitHub',
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

  const columns = [
    {
      title: '名称',
      dataIndex: 'name',
      width: 160,
      render: (name: string) => <span style={{ fontWeight: 500 }}>{name}</span>,
    },
    {
      title: 'Provider',
      dataIndex: 'providerType',
      width: 100,
      render: (val: string) => (
        <Tag size="small">{PROVIDER_LABELS[val] || val}</Tag>
      ),
    },
    {
      title: 'Server URL',
      dataIndex: 'serverUrl',
      ellipsis: true,
      render: (url: string) => <span style={{ color: 'var(--zcid-text-3)', fontFamily: 'var(--zcid-font-mono)', fontSize: 13 }}>{url}</span>,
    },
    {
      title: 'Token',
      dataIndex: 'tokenMask',
      width: 120,
      render: (val: string) => (
        <span style={{ fontFamily: 'var(--zcid-font-mono)', color: 'var(--zcid-text-4)', fontSize: 13 }}>{val}</span>
      ),
    },
    {
      title: '状态',
      dataIndex: 'status',
      width: 100,
      render: (val: string) => {
        const cfg = STATUS_CONFIG[val] || { color: 'gray', text: val };
        return <Badge color={cfg.color} text={cfg.text} />;
      },
    },
    {
      title: '描述',
      dataIndex: 'description',
      ellipsis: true,
      render: (v: string) => <span style={{ color: 'var(--zcid-text-3)' }}>{v || '-'}</span>,
    },
    { title: '创建时间', dataIndex: 'createdAt', width: 180 },
    {
      title: '操作',
      width: 200,
      render: (_: unknown, record: GitConnection) => (
        <Space size="mini">
          <Tooltip content="测试连接">
            <Button
              type="text"
              size="small"
              icon={<IconPlayArrow />}
              loading={testingId === record.id}
              onClick={() => handleTest(record.id)}
              style={{ color: 'var(--zcid-success)' }}
            />
          </Tooltip>
          <Tooltip content="复制 Webhook Secret">
            <Button
              type="text"
              size="small"
              icon={<IconCopy />}
              onClick={() => handleCopyWebhookSecret(record.id)}
            />
          </Tooltip>
          <Tooltip content="编辑">
            <Button
              type="text"
              size="small"
              icon={<IconEdit />}
              onClick={() => setEditItem(record)}
              style={{ color: 'var(--zcid-primary)' }}
            />
          </Tooltip>
          <Popconfirm title="确认删除此连接？" onOk={() => handleDelete(record.id)}>
            <Tooltip content="删除">
              <Button type="text" size="small" status="danger" icon={<IconDelete />} />
            </Tooltip>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <AppLayout>
      <div className="page-container">
        <div className="page-header">
          <div>
            <h3 className="page-title">集成管理</h3>
            <p className="page-subtitle">管理 Git 仓库连接和 Webhook 配置</p>
          </div>
          <Space>
            <Button icon={<IconRefresh />} onClick={loadData}>刷新</Button>
            <Button type="primary" icon={<IconPlus />} onClick={() => setCreateVisible(true)}>
              添加连接
            </Button>
          </Space>
        </div>
        <div className="table-card">
          <Table
            columns={columns}
            data={connections}
            rowKey="id"
            loading={loading}
            border={false}
            pagination={{ pageSize: 20, showTotal: true, sizeCanChange: false, style: { padding: '12px 16px' } }}
          />
        </div>
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
