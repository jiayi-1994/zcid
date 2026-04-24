interface SegmentedProps {
  value: string;
  options: (string | { value: string; label: string })[];
  onChange?: (value: string) => void;
}

export function Segmented({ value, options, onChange }: SegmentedProps) {
  return (
    <div className="seg">
      {options.map((o) => {
        const v = typeof o === 'object' ? o.value : o;
        const l = typeof o === 'object' ? o.label : o;
        return (
          <button
            key={v}
            type="button"
            className={v === value ? 'is-on' : ''}
            onClick={() => onChange?.(v)}
          >
            {l}
          </button>
        );
      })}
    </div>
  );
}
