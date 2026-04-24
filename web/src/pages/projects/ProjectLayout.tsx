import { useEffect, useState, type ReactNode } from 'react';
import { Outlet, useLocation, useNavigate, useParams } from 'react-router-dom';
import { useAuthStore } from '../../stores/auth';
import { fetchProject, type Project } from '../../services/project';
import {
  IZap, ILayers, IRocket, IServer, IUsers, IKey, IBell, IChevD, IChevR, IHome, ISearch, IGrid, IFolder, IPlug, IShield, ISettings,
} from '../../components/ui/icons';
import { Avatar } from '../../components/ui/Avatar';
import { logout } from '../../services/auth';
import { type SystemRole } from '../../stores/auth';

const ROLE_LABELS: Record<SystemRole, string> = {
  admin: '管理员',
  project_admin: '项目管理员',
  member: '普通成员',
};

const PROJECT_NAV = [
  { id: 'pipelines',    label: '流水线', icon: <IZap size={14} /> },
  { id: 'environments', label: '环境',   icon: <ILayers size={14} /> },
  { id: 'deployments',  label: '部署',   icon: <IRocket size={14} /> },
  { id: 'services',     label: '服务',   icon: <IServer size={14} /> },
  { id: 'members',      label: '成员',   icon: <IUsers size={14} /> },
  { id: 'variables',    label: '变量',   icon: <IKey size={14} /> },
  { id: 'notifications',label: '通知',   icon: <IBell size={14} /> },
];

function NavItem({ icon, label, path, active, onClick }: { icon: ReactNode; label: string; path: string; active: boolean; onClick: (p: string) => void }) {
  return (
    <button type="button" className={`nav-item${active ? ' is-active' : ''}`} onClick={() => onClick(path)}>
      {icon}<span>{label}</span>
    </button>
  );
}

function Topbar({ crumbs = [] }: { crumbs?: string[] }) {
  return (
    <div className="top">
      <div className="crumb-path">
        <IHome size={13} />
        <span>zcid</span>
        {crumbs.map((c, i) => (
          <span key={i} style={{ display: 'contents' }}>
            <IChevR size={11} style={{ opacity: 0.5 }} />
            {i === crumbs.length - 1 ? <b>{c}</b> : <span>{c}</span>}
          </span>
        ))}
      </div>
      <div style={{ flex: 1 }} />
      <span style={{ fontSize: 11.5, color: 'var(--z-500)', display: 'inline-flex', alignItems: 'center', gap: 6 }}>
        <span className="st-dot st-dot--green" />
        all systems operational
      </span>
      <span className="kbd">⌘ K</span>
    </div>
  );
}

