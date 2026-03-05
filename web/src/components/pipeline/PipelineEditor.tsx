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
import { Button, Space, Message } from '@arco-design/web-react';
import { IconPlus, IconSave } from '@arco-design/web-react/icon';
import { StageNode } from './StageNode';
import { StepNode } from './StepNode';
import { StepConfigPanel } from './StepConfigPanel';
import type { PipelineConfig, StageConfig, StepConfig } from '../../services/pipeline';

interface PipelineEditorProps {
  config: PipelineConfig;
  onSave: (config: PipelineConfig) => void;
  onChange?: (config: PipelineConfig) => void;
  saving?: boolean;
}

const nodeTypes = {
  stage: StageNode,
  step: StepNode,
};

function layoutNodes(stages: StageConfig[], callbacks: {
  onAddStep: (stageId: string) => void;
  onDeleteStage: (stageId: string) => void;
  onSelectStep: (stageId: string, stepId: string) => void;
  onDeleteStep: (stageId: string, stepId: string) => void;
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
        onAddStep: callbacks.onAddStep,
        onDelete: callbacks.onDeleteStage,
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
          type: step.type,
          image: step.image,
          onSelect: callbacks.onSelectStep,
          onDelete: callbacks.onDeleteStep,
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

let idCounter = 0;
function genId() {
  idCounter += 1;
  return `new-${Date.now()}-${idCounter}`;
}

export function PipelineEditor({ config, onSave, onChange, saving }: PipelineEditorProps) {
  const [stages, setStages] = useState<StageConfig[]>(config.stages || []);
  const [selectedStep, setSelectedStep] = useState<{ stageId: string; stepId: string } | null>(null);
  const [panelVisible, setPanelVisible] = useState(false);
  const isFirstRender = useRef(true);

  useEffect(() => {
    if (isFirstRender.current) {
      isFirstRender.current = false;
      return;
    }
    onChange?.({ ...config, stages });
  }, [stages]);

  const handleAddStage = useCallback(() => {
    const newStage: StageConfig = {
      id: genId(),
      name: `Stage ${stages.length + 1}`,
      steps: [],
    };
    setStages((prev) => [...prev, newStage]);
  }, [stages.length]);

  const handleDeleteStage = useCallback((stageId: string) => {
    setStages((prev) => prev.filter((s) => s.id !== stageId));
  }, []);

  const handleAddStep = useCallback((stageId: string) => {
    setStages((prev) =>
      prev.map((stage) => {
        if (stage.id !== stageId) return stage;
        const newStep: StepConfig = {
          id: genId(),
          name: `Step ${stage.steps.length + 1}`,
          type: 'shell',
        };
        return { ...stage, steps: [...stage.steps, newStep] };
      })
    );
  }, []);

  const handleDeleteStep = useCallback((stageId: string, stepId: string) => {
    setStages((prev) =>
      prev.map((stage) => {
        if (stage.id !== stageId) return stage;
        return { ...stage, steps: stage.steps.filter((s) => s.id !== stepId) };
      })
    );
  }, []);

  const handleSelectStep = useCallback((stageId: string, stepId: string) => {
    setSelectedStep({ stageId, stepId });
    setPanelVisible(true);
  }, []);

  const handleSaveStep = useCallback((updatedStep: StepConfig) => {
    if (!selectedStep) return;
    setStages((prev) =>
      prev.map((stage) => {
        if (stage.id !== selectedStep.stageId) return stage;
        return {
          ...stage,
          steps: stage.steps.map((s) => (s.id === updatedStep.id ? updatedStep : s)),
        };
      })
    );
  }, [selectedStep]);

  const handleSave = useCallback(() => {
    const updatedConfig: PipelineConfig = {
      ...config,
      stages,
    };
    onSave(updatedConfig);
    Message.success('流水线配置已保存');
  }, [config, stages, onSave]);

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
      }),
    [stages, handleAddStep, handleDeleteStage, handleSelectStep, handleDeleteStep]
  );

  const [nodes, , onNodesChange] = useNodesState(layoutedNodes);
  const [edges, , onEdgesChange] = useEdgesState(layoutedEdges);

  return (
    <div style={{ width: '100%', height: '100%', position: 'relative' }}>
      <div style={{ position: 'absolute', top: 12, left: 12, zIndex: 10 }}>
        <Space>
          <Button type="primary" icon={<IconPlus />} onClick={handleAddStage}>
            添加 Stage
          </Button>
          <Button type="primary" status="success" icon={<IconSave />} onClick={handleSave} loading={saving}>
            保存
          </Button>
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
