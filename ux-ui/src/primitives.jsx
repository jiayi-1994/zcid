// Shared primitives used everywhere
const { I } = window;

// ─── Button ───
const Btn = ({ variant = 'outline', size, icon, iconOnly, children, onClick, className = '', style, ...rest }) => {
  const klass = [
    'btn',
    variant === 'primary' ? 'btn--primary' : variant === 'ghost' ? 'btn--ghost' : variant === 'danger' ? 'btn--danger' : 'btn--outline',
    size === 'sm' ? 'btn--sm' : size === 'xs' ? 'btn--xs' : '',
    iconOnly ? 'btn--icon' : '',
    className,
  ].join(' ');
  return (
    <button className={klass} onClick={onClick} style={style} {...rest}>
      {icon}
      {children && <span>{children}</span>}
    </button>
  );
};

// ─── Input ───
const Input = ({ icon, className = '', ...rest }) => {
  if (!icon) return <input className={['input', className].join(' ')} {...rest} />;
  return (
    <div className="input-wrap" style={{ width: rest.style?.width || 'auto' }}>
      {icon}
      <input className={['input', 'input--with-icon', className].join(' ')} {...rest} />
    </div>
  );
};

// ─── Badge ───
const Badge = ({ tone = 'grey', dot, pulse, children }) => (
  <span className={`badge badge--${tone}`}>
    {dot && <span className={`st-dot st-dot--${tone === 'accent' ? 'blue' : tone} ${pulse ? 'is-pulse' : ''}`} style={{ width: 6, height: 6 }} />}
    {children}
  </span>
);

// ─── Avatar ───
const Avatar = ({ name = 'U', size = 'md', tone, round, style }) => {
  const letter = (name[0] || 'U').toUpperCase();
  const hue = tone ?? ((name.charCodeAt(0) * 17) % 360);
  const bg = tone === undefined
    ? `linear-gradient(135deg, oklch(0.62 0.17 ${hue}), oklch(0.52 0.19 ${(hue + 24) % 360}))`
    : undefined;
  return (
    <span
      className={`avatar avatar--${size} ${round ? 'avatar--round' : ''}`}
      style={{ background: bg, ...style }}
    >{letter}</span>
  );
};

// ─── Card ───
const Card = ({ title, extra, padding = true, children, style, className = '' }) => (
  <div className={`card ${className}`} style={style}>
    {(title || extra) && (
      <div className="card-hd">
        {typeof title === 'string' ? <h2>{title}</h2> : title}
        {extra}
      </div>
    )}
    <div style={padding ? { padding: '14px 16px' } : null}>{children}</div>
  </div>
);

// ─── Status badge helpers ───
const StatusBadge = ({ status, labels }) => {
  const map = labels || {
    // pipelines
    active: ['green', '已启用'], draft: ['grey', '草稿'], disabled: ['amber', '已停用'],
    // runs
    pending: ['grey', '待执行'], queued: ['blue', '排队中'], running: ['cyan', '运行中'],
    succeeded: ['green', '成功'], failed: ['red', '失败'], cancelled: ['amber', '已取消'],
    // deployments
    not_deployed: ['grey', '待部署'], syncing: ['blue', '同步中'], healthy: ['green', '健康'],
    degraded: ['amber', '异常'], sync_failed: ['red', '失败'], rolled_back: ['grey', '已回滚'],
    // health
    synced: ['green', 'Healthy'], outOfSync: ['amber', 'OutOfSync'], missing: ['red', 'Missing'],
    // env
    env_healthy: ['green', 'Healthy'], env_syncing: ['blue', 'Syncing'],
    env_degraded: ['amber', 'Degraded'], env_down: ['red', 'Down'],
    // integration
    connected: ['green', 'Connected'], token_expired: ['amber', 'Token Expired'], disconnected: ['red', 'Disconnected'],
    // user
    enabled: ['green', '启用'], userDisabled: ['grey', '禁用'],
    // events
    build_success: ['green', '构建成功'], build_failed: ['red', '构建失败'],
    deploy_success: ['blue', '部署成功'], deploy_failed: ['amber', '部署失败'],
    // audit methods
    GET: ['blue', 'GET'], POST: ['green', 'POST'], PUT: ['amber', 'PUT'], DELETE: ['red', 'DELETE'],
    // audit result
    ok: ['green', '成功'], err: ['red', '失败'],
    // roles
    admin: ['blue', '管理员'], project_admin: ['green', '项目管理员'], member: ['grey', '普通成员'],
  };
  const [tone, label] = map[status] || ['grey', status];
  return <Badge tone={tone} dot>{label}</Badge>;
};

