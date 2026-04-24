import { useEffect, useState } from 'react';
import { Message } from '@arco-design/web-react';
import { http, extractErrorMessage } from '../../services/http';
import { ZModal } from '../../components/ui/ZModal';
import { Btn } from '../../components/ui/Btn';
import { Field } from '../../components/ui/Field';
import { ZSelect } from '../../components/ui/ZSelect';

interface UserFormModalProps {
  visible: boolean;
  user?: { id: string; username: string; role: string; status: string } | null;
  onClose: () => void;
  onSuccess: () => void;
}

const EMPTY = { username: '', password: '', role: 'member', status: 'active' };

export function UserFormModal({ visible, user, onClose, onSuccess }: UserFormModalProps) {
  const [form, setForm] = useState(EMPTY);
  const [loading, setLoading] = useState(false);
  const isEdit = !!user;

  useEffect(() => {
    if (visible) {
      setForm(user ? { username: user.username, password: '', role: user.role, status: user.status } : EMPTY);
    }
  }, [visible, user]);

  const handleSubmit = async () => {
    if (!form.username.trim()) { Message.error('请输入用户名'); return; }
    if (!isEdit && !form.password) { Message.error('请输入密码'); return; }
    if (form.password && form.password.length < 6) { Message.error('密码长度至少 6 位'); return; }
    setLoading(true);
    try {
      if (isEdit) {
        const body: Record<string, string> = { role: form.role, status: form.status };
        if (form.password) body.password = form.password;
        await http.put(`/admin/users/${user.id}`, body);
        Message.success('更新成功');
      } else {
        await http.post('/admin/users', { username: form.username.trim(), password: form.password, role: form.role, status: form.status });
        Message.success('创建成功');
      }
      onSuccess();
      onClose();
    } catch (error: unknown) {
      Message.error(extractErrorMessage(error, '操作失败'));
    } finally {
      setLoading(false);
    }
  };

  if (!visible) return null;

  return (
    <ZModal
      title={isEdit ? '编辑用户' : '新建用户'}
      onClose={onClose}
      footer={
        <div style={{ display: 'flex', justifyContent: 'flex-end', gap: 8 }}>
          <Btn onClick={onClose}>取消</Btn>
          <Btn variant="primary" onClick={handleSubmit} disabled={loading}>
            {loading ? '保存中...' : '保存'}
          </Btn>
        </div>
      }
    >
      <div style={{ display: 'flex', flexDirection: 'column', gap: 14 }}>
        <Field label="用户名" required>
          <input
            className="input"
            value={form.username}
            onChange={(e) => setForm({ ...form, username: e.target.value })}
            placeholder="请输入用户名"
            disabled={isEdit}
          />
        </Field>
        <Field label="密码" required={!isEdit} help={isEdit ? '留空则不修改密码' : '至少 6 位'}>
          <input
            className="input"
            type="password"
            value={form.password}
            onChange={(e) => setForm({ ...form, password: e.target.value })}
            placeholder={isEdit ? '留空则不修改' : '请输入密码'}
          />
        </Field>
        <Field label="角色" required>
          <ZSelect
            width={240}
            value={form.role}
            options={[{ value: 'admin', label: '管理员' }, { value: 'member', label: '普通成员' }]}
            onChange={(v) => setForm({ ...form, role: v })}
          />
        </Field>
        <Field label="状态" required>
          <ZSelect
            width={240}
            value={form.status}
            options={[{ value: 'active', label: '启用' }, { value: 'disabled', label: '禁用' }]}
            onChange={(v) => setForm({ ...form, status: v })}
          />
        </Field>
      </div>
    </ZModal>
  );
}
