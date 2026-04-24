import { Segmented } from '../ui/Segmented';

export type EditorMode = 'visual' | 'json';

interface ModeSwitchProps {
  mode: EditorMode;
  onChange: (mode: EditorMode) => void;
}

export function ModeSwitch({ mode, onChange }: ModeSwitchProps) {
  return (
    <Segmented
      value={mode}
      options={[{ value: 'visual', label: '可视化' }, { value: 'json', label: 'JSON' }]}
      onChange={(v) => onChange(v as EditorMode)}
    />
  );
}
