import { type ReactNode } from 'react';

interface PageHeaderProps {
  crumb?: string;
  title: ReactNode;
  sub?: ReactNode;
  actions?: ReactNode;
}

export function PageHeader({ crumb, title, sub, actions }: PageHeaderProps) {
  return (
    <div className="page-hd">
      <div className="meta">
        {crumb && <span className="crumb">{crumb}</span>}
        <h1>{title}</h1>
        {sub && <p className="sub" style={{ marginTop: 4 }}>{sub}</p>}
      </div>
      {actions && <div className="actions">{actions}</div>}
    </div>
  );
}
