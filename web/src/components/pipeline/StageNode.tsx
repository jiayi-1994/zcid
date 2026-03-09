import { memo } from 'react';
import { Handle, Position, type NodeProps } from '@xyflow/react';
import { Button, Typography, Tooltip } from '@arco-design/web-react';
import { IconPlus, IconDelete, IconUp, IconDown } from '@arco-design/web-react/icon';

const { Text } = Typography;

export interface StageNodeData {
  label: string;
  stageId: string;
  stageIndex: number;
  totalStages: number;
  onAddStep?: (stageId: string) => void;
  onDelete?: (stageId: string) => void;
  onMove?: (stageId: string, direction: 'up' | 'down') => void;
}

function StageNodeComponent({ data }: NodeProps) {
  const { label, stageId, stageIndex, totalStages, onAddStep, onDelete, onMove } = data as unknown as StageNodeData;
  const canMoveUp = stageIndex > 0;
  const canMoveDown = stageIndex < totalStages - 1;

  return (
    <div className="zcid-stage-node" tabIndex={0}>
      <Handle type="target" position={Position.Left} style={{ background: 'var(--zcid-color-primary)' }} />
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: 8 }}>
        <Text bold style={{ fontSize: 14, color: 'var(--zcid-color-text-primary)' }}>{label}</Text>
        <div style={{ display: 'flex', gap: 2 }}>
          <Tooltip content="上移">
            <Button
              size="mini"
              type="text"
              icon={<IconUp />}
              disabled={!canMoveUp}
              onClick={(e) => { e.stopPropagation(); onMove?.(stageId, 'up'); }}
            />
          </Tooltip>
          <Tooltip content="下移">
            <Button
              size="mini"
              type="text"
              icon={<IconDown />}
              disabled={!canMoveDown}
              onClick={(e) => { e.stopPropagation(); onMove?.(stageId, 'down'); }}
            />
          </Tooltip>
          <Tooltip content="添加 Step">
            <Button
              size="mini"
              type="text"
              icon={<IconPlus />}
              onClick={(e) => { e.stopPropagation(); onAddStep?.(stageId); }}
            />
          </Tooltip>
          <Tooltip content="删除 Stage">
            <Button
              size="mini"
              type="text"
              status="danger"
              icon={<IconDelete />}
              onClick={(e) => { e.stopPropagation(); onDelete?.(stageId); }}
            />
          </Tooltip>
        </div>
      </div>
      <Handle type="source" position={Position.Right} style={{ background: 'var(--zcid-color-primary)' }} />
    </div>
  );
}

export const StageNode = memo(StageNodeComponent);
