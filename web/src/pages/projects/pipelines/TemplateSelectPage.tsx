import { useEffect, useState, useCallback } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { Card, Button, Space, Typography, Spin, Message, Form, Input, Divider } from '@arco-design/web-react';
import { IconArrowLeft, IconCheck, IconGithub } from '@arco-design/web-react/icon';
import { fetchTemplates, createPipeline, type PipelineTemplate, type PipelineConfig } from '../../../services/pipeline';
import { configToJson } from '../../../lib/pipeline/configJson';

const { Title, Text } = Typography;

const categoryIcons: Record<string, string> = {
  'go': '🐹',
  'java': '☕',
  'frontend': '📦',
  'docker': '🐳',
  'default': '⚙️',
};

const categoryLabels: Record<string, string> = {
  'go': 'Go',
  'java': 'Java',
  'frontend': '前端',
  'docker': '通用 Docker',
  'default': '通用',
};

export default function TemplateSelectPage() {
  const { id: projectId } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [templates, setTemplates] = useState<PipelineTemplate[]>([]);
  const [loading, setLoading] = useState(true);
  const [selectedTemplate, setSelectedTemplate] = useState<PipelineTemplate | null>(null);
  const [templateParams, setTemplateParams] = useState<Record<string, string>>({});
  const [creating, setCreating] = useState(false);
  const [pipelineName, setPipelineName] = useState('');
  const [pipelineDesc, setPipelineDesc] = useState('');

  useEffect(() => {
    fetchTemplates()
      .then((res) => setTemplates(res.items))
      .catch(() => Message.error('加载模板失败'))
      .finally(() => setLoading(false));
  }, []);

  const handleSelectTemplate = useCallback((template: PipelineTemplate) => {
    setSelectedTemplate(template);
    const params: Record<string, string> = {};
    template.params.forEach((p) => {
      if (p.defaultValue) {
        params[p.name] = p.defaultValue;
      }
    });
    setTemplateParams(params);
  }, []);

  const handleCreate = useCallback(async () => {
    if (!projectId || !selectedTemplate) return;
    if (!pipelineName.trim()) {
      Message.error('请输入流水线名称');
      return;
    }

    // Validate required params
    const missingParams = selectedTemplate.params
      .filter((p) => p.required && !templateParams[p.name])
      .map((p) => p.name);
    if (missingParams.length > 0) {
      Message.error(`请填写必填参数: ${missingParams.join(', ')}`);
      return;
    }

    setCreating(true);
    try {
      const created = await createPipeline(projectId, {
        name: pipelineName.trim(),
        description: pipelineDesc.trim(),
        templateId: selectedTemplate.id,
        templateParams,
      });
      Message.success('创建成功');
      navigate(`/projects/${projectId}/pipelines/${created.id}`);
    } catch {
      Message.error('创建失败');
    } finally {
      setCreating(false);
    }
  }, [projectId, selectedTemplate, pipelineName, pipelineDesc, templateParams, navigate]);

  if (loading) {
    return (
      <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '100%' }}>
        <Spin size={40} />
      </div>
    );
  }

  // If a template is selected, show the config form
  if (selectedTemplate) {
    return (
      <div style={{ padding: 24, maxWidth: 800, margin: '0 auto' }}>
        <Button type="text" icon={<IconArrowLeft />} onClick={() => setSelectedTemplate(null)} style={{ marginBottom: 16 }}>
          返回模板列表
        </Button>

        <Card>
          <Title heading={5}>{selectedTemplate.name}</Title>
          <Text type="secondary">{selectedTemplate.description}</Text>

          <Divider />

          <Form layout="vertical">
            <Form.Item label="流水线名称" required>
              <Input
                value={pipelineName}
                onChange={setPipelineName}
                placeholder="例如: my-app-pipeline"
              />
            </Form.Item>
            <Form.Item label="描述">
              <Input.TextArea
                value={pipelineDesc}
                onChange={setPipelineDesc}
                placeholder="可选描述"
                autoSize={{ minRows: 2 }}
              />
            </Form.Item>

            {selectedTemplate.params.length > 0 && (
              <>
                <Divider>模板参数</Divider>
                {selectedTemplate.params.map((param) => (
                  <Form.Item
                    key={param.name}
                    label={param.name + (param.required ? ' *' : '')}
                    extra={param.description}
                  >
                    <Input
                      value={templateParams[param.name] || ''}
                      onChange={(v) => setTemplateParams((prev) => ({ ...prev, [param.name]: v }))}
                      placeholder={param.defaultValue || `请输入 ${param.name}`}
                    />
                  </Form.Item>
                ))}
              </>
            )}

            {selectedTemplate.config && (
              <>
                <Divider>预览配置</Divider>
                <pre style={{
                  background: 'var(--zcid-color-bg-secondary)',
                  padding: 12,
                  borderRadius: 4,
                  fontSize: 12,
                  maxHeight: 200,
                  overflow: 'auto',
                }}>
                  {configToJson(selectedTemplate.config)}
                </pre>
              </>
            )}

            <Divider />

            <Form.Item>
              <Space>
                <Button type="primary" icon={<IconCheck />} onClick={handleCreate} loading={creating}>
                  创建流水线
                </Button>
                <Button onClick={() => setSelectedTemplate(null)}>返回</Button>
              </Space>
            </Form.Item>
          </Form>
        </Card>
      </div>
    );
  }

  // Show template list
  return (
    <div style={{ padding: 24 }}>
      <div style={{ marginBottom: 24 }}>
        <Title heading={4}>选择流水线模板</Title>
        <Text type="secondary">从模板快速创建流水线，或从空白开始</Text>
      </div>

      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(280px, 1fr))', gap: 16 }}>
        {/* Blank template option */}
        <Card
          hoverable
          style={{ cursor: 'pointer', border: '2px dashed var(--zcid-color-border)' }}
          onClick={() => {
            if (!projectId) return;
            navigate(`/projects/${projectId}/pipelines/blank`);
          }}
        >
          <div style={{ textAlign: 'center', padding: 24 }}>
            <div style={{ fontSize: 32, marginBottom: 8 }}>➕</div>
            <Title heading={6} style={{ marginBottom: 4 }}>空白流水线</Title>
            <Text type="secondary">从零开始手动配置</Text>
          </div>
        </Card>

        {/* Template cards */}
        {templates.map((template) => (
          <Card
            key={template.id}
            hoverable
            style={{ cursor: 'pointer' }}
            onClick={() => handleSelectTemplate(template)}
          >
            <div style={{ display: 'flex', alignItems: 'flex-start', gap: 12 }}>
              <div style={{ fontSize: 32 }}>{categoryIcons[template.category] || categoryIcons.default}</div>
              <div style={{ flex: 1 }}>
                <Title heading={6} style={{ marginBottom: 4 }}>{template.name}</Title>
                <Text type="secondary" style={{ fontSize: 12 }}>{template.description}</Text>
                <div style={{ marginTop: 8 }}>
                  <Text type="secondary" style={{ fontSize: 12 }}>
                    {categoryLabels[template.category] || template.category}
                    {template.params.length > 0 && ` • ${template.params.length} 个参数`}
                  </Text>
                </div>
              </div>
            </div>
          </Card>
        ))}
      </div>
    </div>
  );
}
