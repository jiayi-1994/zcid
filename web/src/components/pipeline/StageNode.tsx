import { memo, useState, useRef, useEffect } from 'react';
import { Handle, Position, type NodeProps } from '@xyflow/react';
import { Button, Typography, Tooltip, Input } from '@arco-design/web-react';
import { IconPlus, IconDelete, IconUp, IconDown, IconEdit } from '@arco-design/web-react/icon';

const { Text } = Typography;

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
    if (trimmed && trimmed !== label) {
      onRename?.(stageId, trimmed);
    } else {
      setEditName(label);
    }
    setEditing(false);
  };

  return (
    <div
      style={{
        background: 'linear-gradient(135deg, #d9e2ff 0%, #f8f9fb 100%)',
        borderRadius: 24,
        padding: '10px 18px',
        minWidth: 220,
        boxShadow: '0 2px 8px rgba(0, 87, 194, 0.1)',
        transition: 'all 0.2s',
      }}
      tabIndex={0}
      onMouseEnter={(e) => {
        (e.currentTarget as HTMLElement).style.boxShadow = '0 0 0 4px rgba(0, 87, 194, 0.2), 0 8px 24px rgba(0, 87, 194, 0.15)';
      }}
      onMouseLeave={(e) => {
        (e.currentTarget as HTMLElement).style.boxShadow = '0 2px 8px rgba(0, 87, 194, 0.1)';
      }}
    >
      <Handle type="target" position={Position.Left} style={{ background: '#0057c2', width: 8, height: 8 }} />
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: 6, flex: 1, minWidth: 0 }}>
          <div style={{
            width: 26, height: 26, borderRadius: '50%',
            background: 'linear-gradient(135deg, #0057c2 0%, #006ef2 100%)',
            color: '#fff',
            display: 'flex', alignItems: 'center', justifyContent: 'center',
            fontFamily: "'Manrope', system-ui, sans-serif",
            fontSize: 12, fontWeight: 700, flexShrink: 0,
            boxShadow: '0 2px 6px rgba(0, 87, 194, 0.4)',
          }}>
            {stageIndex + 1}
          </div>
          {editing ? (
            <Input
              ref={inputRef as any}
              size="mini"
              value={editName}
              onChange={setEditName}
              onBlur={handleSaveName}
              onPressEnter={handleSaveName}
              style={{ flex: 1, fontSize: 13, fontWeight: 600, borderRadius: 4 }}
            />
          ) : (
            <Text bold style={{ fontSize: 14, color: '#1D2129', cursor: 'text', flex: 1, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}
              onDoubleClick={(e) => { e.stopPropagation(); setEditing(true); }}
            >
              {label}
            </Text>
          )}
        </div>
        <div style={{ display: 'flex', gap: 1, flexShrink: 0 }}>
          {canMoveUp && (
            <Tooltip content="左移" mini>
              <Button size="mini" type="text" icon={<IconUp />}
                onClick={(e) => { e.stopPropagation(); onMove?.(stageId, 'up'); }}
                style={{ color: '#86909C' }}
              />
            </Tooltip>
          )}
          {canMoveDown && (
            <Tooltip content="右移" mini>
              <Button size="mini" type="text" icon={<IconDown />}
                onClick={(e) => { e.stopPropagation(); onMove?.(stageId, 'down'); }}
                style={{ color: '#86909C' }}
              />
            </Tooltip>
          )}
          <Tooltip content="添加 Step" mini>
            <Button size="mini" type="text" icon={<IconPlus />}
              onClick={(e) => { e.stopPropagation(); onAddStep?.(stageId); }}
              style={{ color: '#0057c2' }}
            />
          </Tooltip>
          <Tooltip content="删除" mini>
            <Button size="mini" type="text" status="danger" icon={<IconDelete />}
              onClick={(e) => { e.stopPropagation(); onDelete?.(stageId); }}
            />
          </Tooltip>
        </div>
      </div>
      <Handle type="source" position={Position.Right} style={{ background: '#0057c2', width: 8, height: 8 }} />
    </div>
  );
}

export const StageNode = memo(StageNodeComponent);
