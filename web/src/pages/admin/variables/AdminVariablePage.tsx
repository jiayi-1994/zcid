import { useCallback, useEffect, useState } from 'react';
import { Message } from '@arco-design/web-react';
import { AppLayout } from '../../../components/layout/AppLayout';
import {
  fetchGlobalVariables, createGlobalVariable, updateGlobalVariable, deleteGlobalVariable,
  type VariableItem,
} from '../../../services/variable';
import { VariableFormModal } from '../../projects/variables/VariableFormModal';
import { PageHeader } from '../../../components/ui/PageHeader';
import { Card } from '../../../components/ui/Card';
import { Btn } from '../../../components/ui/Btn';
import { Badge } from '../../../components/ui/Badge';
import { IPlus, IEdit, ITrash } from '../../../components/ui/icons';

export function AdminVariablePage() {
  const [variables, setVariables] = useState<VariableItem[]>([]);
  const [loading, setLoading] = useState(false);
  const [createVisible, setCreateVisible] = useState(false);
  const [editItem, setEditItem] = useState<VariableItem | null>(null);

  const loadData = useCallback(async () => {
    setLoading(true);
    try { const data = await fetchGlobalVariables(); setVariables(data.items || []); }
    catch { Message.error('加载全局变量失败'); }
    finally { setLoading(false); }
  }, []);

  useEffect(() => { loadData(); }, [loadData]);

  const handleCreate = async (data: { key: string; value: string; varType: string; description: string }) => {
    await createGlobalVariable(data);
    await loadData();
  };

  const handleEdit = async (data: { value: string; description: string }) => {
    if (!editItem) return;
    await updateGlobalVariable(editItem.id, { value: data.value || undefined, description: data.description });
    setEditItem(null);
    await loadData();
  };

  const handleDelete = async (id: string) => {
    try { await deleteGlobalVariable(id); Message.success('全局变量已删除'); await loadData(); }
    catch { Message.error('删除失败'); }
  };

  return (
    <AppLayout>
      <PageHeader
        crumb="System › Variables"
        title="Global Variables"
        sub="System-wide variable management. 管理跨项目共享的全局变量和密钥。"
        actions={
          <Btn size="sm" variant="primary" icon={<IPlus size={13} />} onClick={() => setCreateVisible(true)}>
            Add Variable
          </Btn>
        }
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
                  <th style={{ textAlign: 'right' }}>操作</th>
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
                    <td style={{ textAlign: 'right' }}>
                      <div style={{ display: 'inline-flex', gap: 4 }}>
                        <Btn size="xs" variant="ghost" iconOnly icon={<IEdit size={12} />} onClick={() => setEditItem(v)} />
                        <Btn size="xs" variant="ghost" iconOnly icon={<ITrash size={12} />} onClick={() => handleDelete(v.id)} />
                      </div>
                    </td>
                  </tr>
                ))}
                {variables.length === 0 && !loading && (
                  <tr><td colSpan={6} style={{ textAlign: 'center', padding: '40px 0', color: 'var(--z-400)' }}>暂无全局变量</td></tr>
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
    </AppLayout>
  );
}
