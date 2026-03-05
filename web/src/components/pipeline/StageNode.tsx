import { memo } from 'react';
import { Handle, Position, type NodeProps } from '@xyflow/react';
import { Button, Typography } from '@arco-design/web-react';
import { IconPlus, IconDelete } from '@arco-design/web-react/icon';

const { Text } = Typography;

export interface StageNodeData {
  label: string;
  stageId: string;
  onAddStep?: (stageId: string) => void;
  onDelete?: (stageId: string) => void;
}

function StageNodeComponent({ data }: NodeProps) {
  const { label, stageId, onAddStep, onDelete } = data as unknown as StageNodeData;

  return (
    <div
      style={{
        padding: '12px 16px',
        borderRadius: 'var(--zcid-radius-md)',
        background: 'var(--zcid-color-bg-tertiary)',
        border: '2px solid var(--zcid-color-border)',
        minWidth: 200,
        transition: 'border-color var(--zcid-transition-fast), box-shadow var(--zcid-transition-fast)',
      }}
      onMouseEnter={(e) => {
        e.currentTarget.style.borderColor = 'var(--zcid-color-primary-light)';
        e.currentTarget.style.boxShadow = 'var(--zcid-shadow-md)';
      }}
      onMouseLeave={(e) => {
        e.currentTarget.style.borderColor = 'var(--zcid-color-border)';
        e.currentTarget.style.boxShadow = 'none';
      }}
    >
      <Handle type="target" position={Position.Left} style={{ background: 'var(--zcid-color-primary)' }} />
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: 8 }}>
        <Text bold style={{ fontSize: 14, color: 'var(--zcid-color-text-primary)' }}>{label}</Text>
        <div style={{ display: 'flex', gap: 4 }}>
          <Button
            size="mini"
            type="text"
            icon={<IconPlus />}
            onClick={() => onAddStep?.(stageId)}
            style={{ cursor: 'pointer' }}
          />
          <Button
            size="mini"
            type="text"
            status="danger"
            icon={<IconDelete />}
            onClick={() => onDelete?.(stageId)}
            style={{ cursor: 'pointer' }}
          />
        </div>
      </div>
      <Handle type="source" position={Position.Right} style={{ background: 'var(--zcid-color-primary)' }} />
    </div>
  );
}

export const StageNode = memo(StageNodeComponent);
