import { memo, useEffect, useRef, useState } from 'react';
import { Handle, Position, type NodeProps } from '@xyflow/react';
import { Btn } from '../ui/Btn';
import { IPlus, ITrash, IUp, IDown } from '../ui/icons';

export interface StageNodeData {
  label: string;
  stageId: string;
  stageIndex: number;
  totalStages: number;
  onAddStep?: (stageId: string) => void;
  onDelete?: (stageId: string) => void;
  onMove?: (stageId: string, direction: 'up' | 'down') => void;
  onRename?: (stageId: string, newName: string) => void;
}

function StageNodeComponent({ data }: NodeProps) {
  const { label, stageId, stageIndex, totalStages, onAddStep, onDelete, onMove, onRename } = data as unknown as StageNodeData;
  const canMoveUp = stageIndex > 0;
  const canMoveDown = stageIndex < totalStages - 1;
  const [editing, setEditing] = useState(false);
  const [editName, setEditName] = useState(label);
  const inputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    if (editing && inputRef.current) {
      inputRef.current.focus();
      inputRef.current.select();
    }
  }, [editing]);

  const handleSaveName = () => {
    const trimmed = editName.trim();
    if (trimmed && trimmed !== label) onRename?.(stageId, trimmed);
    else setEditName(label);
    setEditing(false);
  };

  return (
    <div
      className="zc"
      style={{
        background: 'color-mix(in oklch, var(--accent-1), white 90%)',
        border: '1px solid color-mix(in oklch, var(--accent-1), white 75%)',
        borderRadius: 10,
        padding: '8px 12px',
        minWidth: 220,
        boxShadow: 'var(--shadow-sm)',
        fontFamily: 'var(--font-sans)',
        transition: 'box-shadow .15s, border-color .15s',
      }}
      tabIndex={0}
    >
      <Handle type="target" position={Position.Left} style={{ background: 'var(--accent-1)', width: 8, height: 8, border: 'none' }} />
      <div style={{ display: 'flex', alignItems: 'center', gap: 8, justifyContent: 'space-between' }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: 8, flex: 1, minWidth: 0 }}>
          <div style={{
            width: 22, height: 22, borderRadius: '50%',
            background: 'linear-gradient(135deg, var(--accent-1), var(--accent-2))',
            color: '#fff',
            display: 'flex', alignItems: 'center', justifyContent: 'center',
            fontSize: 11, fontWeight: 700, flex: 'none',
            boxShadow: '0 1px 3px rgba(0,0,0,.15)',
          }}>
            {stageIndex + 1}
          </div>
          {editing ? (
            <input
              ref={inputRef}
              className="input"
              style={{ height: 22, fontSize: 12.5, fontWeight: 600, flex: 1, minWidth: 0, padding: '0 6px' }}
              value={editName}
              onChange={(e) => setEditName(e.target.value)}
              onBlur={handleSaveName}
              onKeyDown={(e) => { if (e.key === 'Enter') handleSaveName(); else if (e.key === 'Escape') { setEditName(label); setEditing(false); } }}
            />
          ) : (
            <span
              style={{ fontSize: 13, fontWeight: 600, color: 'var(--z-900)', cursor: 'text', flex: 1, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}
              onDoubleClick={(e) => { e.stopPropagation(); setEditing(true); }}
              title="双击重命名"
            >
              {label}
            </span>
          )}
        </div>
        <div style={{ display: 'flex', gap: 2, flex: 'none' }}>
          {canMoveUp && (
            <Btn size="xs" variant="ghost" iconOnly icon={<IUp size={10} />} title="左移" onClick={(e) => { e.stopPropagation(); onMove?.(stageId, 'up'); }} />
          )}
          {canMoveDown && (
            <Btn size="xs" variant="ghost" iconOnly icon={<IDown size={10} />} title="右移" onClick={(e) => { e.stopPropagation(); onMove?.(stageId, 'down'); }} />
          )}
          <Btn size="xs" variant="ghost" iconOnly icon={<IPlus size={11} />} title="添加 Step" onClick={(e) => { e.stopPropagation(); onAddStep?.(stageId); }} />
          <Btn size="xs" variant="ghost" iconOnly icon={<ITrash size={10} />} title="删除" onClick={(e) => { e.stopPropagation(); onDelete?.(stageId); }} />
        </div>
      </div>
      <Handle type="source" position={Position.Right} style={{ background: 'var(--accent-1)', width: 8, height: 8, border: 'none' }} />
    </div>
  );
}

export const StageNode = memo(StageNodeComponent);
