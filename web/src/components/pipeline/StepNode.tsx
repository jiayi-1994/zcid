import { memo } from 'react';
import { type NodeProps } from '@xyflow/react';
import { Tag, Typography } from '@arco-design/web-react';
import { IconDelete } from '@arco-design/web-react/icon';
import { Button } from '@arco-design/web-react';

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
  type: string;
  image?: string;
  onSelect?: (stageId: string, stepId: string) => void;
  onDelete?: (stageId: string, stepId: string) => void;
}

function StepNodeComponent({ data }: NodeProps) {
  const { label, stepId, stageId, type, image, onSelect, onDelete } = data as unknown as StepNodeData;

  return (
    <div
      onClick={() => onSelect?.(stageId, stepId)}
      style={{
        padding: '8px 12px',
        borderRadius: 'var(--zcid-radius-sm)',
        background: 'var(--zcid-color-bg-elevated)',
        border: '1px solid var(--zcid-color-border)',
        minWidth: 180,
        cursor: 'pointer',
        transition: 'all var(--zcid-transition-fast)',
      }}
      onMouseEnter={(e) => {
        e.currentTarget.style.boxShadow = 'var(--zcid-shadow-md)';
        e.currentTarget.style.borderColor = 'var(--zcid-color-primary-lighter)';
        e.currentTarget.style.transform = 'translateY(-1px)';
      }}
      onMouseLeave={(e) => {
        e.currentTarget.style.boxShadow = 'none';
        e.currentTarget.style.borderColor = 'var(--zcid-color-border)';
        e.currentTarget.style.transform = 'translateY(0)';
      }}
    >
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
        <Text style={{ fontSize: 13, fontWeight: 500 }}>{label}</Text>
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
      </div>
      <div style={{ display: 'flex', gap: 4, marginTop: 4 }}>
        <Tag size="small" color={stepTypeColors[type] || 'gray'}>{type}</Tag>
        {image && <Tag size="small" color="arcoblue">{image}</Tag>}
      </div>
    </div>
  );
}

export const StepNode = memo(StepNodeComponent);
