import { memo } from 'react';
import { type NodeProps } from '@xyflow/react';
import { Tag, Typography, Button, Tooltip } from '@arco-design/web-react';
import { IconDelete, IconUp, IconDown } from '@arco-design/web-react/icon';

const { Text } = Typography;

const stepTypeColors: Record<string, string> = {
  'git-clone': 'blue',
  shell: 'green',
  kaniko: 'orange',
  buildkit: 'purple',
};

export interface StepNodeData {
  label: string;
  stepId: string;
  stageId: string;
  stepIndex: number;
  totalSteps: number;
  type: string;
  image?: string;
  onSelect?: (stageId: string, stepId: string) => void;
  onDelete?: (stageId: string, stepId: string) => void;
  onMove?: (stageId: string, stepId: string, direction: 'up' | 'down') => void;
}

function StepNodeComponent({ data }: NodeProps) {
  const { label, stepId, stageId, stepIndex, totalSteps, type, image, onSelect, onDelete, onMove } = data as unknown as StepNodeData;
  const canMoveUp = stepIndex > 0;
  const canMoveDown = stepIndex < totalSteps - 1;

  return (
    <div
      className="zcid-step-node"
      tabIndex={0}
      role="button"
      onClick={() => onSelect?.(stageId, stepId)}
      onKeyDown={(e) => {
        if (e.key === 'Enter' || e.key === ' ') {
          e.preventDefault();
          onSelect?.(stageId, stepId);
        }
      }}
    >
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
        <Text style={{ fontSize: 13, fontWeight: 500 }}>{label}</Text>
        <div style={{ display: 'flex', gap: 2 }}>
          <Tooltip content="上移">
            <Button
              size="mini"
              type="text"
              icon={<IconUp />}
              disabled={!canMoveUp}
              onClick={(e) => {
                e.stopPropagation();
                onMove?.(stageId, stepId, 'up');
              }}
            />
          </Tooltip>
          <Tooltip content="下移">
            <Button
              size="mini"
              type="text"
              icon={<IconDown />}
              disabled={!canMoveDown}
              onClick={(e) => {
                e.stopPropagation();
                onMove?.(stageId, stepId, 'down');
              }}
            />
          </Tooltip>
          <Tooltip content="删除">
            <Button
              size="mini"
              type="text"
              status="danger"
              icon={<IconDelete />}
              onClick={(e) => {
                e.stopPropagation();
                onDelete?.(stageId, stepId);
              }}
            />
          </Tooltip>
        </div>
      </div>
      <div style={{ display: 'flex', gap: 4, marginTop: 4 }}>
        <Tag size="small" color={stepTypeColors[type] || 'gray'}>{type}</Tag>
        {image && <Tag size="small" color="arcoblue">{image}</Tag>}
      </div>
    </div>
  );
}

export const StepNode = memo(StepNodeComponent);
