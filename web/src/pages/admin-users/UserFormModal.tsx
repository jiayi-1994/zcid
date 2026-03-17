import { Modal, Form, Input, Select, Message } from '@arco-design/web-react';
import { useState } from 'react';
import { http, extractErrorMessage } from '../../services/http';

interface UserFormModalProps {
  visible: boolean;
  user?: { id: string; username: string; role: string; status: string } | null;
  onClose: () => void;
  onSuccess: () => void;
}

export function UserFormModal({ visible, user, onClose, onSuccess }: UserFormModalProps) {
  const [form] = Form.useForm();
  const [loading, setLoading] = useState(false);
  const isEdit = !!user;

  const handleSubmit = async () => {
    try {
      await form.validate();
      const values = form.getFieldsValue();
      setLoading(true);

      if (isEdit) {
        await http.put(`/admin/users/${user.id}`, values);
        Message.success('更新成功');
      } else {
        await http.post('/admin/users', values);
        Message.success('创建成功');
      }

      onSuccess();
      onClose();
      form.resetFields();
    } catch (error: unknown) {
      Message.error(extractErrorMessage(error, '操作失败'));
    } finally {
      setLoading(false);
    }
  };

  return (
    <Modal
      title={isEdit ? '编辑用户' : '新建用户'}
      visible={visible}
      onOk={handleSubmit}
      onCancel={onClose}
      confirmLoading={loading}
    >
      <Form form={form} layout="vertical" initialValues={user || { role: 'member', status: 'active' }}>
        <Form.Item label="用户名" field="username" rules={[{ required: true, message: '请输入用户名' }]}>
          <Input placeholder="请输入用户名" disabled={isEdit} />
        </Form.Item>
        <Form.Item label="密码" field="password" rules={isEdit ? [{ minLength: 6, message: '密码长度至少6位' }] : [{ required: true, message: '请输入密码' }, { minLength: 6, message: '密码长度至少6位' }]}>
          <Input.Password placeholder={isEdit ? '留空则不修改' : '请输入密码'} />
        </Form.Item>
        <Form.Item label="角色" field="role" rules={[{ required: true }]}>
          <Select>
            <Select.Option value="admin">管理员</Select.Option>
            <Select.Option value="member">普通成员</Select.Option>
          </Select>
        </Form.Item>
        <Form.Item label="状态" field="status" rules={[{ required: true }]}>
          <Select>
            <Select.Option value="active">启用</Select.Option>
            <Select.Option value="disabled">禁用</Select.Option>
          </Select>
        </Form.Item>
      </Form>
    </Modal>
  );
}
