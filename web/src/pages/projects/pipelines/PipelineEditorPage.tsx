import { useEffect, useState, useCallback } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { Skeleton, Message, Typography, Button, Space, Input, Form, Modal } from '@arco-design/web-react';
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
  const isNew = pipelineId === 'new';

  useEffect(() => {
    if (!projectId || !pipelineId || isNew) return;
    setLoading(true);
    fetchPipeline(projectId, pipelineId)
      .then((p) => {
        setPipeline(p);
        setName(p.name);
        setDescription(p.description);
        setConfig(p.config);
      })
      .catch(() => Message.error('加载流水线失败'))
      .finally(() => setLoading(false));
  }, [projectId, pipelineId, isNew]);

  const handleSave = useCallback(async (cfg: PipelineConfig) => {
    if (!projectId) return;
    setSaving(true);
    try {
      if (isNew) {
        if (!name.trim()) {
          Message.error('请输入流水线名称');
          setSaving(false);
          return;
        }
        const created = await createPipeline(projectId, { name, description, config: cfg });
        Message.success('创建成功');
        navigate(`/projects/${projectId}/pipelines/${created.id}`, { replace: true });
      } else if (pipelineId) {
        await updatePipeline(projectId, pipelineId, { name, description, config: cfg });
        Message.success('保存成功');
        setConfig(cfg);
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

  const handleConfigChange = useCallback((cfg: PipelineConfig) => {
    setConfig(cfg);
    setJsonValid(true);
  }, []);

  const handleJsonValidationError = useCallback(() => {
    setJsonValid(false);
  }, []);

  const handleSaveFromJson = useCallback(() => {
    if (!jsonValid) {
      Message.warning('请先修正 JSON 格式错误');
      return;
    }
    handleSave(effectiveConfig);
  }, [effectiveConfig, handleSave, jsonValid]);

  return (
    <div style={{ display: 'flex', flexDirection: 'column', height: '100%' }}>
      <div style={{ padding: '12px 24px', borderBottom: '1px solid var(--zcid-color-border)', background: 'var(--zcid-color-bg-primary)' }}>
        <Space align="center">
          <Button
            type="text"
            icon={<IconArrowLeft />}
            onClick={() => navigate(`/projects/${projectId}/pipelines`)}
          />
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
        <div style={{ marginTop: 8, display: 'flex', gap: 12 }}>
          <Form.Item label="名称" style={{ marginBottom: 0, flex: 1 }}>
            <Input value={name} onChange={setName} placeholder="流水线名称" />
          </Form.Item>
          <Form.Item label="描述" style={{ marginBottom: 0, flex: 2 }}>
            <Input value={description} onChange={setDescription} placeholder="描述（可选）" />
          </Form.Item>
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
