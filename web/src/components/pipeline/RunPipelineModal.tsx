import { useState } from 'react';
import { Modal, Typography, Form, Input, Button } from '@arco-design/web-react';
import type { PipelineSummary } from '../../services/pipeline';

const { Paragraph } = Typography;

interface RunPipelineModalProps {
  visible: boolean;
  pipeline: PipelineSummary | null;
  onClose: () => void;
  onSubmit?: (params: { params?: Record<string, string>; gitBranch?: string; gitCommit?: string }) => Promise<void>;
}

export function RunPipelineModal({ visible, pipeline, onClose, onSubmit }: RunPipelineModalProps) {
  const [loading, setLoading] = useState(false);
  const [form] = Form.useForm();

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
        <Paragraph>功能开发中（Pipeline 执行属于 Epic 7）</Paragraph>
      </Modal>
    );
  }

  return (
    <Modal
      title={`运行流水线: ${pipeline?.name ?? ''}`}
      visible={visible}
      onCancel={onClose}
      footer={
        <>
          <Button onClick={onClose}>取消</Button>
          <Button type="primary" loading={loading} onClick={handleSubmit}>
            运行
          </Button>
        </>
      }
    >
      <Form form={form} layout="vertical">
        <Form.Item label="Git 分支" field="gitBranch">
          <Input placeholder="例如 main" />
        </Form.Item>
        <Form.Item label="Git Commit" field="gitCommit">
          <Input placeholder="可选，指定 commit SHA" />
        </Form.Item>
      </Form>
    </Modal>
  );
}
