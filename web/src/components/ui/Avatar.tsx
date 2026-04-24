interface AvatarProps {
  name?: string;
  size?: 'sm' | 'md' | 'lg';
  tone?: number;
  round?: boolean;
  style?: React.CSSProperties;
}

export function Avatar({ name = 'U', size = 'md', tone, round, style }: AvatarProps) {
  const letter = (name[0] || 'U').toUpperCase();
  const hue = tone ?? ((name.charCodeAt(0) * 17) % 360);
  const bg = tone === undefined
    ? `linear-gradient(135deg, oklch(0.62 0.17 ${hue}), oklch(0.52 0.19 ${(hue + 24) % 360}))`
    : undefined;

  return (
    <span
      className={['avatar', size !== 'md' ? `avatar--${size}` : '', round ? 'avatar--round' : ''].filter(Boolean).join(' ')}
      style={{ background: bg, ...style }}
    >
      {letter}
    </span>
  );
}
