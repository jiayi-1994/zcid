import { Button, Table, Message, Modal, Form, Input, Tag, Select } from '@arco-design/web-react';
import { IconPlus } from '@arco-design/web-react/icon';
import { useState, useEffect, useCallback } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import {
  fetchDeployments,
  triggerDeploy,
  type DeploymentSummary,
  type DeploymentList,
} from '../../../services/deployment';
import { fetchEnvironments, type EnvironmentItem } from '../../../services/project';
import { extractErrorMessage } from '../../../services/http';

const FormItem = Form.Item;

const statusColors: Record<string, string> = {
  pending: 'gray',
  syncing: 'arcoblue',
  healthy: 'green',
  degraded: 'orange',
  failed: 'red',
  rolled_back: 'gray',
};

const statusLabels: Record<string, string> = {
  pending: '待部署',
  syncing: '同步中',
  healthy: '健康',
  degraded: '异常',
  failed: '失败',
  rolled_back: '已回滚',
};

export function DeploymentListPage() {
  const { id: projectId } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [data, setData] = useState<DeploymentList>({ items: [], total: 0, page: 1, pageSize: 20 });
  const [loading, setLoading] = useState(false);
  const [page, setPage] = useState(1);
  const [modalVisible, setModalVisible] = useState(false);
  const [form] = Form.useForm();
  const [submitLoading, setSubmitLoading] = useState(false);
  const [envs, setEnvs] = useState<EnvironmentItem[]>([]);

  const loadData = useCallback(async (p: number) => {
    if (!projectId) return;
    setLoading(true);
    try {
      const result = await fetchDeployments(projectId, p, 20);
      setData(result);
    } catch {
      Message.error('加载部署列表失败');
    } finally {
      setLoading(false);
    }
  }, [projectId]);

  const loadEnvs = useCallback(async () => {
    if (!projectId) return;
    try {
      const r = await fetchEnvironments(projectId, 1, 100);
      setEnvs(r.items ?? []);
    } catch {
      Message.error('加载环境列表失败');
    }
  }, [projectId]);

  useEffect(() => {
    loadData(page);
  }, [page, loadData]);

  useEffect(() => {
    if (modalVisible && projectId) loadEnvs();
  }, [modalVisible, projectId, loadEnvs]);

  const handleTrigger = () => {
    setModalVisible(true);
  };

  const handleSubmit = async () => {
    try {
      const values = await form.validate();
      if (!projectId) return;
      setSubmitLoading(true);
      await triggerDeploy(projectId, {
        environmentId: values.environmentId,
        image: values.image,
        pipelineRunId: values.pipelineRunId || undefined,
      });
      Message.success('部署已触发');
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

  const columns = [
    { title: '镜像', dataIndex: 'image' },
    { title: '环境', dataIndex: 'environmentId', render: (id: string) => id || '-' },
    {
      title: '状态',
      dataIndex: 'status',
      render: (s: string) => <Tag size="small" color={statusColors[s] || 'default'}>{statusLabels[s] || s}</Tag>,
    },
    { title: '同步状态', dataIndex: 'syncStatus', render: (v: string) => v ?? '-' },
    { title: '健康状态', dataIndex: 'healthStatus', render: (v: string) => v ?? '-' },
    { title: '部署人', dataIndex: 'deployedBy' },
    { title: '时间', dataIndex: 'createdAt', render: (t: string) => new Date(t).toLocaleString() },
    {
      title: '操作',
      render: (_: unknown, record: DeploymentSummary) => (
        <Button
          type="text"
          size="small"
          style={{ color: 'var(--zcid-primary)' }}
          onClick={() => navigate(`/projects/${projectId}/deployments/${record.id}`)}
        >
          详情
        </Button>
      ),
    },
  ];

  return (
    <div className="page-container">
      <div className="page-header">
        <h3 className="page-title">部署管理</h3>
        <Button type="primary" icon={<IconPlus />} size="small" onClick={handleTrigger}>
          部署
        </Button>
      </div>
      <div className="table-card">
        <Table
          columns={columns}
          data={data.items}
          loading={loading}
          rowKey="id"
          border={false}
          pagination={{ current: page, total: data.total, pageSize: 20, onChange: setPage, style: { padding: '12px 16px' } }}
          noDataElement={
            <div className="empty-state">
              <div className="empty-state-title">暂无部署</div>
              <div className="empty-state-desc">触发部署后这里会显示列表</div>
            </div>
          }
        />
      </div>
      <Modal
        title="触发部署"
        visible={modalVisible}
        onOk={handleSubmit}
        onCancel={() => {
          form.resetFields();
          setModalVisible(false);
        }}
        confirmLoading={submitLoading}
        unmountOnExit
      >
        <Form form={form} layout="vertical">
          <FormItem label="环境" field="environmentId" rules={[{ required: true, message: '请选择环境' }]}>
            <Select
              placeholder="选择环境"
              options={envs.map((e) => ({ label: `${e.name} (${e.namespace})`, value: e.id }))}
            />
          </FormItem>
          <FormItem label="镜像" field="image" rules={[{ required: true, message: '请输入镜像地址' }]}>
            <Input placeholder="如 nginx:latest 或 registry.io/app:v1" />
          </FormItem>
          <FormItem label="流水线运行 ID (可选)" field="pipelineRunId">
            <Input placeholder="关联的 Pipeline Run ID" />
          </FormItem>
        </Form>
      </Modal>
    </div>
  );
}
