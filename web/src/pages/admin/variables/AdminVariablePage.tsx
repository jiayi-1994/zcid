import { Button, Message, Popconfirm, Table, Tag } from '@arco-design/web-react';
import { useCallback, useEffect, useState } from 'react';
import { AppLayout } from '../../../components/layout/AppLayout';
import {
  fetchGlobalVariables,
  createGlobalVariable,
  updateGlobalVariable,
  deleteGlobalVariable,
  type VariableItem,
} from '../../../services/variable';
import { VariableFormModal } from '../../projects/variables/VariableFormModal';

export function AdminVariablePage() {
  const [variables, setVariables] = useState<VariableItem[]>([]);
  const [loading, setLoading] = useState(false);
  const [createVisible, setCreateVisible] = useState(false);
  const [editItem, setEditItem] = useState<VariableItem | null>(null);

  const loadData = useCallback(async () => {
    setLoading(true);
    try {
      const data = await fetchGlobalVariables();
      setVariables(data.items || []);
    } catch {
      Message.error('加载全局变量失败');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => { loadData(); }, [loadData]);

  const handleCreate = async (data: { key: string; value: string; varType: string; description: string }) => {
    await createGlobalVariable(data);
    await loadData();
  };

  const handleEdit = async (data: { value: string; description: string }) => {
    if (!editItem) return;
    await updateGlobalVariable(editItem.id, {
      value: data.value || undefined,
      description: data.description,
    });
    setEditItem(null);
    await loadData();
  };

  const handleDelete = async (id: string) => {
    try {
      await deleteGlobalVariable(id);
      Message.success('全局变量已删除');
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
        <Tag color={val === 'secret' ? 'red' : 'blue'}>{val === 'secret' ? '密钥' : '普通'}</Tag>
      ),
    },
    { title: '描述', dataIndex: 'description' },
    { title: '创建时间', dataIndex: 'createdAt' },
    {
      title: '操作',
      render: (_: unknown, record: VariableItem) => (
        <>
          <Button type="text" size="small" onClick={() => setEditItem(record)}>
            编辑
          </Button>
          <Popconfirm title="确认删除此全局变量？" onOk={() => handleDelete(record.id)}>
            <Button type="text" size="small" status="danger">删除</Button>
          </Popconfirm>
        </>
      ),
    },
  ];

  return (
    <AppLayout>
      <div className="page-container">
        <div className="page-header">
          <h3 className="page-title">全局变量管理</h3>
          <Button type="primary" onClick={() => setCreateVisible(true)}>新建全局变量</Button>
        </div>
        <Table
          columns={columns}
          data={variables}
          rowKey="id"
          loading={loading}
          pagination={false}
        />
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
    </AppLayout>
  );
}
