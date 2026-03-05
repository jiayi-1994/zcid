import {
  Button,
  Form,
  Input,
  Message,
  Modal,
  Popconfirm,
  Select,
  Space,
  Switch,
  Table,
  Tag,
} from '@arco-design/web-react';
import { IconDelete, IconEdit, IconPlus } from '@arco-design/web-react/icon';
import { useCallback, useEffect, useState } from 'react';
import { useParams } from 'react-router-dom';
import {
  createNotificationRule,
  deleteNotificationRule,
  fetchNotificationRules,
  updateNotificationRule,
  type NotificationRule,
} from '../../../services/notification';

const EVENT_TYPES = [
  { value: 'build_success', label: '构建成功' },
  { value: 'build_failed', label: '构建失败' },
  { value: 'deploy_success', label: '部署成功' },
  { value: 'deploy_failed', label: '部署失败' },
];

const EVENT_LABELS: Record<string, { text: string; color: string }> = {
  build_success: { text: '构建成功', color: 'green' },
  build_failed: { text: '构建失败', color: 'red' },
  deploy_success: { text: '部署成功', color: 'blue' },
  deploy_failed: { text: '部署失败', color: 'orange' },
};

export function NotificationRulesPage() {
  const { id: projectId } = useParams<{ id: string }>();
  const [rules, setRules] = useState<NotificationRule[]>([]);
  const [loading, setLoading] = useState(false);
  const [modalVisible, setModalVisible] = useState(false);
  const [editRule, setEditRule] = useState<NotificationRule | null>(null);
  const [form] = Form.useForm();

  const load = useCallback(async () => {
    if (!projectId) return;
    setLoading(true);
    try {
      const data = await fetchNotificationRules(projectId);
      setRules(data.items || []);
    } catch {
      Message.error('加载通知规则失败');
    } finally {
      setLoading(false);
    }
  }, [projectId]);

  useEffect(() => { load(); }, [load]);

  const openCreate = () => {
    setEditRule(null);
    form.resetFields();
    form.setFieldValue('enabled', true);
    setModalVisible(true);
  };

  const openEdit = (rule: NotificationRule) => {
    setEditRule(rule);
    form.setFieldsValue({
      name: rule.name,
      eventType: rule.eventType,
      webhookUrl: rule.webhookUrl,
      enabled: rule.enabled,
    });
    setModalVisible(true);
  };

  const handleSubmit = async () => {
    if (!projectId) return;
    try {
      const values = await form.validate();
      if (editRule) {
        await updateNotificationRule(projectId, editRule.id, values);
        Message.success('通知规则已更新');
      } else {
        await createNotificationRule(projectId, values);
        Message.success('通知规则已创建');
      }
      setModalVisible(false);
      load();
    } catch {
      /* validation error */
    }
  };

  const handleDelete = async (ruleId: string) => {
    if (!projectId) return;
    try {
      await deleteNotificationRule(projectId, ruleId);
      Message.success('通知规则已删除');
      load();
    } catch {
      Message.error('删除失败');
    }
  };

  const handleToggle = async (rule: NotificationRule, enabled: boolean) => {
    if (!projectId) return;
    try {
      await updateNotificationRule(projectId, rule.id, { enabled });
      load();
    } catch {
      Message.error('更新状态失败');
    }
  };

  const columns = [
    { title: '名称', dataIndex: 'name', width: 180 },
    {
      title: '事件类型',
      dataIndex: 'eventType',
      width: 120,
      render: (val: string) => {
        const cfg = EVENT_LABELS[val] || { text: val, color: 'gray' };
        return <Tag color={cfg.color}>{cfg.text}</Tag>;
      },
    },
    {
      title: 'Webhook URL',
      dataIndex: 'webhookUrl',
      ellipsis: true,
      render: (val: string) => (
        <span style={{ fontFamily: 'monospace', fontSize: 12 }}>{val}</span>
      ),
    },
    {
      title: '状态',
      dataIndex: 'enabled',
      width: 80,
      render: (val: boolean, record: NotificationRule) => (
        <Switch size="small" checked={val} onChange={(v) => handleToggle(record, v)} />
      ),
    },
    { title: '创建时间', dataIndex: 'createdAt', width: 180 },
    {
      title: '操作',
      width: 120,
      render: (_: unknown, record: NotificationRule) => (
        <Space size="mini">
          <Button type="text" size="small" icon={<IconEdit />} onClick={() => openEdit(record)} />
          <Popconfirm title="确认删除该通知规则？" onOk={() => handleDelete(record.id)}>
            <Button type="text" size="small" status="danger" icon={<IconDelete />} />
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <div className="page-container">
      <div className="page-header">
        <h3 className="page-title">通知规则</h3>
        <Button type="primary" icon={<IconPlus />} onClick={openCreate}>
          创建规则
        </Button>
      </div>
      <Table
        columns={columns}
        data={rules}
        rowKey="id"
        loading={loading}
        pagination={{ pageSize: 20, showTotal: true }}
      />
      <Modal
        title={editRule ? '编辑通知规则' : '创建通知规则'}
        visible={modalVisible}
        onCancel={() => setModalVisible(false)}
        onOk={handleSubmit}
        unmountOnExit
      >
        <Form form={form} layout="vertical">
          <Form.Item label="名称" field="name" rules={[{ required: true, message: '请输入名称' }]}>
            <Input placeholder="例如：构建失败通知" />
          </Form.Item>
          <Form.Item label="事件类型" field="eventType" rules={[{ required: true, message: '请选择事件类型' }]}>
            <Select options={EVENT_TYPES} placeholder="选择触发事件" />
          </Form.Item>
          <Form.Item
            label="Webhook URL"
            field="webhookUrl"
            rules={[
              { required: true, message: '请输入 Webhook URL' },
              { match: /^https?:\/\//, message: '请输入合法 URL' },
            ]}
          >
            <Input placeholder="https://hooks.example.com/callback" />
          </Form.Item>
          <Form.Item label="启用" field="enabled" triggerPropName="checked">
            <Switch />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
}

export default NotificationRulesPage;
