import { useCallback, useEffect, useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { Message } from '@arco-design/web-react';
import { fetchDeployment, fetchDeployStatus, resyncDeploy, rollbackDeploy, type Deployment } from '../../../services/deployment';
import { extractErrorMessage } from '../../../services/http';
import { PageHeader } from '../../../components/ui/PageHeader';
import { Card } from '../../../components/ui/Card';
import { Btn } from '../../../components/ui/Btn';
import { StatusBadge } from '../../../components/ui/StatusBadge';
import { IArrL, IRefresh, ISync } from '../../../components/ui/icons';

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
      <>
        <PageHeader crumb="Project · Delivery" title="部署详情" />
        <div style={{ padding: '40px 24px', color: 'var(--z-400)' }}>加载中...</div>
      </>
    );
  }

  if (!deploy) {
    return (
      <>
        <PageHeader crumb="Project · Delivery" title="部署详情" />
        <div style={{ padding: '40px 24px', color: 'var(--z-400)' }}>部署记录不存在</div>
      </>
    );
  }

  const rows: { label: string; value: React.ReactNode }[] = [
    { label: 'ID', value: <span className="mono" style={{ fontSize: 11.5 }}>{deploy.id}</span> },
    { label: '镜像', value: <span className="code" style={{ fontSize: 11.5 }}>{deploy.image}</span> },
    { label: '环境 ID', value: <span className="sub">{deploy.environmentId}</span> },
    { label: '同步状态', value: <span className="sub">{deploy.syncStatus ?? '-'}</span> },
    { label: '健康状态', value: <span className="sub">{deploy.healthStatus ?? '-'}</span> },
    { label: 'ArgoCD 应用', value: <span className="mono sub" style={{ fontSize: 11.5 }}>{deploy.argoAppName ?? '-'}</span> },
    { label: '部署人', value: <span className="sub">{deploy.deployedBy}</span> },
    ...(deploy.errorMessage ? [{ label: '错误信息', value: <span style={{ color: 'var(--red-ink)', fontSize: 12 }}>{deploy.errorMessage}</span> }] : []),
    { label: '开始时间', value: <span className="mono sub" style={{ fontSize: 11 }}>{deploy.startedAt ? new Date(deploy.startedAt).toLocaleString() : '-'}</span> },
    { label: '完成时间', value: <span className="mono sub" style={{ fontSize: 11 }}>{deploy.finishedAt ? new Date(deploy.finishedAt).toLocaleString() : '-'}</span> },
    { label: '创建时间', value: <span className="mono sub" style={{ fontSize: 11 }}>{new Date(deploy.createdAt).toLocaleString()}</span> },
  ];

  return (
    <>
      <PageHeader
        crumb="Project · Delivery"
        title="部署详情"
        sub={<span className="code" style={{ fontSize: 12 }}>{deploy.image}</span>}
        actions={
          <>
            <Btn size="sm" icon={<IArrL size={13} />} onClick={() => navigate(`/projects/${projectId}/deployments`)}>返回</Btn>
            <Btn size="sm" icon={<IRefresh size={13} />} onClick={handleRefresh}>刷新状态</Btn>
            <Btn size="sm" icon={<ISync size={13} />} onClick={handleResync}>重新同步</Btn>
            <Btn size="sm" variant="primary" onClick={handleRollback}>回滚</Btn>
          </>
        }
      />
      <div style={{ padding: 24 }}>
        <Card>
          <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: 16 }}>
            <div style={{ fontSize: 13, fontWeight: 600 }}>基本信息</div>
            <StatusBadge status={deploy.status} />
          </div>
          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', rowGap: 12, columnGap: 20 }}>
            {rows.map(({ label, value }) => (
              <div key={label} style={{ display: 'flex', flexDirection: 'column', gap: 2 }}>
                <span style={{ fontSize: 11, color: 'var(--z-400)', fontWeight: 500, textTransform: 'uppercase', letterSpacing: '0.05em' }}>{label}</span>
                <div style={{ fontSize: 12.5 }}>{value}</div>
              </div>
            ))}
          </div>
        </Card>
      </div>
    </>
  );
}
