import { useEffect, useState, useCallback } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { Skeleton, Message, Button, Space, Input, Form, Divider, Tooltip } from '@arco-design/web-react';
import { IconArrowLeft, IconSettings, IconSave, IconCode } from '@arco-design/web-react/icon';
import { PipelineEditor } from '../../../components/pipeline/PipelineEditor';
import { YamlEditor } from '../../../components/pipeline/YamlEditor';
import { ModeSwitch, type EditorMode } from '../../../components/pipeline/ModeSwitch';
import { PipelineSettingsPanel } from '../../../components/pipeline/PipelineSettingsPanel';
import { fetchPipeline, updatePipeline, createPipeline, type Pipeline, type PipelineConfig } from '../../../services/pipeline';

export default function PipelineEditorPage() {
  const { id: projectId, pipelineId } = useParams<{ id: string; pipelineId: string }>();
  const navigate = useNavigate();
  const [pipeline, setPipeline] = useState<Pipeline | null>(null);
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [editorMode, setEditorMode] = useState<EditorMode>('visual');
  const [config, setConfig] = useState<PipelineConfig | null>(null);
  const [jsonValid, setJsonValid] = useState(true);
  const [settingsVisible, setSettingsVisible] = useState(false);
  const [dirty, setDirty] = useState(false);
  const isNew = !pipelineId;

  const [headerForm] = Form.useForm();

  useEffect(() => {
    if (!projectId || !pipelineId || isNew) return;
    setLoading(true);
    fetchPipeline(projectId, pipelineId)
      .then((p) => {
        setPipeline(p);
        setName(p.name);
        setDescription(p.description);
        setConfig(p.config);
        headerForm.setFieldsValue({ name: p.name });
        setDirty(false);
      })
      .catch(() => Message.error('加载流水线失败'))
      .finally(() => setLoading(false));
  }, [projectId, pipelineId, isNew, headerForm]);

  useEffect(() => {
    if (!dirty) return;
    const handler = (e: BeforeUnloadEvent) => { e.preventDefault(); };
    window.addEventListener('beforeunload', handler);
    return () => window.removeEventListener('beforeunload', handler);
  }, [dirty]);

  const markDirty = useCallback(() => { if (!dirty) setDirty(true); }, [dirty]);

  const handleNameChange = useCallback((v: string) => {
    setName(v);
    markDirty();
  }, [markDirty]);

  const handleSave = useCallback(async (cfg: PipelineConfig) => {
    if (!projectId) return;
    const currentName = name.trim();
    if (!currentName) { Message.error('请输入流水线名称'); return; }
    setSaving(true);
    try {
      if (isNew) {
        const created = await createPipeline(projectId, { name: currentName, description, config: cfg });
        Message.success('创建成功');
        setDirty(false);
        navigate(`/projects/${projectId}/pipelines/${created.id}`, { replace: true });
      } else if (pipelineId) {
        await updatePipeline(projectId, pipelineId, { name: currentName, description, config: cfg });
        Message.success('保存成功');
        setConfig(cfg);
        setDirty(false);
      }
    } catch { Message.error('保存失败'); }
    finally { setSaving(false); }
  }, [projectId, pipelineId, isNew, name, description, navigate]);

  const handleSaveSettings = useCallback(
    async (data: { triggerType: string; concurrencyPolicy: string; description: string }) => {
      if (!projectId || !pipelineId || isNew) return;
      setSaving(true);
      try {
        await updatePipeline(projectId, pipelineId, { triggerType: data.triggerType, concurrencyPolicy: data.concurrencyPolicy, description: data.description });
        setDescription(data.description);
        if (pipeline) setPipeline({ ...pipeline, ...data });
        Message.success('设置已保存');
      } catch { Message.error('保存设置失败'); }
      finally { setSaving(false); }
    }, [projectId, pipelineId, isNew, pipeline]
  );

  const defaultConfig: PipelineConfig = { schemaVersion: '1.0', stages: [] };
  const effectiveConfig = config ?? pipeline?.config ?? defaultConfig;

  const handleConfigChange = (cfg: PipelineConfig) => { setConfig(cfg); setJsonValid(true); markDirty(); };
  const handleJsonValidationError = () => { setJsonValid(false); };
  const handleSaveFromJson = () => {
    if (!jsonValid) { Message.warning('请先修正 JSON 格式错误'); return; }
    handleSave(effectiveConfig);
  };
  const handleBack = () => {
    if (dirty && !window.confirm('有未保存的更改，确定离开吗？')) return;
    navigate(`/projects/${projectId}/pipelines`);
  };

  if (loading) {
    return (
      <div style={{ position: 'fixed', inset: 0, zIndex: 1000, background: '#fff', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
        <Skeleton text={{ rows: 6 }} animation style={{ width: 600 }} />
      </div>
    );
  }

  return (
    <div style={{ position: 'fixed', inset: 0, zIndex: 1000, display: 'flex', flexDirection: 'column', background: 'var(--surface)' }}>
      <div style={{
        height: 56, flexShrink: 0,
        display: 'flex', alignItems: 'center', justifyContent: 'space-between',
        padding: '0 20px',
        background: 'var(--glass-fill)',
        backdropFilter: 'var(--glass-blur)',
        WebkitBackdropFilter: 'var(--glass-blur)',
      }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: 12, flex: 1 }}>
          <Tooltip content="返回流水线列表">
            <Button type="text" icon={<IconArrowLeft />} onClick={handleBack} style={{ color: 'var(--on-surface-variant)' }} />
          </Tooltip>
          <Divider type="vertical" style={{ height: 20, borderColor: 'var(--ghost-border-strong)' }} />
          <div style={{
            width: 32, height: 32, borderRadius: 'var(--radius-md)',
            background: 'var(--primary-gradient)',
            display: 'flex', alignItems: 'center', justifyContent: 'center',
            boxShadow: '0 4px 12px rgba(0, 87, 194, 0.3)',
          }}>
            <IconCode style={{ color: '#fff', fontSize: 15 }} />
          </div>
          <Form form={headerForm} style={{ marginBottom: 0, flex: 1, maxWidth: 320 }}>
            <Form.Item field="name" rules={[{ required: true, message: '请输入名称' }]} style={{ marginBottom: 0 }}>
              <Input
                value={name}
                onChange={handleNameChange}
                placeholder="流水线名称"
                style={{
                  border: 'none', background: 'transparent',
                  fontFamily: 'var(--font-display)',
                  fontSize: 17, fontWeight: 700,
                  letterSpacing: '-0.015em',
                  padding: '4px 8px',
                  color: 'var(--on-surface)',
                }}
              />
            </Form.Item>
          </Form>
          {description && (
            <span style={{ fontSize: 12, color: 'var(--on-surface-variant)', maxWidth: 200, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
              {description}
            </span>
          )}
        </div>

        <div>
          <ModeSwitch mode={editorMode} onChange={setEditorMode} />
        </div>

        <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
          {!isNew && pipeline && (
            <Button type="text" icon={<IconSettings />} onClick={() => setSettingsVisible(true)}>
              设置
            </Button>
          )}
          {editorMode === 'json' && (
            <Button type="primary" icon={<IconSave />} onClick={handleSaveFromJson} loading={saving} disabled={!jsonValid}>
              保存
            </Button>
          )}
        </div>
      </div>

      <div style={{ flex: 1, position: 'relative', overflow: 'hidden', background: 'var(--surface)' }}>
        {editorMode === 'visual' ? (
          <PipelineEditor config={effectiveConfig} onSave={handleSave} onChange={handleConfigChange} saving={saving} />
        ) : (
          <div style={{ padding: 20, height: '100%', display: 'flex', flexDirection: 'column', background: 'var(--surface)' }}>
            <div style={{ flex: 1, minHeight: 0 }}>
              <YamlEditor config={effectiveConfig} onChange={handleConfigChange} onValidationError={handleJsonValidationError} />
            </div>
          </div>
        )}
      </div>

      <PipelineSettingsPanel visible={settingsVisible} pipeline={pipeline} onClose={() => setSettingsVisible(false)} onSave={handleSaveSettings} saving={saving} />
    </div>
  );
}
