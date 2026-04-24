import { type ReactNode } from 'react';

interface FieldProps {
  label: ReactNode;
  required?: boolean;
  help?: string;
  children: ReactNode;
}

export function Field({ label, required, help, children }: FieldProps) {
  return (
    <div>
      <div className="field-label">
        {label}
        {required && <span className="req">*</span>}
      </div>
      {children}
      {help && <div className="help">{help}</div>}
    </div>
  );
}
