interface ZSelectProps {
  value?: string;
  options: (string | { value: string; label: string })[];
  onChange?: (value: string) => void;
  style?: React.CSSProperties;
  width?: number;
}

export function ZSelect({ value, options, onChange, style, width = 140 }: ZSelectProps) {
  return (
    <select
      className="input"
      style={{ width, ...style }}
      value={value}
      onChange={(e) => onChange?.(e.target.value)}
    >
      {options.map((o) => {
        const v = typeof o === 'object' ? o.value : o;
        const l = typeof o === 'object' ? o.label : o;
        return <option key={v} value={v}>{l}</option>;
      })}
    </select>
  );
}
