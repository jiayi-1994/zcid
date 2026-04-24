import { useCallback, useEffect, useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { Message } from '@arco-design/web-react';
import { PipelineEditor } from '../../../components/pipeline/PipelineEditor';
import { YamlEditor } from '../../../components/pipeline/YamlEditor';
import { ModeSwitch, type EditorMode } from '../../../components/pipeline/ModeSwitch';
import { PipelineSettingsPanel } from '../../../components/pipeline/PipelineSettingsPanel';
import { fetchPipeline, updatePipeline, createPipeline, type Pipeline, type PipelineConfig } from '../../../services/pipeline';
import { Btn } from '../../../components/ui/Btn';
import { IArrL, ISettings, ICheck, ICode } from '../../../components/ui/icons';

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

  useEffect(() => {
    if (!projectId || !pipelineId || isNew) return;
    setLoading(true);
    fetchPipeline(projectId, pipelineId)
      .then((p) => {
        setPipeline(p);
        setName(p.name);
        setDescription(p.description);
        setConfig(p.config);
        setDirty(false);
      })
      .catch(() => Message.error('加载流水线失败'))
      .finally(() => setLoading(false));
  }, [projectId, pipelineId, isNew]);

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

  const stageCount = effectiveConfig.stages.length;
  const stepCount = effectiveConfig.stages.reduce((a, s) => a + s.steps.length, 0);

  if (loading) {
    return (
      <div className="zc" style={{ position: 'fixed', inset: 0, zIndex: 1000, background: 'var(--z-0)', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
        <div style={{ color: 'var(--z-500)', fontFamily: 'var(--font-sans)' }}>加载中...</div>
      </div>
    );
  }

  return (
    <div className="zc" style={{ position: 'fixed', inset: 0, zIndex: 1000, display: 'flex', flexDirection: 'column', background: 'var(--z-0)', fontFamily: 'var(--font-sans)' }}>
      {/* Header */}
      <div style={{
        height: 56, flex: 'none',
        display: 'flex', alignItems: 'center', gap: 12,
        padding: '0 16px',
        background: 'rgba(255,255,255,0.75)',
        backdropFilter: 'blur(10px)',
        borderBottom: '1px solid var(--z-150)',
        zIndex: 5,
      }}>
        <Btn size="sm" variant="ghost" iconOnly icon={<IArrL size={13} />} onClick={handleBack} title="返回" />
        <div style={{ width: 1, height: 22, background: 'var(--z-200)' }} />
        <div style={{
          width: 32, height: 32, borderRadius: 8,
          background: 'linear-gradient(135deg, var(--accent-1), var(--accent-2))',
          color: '#fff',
          display: 'flex', alignItems: 'center', justifyContent: 'center',
          flex: 'none',
        }}>
          <ICode size={14} />
        </div>
        <div style={{ display: 'flex', flexDirection: 'column', gap: 1, minWidth: 0 }}>
          <input
            value={name}
            onChange={(e) => handleNameChange(e.target.value)}
            placeholder="流水线名称"
            style={{
              border: 0, outline: 'none', background: 'transparent',
              fontSize: 15, fontWeight: 600, letterSpacing: '-0.01em',
              color: 'var(--z-900)', padding: 0, minWidth: 180,
              fontFamily: 'var(--font-sans)',
            }}
          />
          <span className="sub mono" style={{ fontSize: 10.5 }}>
            {pipeline?.triggerType ?? (isNew ? 'manual' : '...')} · {stageCount} stages · {stepCount} steps
            {dirty && <>  · <span style={{ color: 'var(--amber-ink)' }}>● 未保存</span></>}
          </span>
        </div>

        <div style={{ flex: 1 }} />

        <ModeSwitch mode={editorMode} onChange={setEditorMode} />

        <div style={{ width: 1, height: 22, background: 'var(--z-200)' }} />

        {!isNew && pipeline && (
          <Btn size="sm" variant="ghost" icon={<ISettings size={13} />} onClick={() => setSettingsVisible(true)}>
            设置
          </Btn>
        )}
        <Btn size="sm" variant="outline" onClick={handleBack}>取消</Btn>
        {editorMode === 'json' ? (
          <Btn size="sm" variant="primary" icon={<ICheck size={13} />} onClick={handleSaveFromJson} disabled={!jsonValid || saving}>
            {saving ? '保存中...' : '保存'}
          </Btn>
        ) : (
          <Btn size="sm" variant="primary" icon={<ICheck size={13} />} onClick={() => handleSave(effectiveConfig)} disabled={saving}>
            {saving ? '保存中...' : '保存'}
          </Btn>
        )}
      </div>

      {/* Main area */}
      <div style={{ flex: 1, position: 'relative', overflow: 'hidden', background: 'var(--z-25)' }}>
        {editorMode === 'visual' ? (
          <PipelineEditor config={effectiveConfig} onSave={handleSave} onChange={handleConfigChange} saving={saving} />
        ) : (
          <div style={{ padding: 20, height: '100%', display: 'flex', flexDirection: 'column', background: 'var(--z-0)' }}>
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
