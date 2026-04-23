import { useEffect, useState, useCallback, type ReactNode } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import {
  Skeleton,
  Message,
  Button,
  Space,
  Descriptions,
  Popconfirm,
} from '@arco-design/web-react';
import { IconArrowLeft, IconRefresh, IconRotateLeft, IconSync } from '@arco-design/web-react/icon';
import {
  fetchDeployment,
  fetchDeployStatus,
  resyncDeploy,
  rollbackDeploy,
  type Deployment,
} from '../../../services/deployment';
import { extractErrorMessage } from '../../../services/http';

const STATUS_CLS: Record<string, string> = {
  pending: 'pipeline-status-badge--pending',
  syncing: 'pipeline-status-badge--running',
  healthy: 'pipeline-status-badge--success',
  degraded: 'pipeline-status-badge--cancelled',
  failed: 'pipeline-status-badge--failed',
  rolled_back: 'pipeline-status-badge--pending',
};

const STATUS_LABEL: Record<string, string> = {
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

  useEffect(() => { loadData(); }, [loadData]);

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
    } catch (err: unknown) {
      Message.error(extractErrorMessage(err, '重新同步失败'));
    }
  };

  const handleRollback = async () => {
    if (!projectId || !deployId) return;
    try {
      const d = await rollbackDeploy(projectId, deployId);
      setDeploy(d);
      Message.success('已触发回滚');
      loadData();
    } catch (err: unknown) {
      Message.error(extractErrorMessage(err, '回滚失败'));
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
        <div className="empty-state">
          <div className="empty-state-title">部署记录不存在</div>
        </div>
      </div>
    );
  }

  const statusCls = STATUS_CLS[deploy.status] || 'pipeline-status-badge--pending';
  const statusLabel = STATUS_LABEL[deploy.status] || deploy.status;

  return (
    <div className="page-container">
      <div className="page-header">
        <div>
          <Button
            type="text"
            icon={<IconArrowLeft />}
            onClick={() => navigate(`/projects/${projectId}/deployments`)}
            style={{ marginBottom: 4 }}
          >
            返回列表
          </Button>
          <h1 className="page-title">部署详情</h1>
          <p className="page-subtitle">
            <code className="mono">{deploy.image}</code>
          </p>
        </div>
        <Space>
          <Button icon={<IconRefresh />} onClick={handleRefresh}>
            刷新状态
          </Button>
          <Button icon={<IconSync />} onClick={handleResync}>
            重新同步
          </Button>
          <Popconfirm title="确定回滚到上一个版本？" onOk={handleRollback}>
            <Button status="warning" icon={<IconRotateLeft />}>
              回滚
            </Button>
          </Popconfirm>
        </Space>
      </div>

      <div className="zcid-card" style={{ padding: 'var(--space-6)' }}>
        <div
          style={{
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
            marginBottom: 'var(--space-4)',
          }}
        >
          <h2 className="section-title" style={{ margin: 0 }}>基本信息</h2>
          <span className={`pipeline-status-badge ${statusCls}`}>{statusLabel}</span>
        </div>
        <Descriptions
          column={1}
          data={
            [
              { label: 'ID', value: <code className="mono">{deploy.id}</code> },
              { label: '镜像', value: <code className="mono">{deploy.image}</code> },
              { label: '环境 ID', value: deploy.environmentId },
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
            ].filter(Boolean) as { label: string; value: ReactNode }[]
          }
        />
      </div>
    </div>
  );
}
