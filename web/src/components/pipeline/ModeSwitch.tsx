import { Radio } from '@arco-design/web-react';

export type EditorMode = 'visual' | 'json';

interface ModeSwitchProps {
  mode: EditorMode;
  onChange: (mode: EditorMode) => void;
}

const options = [
  { label: '可视化', value: 'visual' as EditorMode },
  { label: 'JSON 模式', value: 'json' as EditorMode },
];

export function ModeSwitch({ mode, onChange }: ModeSwitchProps) {
  return (
    <Radio.Group
      type="button"
      value={mode}
      onChange={(v) => onChange(v as EditorMode)}
      options={options}
    />
  );
}
