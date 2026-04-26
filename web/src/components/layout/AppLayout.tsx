import { type ReactNode, useState } from 'react';
import { useLocation, useNavigate } from 'react-router-dom';
import { Message } from '@arco-design/web-react';
import { logout } from '../../services/auth';
import { type SystemRole, useAuthStore } from '../../stores/auth';
import {
  IGrid, IFolder, IUsers, IKey, IPlug, IShield, ISettings,
  ISearch, IChevD, IChevR, IHome,
} from '../ui/icons';
import { Avatar } from '../ui/Avatar';

const ROLE_LABELS: Record<SystemRole, string> = {
  admin: '管理员',
  project_admin: '项目管理员',
  member: '普通成员',
};

interface AppLayoutProps {
  children?: ReactNode;
}

interface NavItemProps {
  icon: ReactNode;
  label: string;
  path: string;
  active: boolean;
  onClick: (path: string) => void;
}

function NavItem({ icon, label, path, active, onClick }: NavItemProps) {
  return (
    <button
      type="button"
      className={`nav-item${active ? ' is-active' : ''}`}
      onClick={() => onClick(path)}
    >
      {icon}
      <span>{label}</span>
    </button>
  );
}

interface TopbarProps {
  crumbs?: string[];
}

function Topbar({ crumbs = [] }: TopbarProps) {
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

export function AppLayout({ children }: AppLayoutProps) {
  const navigate = useNavigate();
  const location = useLocation();
  const canViewDashboard        = useAuthStore((s) => s.hasPermission('route:dashboard:view'));
  const canViewAdminUsers       = useAuthStore((s) => s.hasPermission('route:admin-users:view'));
  const canViewAdminVariables   = useAuthStore((s) => s.hasPermission('route:admin-variables:view'));
  const canViewAdminIntegrations = useAuthStore((s) => s.hasPermission('route:admin-integrations:view'));
  const canViewAuditLogs        = useAuthStore((s) => s.hasPermission('route:admin-audit:view'));
  const canViewAccessTokens     = useAuthStore((s) => s.hasPermission('route:access-tokens:view'));
  const canViewSystemSettings   = useAuthStore((s) => s.hasPermission('route:admin-settings:view'));
  const user = useAuthStore((s) => s.user);
  const refreshToken = useAuthStore((s) => s.refreshToken);
  const clearSession = useAuthStore((s) => s.clearSession);

  const [userMenuOpen, setUserMenuOpen] = useState(false);

  const roleLabel = user ? ROLE_LABELS[user.role] : '';

  const hasAdminSection = canViewAdminUsers || canViewAdminVariables || canViewAdminIntegrations || canViewAuditLogs || canViewAccessTokens || canViewSystemSettings;

  const isActive = (path: string) =>
    location.pathname === path || location.pathname.startsWith(path + '/');

  const handleLogout = async () => {
    try { if (refreshToken) await logout(refreshToken); } catch { /* ignore */ }
    finally { clearSession(); navigate('/login', { replace: true }); }
  };

  const crumbs = getBreadcrumbs(location.pathname);

  return (
    <div className="zc">
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
              <NavItem icon={<IGrid size={14} />} label="Dashboard" path="/dashboard" active={isActive('/dashboard')} onClick={navigate} />
            )}
            <NavItem icon={<IFolder size={14} />} label="项目管理" path="/projects" active={isActive('/projects')} onClick={navigate} />
          </div>

          {hasAdminSection && (
            <div className="nav-group">
              <div className="nav-title">System</div>
              {canViewAdminUsers && (
                <NavItem icon={<IUsers size={14} />} label="用户管理" path="/admin/users" active={isActive('/admin/users')} onClick={navigate} />
              )}
              {canViewAdminVariables && (
                <NavItem icon={<IKey size={14} />} label="全局变量" path="/admin/variables" active={isActive('/admin/variables')} onClick={navigate} />
              )}
              {canViewAdminIntegrations && (
                <NavItem icon={<IPlug size={14} />} label="集成管理" path="/admin/integrations" active={isActive('/admin/integrations')} onClick={navigate} />
              )}
              {canViewAuditLogs && (
                <NavItem icon={<IShield size={14} />} label="审计日志" path="/admin/audit-logs" active={isActive('/admin/audit-logs')} onClick={navigate} />
              )}
              {canViewAccessTokens && (
                <NavItem icon={<IKey size={14} />} label="访问令牌" path="/admin/access-tokens" active={isActive('/admin/access-tokens')} onClick={navigate} />
              )}
              {canViewSystemSettings && (
                <NavItem icon={<ISettings size={14} />} label="系统设置" path="/admin/settings" active={isActive('/admin/settings')} onClick={navigate} />
              )}
            </div>
          )}
        </div>

        <div className="nav-ft">
          <div style={{ position: 'relative' }}>
            <button type="button" className="nav-user user-entry" onClick={() => setUserMenuOpen(!userMenuOpen)}>
              <Avatar name={user?.username ?? 'U'} size="sm" round />
              <div className="who">
                <b>{user?.username}</b>
                <small>{roleLabel}</small>
              </div>
              <IChevD size={13} style={{ color: 'var(--z-500)', flex: 'none' }} />
            </button>
            {userMenuOpen && (
              <div style={{
                position: 'absolute', bottom: '100%', left: 0, right: 0, marginBottom: 4,
                background: 'var(--z-0)', border: '1px solid var(--z-200)', borderRadius: 8,
                boxShadow: 'var(--shadow-md)', overflow: 'hidden', zIndex: 100,
              }}>
                <button
                  type="button"
                  className="nav-item"
                  role="menuitem"
                  style={{ color: 'var(--red)' }}
                  onClick={() => { setUserMenuOpen(false); void handleLogout(); }}
                >
                  退出登录
                </button>
              </div>
            )}
          </div>
        </div>
      </div>

      <div className="main">
        <Topbar crumbs={crumbs} />
        <div className="scroll">{children}</div>
      </div>
    </div>
  );
}

function getBreadcrumbs(pathname: string): string[] {
  if (pathname.startsWith('/dashboard')) return ['Dashboard'];
  if (pathname.startsWith('/projects')) return ['项目管理'];
  if (pathname.startsWith('/admin/users')) return ['用户管理'];
  if (pathname.startsWith('/admin/variables')) return ['全局变量'];
  if (pathname.startsWith('/admin/integrations')) return ['集成管理'];
  if (pathname.startsWith('/admin/audit-logs')) return ['审计日志'];
  if (pathname.startsWith('/admin/access-tokens')) return ['访问令牌'];
  if (pathname.startsWith('/admin/settings')) return ['系统设置'];
  return [];
}
