import { type ReactNode } from 'react';

interface EmptyProps {
  icon?: ReactNode;
  title: string;
  sub?: string;
  action?: ReactNode;
}

export function Empty({ icon, title, sub, action }: EmptyProps) {
  return (
    <div style={{ padding: '48px 24px', textAlign: 'center', color: 'var(--z-500)' }}>
      {icon && (
        <div style={{ color: 'var(--z-400)', marginBottom: 12, display: 'flex', justifyContent: 'center' }}>
          {icon}
        </div>
      )}
      <div style={{ fontSize: 14, fontWeight: 500, color: 'var(--z-800)', marginBottom: 4 }}>{title}</div>
      {sub && <div style={{ fontSize: 12.5, marginBottom: 14 }}>{sub}</div>}
      {action}
    </div>
  );
}
