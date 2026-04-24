// Shell: AppLayout (main sidebar + topbar) + ProjectLayout (nested project nav)
const { I, Avatar, Btn } = window;

// Primary nav items — clicking dispatches routeChange for the demo
const NAV_WORKSPACE = [
  { id: 'dashboard',    label: 'Dashboard',     icon: <I.grid /> },
  { id: 'projects',     label: '项目管理',       icon: <I.folder /> },
];
const NAV_SYSTEM = [
  { id: 'users',        label: '用户管理',       icon: <I.users /> },
  { id: 'globalvars',   label: '全局变量',       icon: <I.key /> },
  { id: 'integrations', label: '集成管理',       icon: <I.plug /> },
  { id: 'audit',        label: '审计日志',       icon: <I.shield /> },
  { id: 'systemset',    label: '系统设置',       icon: <I.settings /> },
];

const PROJECT_NAV = [
  { id: 'pipelines',     label: '流水线',     icon: <I.zap /> },
  { id: 'environments',  label: '环境',       icon: <I.layers /> },
  { id: 'deployments',   label: '部署',       icon: <I.rocket /> },
  { id: 'services',      label: '服务',       icon: <I.server /> },
  { id: 'members',       label: '成员',       icon: <I.users /> },
  { id: 'variables',     label: '变量',       icon: <I.key /> },
  { id: 'notifications', label: '通知',       icon: <I.bell /> },
];

function NavSidebar({ active, onNavigate }) {
  return (
    <div className="nav">
      <div className="nav-hd">
        <div className="logo">Z</div>
        <div style={{ display: 'flex', flexDirection: 'column', gap: 1 }}>
          <span style={{ fontSize: 14, fontWeight: 600, letterSpacing: -0.01 }}>zcid</span>
          <span style={{ fontSize: 10.5, color: 'var(--z-500)' }}>Cloud-Native CI/CD</span>
        </div>
      </div>
      <div style={{ padding: '4px 10px 6px' }}>
        <div className="input-wrap">
          <I.search size={13} />
          <input className="input input--with-icon" style={{ height: 28, fontSize: 12 }} placeholder="搜索... " />
        </div>
      </div>
      <div className="nav-group">
        <div className="nav-title">Workspace</div>
        {NAV_WORKSPACE.map((it) => (
          <a key={it.id} className={`nav-item ${active === it.id ? 'is-active' : ''}`}
             onClick={() => onNavigate && onNavigate(it.id)}>
            {it.icon}<span>{it.label}</span>
          </a>
        ))}
      </div>
      <div className="nav-group">
        <div className="nav-title">System</div>
        {NAV_SYSTEM.map((it) => (
          <a key={it.id} className={`nav-item ${active === it.id ? 'is-active' : ''}`}
             onClick={() => onNavigate && onNavigate(it.id)}>
            {it.icon}<span>{it.label}</span>
          </a>
        ))}
      </div>
      <div className="nav-ft">
        <div className="nav-user">
          <Avatar name="admin" size="sm" round />
          <div className="who">
            <b>admin</b>
            <small>管理员 · zcid.local</small>
          </div>
          <I.chevD size={13} />
        </div>
      </div>
    </div>
  );
}

function Topbar({ crumbs = [] }) {
  return (
    <div className="top">
      <div className="crumb-path">
        <I.home size={13} />
        <span>zcid</span>
        {crumbs.map((c, i) => (
          <React.Fragment key={i}>
            <I.chevR size={11} style={{ opacity: 0.5 }} />
            {i === crumbs.length - 1 ? <b>{c}</b> : <span>{c}</span>}
          </React.Fragment>
        ))}
      </div>
      <div style={{ flex: 1 }} />
      <span style={{ fontSize: 11.5, color: 'var(--z-500)', display: 'inline-flex', alignItems: 'center', gap: 6 }}>
        <span className="st-dot st-dot--green" /> all systems operational
      </span>
      <span className="kbd">⌘ K</span>
    </div>
  );
}

function AppShell({ active, onNavigate, crumbs, children, topbarRight }) {
  return (
    <div className="zc">
      <NavSidebar active={active} onNavigate={onNavigate} />
      <div className="main">
        <Topbar crumbs={crumbs} />
        <div className="scroll">{children}</div>
      </div>
    </div>
  );
}

// Project layout: app nav + project sub-nav
function ProjectShell({ appActive = 'projects', projectActive, project, onNavigate, onProjectNavigate, crumbs, children }) {
  return (
    <div className="zc">
      <NavSidebar active={appActive} onNavigate={onNavigate} />
      <div className="subnav">
        <div className="subnav-hd">
          <Avatar name={project.name} size="sm" tone={project.hue} />
          <div style={{ minWidth: 0 }}>
            <div style={{ fontSize: 13, fontWeight: 600, color: 'var(--z-900)', overflow: 'hidden', textOverflow: 'ellipsis' }}>{project.name}</div>
            <div style={{ fontSize: 11, color: 'var(--z-500)', display: 'flex', alignItems: 'center', gap: 4 }}>
              <span className="st-dot st-dot--green" style={{ width: 6, height: 6 }} /> active
            </div>
          </div>
        </div>
        <div className="nav-group" style={{ padding: 8 }}>
          {PROJECT_NAV.map((it) => (
            <a key={it.id} className={`nav-item ${projectActive === it.id ? 'is-active' : ''}`}
               onClick={() => onProjectNavigate && onProjectNavigate(it.id)}>
              {it.icon}<span>{it.label}</span>
            </a>
          ))}
        </div>
        <div style={{ marginTop: 'auto', padding: 12, borderTop: '1px solid var(--z-150)' }}>
          <a className="nav-item" onClick={() => onNavigate && onNavigate('projects')} style={{ fontSize: 11.5, color: 'var(--z-500)' }}>
            <I.arrL size={12} /><span>所有项目</span>
          </a>
        </div>
      </div>
      <div className="main">
        <Topbar crumbs={[project.name, ...(crumbs || [])]} />
        <div className="scroll">{children}</div>
      </div>
    </div>
  );
}

Object.assign(window, { NavSidebar, Topbar, AppShell, ProjectShell, NAV_WORKSPACE, NAV_SYSTEM, PROJECT_NAV });
