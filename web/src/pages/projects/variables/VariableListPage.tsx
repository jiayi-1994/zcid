import { useCallback, useEffect, useState } from 'react';
import { Message } from '@arco-design/web-react';
import { useParams } from 'react-router-dom';
import { useAuthStore } from '../../../stores/auth';
import {
  fetchProjectVariables, createProjectVariable, updateProjectVariable, deleteProjectVariable,
  type VariableItem,
} from '../../../services/variable';
import { VariableFormModal } from './VariableFormModal';
import { PageHeader } from '../../../components/ui/PageHeader';
import { Card } from '../../../components/ui/Card';
import { Btn } from '../../../components/ui/Btn';
import { Badge } from '../../../components/ui/Badge';
import { IPlus, IEdit, ITrash } from '../../../components/ui/icons';

export function VariableListPage() {
  const { id: projectId } = useParams<{ id: string }>();
  const [variables, setVariables] = useState<VariableItem[]>([]);
  const [loading, setLoading] = useState(false);
  const [createVisible, setCreateVisible] = useState(false);
  const [editItem, setEditItem] = useState<VariableItem | null>(null);
  const user = useAuthStore((s) => s.user);
  const canManage = user?.role === 'admin' || user?.role === 'project_admin';

  const loadData = useCallback(async () => {
    if (!projectId) return;
    setLoading(true);
    try {
      const data = await fetchProjectVariables(projectId);
      setVariables(data.items || []);
    } catch {
      Message.error('加载变量列表失败');
    } finally {
      setLoading(false);
    }
  }, [projectId]);

  useEffect(() => { loadData(); }, [loadData]);

  const handleCreate = async (data: { key: string; value: string; varType: string; description: string }) => {
    if (!projectId) return;
    await createProjectVariable(projectId, data);
    await loadData();
  };

  const handleEdit = async (data: { value: string; description: string }) => {
    if (!projectId || !editItem) return;
    await updateProjectVariable(projectId, editItem.id, { value: data.value || undefined, description: data.description });
    setEditItem(null);
    await loadData();
  };

  const handleDelete = async (id: string) => {
    if (!projectId) return;
    try {
      await deleteProjectVariable(projectId, id);
      Message.success('变量已删除');
      await loadData();
    } catch {
      Message.error('删除失败');
    }
  };

  return (
    <>
      <PageHeader
        crumb="Project › Variables"
        title="Project Variables"
        sub="Secure variable management for the project. 项目级环境变量与密钥管理。"
        actions={canManage && (
          <Btn size="sm" variant="primary" icon={<IPlus size={13} />} onClick={() => setCreateVisible(true)}>
            Add Variable
          </Btn>
        )}
      />
      <div style={{ padding: 24 }}>
        <Card padding={false}>
          {loading ? (
            <div style={{ padding: '40px 0', textAlign: 'center', color: 'var(--z-400)' }}>加载中...</div>
          ) : (
            <table className="table">
              <thead>
                <tr>
                  <th>变量名</th><th>值</th><th>类型</th><th>描述</th><th>创建时间</th>
                  {canManage && <th style={{ textAlign: 'right' }}>操作</th>}
                </tr>
              </thead>
              <tbody>
                {variables.map((v) => (
                  <tr key={v.id}>
                    <td><span className="code">{v.key}</span></td>
                    <td>
                      <span className="mono" style={{ color: v.varType === 'secret' ? 'var(--z-400)' : 'var(--z-800)' }}>
                        {v.varType === 'secret' ? '••••••••' : v.value}
                      </span>
                    </td>
                    <td><Badge tone={v.varType === 'secret' ? 'red' : 'blue'}>{v.varType === 'secret' ? 'Secret' : 'Variable'}</Badge></td>
                    <td><span className="sub">{v.description}</span></td>
                    <td><span className="sub mono" style={{ fontSize: 11.5 }}>{v.createdAt}</span></td>
                    {canManage && (
                      <td style={{ textAlign: 'right' }}>
                        <div style={{ display: 'inline-flex', gap: 4 }}>
                          <Btn size="xs" variant="ghost" iconOnly icon={<IEdit size={12} />} onClick={() => setEditItem(v)} />
                          <Btn size="xs" variant="ghost" iconOnly icon={<ITrash size={12} />} onClick={() => handleDelete(v.id)} />
                        </div>
                      </td>
                    )}
                  </tr>
                ))}
                {variables.length === 0 && !loading && (
                  <tr>
                    <td colSpan={canManage ? 6 : 5} style={{ textAlign: 'center', padding: '40px 0', color: 'var(--z-400)' }}>
                      暂无变量
                    </td>
                  </tr>
                )}
              </tbody>
            </table>
          )}
        </Card>
      </div>

      <VariableFormModal visible={createVisible} onClose={() => setCreateVisible(false)} onSubmit={handleCreate} />
      {editItem && (
        <VariableFormModal
          visible={!!editItem} onClose={() => setEditItem(null)} onSubmit={handleEdit}
          editMode isSecret={editItem.varType === 'secret'}
          initialValues={{ key: editItem.key, description: editItem.description }}
        />
      )}
    </>
  );
}
