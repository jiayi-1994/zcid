import { type ReactNode, type InputHTMLAttributes } from 'react';

interface InputProps extends InputHTMLAttributes<HTMLInputElement> {
  icon?: ReactNode;
  wrapStyle?: React.CSSProperties;
  wrapClassName?: string;
}

export function Input({ icon, className = '', wrapStyle, wrapClassName = '', ...rest }: InputProps) {
  if (!icon) {
    return <input className={['input', className].filter(Boolean).join(' ')} {...rest} />;
  }
  return (
    <div className={['input-wrap', wrapClassName].filter(Boolean).join(' ')} style={wrapStyle}>
      {icon}
      <input className={['input', 'input--with-icon', className].filter(Boolean).join(' ')} {...rest} />
    </div>
  );
}
