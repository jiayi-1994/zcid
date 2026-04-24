interface ZSwitchProps {
  on: boolean;
  onChange?: (value: boolean) => void;
}

export function ZSwitch({ on, onChange }: ZSwitchProps) {
  return (
    <button
      className="sw"
      data-on={on ? '1' : '0'}
      onClick={() => onChange?.(!on)}
      type="button"
    >
      <i />
    </button>
  );
}
