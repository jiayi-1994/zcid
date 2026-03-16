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
    <div
      style={{
        background: 'linear-gradient(135deg, #E8F3FF 0%, #F2F3F5 100%)',
        border: '2px solid #BEDAFF',
        borderRadius: 12,
        padding: '10px 14px',
        minWidth: 200,
        boxShadow: '0 2px 8px rgba(22, 93, 255, 0.08)',
        transition: 'box-shadow 0.2s, border-color 0.2s',
      }}
      tabIndex={0}
      onMouseEnter={(e) => {
        (e.currentTarget as HTMLElement).style.boxShadow = '0 4px 16px rgba(22, 93, 255, 0.15)';
        (e.currentTarget as HTMLElement).style.borderColor = '#165DFF';
      }}
      onMouseLeave={(e) => {
        (e.currentTarget as HTMLElement).style.boxShadow = '0 2px 8px rgba(22, 93, 255, 0.08)';
        (e.currentTarget as HTMLElement).style.borderColor = '#BEDAFF';
      }}
    >
      <Handle type="target" position={Position.Left} style={{ background: '#165DFF', width: 8, height: 8 }} />
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
          <div style={{
            width: 24, height: 24, borderRadius: 6,
            background: '#165DFF', color: '#fff',
            display: 'flex', alignItems: 'center', justifyContent: 'center',
            fontSize: 11, fontWeight: 700,
          }}>
            {stageIndex + 1}
          </div>
          <Text bold style={{ fontSize: 14, color: '#1D2129' }}>{label}</Text>
        </div>
        <div style={{ display: 'flex', gap: 1 }}>
          {canMoveUp && (
            <Tooltip content="上移" mini>
              <Button size="mini" type="text" icon={<IconUp />}
                onClick={(e) => { e.stopPropagation(); onMove?.(stageId, 'up'); }}
                style={{ color: '#86909C' }}
              />
            </Tooltip>
          )}
          {canMoveDown && (
            <Tooltip content="下移" mini>
              <Button size="mini" type="text" icon={<IconDown />}
                onClick={(e) => { e.stopPropagation(); onMove?.(stageId, 'down'); }}
                style={{ color: '#86909C' }}
              />
            </Tooltip>
          )}
          <Tooltip content="添加 Step" mini>
            <Button size="mini" type="text" icon={<IconPlus />}
              onClick={(e) => { e.stopPropagation(); onAddStep?.(stageId); }}
              style={{ color: '#165DFF' }}
            />
          </Tooltip>
          <Tooltip content="删除" mini>
            <Button size="mini" type="text" status="danger" icon={<IconDelete />}
              onClick={(e) => { e.stopPropagation(); onDelete?.(stageId); }}
            />
          </Tooltip>
        </div>
      </div>
      <Handle type="source" position={Position.Right} style={{ background: '#165DFF', width: 8, height: 8 }} />
    </div>
  );
}

export const StageNode = memo(StageNodeComponent);
