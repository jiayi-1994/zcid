import { useState, useEffect, useCallback } from 'react';
import { Modal, Form, Input, Button, Select, Space, Tag, Divider, Typography } from '@arco-design/web-react';
import { IconBranch, IconCode } from '@arco-design/web-react/icon';
import type { PipelineSummary, PipelineConfig } from '../../services/pipeline';
import { fetchConnections, listBranches, type GitBranch, type GitConnection } from '../../services/integration';

const { Text } = Typography;
const DEFAULT_BRANCHES = ['main', 'master', 'develop', 'staging', 'release'];

interface RunPipelineModalProps {
  visible: boolean;
  pipeline: PipelineSummary | null;
  pipelineConfig?: PipelineConfig | null;
  onClose: () => void;
  onSubmit?: (params: { params?: Record<string, string>; gitBranch?: string; gitCommit?: string }) => Promise<void>;
}

function extractRepoUrl(config?: PipelineConfig | null): string | undefined {
  if (!config?.stages) return undefined;
  for (const stage of config.stages) {
    for (const step of stage.steps) {
      if (step.type === 'git-clone' && step.config) {
        const url = step.config.repoUrl;
        if (typeof url === 'string' && url) return url;
      }
    }
  }
  return undefined;
}

function extractConfiguredBranch(config?: PipelineConfig | null): string | undefined {
  if (!config?.stages) return undefined;
  for (const stage of config.stages) {
    for (const step of stage.steps) {
      if (step.type === 'git-clone' && step.config) {
        const branch = step.config.branch;
        if (typeof branch === 'string' && branch) return branch;
      }
    }
  }
  return undefined;
}

function repoFullNameFromUrl(repoUrl: string): string | undefined {
  const m = repoUrl.match(/[:/]([^/:]+\/[^/.]+?)(?:\.git)?$/);
  return m?.[1];
}

function matchConnection(repoUrl: string, connections: GitConnection[]): GitConnection | undefined {
  try {
    const urlObj = new URL(repoUrl);
    return connections.find((c) => {
      try {
        const connUrl = new URL(c.serverUrl);
        return connUrl.hostname === urlObj.hostname;
      } catch {
        return false;
      }
    });
  } catch {
    return undefined;
  }
}

const triggerLabels: Record<string, string> = {
  manual: '手动触发',
  webhook: 'Webhook 触发',
  scheduled: '定时触发',
};

export function RunPipelineModal({ visible, pipeline, pipelineConfig, onClose, onSubmit }: RunPipelineModalProps) {
  const [loading, setLoading] = useState(false);
  const [form] = Form.useForm();
  const [branchOptions, setBranchOptions] = useState<string[]>(DEFAULT_BRANCHES);
  const [branchLoading, setBranchLoading] = useState(false);

  const loadBranches = useCallback(async () => {
    const repoUrl = extractRepoUrl(pipelineConfig);
    if (!repoUrl) {
      const configured = extractConfiguredBranch(pipelineConfig);
      if (configured) {
        const opts = new Set(DEFAULT_BRANCHES);
        opts.add(configured);
        setBranchOptions(Array.from(opts));
      }
      return;
    }

    const repoFullName = repoFullNameFromUrl(repoUrl);
    if (!repoFullName) return;

    setBranchLoading(true);
    try {
      const connList = await fetchConnections();
      const conn = matchConnection(repoUrl, connList.items ?? []);
      if (!conn) {
        const configured = extractConfiguredBranch(pipelineConfig);
        if (configured) {
          const opts = new Set(DEFAULT_BRANCHES);
          opts.add(configured);
          setBranchOptions(Array.from(opts));
        }
        return;
      }
      const branches = await listBranches(conn.id, repoFullName);
      if (branches.length > 0) {
        const names = branches.map((b: GitBranch) => b.name);
        const defaultBranch = branches.find((b: GitBranch) => b.isDefault);
        setBranchOptions(names);
        if (defaultBranch && !form.getFieldValue('gitBranch')) {
          form.setFieldValue('gitBranch', defaultBranch.name);
        }
      }
    } catch {
      const configured = extractConfiguredBranch(pipelineConfig);
      if (configured) {
        const opts = new Set(DEFAULT_BRANCHES);
        opts.add(configured);
        setBranchOptions(Array.from(opts));
      }
    } finally {
      setBranchLoading(false);
    }
  }, [pipelineConfig, form]);

  useEffect(() => {
    if (visible) loadBranches();
  }, [visible, loadBranches]);

  const handleSubmit = async () => {
    if (!onSubmit) return;
    try {
      await form.validate();
      setLoading(true);
      await onSubmit({
        gitBranch: form.getFieldValue('gitBranch') || undefined,
        gitCommit: form.getFieldValue('gitCommit') || undefined,
        params: {},
      });
      form.resetFields();
    } catch {
      // validation or submit error
    } finally {
      setLoading(false);
    }
  };

  if (!onSubmit) {
    return (
      <Modal title={`运行流水线: ${pipeline?.name ?? ''}`} visible={visible} onCancel={onClose} footer={null}>
        <p>功能开发中（Pipeline 执行属于 Epic 7）</p>
      </Modal>
    );
  }

  return (
    <Modal
      title={null}
      visible={visible}
      onCancel={onClose}
      style={{ borderRadius: 12, overflow: 'hidden' }}
      footer={
        <div style={{ display: 'flex', justifyContent: 'flex-end', gap: 8, padding: '4px 0' }}>
          <Button onClick={onClose} style={{ borderRadius: 6 }}>取消</Button>
          <Button type="primary" loading={loading} onClick={handleSubmit} icon={<IconCode />} style={{ borderRadius: 6 }}>
            开始运行
          </Button>
        </div>
      }
    >
      <div style={{ marginBottom: 20 }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: 12, marginBottom: 8 }}>
          <div style={{
            width: 40, height: 40, borderRadius: 10,
            background: 'linear-gradient(135deg, #165DFF 0%, #0FC6C2 100%)',
            display: 'flex', alignItems: 'center', justifyContent: 'center',
          }}>
            <IconCode style={{ color: '#fff', fontSize: 20 }} />
          </div>
          <div>
            <div style={{ fontSize: 16, fontWeight: 600, color: 'var(--color-text-1)' }}>
              运行 {pipeline?.name ?? '流水线'}
            </div>
            <Text type="secondary" style={{ fontSize: 13 }}>
              {triggerLabels[pipeline?.triggerType ?? ''] || '手动触发'} · {pipeline?.concurrencyPolicy === 'reject' ? '拒绝并发' : '队列模式'}
            </Text>
          </div>
        </div>
      </div>

      <Form form={form} layout="vertical">
        <Form.Item label={
          <Space size={4}><IconBranch style={{ fontSize: 14 }} /><span>构建分支</span></Space>
        } field="gitBranch">
          <Select
            showSearch
            allowCreate
            placeholder="选择或输入分支名"
            loading={branchLoading}
            options={branchOptions.map((b) => ({ label: b, value: b }))}
            style={{ borderRadius: 6 }}
          />
        </Form.Item>
        <Form.Item label="Git Commit（可选）" field="gitCommit">
          <Input placeholder="指定 commit SHA，留空则使用分支最新" style={{ borderRadius: 6 }} />
        </Form.Item>
      </Form>
    </Modal>
  );
}
