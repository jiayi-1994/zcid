import { memo } from 'react';
import { type NodeProps } from '@xyflow/react';
import { Tag, Typography, Button, Tooltip } from '@arco-design/web-react';
import { IconDelete, IconUp, IconDown } from '@arco-design/web-react/icon';

const { Text } = Typography;

const stepTypeConfig: Record<string, { color: string; bg: string; icon: string; label: string }> = {
  'git-clone': { color: '#165DFF', bg: '#E8F3FF', icon: '📥', label: 'Git Clone' },
  shell:       { color: '#00B42A', bg: '#E8FFEA', icon: '💻', label: 'Shell' },
  kaniko:      { color: '#FF7D00', bg: '#FFF7E8', icon: '🐳', label: 'Kaniko' },
  buildkit:    { color: '#722ED1', bg: '#F5E8FF', icon: '🔨', label: 'BuildKit' },
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
  const cfg = stepTypeConfig[type] || { color: '#86909C', bg: '#F2F3F5', icon: '⚙️', label: type };

  return (
    <div
      style={{
        background: '#fff',
        border: `1.5px solid ${cfg.color}30`,
        borderRadius: 10,
        padding: '8px 12px',
        minWidth: 180,
        cursor: 'pointer',
        transition: 'all 0.2s',
        boxShadow: '0 1px 4px rgba(0,0,0,0.04)',
        borderLeft: `4px solid ${cfg.color}`,
      }}
      tabIndex={0}
      role="button"
      onClick={() => onSelect?.(stageId, stepId)}
      onKeyDown={(e) => {
        if (e.key === 'Enter' || e.key === ' ') {
          e.preventDefault();
          onSelect?.(stageId, stepId);
        }
      }}
      onMouseEnter={(e) => {
        (e.currentTarget as HTMLElement).style.boxShadow = `0 4px 12px ${cfg.color}20`;
        (e.currentTarget as HTMLElement).style.borderColor = cfg.color;
      }}
      onMouseLeave={(e) => {
        (e.currentTarget as HTMLElement).style.boxShadow = '0 1px 4px rgba(0,0,0,0.04)';
        (e.currentTarget as HTMLElement).style.borderColor = `${cfg.color}30`;
      }}
    >
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
          <span style={{ fontSize: 14 }}>{cfg.icon}</span>
          <Text style={{ fontSize: 13, fontWeight: 600, color: '#1D2129' }}>{label}</Text>
        </div>
        <div style={{ display: 'flex', gap: 1 }}>
          {canMoveUp && (
            <Tooltip content="上移" mini>
              <Button size="mini" type="text" icon={<IconUp />}
                onClick={(e) => { e.stopPropagation(); onMove?.(stageId, stepId, 'up'); }}
                style={{ color: '#86909C' }}
              />
            </Tooltip>
          )}
          {canMoveDown && (
            <Tooltip content="下移" mini>
              <Button size="mini" type="text" icon={<IconDown />}
                onClick={(e) => { e.stopPropagation(); onMove?.(stageId, stepId, 'down'); }}
                style={{ color: '#86909C' }}
              />
            </Tooltip>
          )}
          <Tooltip content="删除" mini>
            <Button size="mini" type="text" status="danger" icon={<IconDelete />}
              onClick={(e) => { e.stopPropagation(); onDelete?.(stageId, stepId); }}
            />
          </Tooltip>
        </div>
      </div>
      <div style={{ display: 'flex', gap: 6, marginTop: 6, alignItems: 'center' }}>
        <span style={{
          display: 'inline-flex', alignItems: 'center', gap: 4,
          padding: '2px 8px', borderRadius: 10,
          background: cfg.bg, color: cfg.color,
          fontSize: 11, fontWeight: 600,
        }}>
          {cfg.label}
        </span>
        {image && (
          <span style={{
            padding: '2px 8px', borderRadius: 10,
            background: '#F2F3F5', color: '#4E5969',
            fontSize: 11, maxWidth: 120, overflow: 'hidden',
            textOverflow: 'ellipsis', whiteSpace: 'nowrap',
          }}>
            {image}
          </span>
        )}
      </div>
    </div>
  );
}

export const StepNode = memo(StepNodeComponent);
