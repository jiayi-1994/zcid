import { Button, Table, Message, Popconfirm, Modal, Form, Input, Select, Tag } from '@arco-design/web-react';
import { IconPlus } from '@arco-design/web-react/icon';
import { useState, useEffect, useCallback } from 'react';
import { useParams } from 'react-router-dom';
import { useAuthStore } from '../../../stores/auth';
import { fetchMembers, addMember, removeMember, updateMemberRole, type MemberItem } from '../../../services/project';

const FormItem = Form.Item;

const ROLE_LABELS: Record<string, string> = {
  project_admin: '项目管理员',
  member: '普通成员',
};

export function MemberListPage() {
  const { id: projectId } = useParams<{ id: string }>();
  const [members, setMembers] = useState<MemberItem[]>([]);
  const [loading, setLoading] = useState(false);
  const [modalVisible, setModalVisible] = useState(false);
  const [form] = Form.useForm();
  const [submitLoading, setSubmitLoading] = useState(false);
  const isAdmin = useAuthStore((s) => s.user?.role === 'admin');

  const loadData = useCallback(async () => {
    if (!projectId) return;
    setLoading(true);
    try {
      const data = await fetchMembers(projectId);
      setMembers(data.items ?? []);
    } catch (err: any) {
      Message.error(err.response?.data?.message || '加载成员列表失败');
    } finally {
      setLoading(false);
    }
  }, [projectId]);

  useEffect(() => { loadData(); }, [loadData]);

  const handleAdd = async () => {
    try {
      const values = await form.validate();
      setSubmitLoading(true);
      await addMember(projectId!, values.userId, values.role);
      Message.success('成员添加成功');
      form.resetFields();
      setModalVisible(false);
      loadData();
    } catch (err: any) {
      if (err.response?.data?.message) Message.error(err.response.data.message);
    } finally {
      setSubmitLoading(false);
    }
  };

  const handleRemove = async (userId: string) => {
    try {
      await removeMember(projectId!, userId);
      Message.success('成员已移除');
      loadData();
    } catch (err: any) {
      Message.error(err.response?.data?.message || '移除失败');
    }
  };

  const handleRoleChange = async (userId: string, newRole: string) => {
    try {
      await updateMemberRole(projectId!, userId, newRole);
      Message.success('角色已更新');
      loadData();
    } catch (err: any) {
      Message.error(err.response?.data?.message || '更新失败');
    }
  };

  const columns = [
    { title: '用户名', dataIndex: 'username' },
    {
      title: '角色',
      dataIndex: 'role',
      render: (role: string, record: MemberItem) => isAdmin ? (
        <Select
          value={role}
          size="small"
          style={{ width: 140 }}
          onChange={(val) => handleRoleChange(record.userId, val)}
          options={[
            { label: '项目管理员', value: 'project_admin' },
            { label: '普通成员', value: 'member' },
          ]}
        />
      ) : (
        <Tag>{ROLE_LABELS[role] || role}</Tag>
      ),
    },
    { title: '加入时间', dataIndex: 'joinedAt' },
    {
      title: '操作',
      render: (_: any, record: MemberItem) => isAdmin ? (
        <Popconfirm title="确定移除？" onOk={() => handleRemove(record.userId)}>
          <Button type="text" size="small" status="danger">移除</Button>
        </Popconfirm>
      ) : null,
    },
  ];

  return (
    <div className="page-container">
      <div className="page-header">
        <h3 className="page-title">成员管理</h3>
        {isAdmin && (
          <Button type="primary" icon={<IconPlus />} size="small" onClick={() => setModalVisible(true)}>
            添加成员
          </Button>
        )}
      </div>
      <Table
        columns={columns} data={members} loading={loading} rowKey="userId"
        noDataElement={<div className="empty-state">暂无成员</div>}
      />
      <Modal
        title="添加成员" visible={modalVisible}
        onOk={handleAdd}
        onCancel={() => { form.resetFields(); setModalVisible(false); }}
        confirmLoading={submitLoading} unmountOnExit
      >
        <Form form={form} layout="vertical">
          <FormItem label="用户 ID" field="userId" rules={[{ required: true, message: '请输入用户 ID' }]}>
            <Input placeholder="输入要添加的用户 ID" />
          </FormItem>
          <FormItem label="角色" field="role" rules={[{ required: true, message: '请选择角色' }]} initialValue="member">
            <Select
              options={[
                { label: '项目管理员', value: 'project_admin' },
                { label: '普通成员', value: 'member' },
              ]}
            />
          </FormItem>
        </Form>
      </Modal>
    </div>
  );
}
