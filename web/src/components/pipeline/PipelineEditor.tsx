import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import {
  ReactFlow,
  Controls,
  Background,
  BackgroundVariant,
  useNodesState,
  useEdgesState,
  type Node,
  type Edge,
} from '@xyflow/react';
import '@xyflow/react/dist/style.css';
import { Button, Space, Tooltip, Dropdown, Menu } from '@arco-design/web-react';
import { IconPlus, IconSave, IconLeft, IconRight } from '@arco-design/web-react/icon';
import { StageNode, type StageNodeData } from './StageNode';
import { StepNode, type StepNodeData } from './StepNode';
import { StepConfigPanel } from './StepConfigPanel';
import type { PipelineConfig, StageConfig, StepConfig } from '../../services/pipeline';

interface PipelineEditorProps {
  config: PipelineConfig;
  onSave: (config: PipelineConfig) => void | Promise<void>;
  onChange?: (config: PipelineConfig) => void;
  saving?: boolean;
}

const MAX_HISTORY_SIZE = 50;
const STAGE_WIDTH = 240;
const STEP_HEIGHT = 64;
const STAGE_HEADER_H = 52;
const STAGE_GAP_X = 80;
const STEP_GAP_Y = 12;
const CANVAS_PADDING_X = 60;
const CANVAS_PADDING_Y = 40;

const nodeTypes = { stage: StageNode, step: StepNode };

const STEP_TEMPLATES: { type: string; name: string; icon: string; image?: string }[] = [
  { type: 'git-clone', name: 'Git Clone', icon: '📥' },
  { type: 'shell', name: 'Shell 脚本', icon: '💻' },
  { type: 'kaniko', name: 'Kaniko 构建', icon: '🐳' },
  { type: 'buildkit', name: 'BuildKit 构建', icon: '🔨' },
];

function layoutNodes(stages: StageConfig[], callbacks: {
  onAddStep: (stageId: string) => void;
  onDeleteStage: (stageId: string) => void;
  onSelectStep: (stageId: string, stepId: string) => void;
  onDeleteStep: (stageId: string, stepId: string) => void;
  onMoveStage: (stageId: string, direction: 'up' | 'down') => void;
  onMoveStep: (stageId: string, stepId: string, direction: 'up' | 'down') => void;
  onRenameStage: (stageId: string, newName: string) => void;
}): { nodes: Node[]; edges: Edge[] } {
  const nodes: Node[] = [];
  const edges: Edge[] = [];

  stages.forEach((stage, stageIndex) => {
    const stageX = CANVAS_PADDING_X + stageIndex * (STAGE_WIDTH + STAGE_GAP_X);
    const stageY = CANVAS_PADDING_Y;
    const stageNodeId = `stage-${stage.id}`;

    nodes.push({
      id: stageNodeId,
      type: 'stage',
      position: { x: stageX, y: stageY },
      data: {
        label: stage.name,
        stageId: stage.id,
        stageIndex,
        totalStages: stages.length,
        onAddStep: callbacks.onAddStep,
        onDelete: callbacks.onDeleteStage,
        onMove: callbacks.onMoveStage,
        onRename: callbacks.onRenameStage,
      },
    });

    stage.steps.forEach((step, stepIndex) => {
      const stepNodeId = `step-${stage.id}-${step.id}`;
      const stepY = stageY + STAGE_HEADER_H + 16 + stepIndex * (STEP_HEIGHT + STEP_GAP_Y);

      nodes.push({
        id: stepNodeId,
        type: 'step',
        position: { x: stageX + 10, y: stepY },
        data: {
          label: step.name,
          stepId: step.id,
          stageId: stage.id,
          stepIndex,
          totalSteps: stage.steps.length,
          type: step.type,
          image: step.image,
          onSelect: callbacks.onSelectStep,
          onDelete: callbacks.onDeleteStep,
          onMove: callbacks.onMoveStep,
        },
      });

      if (stepIndex > 0) {
        const prevStep = stage.steps[stepIndex - 1];
        edges.push({
          id: `e-step-${stage.id}-${prevStep.id}-${step.id}`,
          source: `step-${stage.id}-${prevStep.id}`,
          target: stepNodeId,
          type: 'smoothstep',
          style: { stroke: '#C9CDD4', strokeWidth: 1 },
        });
      }
    });

    if (stageIndex > 0) {
      const prevStage = stages[stageIndex - 1];
      edges.push({
        id: `e-stage-${prevStage.id}-${stage.id}`,
        source: `stage-${prevStage.id}`,
        target: stageNodeId,
        type: 'smoothstep',
        style: { stroke: '#165DFF', strokeWidth: 2 },
        animated: true,
      });
    }
  });

  return { nodes, edges };
}

