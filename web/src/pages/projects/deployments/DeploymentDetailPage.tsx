import { useEffect, useState, useCallback, type ReactNode } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import {
  Skeleton,
  Message,
  Button,
  Space,
  Descriptions,
  Tag,
  Card,
  Popconfirm,
} from '@arco-design/web-react';
import { IconArrowLeft, IconRefresh, IconRotateLeft } from '@arco-design/web-react/icon';
import {
  fetchDeployment,
  fetchDeployStatus,
  resyncDeploy,
  rollbackDeploy,
  type Deployment,
} from '../../../services/deployment';

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

export function DeploymentDetailPage() {
  const { id: projectId, deployId } = useParams<{ id: string; deployId: string }>();
  const navigate = useNavigate();
  const [deploy, setDeploy] = useState<Deployment | null>(null);
  const [loading, setLoading] = useState(false);

  const loadData = useCallback(async () => {
    if (!projectId || !deployId) return;
    setLoading(true);
    try {
      const d = await fetchDeployment(projectId, deployId);
      setDeploy(d);
    } catch {
      Message.error('加载部署详情失败');
    } finally {
      setLoading(false);
    }
  }, [projectId, deployId]);

  useEffect(() => {
    loadData();
  }, [loadData]);

  const handleRefresh = async () => {
    if (!projectId || !deployId) return;
    try {
      const d = await fetchDeployStatus(projectId, deployId);
      setDeploy(d);
      Message.success('状态已刷新');
    } catch {
      Message.error('刷新状态失败');
    }
  };

  const handleResync = async () => {
    if (!projectId || !deployId) return;
    try {
      const d = await resyncDeploy(projectId, deployId);
      setDeploy(d);
      Message.success('已触发重新同步');
    } catch (err: any) {
      Message.error(err.response?.data?.message || '重新同步失败');
    }
  };

  const handleRollback = async () => {
    if (!projectId || !deployId) return;
    try {
      const d = await rollbackDeploy(projectId, deployId);
      setDeploy(d);
      Message.success('已触发回滚');
      loadData();
    } catch (err: any) {
      Message.error(err.response?.data?.message || '回滚失败');
    }
  };

  if (loading && !deploy) {
    return (
      <div className="page-container">
        <Skeleton text={{ rows: 4 }} animation />
      </div>
    );
  }

  if (!deploy) {
    return (
      <div className="page-container">
        <Message type="error">部署记录不存在</Message>
      </div>
    );
  }

  return (
    <div className="page-container">
      <div className="page-header">
        <Space>
          <Button type="text" icon={<IconArrowLeft />} onClick={() => navigate(`/projects/${projectId}/deployments`)}>
            返回
          </Button>
          <h3 className="page-title">部署详情</h3>
        </Space>
        <Space>
          <Button type="outline" icon={<IconRefresh />} onClick={handleRefresh}>
            刷新状态
          </Button>
          <Button type="outline" icon={<IconRefresh />} onClick={handleResync}>
            重新同步
          </Button>
          <Popconfirm title="确定回滚到上一个版本？" onOk={handleRollback}>
            <Button type="outline" status="warning" icon={<IconRotateLeft />}>
              回滚
            </Button>
          </Popconfirm>
        </Space>
      </div>

      <Card title="基本信息" style={{ marginBottom: 16 }}>
        <Descriptions
          column={1}
          data={[
            { label: 'ID', value: deploy.id },
            { label: '镜像', value: deploy.image },
            { label: '环境 ID', value: deploy.environmentId },
            {
              label: '状态',
              value: <Tag color={statusColors[deploy.status] || 'default'}>{statusLabels[deploy.status] || deploy.status}</Tag>,
            },
            { label: '同步状态', value: deploy.syncStatus ?? '-' },
            { label: '健康状态', value: deploy.healthStatus ?? '-' },
            { label: 'ArgoCD 应用', value: deploy.argoAppName ?? '-' },
            { label: '部署人', value: deploy.deployedBy },
            deploy.errorMessage ? { label: '错误信息', value: deploy.errorMessage } : null,
            {
              label: '开始时间',
              value: deploy.startedAt ? new Date(deploy.startedAt).toLocaleString() : '-',
            },
            {
              label: '完成时间',
              value: deploy.finishedAt ? new Date(deploy.finishedAt).toLocaleString() : '-',
            },
            { label: '创建时间', value: new Date(deploy.createdAt).toLocaleString() },
          ].filter(Boolean) as { label: string; value: ReactNode }[]}
        />
      </Card>
    </div>
  );
}