export function ProjectLayout() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const location = useLocation();
  const [project, setProject] = useState<Project | null>(null);
  const [userMenuOpen, setUserMenuOpen] = useState(false);

  const canViewDashboard         = useAuthStore((s) => s.hasPermission('route:dashboard:view'));
  const canViewAdminUsers        = useAuthStore((s) => s.hasPermission('route:admin-users:view'));
  const canViewAdminVariables    = useAuthStore((s) => s.hasPermission('route:admin-variables:view'));
  const canViewAdminIntegrations = useAuthStore((s) => s.hasPermission('route:admin-integrations:view'));
  const canViewAuditLogs         = useAuthStore((s) => s.hasPermission('route:admin-audit:view'));
  const canViewSystemSettings    = useAuthStore((s) => s.hasPermission('route:admin-settings:view'));
  const user = useAuthStore((s) => s.user);
  const refreshToken = useAuthStore((s) => s.refreshToken);
  const clearSession = useAuthStore((s) => s.clearSession);

  const hasAdminSection = canViewAdminUsers || canViewAdminVariables || canViewAdminIntegrations || canViewAuditLogs || canViewSystemSettings;
  const roleLabel = user ? ROLE_LABELS[user.role] : '';

  useEffect(() => {
    if (id) fetchProject(id).then(setProject).catch(() => setProject(null));
  }, [id]);

  const basePath = `/projects/${id}`;
  const currentKey = location.pathname.replace(basePath, '').split('/')[1] || 'environments';

  const isGlobalActive = (path: string) =>
    location.pathname === path || location.pathname.startsWith(path + '/');

  const handleLogout = async () => {
    try { if (refreshToken) await logout(refreshToken); } catch { /* ignore */ }
    finally { clearSession(); navigate('/login', { replace: true }); }
  };

  const projectCrumb = currentKey
    ? PROJECT_NAV.find((n) => n.id === currentKey)?.label ?? currentKey
    : '';

  return (
    <div className="zc">
      {/* Main app nav */}
      <div className="nav">
        <div className="nav-hd">
          <div className="logo">Z</div>
          <div style={{ display: 'flex', flexDirection: 'column', gap: 1 }}>
            <span style={{ fontSize: 14, fontWeight: 600, letterSpacing: -0.01 }}>zcid</span>
            <span style={{ fontSize: 10.5, color: 'var(--z-500)' }}>Cloud-Native CI/CD</span>
          </div>
        </div>
        <div style={{ padding: '4px 10px 6px' }}>
          <div className="input-wrap" style={{ width: '100%' }}>
            <ISearch size={13} />
            <input className="input input--with-icon" style={{ height: 28, fontSize: 12 }} placeholder="搜索..." />
          </div>
        </div>
        <div style={{ flex: 1, overflow: 'auto' }}>
          <div className="nav-group">
            <div className="nav-title">Workspace</div>
            {canViewDashboard && (
              <NavItem icon={<IGrid size={14} />} label="Dashboard" path="/dashboard" active={isGlobalActive('/dashboard')} onClick={navigate} />
            )}
            <NavItem icon={<IFolder size={14} />} label="项目管理" path="/projects" active={isGlobalActive('/projects')} onClick={navigate} />
          </div>
          {hasAdminSection && (
            <div className="nav-group">
              <div className="nav-title">System</div>
              {canViewAdminUsers && <NavItem icon={<IUsers size={14} />} label="用户管理" path="/admin/users" active={isGlobalActive('/admin/users')} onClick={navigate} />}
              {canViewAdminVariables && <NavItem icon={<IKey size={14} />} label="全局变量" path="/admin/variables" active={isGlobalActive('/admin/variables')} onClick={navigate} />}
              {canViewAdminIntegrations && <NavItem icon={<IPlug size={14} />} label="集成管理" path="/admin/integrations" active={isGlobalActive('/admin/integrations')} onClick={navigate} />}
              {canViewAuditLogs && <NavItem icon={<IShield size={14} />} label="审计日志" path="/admin/audit-logs" active={isGlobalActive('/admin/audit-logs')} onClick={navigate} />}
              {canViewSystemSettings && <NavItem icon={<ISettings size={14} />} label="系统设置" path="/admin/settings" active={isGlobalActive('/admin/settings')} onClick={navigate} />}
            </div>
          )}
        </div>
        <div className="nav-ft">
          <div style={{ position: 'relative' }}>
            <button type="button" className="nav-user" onClick={() => setUserMenuOpen(!userMenuOpen)}>
              <Avatar name={user?.username ?? 'U'} size="sm" round />
              <div className="who"><b>{user?.username}</b><small>{roleLabel}</small></div>
              <IChevD size={13} style={{ color: 'var(--z-500)', flex: 'none' }} />
            </button>
            {userMenuOpen && (
              <div style={{ position: 'absolute', bottom: '100%', left: 0, right: 0, marginBottom: 4, background: 'var(--z-0)', border: '1px solid var(--z-200)', borderRadius: 8, boxShadow: 'var(--shadow-md)', overflow: 'hidden', zIndex: 100 }}>
                <button type="button" className="nav-item" style={{ color: 'var(--red)' }} onClick={() => { setUserMenuOpen(false); void handleLogout(); }}>退出登录</button>
              </div>
            )}
          </div>
        </div>
      </div>

      {/* Project sub-nav */}
      <div className="subnav">
        <div className="subnav-hd">
          {project ? (
            <>
              <Avatar name={project.name} size="sm" />
              <div style={{ minWidth: 0 }}>
                <div style={{ fontSize: 13, fontWeight: 600, color: 'var(--z-900)', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>{project.name}</div>
                <div style={{ fontSize: 11, color: 'var(--z-500)', display: 'flex', alignItems: 'center', gap: 4 }}>
                  <span className="st-dot st-dot--green" style={{ width: 6, height: 6 }} /> active
                </div>
              </div>
            </>
          ) : (
            <div style={{ fontSize: 13, color: 'var(--z-400)' }}>加载中...</div>
          )}
        </div>
        <div className="nav-group" style={{ flex: 1 }}>
          {PROJECT_NAV.map((item) => (
            <NavItem
              key={item.id}
              icon={item.icon}
              label={item.label}
              path={`${basePath}/${item.id}`}
              active={currentKey === item.id}
              onClick={navigate}
            />
          ))}
        </div>
        <div style={{ padding: '8px 8px', borderTop: '1px solid var(--z-150)' }}>
          <button type="button" className="nav-item" style={{ fontSize: 11.5, color: 'var(--z-500)' }} onClick={() => navigate('/projects')}>
            <svg width="12" height="12" viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
              <path d="M10 8H2M5 5L2 8l3 3" />
            </svg>
            <span>所有项目</span>
          </button>
        </div>
      </div>

      {/* Main content */}
      <div className="main">
        <Topbar crumbs={project ? [project.name, projectCrumb].filter(Boolean) : [projectCrumb].filter(Boolean)} />
        <div className="scroll">
          <Outlet />
        </div>
      </div>
    </div>
  );
}
