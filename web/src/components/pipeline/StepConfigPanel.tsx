import { useEffect, useMemo, useState } from 'react';
import { Drawer, Form, Input, Select, Button, Space, Typography, Divider, Tag } from '@arco-design/web-react';
import type { StepConfig } from '../../services/pipeline';

const { TextArea } = Input;
const { Text } = Typography;

interface StepConfigPanelProps {
  visible: boolean;
  step: StepConfig | null;
  onClose: () => void;
  onSave: (step: StepConfig) => void;
}

const stepTypes = [
  { label: 'Git Clone', value: 'git-clone', icon: '📥', desc: '克隆代码仓库' },
  { label: 'Shell 脚本', value: 'shell', icon: '💻', desc: '执行 Shell 命令或脚本' },
  { label: 'Kaniko 构建', value: 'kaniko', icon: '🐳', desc: 'Kaniko 容器镜像构建' },
  { label: 'BuildKit 构建', value: 'buildkit', icon: '🔨', desc: 'BuildKit 容器镜像构建' },
];

const stepTypeColors: Record<string, string> = {
  'git-clone': '#165DFF',
  shell: '#00B42A',
  kaniko: '#FF7D00',
  buildkit: '#722ED1',
};

const CONFIG_FIELDS_BY_TYPE: Record<string, { key: string; label: string; placeholder: string; type?: string }[]> = {
  'git-clone': [
    { key: 'repoUrl', label: '仓库地址', placeholder: 'https://github.com/user/repo.git' },
    { key: 'branch', label: '分支', placeholder: 'main', type: 'branch-select' },
    { key: 'depth', label: '克隆深度', placeholder: '留空为完整克隆，输入数字如 1 表示浅克隆' },
  ],
  kaniko: [
    { key: 'imageName', label: '镜像名称', placeholder: 'registry.example.com/app:latest' },
    { key: 'dockerfile', label: 'Dockerfile 路径', placeholder: 'Dockerfile' },
    { key: 'context', label: '构建上下文', placeholder: '默认为 .（当前目录）' },
    { key: 'buildArgs', label: '构建参数 (Build Args)', placeholder: 'VERSION=1.0.0\nNODE_ENV=production', type: 'textarea' },
    { key: 'target', label: '目标阶段 (Target)', placeholder: '多阶段构建的目标 stage 名称' },
    { key: 'extraTags', label: '额外标签', placeholder: 'registry.example.com/app:v1.0\nregistry.example.com/app:sha-abc123', type: 'textarea' },
  ],
  buildkit: [
    { key: 'imageName', label: '镜像名称', placeholder: 'registry.example.com/app:latest' },
    { key: 'dockerfile', label: 'Dockerfile 路径', placeholder: 'Dockerfile' },
    { key: 'context', label: '构建上下文', placeholder: '默认为 .（当前目录）' },
    { key: 'buildArgs', label: '构建参数 (Build Args)', placeholder: 'VERSION=1.0.0\nNODE_ENV=production', type: 'textarea' },
    { key: 'target', label: '目标阶段 (Target)', placeholder: '多阶段构建的目标 stage 名称' },
    { key: 'extraTags', label: '额外标签', placeholder: 'registry.example.com/app:v1.0', type: 'textarea' },
    { key: 'secrets', label: '构建密钥 (Secrets)', placeholder: 'id=mysecret,src=/path/to/secret', type: 'textarea' },
    { key: 'platform', label: '目标平台', placeholder: 'linux/amd64', type: 'platform-select' },
  ],
};

const PLATFORM_OPTIONS = [
  'linux/amd64',
  'linux/arm64',
  'linux/amd64,linux/arm64',
];

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

function SectionHeader({ title, icon }: { title: string; icon?: string }) {
  return (
    <div style={{
      display: 'flex', alignItems: 'center', gap: 6,
      padding: '8px 0 4px', marginTop: 4,
      fontSize: 13, fontWeight: 600, color: 'var(--color-text-2)',
    }}>
      {icon && <span>{icon}</span>}
      {title}
    </div>
  );
}

