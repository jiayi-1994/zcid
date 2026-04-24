import { useCallback, useEffect, useState } from 'react';
import { Message } from '@arco-design/web-react';
import { useParams } from 'react-router-dom';
import { useAuthStore } from '../../../stores/auth';
import { fetchMembers, addMember, removeMember, updateMemberRole, type MemberItem } from '../../../services/project';
import { extractErrorMessage } from '../../../services/http';
import { PageHeader } from '../../../components/ui/PageHeader';
import { Card } from '../../../components/ui/Card';
import { Btn } from '../../../components/ui/Btn';
import { Avatar } from '../../../components/ui/Avatar';
import { StatusBadge } from '../../../components/ui/StatusBadge';
import { ZSelect } from '../../../components/ui/ZSelect';
import { ZModal } from '../../../components/ui/ZModal';
import { Field } from '../../../components/ui/Field';
import { IPlus } from '../../../components/ui/icons';

export function MemberListPage() {
  const { id: projectId } = useParams<{ id: string }>();
  const [members, setMembers] = useState<MemberItem[]>([]);
  const [loading, setLoading] = useState(false);
  const [modalVisible, setModalVisible] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [form, setForm] = useState({ userId: '', role: 'member' });
  const isAdmin = useAuthStore((s) => s.user?.role === 'admin');

  const loadData = useCallback(async () => {
    if (!projectId) return;
    setLoading(true);
    try {
      const data = await fetchMembers(projectId);
      setMembers(data.items ?? []);
    } catch (err: unknown) {
      Message.error(extractErrorMessage(err, '加载成员列表失败'));
    } finally {
      setLoading(false);
    }
  }, [projectId]);

  useEffect(() => { loadData(); }, [loadData]);

  const handleAdd = async () => {
    if (!form.userId) { Message.error('请输入用户 ID'); return; }
    setSubmitting(true);
    try {
      await addMember(projectId!, form.userId, form.role);
      Message.success('成员添加成功');
      setForm({ userId: '', role: 'member' });
      setModalVisible(false);
      loadData();
    } catch (err: unknown) {
      Message.error(extractErrorMessage(err, '添加失败'));
    } finally {
      setSubmitting(false);
    }
  };

  const handleRemove = async (userId: string) => {
    try {
      await removeMember(projectId!, userId);
      Message.success('成员已移除');
      loadData();
    } catch (err: unknown) {
      Message.error(extractErrorMessage(err, '移除失败'));
    }
  };

  const handleRoleChange = async (userId: string, newRole: string) => {
    try {
      await updateMemberRole(projectId!, userId, newRole);
      Message.success('角色已更新');
      loadData();
    } catch (err: unknown) {
      Message.error(extractErrorMessage(err, '更新失败'));
    }
  };

  return (
    <>
      <PageHeader
        crumb="Project · Access"
        title="成员管理"
        sub="分配项目成员与角色权限。"
        actions={isAdmin && (
          <Btn size="sm" variant="primary" icon={<IPlus size={13} />} onClick={() => setModalVisible(true)}>
            添加成员
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
                  <th>用户名</th>
                  <th>角色</th>
                  <th>加入时间</th>
                  {isAdmin && <th style={{ textAlign: 'right' }}>操作</th>}
                </tr>
              </thead>
              <tbody>
                {members.map((m) => (
                  <tr key={m.userId}>
                    <td>
                      <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                        <Avatar name={m.username} size="sm" round />
                        <span style={{ fontWeight: 500 }}>{m.username}</span>
                      </div>
                    </td>
                    <td>
                      {isAdmin ? (
                        <ZSelect
                          width={150}
                          value={m.role}
                          options={[
                            { value: 'project_admin', label: '项目管理员' },
                            { value: 'member', label: '普通成员' },
                          ]}
                          onChange={(val) => handleRoleChange(m.userId, val)}
                        />
                      ) : (
                        <StatusBadge status={m.role} />
                      )}
                    </td>
                    <td><span className="sub mono" style={{ fontSize: 11.5 }}>{m.joinedAt}</span></td>
                    {isAdmin && (
                      <td style={{ textAlign: 'right' }}>
                        <Btn size="xs" variant="ghost" onClick={() => handleRemove(m.userId)}>移除</Btn>
                      </td>
                    )}
                  </tr>
                ))}
                {members.length === 0 && !loading && (
                  <tr>
                    <td colSpan={isAdmin ? 4 : 3} style={{ textAlign: 'center', padding: '40px 0', color: 'var(--z-400)' }}>
                      暂无成员
                    </td>
                  </tr>
                )}
              </tbody>
            </table>
          )}
        </Card>
      </div>

      {modalVisible && (
        <ZModal
          title="添加成员"
          onClose={() => setModalVisible(false)}
          footer={
            <div style={{ display: 'flex', justifyContent: 'flex-end', gap: 8 }}>
              <Btn onClick={() => setModalVisible(false)}>取消</Btn>
              <Btn variant="primary" onClick={handleAdd} disabled={submitting}>
                {submitting ? '添加中...' : '添加'}
              </Btn>
            </div>
          }
        >
          <div style={{ display: 'flex', flexDirection: 'column', gap: 14 }}>
            <Field label="用户 ID" required>
              <input
                className="input"
                value={form.userId}
                onChange={(e) => setForm({ ...form, userId: e.target.value })}
                placeholder="输入要添加的用户 ID"
              />
            </Field>
            <Field label="角色">
              <ZSelect
                width={200}
                value={form.role}
                options={[
                  { value: 'project_admin', label: '项目管理员' },
                  { value: 'member', label: '普通成员' },
                ]}
                onChange={(val) => setForm({ ...form, role: val })}
              />
            </Field>
          </div>
        </ZModal>
      )}
    </>
  );
}
