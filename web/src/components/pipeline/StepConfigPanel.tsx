import { useEffect, useMemo, useState } from 'react';
import { Drawer, Form, Input, Select, Button, Space, Typography, Divider } from '@arco-design/web-react';
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

const CONFIG_FIELDS_BY_TYPE: Record<string, { key: string; label: string; placeholder: string }[]> = {
  'git-clone': [
    { key: 'repoUrl', label: '仓库地址', placeholder: 'https://github.com/user/repo.git' },
    { key: 'branch', label: '分支', placeholder: 'main' },
  ],
  kaniko: [
    { key: 'imageName', label: '镜像名称', placeholder: 'registry.example.com/app:latest' },
    { key: 'dockerfile', label: 'Dockerfile 路径', placeholder: 'Dockerfile' },
  ],
  buildkit: [
    { key: 'imageName', label: '镜像名称', placeholder: 'registry.example.com/app:latest' },
    { key: 'dockerfile', label: 'Dockerfile 路径', placeholder: 'Dockerfile' },
  ],
};

function envMapToText(env?: Record<string, string>): string {
  if (!env) return '';
  return Object.entries(env).map(([k, v]) => `${k}=${v}`).join('\n');
}

function textToEnvMap(text: string): Record<string, string> | undefined {
  const lines = text.split('\n').filter((l) => l.trim() && l.includes('='));
  if (lines.length === 0) return undefined;
  const result: Record<string, string> = {};
  for (const line of lines) {
    const idx = line.indexOf('=');
    const key = line.substring(0, idx).trim();
    const val = line.substring(idx + 1).trim();
    if (key) result[key] = val;
  }
  return result;
}

export function StepConfigPanel({ visible, step, onClose, onSave }: StepConfigPanelProps) {
  const [form] = Form.useForm();
  const [currentType, setCurrentType] = useState<string>(step?.type ?? 'shell');

  const configFields = useMemo(
    () => CONFIG_FIELDS_BY_TYPE[currentType] ?? [],
    [currentType]
  );

  useEffect(() => {
    if (step && visible) {
      setCurrentType(step.type);
      const configValues: Record<string, string> = {};
      for (const f of CONFIG_FIELDS_BY_TYPE[step.type] ?? []) {
        const val = step.config?.[f.key];
        configValues[`config_${f.key}`] = typeof val === 'string' ? val : '';
      }
      form.setFieldsValue({
        name: step.name,
        type: step.type,
        image: step.image || '',
        command: step.command?.join('\n') || '',
        args: step.args?.join('\n') || '',
        env: envMapToText(step.env),
        ...configValues,
      });
    }
  }, [step, visible, form]);

  const handleValuesChange = (_: unknown, allValues: Record<string, unknown>) => {
    if (typeof allValues.type === 'string' && allValues.type !== currentType) {
      setCurrentType(allValues.type);
    }
  };

  const handleSave = () => {
    form.validate().then((values) => {
      if (!step) return;

      const configMap: Record<string, unknown> = { ...(step.config ?? {}) };
      for (const f of CONFIG_FIELDS_BY_TYPE[values.type] ?? []) {
        const v = values[`config_${f.key}`];
        if (v !== undefined && v !== '') {
          configMap[f.key] = v;
        } else {
          delete configMap[f.key];
        }
      }

      const updated: StepConfig = {
        ...step,
        name: values.name,
        type: values.type,
        image: values.image || undefined,
        command: values.command ? values.command.split('\n').filter(Boolean) : undefined,
        args: values.args ? values.args.split('\n').filter(Boolean) : undefined,
        env: textToEnvMap(values.env || ''),
        config: Object.keys(configMap).length > 0 ? configMap : undefined,
      };
      onSave(updated);
      onClose();
    });
  };

  return (
    <Drawer
      width={440}
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
      <Form form={form} layout="vertical" onValuesChange={handleValuesChange}>
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

        <Divider style={{ margin: '12px 0' }} />

        <Form.Item label="环境变量" field="env">
          <TextArea placeholder={'KEY=value\nDB_HOST=localhost'} autoSize={{ minRows: 2 }} />
        </Form.Item>

        {configFields.length > 0 && (
          <>
            <Divider style={{ margin: '12px 0' }} />
            {configFields.map((f) => (
              <Form.Item key={f.key} label={f.label} field={`config_${f.key}`}>
                <Input placeholder={f.placeholder} />
              </Form.Item>
            ))}
          </>
        )}
      </Form>
    </Drawer>
  );
}
