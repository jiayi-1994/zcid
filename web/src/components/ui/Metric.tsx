import { type ReactNode } from 'react';

interface MetricProps {
  label: string;
  value: ReactNode;
  icon?: ReactNode;
  trend?: ReactNode;
  trendTone?: string;
  active?: boolean;
  onClick?: () => void;
  iconBg?: string;
  iconColor?: string;
}

export function Metric({ label, value, icon, trend, trendTone = 'grey', active, onClick, iconBg, iconColor }: MetricProps) {
  return (
    <div
      className={['metric', active ? 'is-active' : '', onClick ? 'is-clickable' : ''].filter(Boolean).join(' ')}
      onClick={onClick}
    >
      <div className="mx">
        <span className="mlabel">{label}</span>
        {icon && (
          <span className="ico" style={{ background: iconBg, color: iconColor }}>{icon}</span>
        )}
      </div>
      <div className="mval">{value}</div>
      {trend && (
        <div className="mtrend" style={trendTone !== 'grey' ? { color: `var(--${trendTone})` } : undefined}>
          {trend}
        </div>
      )}
    </div>
  );
}
