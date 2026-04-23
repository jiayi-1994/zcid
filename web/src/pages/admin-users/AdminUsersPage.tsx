import { Button, Table, Space, Message, Popconfirm } from '@arco-design/web-react';
import { IconPlus } from '@arco-design/web-react/icon';
import { AppLayout } from '../../components/layout/AppLayout';
import { useState, useEffect } from 'react';
import axios from 'axios';
import { http, extractErrorMessage } from '../../services/http';
import { UserFormModal } from './UserFormModal';

interface User {
  id: string;
  username: string;
  role: string;
  status: string;
  created_at: string;
}

const ROLE_LABELS: Record<string, string> = {
  admin: '管理员',
  project_admin: '项目管理员',
  member: '普通成员',
};

const ROLE_CLS: Record<string, string> = {
  admin: 'pipeline-status-badge--running',
  project_admin: 'pipeline-status-badge--success',
  member: 'pipeline-status-badge--pending',
};

export function AdminUsersPage() {
  const [users, setUsers] = useState<User[]>([]);
  const [loading, setLoading] = useState(false);
  const [modalVisible, setModalVisible] = useState(false);
  const [editingUser, setEditingUser] = useState<User | null>(null);

  const fetchUsers = async () => {
    setLoading(true);
    try {
      const res = await http.get('/admin/users');
      setUsers(res.data.data || []);
    } catch (error: unknown) {
      if (axios.isAxiosError(error) && error.response?.status === 403) {
        Message.error('权限不足，无法访问用户列表');
      } else if (axios.isAxiosError(error) && error.response?.status === 401) {
        Message.error('登录已过期，请重新登录');
      } else {
        Message.error(extractErrorMessage(error, '加载用户列表失败'));
      }
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchUsers();
  }, []);

  const handleCreate = () => {
    setEditingUser(null);
    setModalVisible(true);
  };

  const handleEdit = (user: User) => {
    setEditingUser(user);
    setModalVisible(true);
  };

  const handleToggleStatus = async (user: User) => {
    try {
      const newStatus = user.status === 'active' ? 'disabled' : 'active';
      await http.put(`/admin/users/${user.id}`, { status: newStatus });
      Message.success('操作成功');
      fetchUsers();
    } catch (error: unknown) {
      Message.error(extractErrorMessage(error, '操作失败'));
    }
  };

  const columns = [
    { title: '用户名', dataIndex: 'username' },
    {
      title: '角色',
      dataIndex: 'role',
      width: 140,
      render: (role: string) => (
        <span className={`pipeline-status-badge ${ROLE_CLS[role] || 'pipeline-status-badge--pending'}`}>
          {ROLE_LABELS[role] || role}
        </span>
      ),
    },
    {
      title: '状态',
      dataIndex: 'status',
      width: 120,
      render: (status: string) => (
        <span
          className={`pipeline-status-badge ${
            status === 'active'
              ? 'pipeline-status-badge--success'
              : 'pipeline-status-badge--cancelled'
          }`}
        >
          {status === 'active' ? '启用' : '禁用'}
        </span>
      ),
    },
    { title: '创建时间', dataIndex: 'created_at', width: 200 },
    {
      title: '操作',
      width: 180,
      render: (_: unknown, record: User) => (
        <Space size="mini">
          <Button type="text" size="small" onClick={() => handleEdit(record)}>
            编辑
          </Button>
          <Popconfirm
            title={`确定${record.status === 'active' ? '禁用' : '启用'}该用户？`}
            onOk={() => handleToggleStatus(record)}
          >
            <Button
              type="text"
              size="small"
              status={record.status === 'active' ? 'danger' : 'success'}
            >
              {record.status === 'active' ? '禁用' : '启用'}
            </Button>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <AppLayout>
      <div className="page-container">
        <div className="page-header">
          <div>
            <div className="breadcrumb">System · Access Control</div>
            <h1 className="page-title">用户管理</h1>
            <p className="page-subtitle">管理系统用户账号与角色</p>
          </div>
          <Button type="primary" icon={<IconPlus />} onClick={handleCreate}>
            新建用户
          </Button>
        </div>
        <div className="table-card">
          <Table
            columns={columns}
            data={users}
            loading={loading}
            rowKey="id"
            border={false}
            noDataElement={
              <div className="empty-state">
                <div className="empty-state-title">暂无用户数据</div>
                <div className="empty-state-desc">点击右上角新建第一个用户</div>
              </div>
            }
          />
        </div>
        <UserFormModal
          visible={modalVisible}
          user={editingUser}
          onClose={() => setModalVisible(false)}
          onSuccess={fetchUsers}
        />
      </div>
    </AppLayout>
  );
}
