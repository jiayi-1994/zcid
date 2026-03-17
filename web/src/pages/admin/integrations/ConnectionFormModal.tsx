import { Form, Input, Modal, Select, Message, Typography, Link } from '@arco-design/web-react';
import { useState } from 'react';

const { Text } = Typography;

interface ConnectionFormModalProps {
  visible: boolean;
  onClose: () => void;
  onSubmit: (data: {
    name: string;
    providerType: string;
    serverUrl: string;
    accessToken: string;
    description: string;
  }) => Promise<void>;
  editMode?: boolean;
  initialValues?: { name?: string; description?: string };
}

export function ConnectionFormModal({
  visible,
  onClose,
  onSubmit,
  editMode = false,
  initialValues,
}: ConnectionFormModalProps) {
  const [form] = Form.useForm();
  const [submitting, setSubmitting] = useState(false);

  const handleOk = async () => {
    try {
      const values = await form.validate();
      setSubmitting(true);
      await onSubmit(values);
      Message.success(editMode ? '连接已更新' : '连接已创建');
      form.resetFields();
      onClose();
    } catch {
      if (!submitting) return;
      Message.error(editMode ? '更新失败' : '创建失败');
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <Modal
      title={editMode ? '编辑 Git 连接' : '添加 Git 连接'}
      visible={visible}
      onOk={handleOk}
      onCancel={onClose}
      confirmLoading={submitting}
      afterClose={() => form.resetFields()}
    >
      <Form form={form} layout="vertical" initialValues={initialValues}>
        <Form.Item label="连接名称" field="name" rules={[{ required: true, message: '请输入连接名称' }]}>
          <Input placeholder="例如: my-gitlab" />
        </Form.Item>
        {!editMode && (
          <>
            <Form.Item label="Provider 类型" field="providerType" rules={[{ required: true }]} initialValue="gitlab">
              <Select>
                <Select.Option value="gitlab">GitLab</Select.Option>
                <Select.Option value="github">GitHub</Select.Option>
              </Select>
            </Form.Item>
            <Form.Item
              label="Server URL"
              field="serverUrl"
              rules={[{ required: true, message: '请输入 Server URL' }]}
              extra={<span style={{ fontSize: 12, color: 'var(--muted-foreground)' }}>支持内网地址，如 http://192.168.1.100:8080 或 https://git.internal.company.com</span>}
            >
              <Input placeholder="例如: https://gitlab.example.com 或 http://内网IP:端口" />
            </Form.Item>
          </>
        )}
        <Form.Item
          label="Access Token (PAT)"
          field="accessToken"
          rules={editMode ? [] : [{ required: true, message: '请输入 Access Token' }]}
          extra={!editMode && (
            <div style={{ fontSize: 12, color: 'var(--muted-foreground)', marginTop: 4, lineHeight: 1.6 }}>
              <div style={{ fontWeight: 500, marginBottom: 2 }}>如何获取 PAT：</div>
              <div>• <strong>GitHub</strong>：Settings → Developer settings → Personal access tokens → Generate new token，勾选 <code>repo</code> 权限</div>
              <div>• <strong>GitLab</strong>：Settings → Access Tokens → 勾选 <code>api</code> + <code>read_repository</code> 权限</div>
              <div>• 内网 GitLab 同样支持，只需填写内网 Server URL 即可（如 <code>http://192.168.1.100:8080</code>）</div>
            </div>
          )}
        >
          <Input.Password placeholder={editMode ? '留空则不更新' : '请输入 Personal Access Token'} />
        </Form.Item>
        <Form.Item label="描述" field="description">
          <Input.TextArea placeholder="可选的描述信息" rows={2} />
        </Form.Item>
      </Form>
    </Modal>
  );
}
