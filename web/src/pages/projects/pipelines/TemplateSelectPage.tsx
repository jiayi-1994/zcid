import { useEffect, useState, useCallback } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { Card, Button, Space, Typography, Spin, Message, Form, Input, Divider, Tag, Grid } from '@arco-design/web-react';
import { IconArrowLeft, IconCheck } from '@arco-design/web-react/icon';
import { fetchTemplates, createPipeline, type PipelineTemplate } from '../../../services/pipeline';
import { configToJson } from '../../../lib/pipeline/configJson';

const { Title, Text, Paragraph } = Typography;
const { Row, Col } = Grid;

const categoryConfig: Record<string, { icon: string; color: string; bg: string }> = {
  backend:  { icon: '⚙️', color: '#165DFF', bg: 'linear-gradient(135deg, #E8F3FF 0%, #D6E4FF 100%)' },
  frontend: { icon: '📦', color: '#0FC6C2', bg: 'linear-gradient(135deg, #E8FFFB 0%, #B5F5EC 100%)' },
  general:  { icon: '🐳', color: '#FF7D00', bg: 'linear-gradient(135deg, #FFF7E8 0%, #FFE4BA 100%)' },
};

const categoryLabels: Record<string, string> = {
  backend: '后端',
  frontend: '前端',
  general: '通用',
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
      if (p.defaultValue) params[p.name] = p.defaultValue;
    });
    setTemplateParams(params);
  }, []);

  const handleCreate = useCallback(async () => {
    if (!projectId || !selectedTemplate) return;
    if (!pipelineName.trim()) {
      Message.error('请输入流水线名称');
      return;
    }
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

  if (selectedTemplate) {
    const catCfg = categoryConfig[selectedTemplate.category] || categoryConfig.general;
    return (
      <div style={{ padding: 24, maxWidth: 720, margin: '0 auto' }}>
        <Button type="text" icon={<IconArrowLeft />} onClick={() => setSelectedTemplate(null)} style={{ marginBottom: 16, borderRadius: 6 }}>
          返回模板列表
        </Button>

        <Card style={{ borderRadius: 12, overflow: 'hidden' }} bodyStyle={{ padding: 0 }}>
          {/* Header */}
          <div style={{
            background: catCfg.bg,
            padding: '24px 28px', borderBottom: '1px solid var(--color-border)',
          }}>
            <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
              <div style={{
                width: 48, height: 48, borderRadius: 12,
                background: '#fff', boxShadow: '0 2px 8px rgba(0,0,0,0.08)',
                display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: 24,
              }}>
                {catCfg.icon}
              </div>
              <div>
                <Title heading={5} style={{ margin: 0 }}>{selectedTemplate.name}</Title>
                <Text type="secondary" style={{ fontSize: 13 }}>{selectedTemplate.description}</Text>
              </div>
            </div>
          </div>

          {/* Form */}
          <div style={{ padding: '24px 28px' }}>
            <Form layout="vertical">
              <Row gutter={16}>
                <Col span={14}>
                  <Form.Item label={<span style={{ fontWeight: 600 }}>流水线名称 <span style={{ color: '#F53F3F' }}>*</span></span>}>
                    <Input
                      value={pipelineName}
                      onChange={setPipelineName}
                      placeholder="例如: my-app-pipeline"
                      style={{ borderRadius: 6 }}
                    />
                  </Form.Item>
                </Col>
                <Col span={10}>
                  <Form.Item label="描述">
                    <Input
                      value={pipelineDesc}
                      onChange={setPipelineDesc}
                      placeholder="可选"
                      style={{ borderRadius: 6 }}
                    />
                  </Form.Item>
                </Col>
              </Row>

              {selectedTemplate.params.length > 0 && (
                <>
                  <Divider style={{ margin: '16px 0' }}>
                    <Text type="secondary" style={{ fontSize: 13 }}>模板参数</Text>
                  </Divider>
                  <Row gutter={16}>
                    {selectedTemplate.params.map((param) => (
                      <Col span={param.name === 'repoUrl' ? 24 : 12} key={param.name}>
                        <Form.Item
                          label={
                            <span style={{ fontWeight: 500 }}>
                              {param.name}
                              {param.required && <span style={{ color: '#F53F3F', marginLeft: 2 }}>*</span>}
                            </span>
                          }
                          extra={<span style={{ fontSize: 12 }}>{param.description}</span>}
                        >
                          <Input
                            value={templateParams[param.name] || ''}
                            onChange={(v) => setTemplateParams((prev) => ({ ...prev, [param.name]: v }))}
                            placeholder={param.defaultValue || `请输入 ${param.name}`}
                            style={{ borderRadius: 6 }}
                          />
                        </Form.Item>
                      </Col>
                    ))}
                  </Row>
                </>
              )}

              {selectedTemplate.config && (
                <>
                  <Divider style={{ margin: '16px 0' }}>
                    <Text type="secondary" style={{ fontSize: 13 }}>配置预览</Text>
                  </Divider>
                  <pre style={{
                    background: '#1D2129', color: '#C9CDD4',
                    padding: 16, borderRadius: 8,
                    fontSize: 12, lineHeight: 1.6, maxHeight: 200, overflow: 'auto',
                    fontFamily: '"Fira Code", "Consolas", monospace',
                  }}>
                    {configToJson(selectedTemplate.config)}
                  </pre>
                </>
              )}

              <div style={{ marginTop: 24, display: 'flex', gap: 12 }}>
                <Button type="primary" icon={<IconCheck />} onClick={handleCreate} loading={creating} style={{ borderRadius: 6, height: 40, padding: '0 24px' }}>
                  创建流水线
                </Button>
                <Button onClick={() => setSelectedTemplate(null)} style={{ borderRadius: 6, height: 40 }}>返回</Button>
              </div>
            </Form>
          </div>
        </Card>
      </div>
    );
  }

  return (
    <div style={{ padding: 24, maxWidth: 960, margin: '0 auto' }}>
      <div style={{ marginBottom: 28, textAlign: 'center' }}>
        <Title heading={4} style={{ marginBottom: 4 }}>选择流水线模板</Title>
        <Text type="secondary">从模板快速创建流水线，或从空白开始</Text>
      </div>

      <Row gutter={[16, 16]}>
        {/* Blank pipeline card */}
        <Col span={8}>
          <div
            onClick={() => { if (projectId) navigate(`/projects/${projectId}/pipelines/blank`); }}
            style={{
              cursor: 'pointer', borderRadius: 10, height: 140,
              border: '2px dashed var(--border)',
              display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center',
              transition: 'border-color 0.2s', background: 'var(--card)',
            }}
            onMouseEnter={(e) => { (e.currentTarget as HTMLElement).style.borderColor = '#A1A1AA'; }}
            onMouseLeave={(e) => { (e.currentTarget as HTMLElement).style.borderColor = 'var(--border)'; }}
          >
            <div style={{
              width: 40, height: 40, borderRadius: 10,
              background: 'var(--muted)', border: '1.5px dashed #A1A1AA',
              display: 'flex', alignItems: 'center', justifyContent: 'center',
              fontSize: 20, marginBottom: 8, color: '#71717A',
            }}>
              ＋
            </div>
            <div style={{ fontSize: 14, fontWeight: 600, color: 'var(--foreground)' }}>空白流水线</div>
            <div style={{ fontSize: 12, color: 'var(--muted-foreground)', marginTop: 2 }}>从零开始手动配置</div>
          </div>
        </Col>

        {/* Template cards */}
        {templates.map((template) => {
          const catCfg = categoryConfig[template.category] || categoryConfig.general;
          return (
            <Col span={8} key={template.id}>
              <div
                onClick={() => handleSelectTemplate(template)}
                style={{
                  cursor: 'pointer', borderRadius: 10, height: 140,
                  border: '1px solid var(--border)', padding: '16px',
                  display: 'flex', gap: 12, transition: 'border-color 0.2s',
                  background: 'var(--card)',
                }}
                onMouseEnter={(e) => { (e.currentTarget as HTMLElement).style.borderColor = '#A1A1AA'; }}
                onMouseLeave={(e) => { (e.currentTarget as HTMLElement).style.borderColor = 'var(--border)'; }}
              >
                <div style={{
                  width: 40, height: 40, borderRadius: 8, flexShrink: 0,
                  background: catCfg.bg,
                  display: 'flex', alignItems: 'center', justifyContent: 'center',
                  fontSize: 20,
                }}>
                  {catCfg.icon}
                </div>
                <div style={{ flex: 1, minWidth: 0, display: 'flex', flexDirection: 'column' }}>
                  <div style={{ fontSize: 14, fontWeight: 600, color: 'var(--foreground)', marginBottom: 4 }}>{template.name}</div>
                  <div style={{
                    fontSize: 12, color: 'var(--muted-foreground)', lineHeight: 1.5,
                    flex: 1, overflow: 'hidden',
                    display: '-webkit-box', WebkitLineClamp: 2, WebkitBoxOrient: 'vertical' as const,
                  }}>
                    {template.description}
                  </div>
                  <div style={{ display: 'flex', gap: 6, marginTop: 8 }}>
                    <Tag size="small" style={{ borderRadius: 999, fontSize: 11 }}>
                      {categoryLabels[template.category] || template.category}
                    </Tag>
                    {template.params.length > 0 && (
                      <Tag size="small" color="gray" style={{ borderRadius: 999, fontSize: 11 }}>
                        {template.params.length} 个参数
                      </Tag>
                    )}
                  </div>
                </div>
              </div>
            </Col>
          );
        })}
      </Row>
    </div>
  );
}
