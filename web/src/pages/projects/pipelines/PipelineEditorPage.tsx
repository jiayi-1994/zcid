import { useEffect, useState, useCallback, useRef } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { Skeleton, Message, Typography, Button, Space, Input, Form } from '@arco-design/web-react';
import { IconArrowLeft, IconSettings } from '@arco-design/web-react/icon';
import { PipelineEditor } from '../../../components/pipeline/PipelineEditor';
import { YamlEditor } from '../../../components/pipeline/YamlEditor';
import { ModeSwitch, type EditorMode } from '../../../components/pipeline/ModeSwitch';
import { PipelineSettingsPanel } from '../../../components/pipeline/PipelineSettingsPanel';
import { fetchPipeline, updatePipeline, createPipeline, type Pipeline, type PipelineConfig } from '../../../services/pipeline';

const { Title } = Typography;

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
    const handler = (e: BeforeUnloadEvent) => {
      e.preventDefault();
    };
    window.addEventListener('beforeunload', handler);
    return () => window.removeEventListener('beforeunload', handler);
  }, [dirty]);

  const markDirty = useCallback(() => {
    if (!dirty) setDirty(true);
  }, [dirty]);

  const handleNameChange = useCallback((v: string) => {
    setName(v);
    markDirty();
  }, [markDirty]);

  const handleSave = useCallback(async (cfg: PipelineConfig) => {
    if (!projectId) return;
    const currentName = name.trim();
    if (!currentName) {
      Message.error('请输入流水线名称');
      return;
    }
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
    } catch {
      Message.error('保存失败');
    } finally {
      setSaving(false);
    }
  }, [projectId, pipelineId, isNew, name, description, navigate]);

  const handleSaveSettings = useCallback(
    async (data: { triggerType: string; concurrencyPolicy: string; description: string }) => {
      if (!projectId || !pipelineId || isNew) return;
      setSaving(true);
      try {
        await updatePipeline(projectId, pipelineId, {
          triggerType: data.triggerType,
          concurrencyPolicy: data.concurrencyPolicy,
          description: data.description,
        });
        setDescription(data.description);
        if (pipeline) {
          setPipeline({ ...pipeline, ...data });
        }
        Message.success('设置已保存');
      } catch {
        Message.error('保存设置失败');
      } finally {
        setSaving(false);
      }
    },
    [projectId, pipelineId, isNew, pipeline]
  );

  if (loading) {
    return (
      <div className="page-container">
        <Skeleton text={{ rows: 6 }} animation />
      </div>
    );
  }

  const defaultConfig: PipelineConfig = {
    schemaVersion: '1.0',
    stages: [],
  };

  const effectiveConfig = config ?? pipeline?.config ?? defaultConfig;

  const handleConfigChange = (cfg: PipelineConfig) => {
    setConfig(cfg);
    setJsonValid(true);
    markDirty();
  };

  const handleJsonValidationError = () => {
    setJsonValid(false);
  };

  const handleSaveFromJson = () => {
    if (!jsonValid) {
      Message.warning('请先修正 JSON 格式错误');
      return;
    }
    handleSave(effectiveConfig);
  };

  const handleBack = () => {
    if (dirty && !window.confirm('有未保存的更改，确定离开吗？')) return;
    navigate(`/projects/${projectId}/pipelines`);
  };

  return (
    <div style={{ display: 'flex', flexDirection: 'column', height: '100%' }}>
      <div style={{ padding: '12px 24px', borderBottom: '1px solid var(--zcid-color-border)', background: 'var(--zcid-color-bg-primary)' }}>
        <Space align="center">
          <Button type="text" icon={<IconArrowLeft />} onClick={handleBack} />
          <Title heading={6} style={{ margin: 0 }}>
            {isNew ? '创建流水线' : `编辑: ${name}`}
          </Title>
          <ModeSwitch mode={editorMode} onChange={setEditorMode} />
          {!isNew && pipeline && (
            <Button
              type="text"
              icon={<IconSettings />}
              onClick={() => setSettingsVisible(true)}
            >
              设置
            </Button>
          )}
        </Space>
        <div style={{ marginTop: 8, display: 'flex', gap: 12, alignItems: 'center' }}>
          <Form form={headerForm} style={{ flex: 1, marginBottom: 0 }}>
            <Form.Item
              field="name"
              rules={[{ required: true, message: '请输入流水线名称' }]}
              style={{ marginBottom: 0 }}
            >
              <Input value={name} onChange={handleNameChange} placeholder="流水线名称" />
            </Form.Item>
          </Form>
          <div style={{ flex: 2, color: 'var(--color-text-3)', fontSize: 13 }}>
            {description ? description : <span style={{ opacity: 0.5 }}>无描述（可在设置中编辑）</span>}
          </div>
        </div>
      </div>
      <div style={{ flex: 1, position: 'relative', overflow: 'hidden' }}>
        {editorMode === 'visual' ? (
          <PipelineEditor
            config={effectiveConfig}
            onSave={handleSave}
            onChange={handleConfigChange}
            saving={saving}
          />
        ) : (
          <div style={{ padding: 16, height: '100%', display: 'flex', flexDirection: 'column' }}>
            <div style={{ marginBottom: 12 }}>
              <Button type="primary" onClick={handleSaveFromJson} loading={saving} disabled={!jsonValid}>
                保存
              </Button>
            </div>
            <div style={{ flex: 1, minHeight: 0 }}>
              <YamlEditor
                config={effectiveConfig}
                onChange={handleConfigChange}
                onValidationError={handleJsonValidationError}
              />
            </div>
          </div>
        )}
      </div>
      <PipelineSettingsPanel
        visible={settingsVisible}
        pipeline={pipeline}
        onClose={() => setSettingsVisible(false)}
        onSave={handleSaveSettings}
        saving={saving}
      />
    </div>
  );
}
