import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import {
  ReactFlow,
  Controls,
  MiniMap,
  Background,
  BackgroundVariant,
  useNodesState,
  useEdgesState,
  type Node,
  type Edge,
} from '@xyflow/react';
import '@xyflow/react/dist/style.css';
import dagre from 'dagre';
import { Button, Space, Tooltip } from '@arco-design/web-react';
import { IconPlus, IconSave, IconLeft, IconRight, IconUp, IconDown } from '@arco-design/web-react/icon';
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

const nodeTypes = {
  stage: StageNode,
  step: StepNode,
};

function layoutNodes(stages: StageConfig[], callbacks: {
  onAddStep: (stageId: string) => void;
  onDeleteStage: (stageId: string) => void;
  onSelectStep: (stageId: string, stepId: string) => void;
  onDeleteStep: (stageId: string, stepId: string) => void;
  onMoveStage: (stageId: string, direction: 'up' | 'down') => void;
  onMoveStep: (stageId: string, stepId: string, direction: 'up' | 'down') => void;
}): { nodes: Node[]; edges: Edge[] } {
  const g = new dagre.graphlib.Graph();
  g.setDefaultEdgeLabel(() => ({}));
  g.setGraph({ rankdir: 'LR', nodesep: 60, ranksep: 120 });

  const nodes: Node[] = [];
  const edges: Edge[] = [];

  stages.forEach((stage, stageIndex) => {
    const stageNodeId = `stage-${stage.id}`;
    g.setNode(stageNodeId, { width: 220, height: 60 + stage.steps.length * 60 });

    nodes.push({
      id: stageNodeId,
      type: 'stage',
      position: { x: 0, y: 0 },
      data: {
        label: stage.name,
        stageId: stage.id,
        stageIndex: stageIndex,
        totalStages: stages.length,
        onAddStep: callbacks.onAddStep,
        onDelete: callbacks.onDeleteStage,
        onMove: callbacks.onMoveStage,
      },
    });

    stage.steps.forEach((step, stepIndex) => {
      const stepNodeId = `step-${stage.id}-${step.id}`;
      g.setNode(stepNodeId, { width: 200, height: 50 });

      nodes.push({
        id: stepNodeId,
        type: 'step',
        position: { x: 0, y: 0 },
        data: {
          label: step.name,
          stepId: step.id,
          stageId: stage.id,
          stepIndex: stepIndex,
          totalSteps: stage.steps.length,
          type: step.type,
          image: step.image,
          onSelect: callbacks.onSelectStep,
          onDelete: callbacks.onDeleteStep,
          onMove: callbacks.onMoveStep,
        },
      });

      if (stepIndex === 0) {
        edges.push({
          id: `e-${stageNodeId}-${stepNodeId}`,
          source: stageNodeId,
          target: stepNodeId,
          type: 'smoothstep',
          animated: false,
        });
      } else {
        const prevStep = stage.steps[stepIndex - 1];
        edges.push({
          id: `e-step-${stage.id}-${prevStep.id}-${step.id}`,
          source: `step-${stage.id}-${prevStep.id}`,
          target: stepNodeId,
          type: 'smoothstep',
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
        style: { stroke: '#2563EB', strokeWidth: 2 },
      });
    }
  });

  dagre.layout(g);

  nodes.forEach((node) => {
    const pos = g.node(node.id);
    if (pos) {
      node.position = { x: pos.x - (pos.width || 0) / 2, y: pos.y - (pos.height || 0) / 2 };
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

  // Undo/Redo history state
  const [history, setHistory] = useState<StageConfig[][]>([]);
  const [historyIndex, setHistoryIndex] = useState(-1);
  const canUndo = historyIndex > 0;
  const canRedo = historyIndex < history.length - 1;

  // Push current state to history
  const pushHistory = useCallback((newStages: StageConfig[]) => {
    setHistory(prev => {
      const newHistory = prev.slice(0, historyIndex + 1);
      newHistory.push(newStages);
      if (newHistory.length > MAX_HISTORY_SIZE) {
        newHistory.shift();
        return newHistory;
      }
      return newHistory;
    });
    setHistoryIndex(prev => Math.min(prev + 1, MAX_HISTORY_SIZE - 1));
  }, [historyIndex]);

  const undo = useCallback(() => {
    if (historyIndex > 0) {
      setHistoryIndex(prev => prev - 1);
      setStages(history[historyIndex - 1]);
    }
  }, [history, historyIndex]);

  const redo = useCallback(() => {
    if (historyIndex < history.length - 1) {
      setHistoryIndex(prev => prev + 1);
      setStages(history[historyIndex + 1]);
    }
  }, [history, historyIndex]);

  // Initialize history on mount
  useEffect(() => {
    if (isFirstRender.current) {
      isFirstRender.current = false;
      setHistory([config.stages || []]);
      setHistoryIndex(0);
    }
  }, [config.stages]);

  function genId() {
    idCounterRef.current += 1;
    return `new-${Date.now()}-${idCounterRef.current}`;
  }

  const configRef = useRef(config);
  configRef.current = config;
  const onChangeRef = useRef(onChange);
  onChangeRef.current = onChange;

  // Sync stages when external config changes (e.g., YAML → Visual switch)
  const configJsonRef = useRef(JSON.stringify(config.stages));
  useEffect(() => {
    const newConfigJson = JSON.stringify(config.stages);
    if (newConfigJson !== configJsonRef.current) {
      configJsonRef.current = newConfigJson;
      setStages(config.stages || []);
      // Reset history when config is replaced externally
      setHistory([config.stages || []]);
      setHistoryIndex(0);
    }
  }, [config.stages]);

  // Notify parent when stages change so mode switch and save have latest data
  const prevStagesJsonRef = useRef(JSON.stringify(config.stages));
  useEffect(() => {
    const nextJson = JSON.stringify(stages);
    if (nextJson !== prevStagesJsonRef.current) {
      prevStagesJsonRef.current = nextJson;
      onChangeRef.current?.({
        ...configRef.current,
        stages,
      });
    }
  }, [stages]);

  const handleAddStage = useCallback(() => {
    setStages((prev) => {
      const newStage: StageConfig = {
        id: genId(),
        name: `Stage ${prev.length + 1}`,
        steps: [],
      };
      const newStages = [...prev, newStage];
      pushHistory(newStages);
      return newStages;
    });
  }, [pushHistory]);

  const handleDeleteStage = useCallback((stageId: string) => {
    setStages((prev) => {
      const newStages = prev.filter((s) => s.id !== stageId);
      pushHistory(newStages);
      return newStages;
    });
  }, [pushHistory]);

  const handleAddStep = useCallback((stageId: string) => {
    setStages((prev) => {
      const newStages = prev.map((stage) => {
        if (stage.id !== stageId) return stage;
        const newStep: StepConfig = {
          id: genId(),
          name: `Step ${stage.steps.length + 1}`,
          type: 'shell',
        };
        return { ...stage, steps: [...stage.steps, newStep] };
      });
      pushHistory(newStages);
      return newStages;
    });
  }, [pushHistory]);

  const handleDeleteStep = useCallback((stageId: string, stepId: string) => {
    setStages((prev) => {
      const newStages = prev.map((stage) => {
        if (stage.id !== stageId) return stage;
        return { ...stage, steps: stage.steps.filter((s) => s.id !== stepId) };
      });
      pushHistory(newStages);
      return newStages;
    });
  }, [pushHistory]);

  const handleMoveStage = useCallback((stageId: string, direction: 'up' | 'down') => {
    setStages((prev) => {
      const index = prev.findIndex((s) => s.id === stageId);
      if (index < 0) return prev;
      const newIndex = direction === 'up' ? index - 1 : index + 1;
      if (newIndex < 0 || newIndex >= prev.length) return prev;
      const newStages = [...prev];
      [newStages[index], newStages[newIndex]] = [newStages[newIndex], newStages[index]];
      pushHistory(newStages);
      return newStages;
    });
  }, [pushHistory]);

  const handleMoveStep = useCallback((stageId: string, stepId: string, direction: 'up' | 'down') => {
    setStages((prev) => {
      const newStages = prev.map((stage) => {
        if (stage.id !== stageId) return stage;
        const index = stage.steps.findIndex((s) => s.id === stepId);
        if (index < 0) return stage;
        const newIndex = direction === 'up' ? index - 1 : index + 1;
        if (newIndex < 0 || newIndex >= stage.steps.length) return stage;
        const newSteps = [...stage.steps];
        [newSteps[index], newSteps[newIndex]] = [newSteps[newIndex], newSteps[index]];
        return { ...stage, steps: newSteps };
      });
      pushHistory(newStages);
      return newStages;
    });
  }, [pushHistory]);

  const handleSelectStep = useCallback((stageId: string, stepId: string) => {
    setSelectedStep({ stageId, stepId });
    setPanelVisible(true);
  }, []);

  const handleSaveStep = useCallback((updatedStep: StepConfig) => {
    if (!selectedStep) return;
    setStages((prev) => {
      const newStages = prev.map((stage) => {
        if (stage.id !== selectedStep.stageId) return stage;
        return {
          ...stage,
          steps: stage.steps.map((s) => (s.id === updatedStep.id ? updatedStep : s)),
        };
      });
      pushHistory(newStages);
      return newStages;
    });
  }, [selectedStep, pushHistory]);

  const handleSave = useCallback(() => {
    const updatedConfig: PipelineConfig = {
      ...configRef.current,
      stages,
    };
    onSave(updatedConfig);
  }, [stages, onSave]);

  const currentStep = useMemo(() => {
    if (!selectedStep) return null;
    const stage = stages.find((s) => s.id === selectedStep.stageId);
    return stage?.steps.find((s) => s.id === selectedStep.stepId) || null;
  }, [selectedStep, stages]);

  const { nodes: layoutedNodes, edges: layoutedEdges } = useMemo(
    () =>
      layoutNodes(stages, {
        onAddStep: handleAddStep,
        onDeleteStage: handleDeleteStage,
        onSelectStep: handleSelectStep,
        onDeleteStep: handleDeleteStep,
        onMoveStage: handleMoveStage,
        onMoveStep: handleMoveStep,
      }),
    [stages, handleAddStep, handleDeleteStage, handleSelectStep, handleDeleteStep, handleMoveStage, handleMoveStep]
  );

  const [nodes, setNodes, onNodesChange] = useNodesState(layoutedNodes);
  const [edges, setEdges, onEdgesChange] = useEdgesState(layoutedEdges);

  // Sync React Flow state when layout changes (add/delete/move stage/step)
  useEffect(() => {
    setNodes(layoutedNodes);
    setEdges(layoutedEdges);
  }, [layoutedNodes, layoutedEdges, setNodes, setEdges]);

  // Keyboard shortcuts for undo/redo and save
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      // Ctrl+S or Cmd+S to save
      if ((e.ctrlKey || e.metaKey) && e.key === 's') {
        e.preventDefault();
        handleSave();
        return;
      }
      // Undo: Ctrl+Z
      if ((e.ctrlKey || e.metaKey) && e.key === 'z' && !e.shiftKey) {
        e.preventDefault();
        undo();
        return;
      }
      // Redo: Ctrl+Y or Ctrl+Shift+Z
      if ((e.ctrlKey || e.metaKey) && (e.key === 'y' || (e.key === 'z' && e.shiftKey))) {
        e.preventDefault();
        redo();
        return;
      }
      // Delete selected step: Del or Backspace
      if ((e.key === 'Delete' || e.key === 'Backspace') && selectedStep) {
        e.preventDefault();
        handleDeleteStep(selectedStep.stageId, selectedStep.stepId);
        setPanelVisible(false);
        return;
      }
      // Escape to close panel
      if (e.key === 'Escape' && panelVisible) {
        setPanelVisible(false);
        return;
      }
    };
    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [handleSave, undo, redo, selectedStep, handleDeleteStep]);

  // Validation function for pipeline config
  const validateConfig = useCallback((cfg: PipelineConfig): string | null => {
    if (!cfg.schemaVersion) {
      return '缺少 schemaVersion';
    }
    if (!Array.isArray(cfg.stages) || cfg.stages.length === 0) {
      return '请至少添加一个 Stage';
    }
    for (let i = 0; i < cfg.stages.length; i++) {
      const stage = cfg.stages[i];
      if (!stage.id || !stage.name) {
        return `Stage ${i + 1}: 缺少 id 或 name`;
      }
      if (!Array.isArray(stage.steps)) {
        return `Stage "${stage.name}": steps 必须是数组`;
      }
      for (let j = 0; j < stage.steps.length; j++) {
        const step = stage.steps[j];
        if (!step.id || !step.name) {
          return `Stage "${stage.name}" Step ${j + 1}: 缺少 id 或 name`;
        }
        if (!step.type) {
          return `Step "${step.name}": 缺少 type`;
        }
      }
    }
    return null;
  }, []);

  const validationError = useMemo(() => validateConfig({ ...configRef.current, stages }), [stages, validateConfig]);

  return (
    <div style={{ width: '100%', height: '100%', position: 'relative' }}>
      <div style={{ position: 'absolute', top: 12, left: 12, zIndex: 10 }}>
        <Space>
          <Tooltip content="撤销 (Ctrl+Z)">
            <Button icon={<IconLeft />} onClick={undo} disabled={!canUndo}>
              撤销
            </Button>
          </Tooltip>
          <Tooltip content="重做 (Ctrl+Y)">
            <Button icon={<IconRight />} onClick={redo} disabled={!canRedo}>
              重做
            </Button>
          </Tooltip>
          <Button type="primary" icon={<IconPlus />} onClick={handleAddStage}>
            添加 Stage
          </Button>
          <Button
            type="primary"
            status={validationError ? 'warning' : 'success'}
            icon={<IconSave />}
            onClick={handleSave}
            loading={saving}
            disabled={!!validationError}
          >
            保存
          </Button>
          {validationError && (
            <Tooltip content={validationError}>
              <span style={{ color: 'var(--color-warning-6)', fontSize: 12, padding: '4px 8px', background: 'var(--color-warning-1)', borderRadius: 4 }}>
                {validationError}
              </span>
            </Tooltip>
          )}
        </Space>
      </div>

      <ReactFlow
        nodes={nodes}
        edges={edges}
        onNodesChange={onNodesChange}
        onEdgesChange={onEdgesChange}
        nodeTypes={nodeTypes}
        fitView
        attributionPosition="bottom-left"
        style={{ background: 'var(--zcid-color-bg-secondary)' }}
      >
        <Controls />
        <MiniMap />
        <Background variant={BackgroundVariant.Dots} gap={16} size={1} />
      </ReactFlow>

      <StepConfigPanel
        visible={panelVisible}
        step={currentStep}
        onClose={() => setPanelVisible(false)}
        onSave={handleSaveStep}
      />
    </div>
  );
}
