import { memo, type ReactNode } from 'react';
import { type NodeProps } from '@xyflow/react';
import { Btn } from '../ui/Btn';
import { ITrash, IUp, IDown, IBranch, ITerminal, ICube, ILayers } from '../ui/icons';

type Tone = 'blue' | 'accent' | 'cyan' | 'grey';

const STEP_META: Record<string, { tone: Tone; label: string; icon: ReactNode }> = {
  'git-clone':    { tone: 'blue',   label: 'Git Clone',    icon: <IBranch size={11} /> },
  'shell':        { tone: 'grey',   label: 'Shell',        icon: <ITerminal size={11} /> },
  'kaniko':       { tone: 'accent', label: 'Kaniko',       icon: <ICube size={11} /> },
  'kaniko-build': { tone: 'accent', label: 'Kaniko',       icon: <ICube size={11} /> },
  'buildkit':     { tone: 'cyan',   label: 'BuildKit',     icon: <ILayers size={11} /> },
  'buildkit-build': { tone: 'cyan', label: 'BuildKit',     icon: <ILayers size={11} /> },
};

function toneSoft(tone: Tone): string {
  return tone === 'grey' ? 'var(--z-100)' : tone === 'accent' ? 'color-mix(in oklch, var(--accent-1), white 85%)' : `var(--${tone}-soft)`;
}
function toneInk(tone: Tone): string {
  return tone === 'grey' ? 'var(--z-700)' : tone === 'accent' ? 'var(--accent-ink)' : `var(--${tone}-ink)`;
}
function toneBorder(tone: Tone): string {
  return tone === 'grey' ? 'var(--z-300)' : tone === 'accent' ? 'var(--accent-1)' : `var(--${tone}-ink)`;
}

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
  const meta = STEP_META[type] ?? { tone: 'grey' as Tone, label: type, icon: null };

  return (
    <div
      className="zc"
      style={{
        background: 'var(--z-0)',
        border: '1px solid var(--z-200)',
        borderLeft: `3px solid ${toneBorder(meta.tone)}`,
        borderRadius: 8,
        padding: '9px 11px',
        minWidth: 200,
        cursor: 'pointer',
        boxShadow: 'var(--shadow-xs)',
        fontFamily: 'var(--font-sans)',
        transition: 'box-shadow .15s, border-color .15s',
      }}
      tabIndex={0}
      role="button"
      onClick={() => onSelect?.(stageId, stepId)}
      onKeyDown={(e) => { if (e.key === 'Enter' || e.key === ' ') { e.preventDefault(); onSelect?.(stageId, stepId); } }}
    >
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', gap: 6 }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: 7, flex: 1, minWidth: 0 }}>
          <div style={{
            width: 22, height: 22, borderRadius: 5,
            background: toneSoft(meta.tone), color: toneInk(meta.tone),
            display: 'flex', alignItems: 'center', justifyContent: 'center', flex: 'none',
          }}>
            {meta.icon}
          </div>
          <span style={{ fontSize: 12.5, fontWeight: 500, color: 'var(--z-900)', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
            {label}
          </span>
        </div>
        <div style={{ display: 'flex', gap: 2, flex: 'none' }}>
          {canMoveUp && (
            <Btn size="xs" variant="ghost" iconOnly icon={<IUp size={10} />} title="上移" onClick={(e) => { e.stopPropagation(); onMove?.(stageId, stepId, 'up'); }} />
          )}
          {canMoveDown && (
            <Btn size="xs" variant="ghost" iconOnly icon={<IDown size={10} />} title="下移" onClick={(e) => { e.stopPropagation(); onMove?.(stageId, stepId, 'down'); }} />
          )}
          <Btn size="xs" variant="ghost" iconOnly icon={<ITrash size={10} />} title="删除" onClick={(e) => { e.stopPropagation(); onDelete?.(stageId, stepId); }} />
        </div>
      </div>
      <div style={{ display: 'flex', gap: 5, marginTop: 5, alignItems: 'center' }}>
        <span style={{
          display: 'inline-flex', alignItems: 'center',
          padding: '1px 6px', borderRadius: 4,
          background: toneSoft(meta.tone), color: toneInk(meta.tone),
          fontSize: 10.5, fontWeight: 500,
        }}>
          {meta.label}
        </span>
        {image && (
          <span className="mono" style={{
            padding: '1px 6px', borderRadius: 4,
            background: 'var(--z-50)', color: 'var(--z-600)',
            fontSize: 10.5, maxWidth: 120, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap',
          }}>
            {image}
          </span>
        )}
      </div>
    </div>
  );
}

export const StepNode = memo(StepNodeComponent);
