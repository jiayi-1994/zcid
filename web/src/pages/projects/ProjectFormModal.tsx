import { Modal, Form, Input, Message } from '@arco-design/web-react';
import { useState } from 'react';
import { createProject } from '../../services/project';

const FormItem = Form.Item;

interface Props {
  visible: boolean;
  onClose: () => void;
  onSuccess: () => void;
}

export function ProjectFormModal({ visible, onClose, onSuccess }: Props) {
  const [form] = Form.useForm();
  const [loading, setLoading] = useState(false);

  const handleSubmit = async () => {
    try {
      const values = await form.validate();
      setLoading(true);
      await createProject(values.name, values.description || '');
      Message.success('项目创建成功');
      form.resetFields();
      onSuccess();
    } catch (err: any) {
      if (err.response?.data?.message) {
        Message.error(err.response.data.message);
      }
    } finally {
      setLoading(false);
    }
  };

  return (
    <Modal
      title="新建项目"
      visible={visible}
      onOk={handleSubmit}
      onCancel={() => { form.resetFields(); onClose(); }}
      confirmLoading={loading}
      autoFocus={false}
      unmountOnExit
    >
      <Form form={form} layout="vertical">
        <FormItem label="项目名称" field="name" rules={[{ required: true, message: '请输入项目名称' }]}>
          <Input placeholder="输入项目名称" />
        </FormItem>
        <FormItem label="描述" field="description">
          <Input.TextArea placeholder="输入项目描述" rows={3} />
        </FormItem>
      </Form>
    </Modal>
  );
}
