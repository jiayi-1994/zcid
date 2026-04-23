import { useEffect, useState, useCallback } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { Button, Space, Spin, Message, Form, Input, Grid } from '@arco-design/web-react';
import { IconArrowLeft, IconCheck } from '@arco-design/web-react/icon';
import { fetchTemplates, createPipeline, type PipelineTemplate } from '../../../services/pipeline';
import { configToJson } from '../../../lib/pipeline/configJson';

const { Row, Col } = Grid;

const categoryConfig: Record<string, { icon: string }> = {
  backend:  { icon: '⚙️' },
  frontend: { icon: '📦' },
  general:  { icon: '🐳' },
};

const langIcons: Record<string, string> = {
  'Go': '🔵',
  'Java Maven': '☕',
  'Java JAR': '☕',
  'Node.js': '🟢',
  'Docker': '🐳',
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

  const currentStep = selectedTemplate ? 2 : 1;

  return (
    <div className="page-container" style={{ maxWidth: 960, margin: '0 auto' }}>
      {/* Back button */}
      <Button
        type="text"
        icon={<IconArrowLeft />}
        onClick={() => selectedTemplate ? setSelectedTemplate(null) : navigate(-1)}
        style={{ marginBottom: 16, borderRadius: 6, color: 'var(--muted-foreground)' }}
      >
        {selectedTemplate ? '返回模板列表' : '返回'}
      </Button>

      <div className="page-header">
        <div>
          <div className="breadcrumb">Create Pipeline</div>
          <h1 className="page-title">Architect New Pipeline</h1>
          <p className="page-subtitle">
            Real-time topology preview. 从模板快速创建流水线，或从空白开始构建。
          </p>
        </div>
      </div>

      {/* Wizard Steps */}
      <div className="wizard-steps">
        <div className={`wizard-step ${currentStep >= 1 ? (currentStep > 1 ? 'wizard-step--completed' : 'wizard-step--active') : ''}`}>
          <span className="wizard-step-number">{currentStep > 1 ? '✓' : '1'}</span>
          <div>
            <div className="wizard-step-label">Template Selection</div>
            <div className="wizard-step-desc">选择语言和构建模板</div>
          </div>
        </div>
        <div className={`wizard-step ${currentStep >= 2 ? 'wizard-step--active' : ''}`}>
          <span className="wizard-step-number">2</span>
          <div>
            <div className="wizard-step-label">Configuration</div>
            <div className="wizard-step-desc">配置参数和仓库信息</div>
          </div>
        </div>
      </div>

      {/* Step 1: Template Selection */}
      {!selectedTemplate && (
        <Row gutter={[16, 16]}>
          {/* Blank pipeline */}
          <Col span={6}>
            <div
              className="template-card"
              onClick={() => { if (projectId) navigate(`/projects/${projectId}/pipelines/blank`); }}
              style={{ height: 180 }}
            >
              <div className="template-card-icon">＋</div>
              <div className="template-card-name">Custom</div>
              <div className="template-card-desc">从零开始配置</div>
            </div>
          </Col>

          {templates.map((template) => {
            const icon = langIcons[template.name] || categoryConfig[template.category]?.icon || '📦';
            return (
              <Col span={6} key={template.id}>
                <div
                  className="template-card"
                  onClick={() => handleSelectTemplate(template)}
                  style={{ height: 180 }}
                >
                  <div className="template-card-icon">{icon}</div>
                  <div className="template-card-name">{template.name}</div>
                  <div className="template-card-desc">{template.description}</div>
                </div>
              </Col>
            );
          })}
        </Row>
      )}

      {/* Step 2: Configuration */}
      {selectedTemplate && (
        <Row gutter={24}>
          {/* Config Form */}
          <Col span={14}>
            <div className="config-panel">
              <div className="config-panel-header">
                Configuration Parameters
              </div>
              <div className="config-panel-body">
                <Form layout="vertical">
                  <Row gutter={16}>
                    <Col span={24}>
                      <Form.Item label={<span style={{ fontWeight: 600 }}>流水线名称 <span style={{ color: 'var(--destructive)' }}>*</span></span>}>
                        <Input
                          value={pipelineName}
                          onChange={setPipelineName}
                          placeholder="例如: my-app-pipeline"
                          style={{ borderRadius: 8, height: 40 }}
                        />
                      </Form.Item>
                    </Col>
                    <Col span={24}>
                      <Form.Item label="描述">
                        <Input
                          value={pipelineDesc}
                          onChange={setPipelineDesc}
                          placeholder="可选描述"
                          style={{ borderRadius: 8, height: 40 }}
                        />
                      </Form.Item>
                    </Col>
                  </Row>

                  {selectedTemplate.params.length > 0 && (
                    <>
                      <div style={{
                        fontSize: 11, fontWeight: 600, color: 'var(--on-surface-variant)',
                        textTransform: 'uppercase', letterSpacing: '0.08em',
                        margin: '20px 0 12px', paddingTop: 16,
                      }}>
                        Template Parameters
                      </div>
                      <Row gutter={16}>
                        {selectedTemplate.params.map((param) => (
                          <Col span={param.name === 'repoUrl' ? 24 : 12} key={param.name}>
                            <Form.Item
                              label={
                                <span style={{ fontWeight: 500 }}>
                                  {param.name}
                                  {param.required && <span style={{ color: 'var(--destructive)', marginLeft: 2 }}>*</span>}
                                </span>
                              }
                              extra={<span style={{ fontSize: 12, color: 'var(--muted-foreground)' }}>{param.description}</span>}
                            >
                              <Input
                                value={templateParams[param.name] || ''}
                                onChange={(v) => setTemplateParams((prev) => ({ ...prev, [param.name]: v }))}
                                placeholder={param.defaultValue || `请输入 ${param.name}`}
                                style={{ borderRadius: 8, height: 40 }}
                              />
                            </Form.Item>
                          </Col>
                        ))}
                      </Row>
                    </>
                  )}

                  <div style={{ marginTop: 24, display: 'flex', gap: 12 }}>
                    <Button
                      type="primary"
                      icon={<IconCheck />}
                      onClick={handleCreate}
                      loading={creating}
                      style={{ borderRadius: 8, height: 42, padding: '0 28px', fontWeight: 600 }}
                    >
                      创建流水线
                    </Button>
                    <Button
                      onClick={() => setSelectedTemplate(null)}
                      style={{ borderRadius: 8, height: 42 }}
                    >
                      返回
                    </Button>
                  </div>
                </Form>
              </div>
            </div>
          </Col>

          {/* Stage Preview */}
          <Col span={10}>
            <div className="config-panel">
              <div className="config-panel-header">
                Real-time Stage Preview
              </div>
              <div className="config-panel-body">
                {selectedTemplate.config ? (
                  <>
                    <div className="stage-preview">
                      {selectedTemplate.config.stages.map((stage, i) => (
                        <Space key={stage.id} size={0}>
                          {i > 0 && <div className="stage-preview-connector" />}
                          <div className="stage-preview-node">
                            <div style={{
                              width: 32, height: 32, borderRadius: 8,
                              background: 'var(--primary-light)', color: 'var(--primary)',
                              display: 'flex', alignItems: 'center', justifyContent: 'center',
                              fontSize: 14, fontWeight: 700,
                            }}>
                              {i + 1}
                            </div>
                            <div className="stage-preview-name">{stage.name}</div>
                            <div className="stage-preview-type">{stage.steps.length} step(s)</div>
                          </div>
                        </Space>
                      ))}
                    </div>
                    <div style={{ marginTop: 16 }}>
                      <div style={{
                        fontSize: 11, fontWeight: 600, color: 'var(--on-surface-variant)',
                        textTransform: 'uppercase', letterSpacing: '0.08em', marginBottom: 8,
                      }}>
                        Config Preview
                      </div>
                      <pre style={{
                        background: 'var(--log-bg)', color: 'var(--log-fg)',
                        padding: 'var(--space-4)', borderRadius: 'var(--radius-md)',
                        fontSize: 11, lineHeight: 1.6, maxHeight: 200, overflow: 'auto',
                        fontFamily: 'var(--font-mono)',
                      }}>
                        {configToJson(selectedTemplate.config)}
                      </pre>
                    </div>
                  </>
                ) : (
                  <div style={{
                    padding: 40, textAlign: 'center', color: 'var(--on-surface-variant)',
                    background: 'var(--surface-container-low)', borderRadius: 'var(--radius-md)',
                  }}>
                    暂无配置预览
                  </div>
                )}
              </div>
            </div>
          </Col>
        </Row>
      )}
    </div>
  );
}