// ─── Sw (switch) ───
const Switch = ({ on, onChange, size = 'md' }) => (
  <button className="sw" data-on={on ? '1' : '0'} onClick={() => onChange && onChange(!on)}><i/></button>
);

// ─── Select (lightweight) ───
const Select = ({ value, options, onChange, style, width = 140 }) => (
  <select className="input" style={{ width, ...style }} value={value} onChange={(e) => onChange && onChange(e.target.value)}>
    {options.map((o) => {
      const v = typeof o === 'object' ? o.value : o;
      const l = typeof o === 'object' ? o.label : o;
      return <option key={v} value={v}>{l}</option>;
    })}
  </select>
);

// ─── Segmented ───
const Segmented = ({ value, options, onChange }) => (
  <div className="seg">
    {options.map((o) => {
      const v = typeof o === 'object' ? o.value : o;
      const l = typeof o === 'object' ? o.label : o;
      return (
        <button key={v} className={v === value ? 'is-on' : ''} onClick={() => onChange && onChange(v)}>{l}</button>
      );
    })}
  </div>
);

// ─── Page header ───
const PageHeader = ({ crumb, title, sub, actions }) => (
  <div className="page-hd">
    <div className="meta">
      <span className="crumb">{crumb}</span>
      <h1>{title}</h1>
      {sub && <p className="sub" style={{ marginTop: 4 }}>{sub}</p>}
    </div>
    {actions && <div className="actions">{actions}</div>}
  </div>
);

// ─── Metric card ───
const Metric = ({ label, value, icon, trend, trendTone = 'grey', active, onClick, iconBg, iconColor }) => (
  <div className={`metric ${active ? 'is-active' : ''} ${onClick ? 'is-clickable' : ''}`} onClick={onClick}>
    <div className="mx">
      <span className="mlabel">{label}</span>
      {icon && <span className="ico" style={{ background: iconBg, color: iconColor }}>{icon}</span>}
    </div>
    <div className="mval">{value}</div>
    {trend && <div className="mtrend" style={{ color: trendTone !== 'grey' ? `var(--${trendTone})` : undefined }}>{trend}</div>}
  </div>
);

// ─── Empty state ───
const Empty = ({ icon, title, sub, action }) => (
  <div style={{ padding: '48px 24px', textAlign: 'center', color: 'var(--z-500)' }}>
    {icon && <div style={{ color: 'var(--z-400)', marginBottom: 12, display: 'flex', justifyContent: 'center' }}>{icon}</div>}
    <div style={{ fontSize: 14, fontWeight: 500, color: 'var(--z-800)', marginBottom: 4 }}>{title}</div>
    {sub && <div style={{ fontSize: 12.5, marginBottom: 14 }}>{sub}</div>}
    {action}
  </div>
);

// ─── Modal (static-only for screens) ───
const Modal = ({ title, children, onClose, footer, width }) => (
  <div className="modal-bg" onClick={onClose}>
    <div className="modal" style={{ width }} onClick={(e) => e.stopPropagation()}>
      <div className="modal-hd">
        <div style={{ fontSize: 14, fontWeight: 600 }}>{title}</div>
        <button className="btn btn--ghost btn--icon btn--sm" onClick={onClose}><I.x /></button>
      </div>
      <div className="modal-bd">{children}</div>
      {footer && <div className="modal-ft">{footer}</div>}
    </div>
  </div>
);

const Field = ({ label, required, help, children }) => (
  <div>
    <div className="field-label">{label}{required && <span className="req">*</span>}</div>
    {children}
    {help && <div className="help">{help}</div>}
  </div>
);

// ─── Utility: format time ago ───
const ago = (mins) => {
  if (mins < 1) return '刚刚';
  if (mins < 60) return `${mins} 分钟前`;
  if (mins < 60 * 24) return `${Math.floor(mins / 60)} 小时前`;
  return `${Math.floor(mins / 1440)} 天前`;
};

// ─── Duration ───
const dur = (secs) => {
  const m = Math.floor(secs / 60), s = secs % 60;
  return m ? `${m}m ${s}s` : `${s}s`;
};

Object.assign(window, {
  Btn, Input, Badge, Avatar, Card, StatusBadge, Switch, Select, Segmented,
  PageHeader, Metric, Empty, Modal, Field, ago, dur,
});
