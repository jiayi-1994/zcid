import { type ReactNode } from 'react';
import { Btn } from './Btn';

interface ZModalProps {
  title: string;
  children: ReactNode;
  onClose?: () => void;
  footer?: ReactNode;
  width?: number | string;
}

export function ZModal({ title, children, onClose, footer, width }: ZModalProps) {
  return (
    <div className="modal-bg" onClick={onClose}>
      <div className="modal" style={{ width }} onClick={(e) => e.stopPropagation()}>
        <div className="modal-hd">
          <div style={{ fontSize: 14, fontWeight: 600 }}>{title}</div>
          <Btn variant="ghost" size="sm" iconOnly onClick={onClose} aria-label="close">
            <svg width="14" height="14" viewBox="0 0 14 14" fill="none">
              <path d="M1 1l12 12M13 1L1 13" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" />
            </svg>
          </Btn>
        </div>
        <div className="modal-bd">{children}</div>
        {footer && <div className="modal-ft">{footer}</div>}
      </div>
    </div>
  );
}
