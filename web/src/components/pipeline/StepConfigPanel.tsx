import { useEffect } from 'react';
import { Drawer, Form, Input, Select, Button, Space, Typography } from '@arco-design/web-react';
import type { StepConfig } from '../../services/pipeline';

const { TextArea } = Input;
const { Title } = Typography;

interface StepConfigPanelProps {
  visible: boolean;
  step: StepConfig | null;
  onClose: () => void;
  onSave: (step: StepConfig) => void;
}

const stepTypes = [
  { label: 'Git Clone', value: 'git-clone' },
  { label: 'Shell', value: 'shell' },
  { label: 'Kaniko', value: 'kaniko' },
  { label: 'BuildKit', value: 'buildkit' },
];

export function StepConfigPanel({ visible, step, onClose, onSave }: StepConfigPanelProps) {
  const [form] = Form.useForm();

  useEffect(() => {
    if (step) {
      form.setFieldsValue({
        name: step.name,
        type: step.type,
        image: step.image || '',
        command: step.command?.join('\n') || '',
        args: step.args?.join('\n') || '',
      });
    }
  }, [step, form]);

  const handleSave = () => {
    form.validate().then((values) => {
      if (!step) return;
      const updated: StepConfig = {
        ...step,
        name: values.name,
        type: values.type,
        image: values.image || undefined,
        command: values.command ? values.command.split('\n').filter(Boolean) : undefined,
        args: values.args ? values.args.split('\n').filter(Boolean) : undefined,
      };
      onSave(updated);
      onClose();
    });
  };

  return (
    <Drawer
      width={400}
      title={<Title heading={6}>Step 配置</Title>}
      visible={visible}
      onCancel={onClose}
      footer={
        <Space>
          <Button onClick={onClose}>取消</Button>
          <Button type="primary" onClick={handleSave}>保存</Button>
        </Space>
      }
    >
      <Form form={form} layout="vertical">
        <Form.Item label="名称" field="name" rules={[{ required: true, message: '请输入 Step 名称' }]}>
          <Input placeholder="Step 名称" />
        </Form.Item>
        <Form.Item label="类型" field="type" rules={[{ required: true, message: '请选择 Step 类型' }]}>
          <Select options={stepTypes} placeholder="选择类型" />
        </Form.Item>
        <Form.Item label="镜像" field="image">
          <Input placeholder="例如 golang:1.24" />
        </Form.Item>
        <Form.Item label="命令" field="command">
          <TextArea placeholder="每行一条命令" autoSize={{ minRows: 3 }} />
        </Form.Item>
        <Form.Item label="参数" field="args">
          <TextArea placeholder="每行一个参数" autoSize={{ minRows: 2 }} />
        </Form.Item>
      </Form>
    </Drawer>
  );
}
