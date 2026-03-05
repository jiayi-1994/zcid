import { Form, Input, Modal, Radio, Message } from '@arco-design/web-react';
import { useState } from 'react';

interface VariableFormModalProps {
  visible: boolean;
  onClose: () => void;
  onSubmit: (data: { key: string; value: string; varType: string; description: string }) => Promise<void>;
  editMode?: boolean;
  isSecret?: boolean;
  initialValues?: { key?: string; description?: string };
}

export function VariableFormModal({ visible, onClose, onSubmit, editMode, isSecret, initialValues }: VariableFormModalProps) {
  const [form] = Form.useForm();
  const [loading, setLoading] = useState(false);

  const handleOk = async () => {
    try {
      const values = await form.validate();
      setLoading(true);
      await onSubmit({
        key: values.key,
        value: values.value,
        varType: values.varType || 'plain',
        description: values.description || '',
      });
      Message.success(editMode ? '变量更新成功' : '变量创建成功');
      form.resetFields();
      onClose();
    } catch {
      // validation or submit error
    } finally {
      setLoading(false);
    }
  };

  return (
    <Modal
      title={editMode ? '编辑变量' : '新建变量'}
      visible={visible}
      onOk={handleOk}
      onCancel={() => { form.resetFields(); onClose(); }}
      confirmLoading={loading}
      unmountOnExit
    >
      <Form form={form} autoComplete="off" initialValues={initialValues}>
        {!editMode && (
          <Form.Item label="变量名" field="key" rules={[{ required: true, message: '请输入变量名' }]}>
            <Input placeholder="例如: DB_HOST" />
          </Form.Item>
        )}
        <Form.Item
          label="值"
          field="value"
          rules={[{ required: !editMode, message: '请输入变量值' }]}
        >
          <Input.Password
            placeholder={isSecret ? '输入新的密钥值' : '变量值'}
            visibilityToggle={!isSecret}
          />
        </Form.Item>
        {!editMode && (
          <Form.Item label="类型" field="varType" initialValue="plain">
            <Radio.Group>
              <Radio value="plain">普通</Radio>
              <Radio value="secret">密钥</Radio>
            </Radio.Group>
          </Form.Item>
        )}
        <Form.Item label="描述" field="description">
          <Input.TextArea placeholder="变量描述（可选）" />
        </Form.Item>
      </Form>
    </Modal>
  );
}
