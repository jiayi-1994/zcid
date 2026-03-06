import { Button, Message, Popconfirm, Table, Tag } from '@arco-design/web-react';
import { useCallback, useEffect, useState } from 'react';
import { useParams } from 'react-router-dom';
import { useAuthStore } from '../../../stores/auth';
import {
  fetchProjectVariables,
  createProjectVariable,
  updateProjectVariable,
  deleteProjectVariable,
  type VariableItem,
} from '../../../services/variable';
import { VariableFormModal } from './VariableFormModal';

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
    await updateProjectVariable(projectId, editItem.id, {
      value: data.value || undefined,
      description: data.description,
    });
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

  const columns = [
    { title: '变量名', dataIndex: 'key' },
    {
      title: '值',
      dataIndex: 'value',
      render: (val: string, record: VariableItem) =>
        record.varType === 'secret' ? <span style={{ color: 'var(--color-text-3)' }}>******</span> : val,
    },
    {
      title: '类型',
      dataIndex: 'varType',
      render: (val: string) => (
        <Tag size="small" color={val === 'secret' ? 'red' : 'blue'}>{val === 'secret' ? '密钥' : '普通'}</Tag>
      ),
    },
    { title: '描述', dataIndex: 'description' },
    { title: '创建时间', dataIndex: 'createdAt' },
    ...(canManage
      ? [{
          title: '操作',
          render: (_: unknown, record: VariableItem) => (
            <>
              <Button type="text" size="small" style={{ color: 'var(--zcid-primary)' }} onClick={() => setEditItem(record)}>
                编辑
              </Button>
              <Popconfirm title="确认删除此变量？" onOk={() => handleDelete(record.id)}>
                <Button type="text" size="small" status="danger">删除</Button>
              </Popconfirm>
            </>
          ),
        }]
      : []),
  ];

  return (
    <div className="page-container">
      <div className="page-header">
        <h3 className="page-title">项目变量</h3>
        {canManage && (
          <Button type="primary" onClick={() => setCreateVisible(true)}>新建变量</Button>
        )}
      </div>
      <div className="table-card">
        <Table
          columns={columns}
          data={variables}
          rowKey="id"
          loading={loading}
          border={false}
          pagination={false}
        />
      </div>
      <VariableFormModal
        visible={createVisible}
        onClose={() => setCreateVisible(false)}
        onSubmit={handleCreate}
      />
      {editItem && (
        <VariableFormModal
          visible={!!editItem}
          onClose={() => setEditItem(null)}
          onSubmit={handleEdit}
          editMode
          isSecret={editItem.varType === 'secret'}
          initialValues={{ key: editItem.key, description: editItem.description }}
        />
      )}
    </div>
  );
}
