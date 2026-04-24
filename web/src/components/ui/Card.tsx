import { type ReactNode } from 'react';

interface CardProps {
  title?: ReactNode;
  extra?: ReactNode;
  padding?: boolean;
  children?: ReactNode;
  style?: React.CSSProperties;
  className?: string;
}

export function Card({ title, extra, padding = true, children, style, className = '' }: CardProps) {
  return (
    <div className={['card', className].filter(Boolean).join(' ')} style={style}>
      {(title || extra) && (
        <div className="card-hd">
          {typeof title === 'string' ? <h2>{title}</h2> : title}
          {extra}
        </div>
      )}
      <div style={padding ? { padding: '14px 16px' } : undefined}>{children}</div>
    </div>
  );
}
