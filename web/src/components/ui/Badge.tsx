import { type ReactNode } from 'react';

type Tone = 'green' | 'red' | 'amber' | 'blue' | 'cyan' | 'accent' | 'grey';

interface BadgeProps {
  tone?: Tone;
  dot?: boolean;
  pulse?: boolean;
  children?: ReactNode;
}

export function Badge({ tone = 'grey', dot, pulse, children }: BadgeProps) {
  return (
    <span className={`badge badge--${tone}`}>
      {dot && (
        <span
          className={`st-dot st-dot--${tone === 'grey' ? 'grey' : tone}${pulse ? ' is-pulse' : ''}`}
          style={{ width: 6, height: 6 }}
        />
      )}
      {children}
    </span>
  );
}
