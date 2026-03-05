import { useEffect } from 'react';
import { Drawer, Form, Input, Select, Button, Space, Typography } from '@arco-design/web-react';
import type { PipelineSummary } from '../../services/pipeline';

const { TextArea } = Input;
const { Title } = Typography;

export type TriggerType = 'manual' | 'webhook' | 'scheduled';
export type ConcurrencyPolicy = 'queue' | 'cancel_old' | 'reject';

export interface PipelineSettingsForm {
  triggerType: TriggerType;
  concurrencyPolicy: ConcurrencyPolicy;
  description: string;
}

interface PipelineSettingsPanelProps {
  visible: boolean;
  pipeline: PipelineSummary | null;
  onClose: () => void;
  onSave: (data: PipelineSettingsForm) => void | Promise<void>;
  saving?: boolean;
}

const triggerOptions = [
  { label: '手动', value: 'manual' },
  { label: 'Webhook', value: 'webhook' },
  { label: '定时', value: 'scheduled' },
];

const concurrencyOptions = [
  { label: '排队', value: 'queue' },
  { label: '取消旧任务', value: 'cancel_old' },
  { label: '拒绝', value: 'reject' },
];

export function PipelineSettingsPanel({
  visible,
  pipeline,
  onClose,
  onSave,
  saving = false,
}: PipelineSettingsPanelProps) {
  const [form] = Form.useForm<PipelineSettingsForm>();

  useEffect(() => {
    if (visible && pipeline) {
      form.setFieldsValue({
        triggerType: (pipeline.triggerType as TriggerType) || 'manual',
        concurrencyPolicy: (pipeline.concurrencyPolicy as ConcurrencyPolicy) || 'queue',
        description: pipeline.description || '',
      });
    }
  }, [visible, pipeline, form]);

  const handleSave = () => {
    form.validate().then(async (values) => {
      await Promise.resolve(onSave(values));
      onClose();
    }).catch(() => {
      // Arco Form shows inline errors
    });
  };

  return (
    <Drawer
      width={400}
      title={<Title heading={6}>流水线设置</Title>}
      visible={visible}
      onCancel={onClose}
      footer={
        <Space>
          <Button onClick={onClose}>取消</Button>
          <Button type="primary" onClick={handleSave} loading={saving}>
            保存
          </Button>
        </Space>
      }
    >
      <Form form={form} layout="vertical">
        <Form.Item label="触发方式" field="triggerType" rules={[{ required: true }]}>
          <Select options={triggerOptions} placeholder="选择触发方式" />
        </Form.Item>
        <Form.Item label="并发策略" field="concurrencyPolicy" rules={[{ required: true }]}>
          <Select options={concurrencyOptions} placeholder="选择并发策略" />
        </Form.Item>
        <Form.Item label="描述" field="description">
          <TextArea placeholder="流水线描述（可选）" autoSize={{ minRows: 3 }} />
        </Form.Item>
      </Form>
    </Drawer>
  );
}
