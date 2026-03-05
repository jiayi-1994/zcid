import { Button, Table, Message, Popconfirm, Modal, Form, Input } from '@arco-design/web-react';
import { IconPlus } from '@arco-design/web-react/icon';
import { useState, useEffect, useCallback } from 'react';
import { useParams } from 'react-router-dom';
import { useAuthStore } from '../../../stores/auth';
import { fetchServices, createService, deleteService, type ServiceItem } from '../../../services/project';

const FormItem = Form.Item;

export function ServiceListPage() {
  const { id: projectId } = useParams<{ id: string }>();
  const [svcs, setSvcs] = useState<ServiceItem[]>([]);
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
      const data = await fetchServices(projectId, p, 20);
      setSvcs(data.items ?? []);
      setTotal(data.total);
    } catch (err: any) {
      Message.error(err.response?.data?.message || '加载服务列表失败');
    } finally {
      setLoading(false);
    }
  }, [projectId]);

  useEffect(() => { loadData(page); }, [page, loadData]);

  const handleCreate = async () => {
    try {
      const values = await form.validate();
      setSubmitLoading(true);
      await createService(projectId!, values.name, values.description || '', values.repoUrl || '');
      Message.success('服务创建成功');
      form.resetFields();
      setModalVisible(false);
      loadData(page);
    } catch (err: any) {
      if (err.response?.data?.message) Message.error(err.response.data.message);
    } finally {
      setSubmitLoading(false);
    }
  };

  const handleDelete = async (svcId: string) => {
    try {
      await deleteService(projectId!, svcId);
      Message.success('服务已删除');
      loadData(page);
    } catch (err: any) {
      Message.error(err.response?.data?.message || '删除失败');
    }
  };

  const columns = [
    { title: '服务名称', dataIndex: 'name' },
    { title: '描述', dataIndex: 'description' },
    { title: '仓库地址', dataIndex: 'repoUrl', render: (url: string) => url || '-' },
    { title: '创建时间', dataIndex: 'createdAt' },
    {
      title: '操作',
      render: (_: any, record: ServiceItem) => isAdmin ? (
        <Popconfirm title="确定删除？" onOk={() => handleDelete(record.id)}>
          <Button type="text" size="small" status="danger">删除</Button>
        </Popconfirm>
      ) : null,
    },
  ];

  return (
    <div className="page-container">
      <div className="page-header">
        <h3 className="page-title">服务管理</h3>
        {isAdmin && (
          <Button type="primary" icon={<IconPlus />} size="small" onClick={() => setModalVisible(true)}>
            新建服务
          </Button>
        )}
      </div>
      <Table
        columns={columns} data={svcs} loading={loading} rowKey="id"
        pagination={{ current: page, total, pageSize: 20, onChange: setPage }}
        noDataElement={<div className="empty-state">暂无服务</div>}
      />
      <Modal
        title="新建服务" visible={modalVisible}
        onOk={handleCreate}
        onCancel={() => { form.resetFields(); setModalVisible(false); }}
        confirmLoading={submitLoading} unmountOnExit
      >
        <Form form={form} layout="vertical">
          <FormItem label="服务名称" field="name" rules={[{ required: true, message: '请输入服务名称' }]}>
            <Input placeholder="如 api-gateway" />
          </FormItem>
          <FormItem label="描述" field="description">
            <Input.TextArea placeholder="服务描述" rows={2} />
          </FormItem>
          <FormItem label="仓库地址" field="repoUrl">
            <Input placeholder="如 https://github.com/org/repo" />
          </FormItem>
        </Form>
      </Modal>
    </div>
  );
}
