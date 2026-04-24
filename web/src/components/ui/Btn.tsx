import { type ReactNode, type ButtonHTMLAttributes } from 'react';

interface BtnProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: 'outline' | 'ghost' | 'primary' | 'danger';
  size?: 'sm' | 'xs';
  icon?: ReactNode;
  iconOnly?: boolean;
  children?: ReactNode;
}

export function Btn({ variant = 'outline', size, icon, iconOnly, children, className = '', ...rest }: BtnProps) {
  const cls = [
    'btn',
    variant === 'primary' ? 'btn--primary'
      : variant === 'ghost' ? 'btn--ghost'
      : variant === 'danger' ? 'btn--danger'
      : 'btn--outline',
    size === 'sm' ? 'btn--sm' : size === 'xs' ? 'btn--xs' : '',
    iconOnly ? 'btn--icon' : '',
    className,
  ].filter(Boolean).join(' ');

  return (
    <button className={cls} {...rest}>
      {icon}
      {children && <span>{children}</span>}
    </button>
  );
}