export function StepConfigPanel({ visible, step, onClose, onSave }: StepConfigPanelProps) {
  const [form] = Form.useForm();
  const [currentType, setCurrentType] = useState<string>(step?.type ?? 'shell');

  const configFields = useMemo(
    () => CONFIG_FIELDS_BY_TYPE[currentType] ?? [],
    [currentType]
  );

  const isDockerType = currentType === 'kaniko' || currentType === 'buildkit';
  const isShellType = currentType === 'shell';
  const isGitClone = currentType === 'git-clone';
  const typeInfo = stepTypes.find((t) => t.value === currentType);

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

  function renderConfigField(f: { key: string; label: string; placeholder: string; type?: string }) {
    if (f.type === 'branch-select') {
      return (
        <Select
          showSearch
          allowCreate
          placeholder={f.placeholder}
          options={['main', 'master', 'develop', 'staging', 'release'].map((b) => ({ label: b, value: b }))}
        />
      );
    }
    if (f.type === 'platform-select') {
      return (
        <Select
          showSearch
          allowCreate
          placeholder={f.placeholder}
          options={PLATFORM_OPTIONS.map((p) => ({ label: p, value: p }))}
        />
      );
    }
    if (f.type === 'textarea') {
      return <TextArea placeholder={f.placeholder} autoSize={{ minRows: 2, maxRows: 6 }} style={{ fontFamily: 'monospace', fontSize: 13 }} />;
    }
    return <Input placeholder={f.placeholder} />;
  }

  return (
    <Drawer
      width={480}
      title={
        <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
          <div style={{
            width: 32, height: 32, borderRadius: 8,
            background: stepTypeColors[currentType] || '#86909C',
            display: 'flex', alignItems: 'center', justifyContent: 'center',
            fontSize: 16,
          }}>
            {typeInfo?.icon || '⚙️'}
          </div>
          <div>
            <div style={{ fontSize: 15, fontWeight: 600 }}>Step 配置</div>
            <Text type="secondary" style={{ fontSize: 12 }}>{typeInfo?.desc || '配置步骤参数'}</Text>
          </div>
        </div>
      }
      visible={visible}
      onCancel={onClose}
      footer={
        <div style={{ display: 'flex', justifyContent: 'flex-end', gap: 8 }}>
          <Button onClick={onClose} style={{ borderRadius: 6 }}>取消</Button>
          <Button type="primary" onClick={handleSave} style={{ borderRadius: 6 }}>保存</Button>
        </div>
      }
    >
      <Form form={form} layout="vertical" onValuesChange={handleValuesChange}>
        {/* Basic Info */}
        <SectionHeader title="基本信息" icon="📋" />
        <Form.Item label="名称" field="name" rules={[{ required: true, message: '请输入 Step 名称' }]}>
          <Input placeholder="Step 名称" />
        </Form.Item>
        <Form.Item label="类型" field="type" rules={[{ required: true, message: '请选择 Step 类型' }]}>
          <Select placeholder="选择类型">
            {stepTypes.map((t) => (
              <Select.Option key={t.value} value={t.value}>
                <span style={{ marginRight: 8 }}>{t.icon}</span>
                {t.label}
                <Text type="secondary" style={{ marginLeft: 8, fontSize: 12 }}>{t.desc}</Text>
              </Select.Option>
            ))}
          </Select>
        </Form.Item>

        {/* Type-specific config fields */}
        {configFields.length > 0 && (
          <>
            <Divider style={{ margin: '12px 0' }} />
            <SectionHeader
              title={isDockerType ? '容器构建配置' : isGitClone ? 'Git 仓库配置' : '专用配置'}
              icon={isDockerType ? '🐳' : isGitClone ? '📥' : '⚙️'}
            />
            {configFields.map((f) => (
              <Form.Item key={f.key} label={f.label} field={`config_${f.key}`}>
                {renderConfigField(f)}
              </Form.Item>
            ))}
          </>
        )}

        {/* Shell/execution section - only for shell and docker types */}
        {(isShellType || isDockerType) && (
          <>
            <Divider style={{ margin: '12px 0' }} />
            <SectionHeader title="运行环境" icon="💻" />
            <Form.Item label="容器镜像" field="image" extra={isShellType ? '执行命令的容器镜像' : '构建工具的基础镜像（通常自动设置）'}>
              <Input placeholder="例如 golang:1.24, node:20-alpine, ubuntu:22.04" />
            </Form.Item>
          </>
        )}

        {isShellType && (
          <>
            <Divider style={{ margin: '12px 0' }} />
            <SectionHeader title="Shell 脚本" icon="📝" />
            <Form.Item label="命令 / 脚本" field="command" extra="每行一条命令，将按顺序执行">
              <TextArea
                placeholder={'#!/bin/bash\nset -e\n\n# 编译\ngo build -o app ./cmd/server\n\n# 单元测试\ngo test ./... -v'}
                autoSize={{ minRows: 6, maxRows: 20 }}
                style={{
                  fontFamily: '"Fira Code", "Consolas", "Monaco", monospace',
                  fontSize: 13, lineHeight: 1.5,
                  background: '#1D2129', color: '#C9CDD4',
                  borderRadius: 8, padding: 12,
                }}
              />
            </Form.Item>
            <Form.Item label="参数" field="args" extra="传递给命令的参数，每行一个">
              <TextArea
                placeholder={'--verbose\n--output=./dist'}
                autoSize={{ minRows: 2, maxRows: 6 }}
                style={{ fontFamily: 'monospace', fontSize: 13 }}
              />
            </Form.Item>
          </>
        )}

        {/* Environment variables - all types except git-clone */}
        {!isGitClone && (
          <>
            <Divider style={{ margin: '12px 0' }} />
            <SectionHeader title="环境变量" icon="🔧" />
            <Form.Item label="变量列表" field="env" extra="格式：KEY=value，每行一个">
              <TextArea
                placeholder={'GOPROXY=https://goproxy.cn,direct\nCGO_ENABLED=0\nGOOS=linux'}
                autoSize={{ minRows: 3, maxRows: 8 }}
                style={{ fontFamily: 'monospace', fontSize: 13 }}
              />
            </Form.Item>
          </>
        )}
      </Form>
    </Drawer>
  );
}
