import { useState, useEffect } from 'react';
import { Message } from '@arco-design/web-react';
import axios from 'axios';
import { AppLayout } from '../../components/layout/AppLayout';
import { http, extractErrorMessage } from '../../services/http';
import { UserFormModal } from './UserFormModal';
import { PageHeader } from '../../components/ui/PageHeader';
import { Card } from '../../components/ui/Card';
import { Btn } from '../../components/ui/Btn';
import { StatusBadge } from '../../components/ui/StatusBadge';
import { Avatar } from '../../components/ui/Avatar';
import { IPlus, IEdit } from '../../components/ui/icons';

interface User {
  id: string;
  username: string;
  role: string;
  status: string;
  created_at: string;
}

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
      } else {
        Message.error(extractErrorMessage(error, '加载用户列表失败'));
      }
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => { fetchUsers(); }, []);

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

  return (
    <AppLayout>
      <PageHeader
        crumb="System · Access Control"
        title="用户管理"
        sub="管理系统用户账号与角色。"
        actions={
          <Btn size="sm" variant="primary" icon={<IPlus size={13} />} onClick={() => { setEditingUser(null); setModalVisible(true); }}>
            新建用户
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
                  <th>用户名</th>
                  <th>角色</th>
                  <th>状态</th>
                  <th>创建时间</th>
                  <th style={{ textAlign: 'right' }}>操作</th>
                </tr>
              </thead>
              <tbody>
                {users.map((u) => (
                  <tr key={u.id}>
                    <td>
                      <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                        <Avatar name={u.username} size="sm" round />
                        <b style={{ fontWeight: 500 }}>{u.username}</b>
                      </div>
                    </td>
                    <td><StatusBadge status={u.role} /></td>
                    <td><StatusBadge status={u.status === 'active' ? 'enabled' : 'userDisabled'} /></td>
                    <td><span className="sub mono" style={{ fontSize: 11.5 }}>{u.created_at}</span></td>
                    <td style={{ textAlign: 'right' }}>
                      <div style={{ display: 'inline-flex', gap: 4 }}>
                        <Btn size="xs" variant="ghost" iconOnly icon={<IEdit size={12} />} onClick={() => { setEditingUser(u); setModalVisible(true); }} />
                        <Btn size="xs" variant="ghost" onClick={() => handleToggleStatus(u)}>
                          {u.status === 'active' ? '禁用' : '启用'}
                        </Btn>
                      </div>
                    </td>
                  </tr>
                ))}
                {users.length === 0 && !loading && (
                  <tr>
                    <td colSpan={5} style={{ textAlign: 'center', padding: '40px 0', color: 'var(--z-400)' }}>暂无用户数据</td>
                  </tr>
                )}
              </tbody>
            </table>
          )}
        </Card>
      </div>
      <UserFormModal
        visible={modalVisible}
        user={editingUser}
        onClose={() => setModalVisible(false)}
        onSuccess={fetchUsers}
      />
    </AppLayout>
  );
}
