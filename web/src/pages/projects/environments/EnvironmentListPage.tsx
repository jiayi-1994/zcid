import { Button, Table, Message, Popconfirm, Modal, Form, Input, Tag } from '@arco-design/web-react';
import { IconPlus } from '@arco-design/web-react/icon';
import { useState, useEffect, useCallback } from 'react';
import { useParams } from 'react-router-dom';
import { useAuthStore } from '../../../stores/auth';
import { fetchEnvironments, createEnvironment, deleteEnvironment, type EnvironmentItem } from '../../../services/project';
import { extractErrorMessage } from '../../../services/http';

const FormItem = Form.Item;

export function EnvironmentListPage() {
  const { id: projectId } = useParams<{ id: string }>();
  const [envs, setEnvs] = useState<EnvironmentItem[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
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

  const columns = [
    { title: '名称', dataIndex: 'name' },
    { title: 'Namespace', dataIndex: 'namespace', render: (ns: string) => <Tag size="small">{ns}</Tag> },
    { title: '描述', dataIndex: 'description' },
    { title: '创建时间', dataIndex: 'createdAt' },
    {
      title: '操作',
      render: (_: unknown, record: EnvironmentItem) => isAdmin ? (
        <Popconfirm title="确定删除？" onOk={() => handleDelete(record.id)}>
          <Button type="text" size="small" status="danger">删除</Button>
        </Popconfirm>
      ) : null,
    },
  ];

  return (
    <div className="page-container">
      <div className="page-header">
        <h3 className="page-title">环境管理</h3>
        {isAdmin && (
          <Button type="primary" icon={<IconPlus />} size="small" onClick={() => setModalVisible(true)}>
            新建环境
          </Button>
        )}
      </div>
      <div className="table-card">
        <Table
          columns={columns} data={envs} loading={loading} rowKey="id" border={false}
          pagination={{ current: page, total, pageSize: 20, onChange: setPage, style: { padding: '12px 16px' } }}
          noDataElement={
            <div className="empty-state">
              <div className="empty-state-title">暂无环境</div>
              <div className="empty-state-desc">创建第一个环境，开始管理部署目标</div>
            </div>
          }
        />
      </div>
      <Modal
        title="新建环境" visible={modalVisible}
        onOk={handleCreate}
        onCancel={() => { form.resetFields(); setModalVisible(false); }}
        confirmLoading={submitLoading} unmountOnExit
      >
        <Form form={form} layout="vertical">
          <FormItem label="环境名称" field="name" rules={[{ required: true, message: '请输入环境名称' }]}>
            <Input placeholder="如 dev / staging / prod" />
          </FormItem>
          <FormItem label="K8s Namespace" field="namespace" rules={[{ required: true, message: '请输入 Namespace' }]}>
            <Input placeholder="如 zcid-dev" />
          </FormItem>
          <FormItem label="描述" field="description">
            <Input.TextArea placeholder="环境描述" rows={2} />
          </FormItem>
        </Form>
      </Modal>
    </div>
  );
}
