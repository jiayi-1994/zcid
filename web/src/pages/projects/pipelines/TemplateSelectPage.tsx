import { useCallback, useEffect, useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { Message } from '@arco-design/web-react';
import { fetchTemplates, createPipeline, type PipelineTemplate } from '../../../services/pipeline';
import { configToJson } from '../../../lib/pipeline/configJson';
import { PageHeader } from '../../../components/ui/PageHeader';
import { Card } from '../../../components/ui/Card';
import { Btn } from '../../../components/ui/Btn';
import { Field } from '../../../components/ui/Field';
import { Badge } from '../../../components/ui/Badge';
import { IArrL, ICheck } from '../../../components/ui/icons';

const LANG_ICONS: Record<string, string> = {
  'Go': '🔵', 'Java Maven': '☕', 'Java JAR': '☕', 'Node.js': '🟢', 'Docker': '🐳',
};

const CATEGORY_ICONS: Record<string, string> = {
  backend: '⚙️', frontend: '📦', general: '🐳',
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
    template.params.forEach((p) => { if (p.defaultValue) params[p.name] = p.defaultValue; });
    setTemplateParams(params);
  }, []);

  const handleCreate = useCallback(async () => {
    if (!projectId || !selectedTemplate) return;
    if (!pipelineName.trim()) { Message.error('请输入流水线名称'); return; }
    const missing = selectedTemplate.params.filter((p) => p.required && !templateParams[p.name]).map((p) => p.name);
    if (missing.length > 0) { Message.error(`请填写必填参数: ${missing.join(', ')}`); return; }
    setCreating(true);
    try {
      const created = await createPipeline(projectId, { name: pipelineName.trim(), description: pipelineDesc.trim(), templateId: selectedTemplate.id, templateParams });
      Message.success('创建成功');
      navigate(`/projects/${projectId}/pipelines/${created.id}`);
    } catch {
      Message.error('创建失败');
    } finally {
      setCreating(false);
    }
  }, [projectId, selectedTemplate, pipelineName, pipelineDesc, templateParams, navigate]);

  const currentStep = selectedTemplate ? 2 : 1;

  return (
    <>
      <PageHeader
        crumb="Create Pipeline"
        title="Architect New Pipeline"
        sub="从模板快速创建流水线，或从空白开始构建。"
        actions={
          <Btn size="sm" icon={<IArrL size={13} />} onClick={() => selectedTemplate ? setSelectedTemplate(null) : navigate(-1)}>
            {selectedTemplate ? '返回模板列表' : '返回'}
          </Btn>
        }
      />
      <div style={{ padding: '24px 24px 48px', maxWidth: 960 }}>
        {/* Wizard steps */}
        <div style={{ display: 'flex', gap: 0, marginBottom: 28 }}>
          {[
            { n: 1, label: 'Template Selection', desc: '选择语言和构建模板' },
            { n: 2, label: 'Configuration', desc: '配置参数和仓库信息' },
          ].map((step, i) => {
            const done = currentStep > step.n;
            const active = currentStep === step.n;
            return (
              <div key={step.n} style={{ display: 'flex', alignItems: 'center', flex: 1 }}>
                {i > 0 && <div style={{ height: 2, flex: 'none', width: 32, background: done ? 'var(--accent-1)' : 'var(--z-200)', marginRight: 12 }} />}
                <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
                  <div style={{
                    width: 28, height: 28, borderRadius: '50%',
                    background: done || active ? 'var(--accent-1)' : 'var(--z-100)',
                    color: done || active ? '#fff' : 'var(--z-500)',
                    display: 'flex', alignItems: 'center', justifyContent: 'center',
                    fontSize: 12, fontWeight: 600, flex: 'none',
                  }}>
                    {done ? '✓' : step.n}
                  </div>
                  <div>
                    <div style={{ fontSize: 12.5, fontWeight: active ? 600 : 400, color: active ? 'var(--z-900)' : 'var(--z-500)' }}>{step.label}</div>
                    <div style={{ fontSize: 11, color: 'var(--z-400)' }}>{step.desc}</div>
                  </div>
                </div>
              </div>
            );
          })}
        </div>

        {/* Step 1: Template grid */}
        {!selectedTemplate && (
          loading ? (
            <div style={{ padding: '40px 0', textAlign: 'center', color: 'var(--z-400)' }}>加载中...</div>
          ) : (
            <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(160px, 1fr))', gap: 14 }}>
              <div
                className="card"
                style={{ padding: 20, cursor: 'pointer', textAlign: 'center', minHeight: 160, display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center', gap: 8 }}
                onClick={() => projectId && navigate(`/projects/${projectId}/pipelines/blank`)}
              >
                <div style={{ fontSize: 28 }}>＋</div>
                <div style={{ fontSize: 13, fontWeight: 600 }}>Custom</div>
                <div style={{ fontSize: 11.5, color: 'var(--z-500)' }}>从零开始配置</div>
              </div>
              {templates.map((template) => {
                const icon = LANG_ICONS[template.name] || CATEGORY_ICONS[template.category] || '📦';
                return (
                  <div
                    key={template.id}
                    className="card"
                    style={{ padding: 20, cursor: 'pointer', textAlign: 'center', minHeight: 160, display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center', gap: 8 }}
                    onClick={() => handleSelectTemplate(template)}
                  >
                    <div style={{ fontSize: 28 }}>{icon}</div>
                    <div style={{ fontSize: 13, fontWeight: 600 }}>{template.name}</div>
                    <div style={{ fontSize: 11.5, color: 'var(--z-500)' }}>{template.description}</div>
                  </div>
                );
              })}
            </div>
          )
        )}

        {/* Step 2: Config + preview */}
        {selectedTemplate && (
          <div style={{ display: 'grid', gridTemplateColumns: '1fr 380px', gap: 18 }}>
            <Card title="Configuration Parameters">
              <div style={{ display: 'flex', flexDirection: 'column', gap: 14 }}>
                <Field label="流水线名称" required>
                  <input className="input" value={pipelineName} onChange={(e) => setPipelineName(e.target.value)} placeholder="my-app-pipeline" />
                </Field>
                <Field label="描述">
                  <input className="input" value={pipelineDesc} onChange={(e) => setPipelineDesc(e.target.value)} placeholder="可选描述" />
                </Field>

                {selectedTemplate.params.length > 0 && (
                  <>
                    <div style={{ fontSize: 11, fontWeight: 600, color: 'var(--z-400)', textTransform: 'uppercase', letterSpacing: '0.06em', marginTop: 4 }}>
                      Template Parameters
                    </div>
                    <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 12 }}>
                      {selectedTemplate.params.map((param) => (
                        <div key={param.name} style={{ gridColumn: param.name === 'repoUrl' ? '1 / -1' : undefined }}>
                          <Field label={<>{param.name}{param.required && <span style={{ color: 'var(--red-ink)', marginLeft: 2 }}>*</span>}</>}>
                            <input
                              className="input"
                              value={templateParams[param.name] || ''}
                              onChange={(e) => setTemplateParams((prev) => ({ ...prev, [param.name]: e.target.value }))}
                              placeholder={param.defaultValue || `请输入 ${param.name}`}
                            />
                            {param.description && <div style={{ fontSize: 11, color: 'var(--z-400)', marginTop: 3 }}>{param.description}</div>}
                          </Field>
                        </div>
                      ))}
                    </div>
                  </>
                )}

                <div style={{ display: 'flex', gap: 8, marginTop: 8 }}>
                  <Btn variant="primary" icon={<ICheck size={13} />} onClick={handleCreate} disabled={creating}>
                    {creating ? '创建中...' : '创建流水线'}
                  </Btn>
                  <Btn onClick={() => setSelectedTemplate(null)}>返回</Btn>
                </div>
              </div>
            </Card>

            <Card title="Stage Preview">
              {selectedTemplate.config ? (
                <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
                  <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
                    {selectedTemplate.config.stages.map((stage, i) => (
                      <div key={stage.id} style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
                        <div style={{ width: 24, height: 24, borderRadius: 6, background: 'var(--blue-soft)', color: 'var(--blue-ink)', display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: 11, fontWeight: 700, flex: 'none' }}>{i + 1}</div>
                        <div style={{ flex: 1 }}>
                          <div style={{ fontSize: 12.5, fontWeight: 500 }}>{stage.name}</div>
                          <div style={{ fontSize: 11, color: 'var(--z-500)' }}>{stage.steps.length} step(s)</div>
                        </div>
                        <Badge tone="blue">{stage.steps.length}</Badge>
                      </div>
                    ))}
                  </div>
                  <div>
                    <div style={{ fontSize: 11, fontWeight: 600, color: 'var(--z-400)', textTransform: 'uppercase', letterSpacing: '0.06em', marginBottom: 6 }}>Config Preview</div>
                    <pre style={{ background: '#0d1117', color: '#e6edf3', padding: 12, borderRadius: 6, fontSize: 11, lineHeight: 1.6, maxHeight: 200, overflow: 'auto', fontFamily: 'var(--font-mono)', margin: 0 }}>
                      {configToJson(selectedTemplate.config)}
                    </pre>
                  </div>
                </div>
              ) : (
                <div style={{ padding: '32px 0', textAlign: 'center', color: 'var(--z-400)' }}>暂无配置预览</div>
              )}
            </Card>
          </div>
        )}
      </div>
    </>
  );
}