export function PipelineEditor({ config, onSave, onChange, saving }: PipelineEditorProps) {
  const [stages, setStages] = useState<StageConfig[]>(config.stages || []);
  const [selectedStep, setSelectedStep] = useState<{ stageId: string; stepId: string } | null>(null);
  const [panelVisible, setPanelVisible] = useState(false);
  const isFirstRender = useRef(true);
  const idCounterRef = useRef(0);

  const [history, setHistory] = useState<StageConfig[][]>([]);
  const [historyIndex, setHistoryIndex] = useState(-1);
  const canUndo = historyIndex > 0;
  const canRedo = historyIndex < history.length - 1;

  const pushHistory = useCallback((newStages: StageConfig[]) => {
    setHistory(prev => {
      const newHistory = prev.slice(0, historyIndex + 1);
      newHistory.push(newStages);
      if (newHistory.length > MAX_HISTORY_SIZE) newHistory.shift();
      return newHistory;
    });
    setHistoryIndex(prev => Math.min(prev + 1, MAX_HISTORY_SIZE - 1));
  }, [historyIndex]);

  const undo = useCallback(() => {
    if (historyIndex > 0) { setHistoryIndex(prev => prev - 1); setStages(history[historyIndex - 1]); }
  }, [history, historyIndex]);

  const redo = useCallback(() => {
    if (historyIndex < history.length - 1) { setHistoryIndex(prev => prev + 1); setStages(history[historyIndex + 1]); }
  }, [history, historyIndex]);

  useEffect(() => {
    if (isFirstRender.current) { isFirstRender.current = false; setHistory([config.stages || []]); setHistoryIndex(0); }
  }, [config.stages]);

  function genId() { idCounterRef.current += 1; return `new-${Date.now()}-${idCounterRef.current}`; }

  const configRef = useRef(config);
  configRef.current = config;
  const onChangeRef = useRef(onChange);
  onChangeRef.current = onChange;

  const configJsonRef = useRef(JSON.stringify(config.stages));
  useEffect(() => {
    const newConfigJson = JSON.stringify(config.stages);
    if (newConfigJson !== configJsonRef.current) {
      configJsonRef.current = newConfigJson;
      setStages(config.stages || []);
      setHistory([config.stages || []]);
      setHistoryIndex(0);
    }
  }, [config.stages]);

  const prevStagesJsonRef = useRef(JSON.stringify(config.stages));
  useEffect(() => {
    const nextJson = JSON.stringify(stages);
    if (nextJson !== prevStagesJsonRef.current) {
      prevStagesJsonRef.current = nextJson;
      onChangeRef.current?.({ ...configRef.current, stages });
    }
  }, [stages]);

  const handleAddStage = useCallback(() => {
    setStages((prev) => {
      const newStage: StageConfig = { id: genId(), name: `Stage ${prev.length + 1}`, steps: [] };
      const newStages = [...prev, newStage];
      pushHistory(newStages);
      return newStages;
    });
  }, [pushHistory]);

  const handleDeleteStage = useCallback((stageId: string) => {
    setStages((prev) => { const ns = prev.filter((s) => s.id !== stageId); pushHistory(ns); return ns; });
  }, [pushHistory]);

  const handleAddStepWithType = useCallback((stageId: string, stepType: string, stepName: string) => {
    setStages((prev) => {
      const newStages = prev.map((stage) => {
        if (stage.id !== stageId) return stage;
        const newStep: StepConfig = { id: genId(), name: stepName, type: stepType };
        return { ...stage, steps: [...stage.steps, newStep] };
      });
      pushHistory(newStages);
      return newStages;
    });
  }, [pushHistory]);

  const handleAddStep = useCallback((stageId: string) => {
    handleAddStepWithType(stageId, 'shell', `Step ${(stages.find(s => s.id === stageId)?.steps.length ?? 0) + 1}`);
  }, [handleAddStepWithType, stages]);

  const handleDeleteStep = useCallback((stageId: string, stepId: string) => {
    setStages((prev) => {
      const ns = prev.map((stage) => {
        if (stage.id !== stageId) return stage;
        return { ...stage, steps: stage.steps.filter((s) => s.id !== stepId) };
      });
      pushHistory(ns);
      return ns;
    });
  }, [pushHistory]);

  const handleMoveStage = useCallback((stageId: string, direction: 'up' | 'down') => {
    setStages((prev) => {
      const index = prev.findIndex((s) => s.id === stageId);
      if (index < 0) return prev;
      const newIndex = direction === 'up' ? index - 1 : index + 1;
      if (newIndex < 0 || newIndex >= prev.length) return prev;
      const ns = [...prev];
      [ns[index], ns[newIndex]] = [ns[newIndex], ns[index]];
      pushHistory(ns);
      return ns;
    });
  }, [pushHistory]);

  const handleMoveStep = useCallback((stageId: string, stepId: string, direction: 'up' | 'down') => {
    setStages((prev) => {
      const ns = prev.map((stage) => {
        if (stage.id !== stageId) return stage;
        const index = stage.steps.findIndex((s) => s.id === stepId);
        if (index < 0) return stage;
        const newIndex = direction === 'up' ? index - 1 : index + 1;
        if (newIndex < 0 || newIndex >= stage.steps.length) return stage;
        const newSteps = [...stage.steps];
        [newSteps[index], newSteps[newIndex]] = [newSteps[newIndex], newSteps[index]];
        return { ...stage, steps: newSteps };
      });
      pushHistory(ns);
      return ns;
    });
  }, [pushHistory]);

  const handleRenameStage = useCallback((stageId: string, newName: string) => {
    setStages((prev) => {
      const ns = prev.map((s) => s.id === stageId ? { ...s, name: newName } : s);
      pushHistory(ns);
      return ns;
    });
  }, [pushHistory]);

  const handleSelectStep = useCallback((stageId: string, stepId: string) => {
    setSelectedStep({ stageId, stepId });
    setPanelVisible(true);
  }, []);

  const handleSaveStep = useCallback((updatedStep: StepConfig) => {
    if (!selectedStep) return;
    setStages((prev) => {
      const ns = prev.map((stage) => {
        if (stage.id !== selectedStep.stageId) return stage;
        return { ...stage, steps: stage.steps.map((s) => (s.id === updatedStep.id ? updatedStep : s)) };
      });
      pushHistory(ns);
      return ns;
    });
  }, [selectedStep, pushHistory]);

  const handleSave = useCallback(() => {
    onSave({ ...configRef.current, stages });
  }, [stages, onSave]);

  const currentStep = useMemo(() => {
    if (!selectedStep) return null;
    const stage = stages.find((s) => s.id === selectedStep.stageId);
    return stage?.steps.find((s) => s.id === selectedStep.stepId) || null;
  }, [selectedStep, stages]);

  const { nodes: layoutedNodes, edges: layoutedEdges } = useMemo(
    () => layoutNodes(stages, {
      onAddStep: handleAddStep,
      onDeleteStage: handleDeleteStage,
      onSelectStep: handleSelectStep,
      onDeleteStep: handleDeleteStep,
      onMoveStage: handleMoveStage,
      onMoveStep: handleMoveStep,
      onRenameStage: handleRenameStage,
    }),
    [stages, handleAddStep, handleDeleteStage, handleSelectStep, handleDeleteStep, handleMoveStage, handleMoveStep, handleRenameStage]
  );

  const [nodes, setNodes, onNodesChange] = useNodesState(layoutedNodes);
  const [edges, setEdges, onEdgesChange] = useEdgesState(layoutedEdges);

  useEffect(() => { setNodes(layoutedNodes); setEdges(layoutedEdges); }, [layoutedNodes, layoutedEdges, setNodes, setEdges]);

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if ((e.ctrlKey || e.metaKey) && e.key === 's') { e.preventDefault(); handleSave(); return; }
      if ((e.ctrlKey || e.metaKey) && e.key === 'z' && !e.shiftKey) { e.preventDefault(); undo(); return; }
      if ((e.ctrlKey || e.metaKey) && (e.key === 'y' || (e.key === 'z' && e.shiftKey))) { e.preventDefault(); redo(); return; }
      if ((e.key === 'Delete' || e.key === 'Backspace') && selectedStep) { e.preventDefault(); handleDeleteStep(selectedStep.stageId, selectedStep.stepId); setPanelVisible(false); return; }
      if (e.key === 'Escape' && panelVisible) { setPanelVisible(false); return; }
    };
    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [handleSave, undo, redo, selectedStep, handleDeleteStep, panelVisible]);

  const validateConfig = useCallback((cfg: PipelineConfig): string | null => {
    if (!cfg.schemaVersion) return '缺少 schemaVersion';
    if (!Array.isArray(cfg.stages) || cfg.stages.length === 0) return '请至少添加一个 Stage';
    for (let i = 0; i < cfg.stages.length; i++) {
      const stage = cfg.stages[i];
      if (!stage.id || !stage.name) return `Stage ${i + 1}: 缺少 id 或 name`;
      if (!Array.isArray(stage.steps)) return `Stage "${stage.name}": steps 必须是数组`;
      for (let j = 0; j < stage.steps.length; j++) {
        const step = stage.steps[j];
        if (!step.id || !step.name) return `Stage "${stage.name}" Step ${j + 1}: 缺少 id 或 name`;
        if (!step.type) return `Step "${step.name}": 缺少 type`;
      }
    }
    return null;
  }, []);

  const validationError = useMemo(() => validateConfig({ ...configRef.current, stages }), [stages, validateConfig]);

  const addStageDroplist = (
    <Menu onClickMenuItem={(key) => {
      const newStageId = genId();
      setStages((prev) => {
        const newStage: StageConfig = { id: newStageId, name: `Stage ${prev.length + 1}`, steps: [] };
        const newStages = [...prev, newStage];
        pushHistory(newStages);
        if (key !== 'empty') {
          const tpl = STEP_TEMPLATES.find(t => t.type === key);
          if (tpl) {
            const stepStages = newStages.map(s => {
              if (s.id !== newStageId) return s;
              return { ...s, steps: [{ id: genId(), name: tpl.name, type: tpl.type } as StepConfig] };
            });
            pushHistory(stepStages);
            return stepStages;
          }
        }
        return newStages;
      });
    }}>
      <Menu.Item key="empty"><span style={{ marginRight: 8 }}>📋</span>空白 Stage</Menu.Item>
      <Menu.ItemGroup title="快速创建 Stage + Step">
        {STEP_TEMPLATES.map(t => (
          <Menu.Item key={t.type}><span style={{ marginRight: 8 }}>{t.icon}</span>{t.name}</Menu.Item>
        ))}
      </Menu.ItemGroup>
    </Menu>
  );

  return (
    <div style={{ width: '100%', height: '100%', position: 'relative' }}>
      {/* Floating toolbar */}
      <div style={{
        position: 'absolute', top: 12, left: 12, zIndex: 10,
        display: 'flex', alignItems: 'center', gap: 6,
        background: 'rgba(255,255,255,0.95)', padding: '6px 12px',
        borderRadius: 10, boxShadow: '0 2px 12px rgba(0,0,0,0.08)',
        backdropFilter: 'blur(8px)',
      }}>
        <Tooltip content="撤销 (Ctrl+Z)">
          <Button size="small" icon={<IconLeft />} onClick={undo} disabled={!canUndo} style={{ borderRadius: 6 }} />
        </Tooltip>
        <Tooltip content="重做 (Ctrl+Y)">
          <Button size="small" icon={<IconRight />} onClick={redo} disabled={!canRedo} style={{ borderRadius: 6 }} />
        </Tooltip>
        <div style={{ width: 1, height: 20, background: '#E5E6EB', margin: '0 4px' }} />
        <Dropdown droplist={addStageDroplist} position="bl">
          <Button size="small" type="primary" icon={<IconPlus />} style={{ borderRadius: 6 }}>
            添加 Stage
          </Button>
        </Dropdown>
        <div style={{ width: 1, height: 20, background: '#E5E6EB', margin: '0 4px' }} />
        <Button
          size="small"
          type="primary"
          status={validationError ? 'warning' : 'success'}
          icon={<IconSave />}
          onClick={handleSave}
          loading={saving}
          disabled={!!validationError}
          style={{ borderRadius: 6 }}
        >
          保存
        </Button>
        {validationError && (
          <Tooltip content={validationError}>
            <span style={{ color: '#FF7D00', fontSize: 11, padding: '2px 8px', background: '#FFF7E8', borderRadius: 10 }}>
              {validationError}
            </span>
          </Tooltip>
        )}
      </div>

      {/* Empty state hint */}
      {stages.length === 0 && (
        <div style={{
          position: 'absolute', top: '50%', left: '50%', transform: 'translate(-50%, -50%)',
          zIndex: 5, textAlign: 'center', color: '#86909C',
        }}>
          <div style={{ fontSize: 48, marginBottom: 12 }}>⚡</div>
          <div style={{ fontSize: 16, fontWeight: 500, marginBottom: 4 }}>开始构建你的流水线</div>
          <div style={{ fontSize: 13, marginBottom: 16 }}>点击上方「添加 Stage」开始，流程从左到右执行</div>
        </div>
      )}

      <ReactFlow
        nodes={nodes}
        edges={edges}
        onNodesChange={onNodesChange}
        onEdgesChange={onEdgesChange}
        nodeTypes={nodeTypes}
        fitView
        fitViewOptions={{ padding: 0.2 }}
        minZoom={0.3}
        maxZoom={2}
        attributionPosition="bottom-left"
        style={{ background: '#F7F8FA' }}
        nodesDraggable={false}
      >
        <Controls position="bottom-left" style={{ borderRadius: 8 }} />
        <Background variant={BackgroundVariant.Dots} gap={20} size={1} color="#E5E6EB" />
      </ReactFlow>

      <StepConfigPanel visible={panelVisible} step={currentStep} onClose={() => setPanelVisible(false)} onSave={handleSaveStep} />
    </div>
  );
}
