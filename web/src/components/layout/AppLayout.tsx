import { Dropdown, Layout, Menu } from '@arco-design/web-react';
import { IconDashboard, IconDown, IconUser, IconApps, IconLock, IconLink, IconFile, IconSettings } from '@arco-design/web-react/icon';
import { type ReactNode } from 'react';
import { useLocation, useNavigate } from 'react-router-dom';
import { logout } from '../../services/auth';
import { type SystemRole, useAuthStore } from '../../stores/auth';

const { Header, Content } = Layout;
const MenuItem = Menu.Item;

const ROLE_LABELS: Record<SystemRole, string> = {
  admin: '管理员',
  project_admin: '项目管理员',
  member: '普通成员',
};

interface AppLayoutProps {
  children?: ReactNode;
}

function NavItem({ icon, label, path, active, onClick }: {
  icon: ReactNode; label: string; path: string; active: boolean; onClick: (path: string) => void;
}) {
  return (
    <div
      className={`nav-item ${active ? 'nav-item-active' : ''}`}
      onClick={() => onClick(path)}
      role="button"
      tabIndex={0}
      onKeyDown={(e) => { if (e.key === 'Enter') onClick(path); }}
    >
      <span className="nav-item-icon">{icon}</span>
      <span className="nav-item-label">{label}</span>
    </div>
  );
}

export function AppLayout({ children }: AppLayoutProps) {
  const navigate = useNavigate();
  const location = useLocation();
  const canViewDashboard = useAuthStore((state) => state.hasPermission('route:dashboard:view'));
  const canViewAdminUsers = useAuthStore((state) => state.hasPermission('route:admin-users:view'));
  const canViewAdminVariables = useAuthStore((state) => state.hasPermission('route:admin-variables:view'));
  const canViewAdminIntegrations = useAuthStore((state) => state.hasPermission('route:admin-integrations:view'));
  const canViewAuditLogs = useAuthStore((state) => state.hasPermission('route:admin-audit:view'));
  const canViewSystemSettings = useAuthStore((state) => state.hasPermission('route:admin-settings:view'));
  const user = useAuthStore((state) => state.user);
  const refreshToken = useAuthStore((state) => state.refreshToken);
  const clearSession = useAuthStore((state) => state.clearSession);

  const roleLabel = user ? ROLE_LABELS[user.role] : '';
  const userInitial = user?.username?.charAt(0).toUpperCase() || 'U';

  const hasAdminSection = canViewAdminUsers || canViewAdminVariables || canViewAdminIntegrations || canViewAuditLogs || canViewSystemSettings;

  const handleLogout = async () => {
    try { if (refreshToken) await logout(refreshToken); }
    catch { /* ignore */ }
    finally { clearSession(); navigate('/login', { replace: true }); }
  };

  const isActive = (path: string) => location.pathname === path || location.pathname.startsWith(path + '/');

  return (
    <div className="app-root" style={{ display: 'flex', minHeight: '100vh' }}>
      {/* Custom sidebar */}
      <div className="app-sider" style={{ width: 220, flexShrink: 0, display: 'flex', flexDirection: 'column' }}>
        <div className="sider-logo">
          <div className="sider-logo-icon">Z</div>
          <span className="sider-logo-text">zcid</span>
        </div>

        <div style={{ flex: 1, overflow: 'auto', padding: '8px 12px' }}>
          <div className="sider-section-label">工作台</div>
          {canViewDashboard && <NavItem icon={<IconDashboard />} label="Dashboard" path="/dashboard" active={isActive('/dashboard')} onClick={navigate} />}
          <NavItem icon={<IconApps />} label="项目管理" path="/projects" active={isActive('/projects')} onClick={navigate} />

          {hasAdminSection && (
            <>
              <div className="sider-section-label" style={{ marginTop: 12 }}>系统管理</div>
              {canViewAdminUsers && <NavItem icon={<IconUser />} label="用户管理" path="/admin/users" active={isActive('/admin/users')} onClick={navigate} />}
              {canViewAdminVariables && <NavItem icon={<IconLock />} label="全局变量" path="/admin/variables" active={isActive('/admin/variables')} onClick={navigate} />}
              {canViewAdminIntegrations && <NavItem icon={<IconLink />} label="集成管理" path="/admin/integrations" active={isActive('/admin/integrations')} onClick={navigate} />}
              {canViewAuditLogs && <NavItem icon={<IconFile />} label="审计日志" path="/admin/audit-logs" active={isActive('/admin/audit-logs')} onClick={navigate} />}
              {canViewSystemSettings && <NavItem icon={<IconSettings />} label="系统设置" path="/admin/settings" active={isActive('/admin/settings')} onClick={navigate} />}
            </>
          )}
        </div>

        {/* Bottom user section */}
        <div style={{ padding: '12px', borderTop: '1px solid var(--border)', flexShrink: 0 }}>
          <Dropdown
            trigger="click"
            position="tr"
            droplist={(
              <Menu onClickMenuItem={(key) => key === 'logout' && void handleLogout()}>
                <MenuItem key="logout">退出登录</MenuItem>
              </Menu>
            )}
          >
            <div className="user-entry" style={{
              display: 'flex', alignItems: 'center', gap: 10,
              padding: '8px 10px', borderRadius: 8,
              cursor: 'pointer', transition: 'background 0.15s',
              color: 'var(--foreground)',
            }}
              onMouseEnter={(e) => { (e.currentTarget as HTMLElement).style.background = 'var(--accent)'; }}
              onMouseLeave={(e) => { (e.currentTarget as HTMLElement).style.background = 'transparent'; }}
            >
              <div style={{
                width: 28, height: 28, borderRadius: 6,
                background: '#18181B',
                color: '#fff', display: 'flex', alignItems: 'center', justifyContent: 'center',
                fontSize: 12, fontWeight: 600, flexShrink: 0,
              }}>
                {userInitial}
              </div>
              <div style={{ flex: 1, minWidth: 0 }}>
                <div style={{ fontSize: 13, fontWeight: 500, color: 'var(--foreground)', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                  {user?.username}
                </div>
                <div style={{ fontSize: 11, color: 'var(--muted-foreground)' }}>{roleLabel}</div>
              </div>
              <IconDown style={{ fontSize: 10, color: 'var(--muted-foreground)' }} />
            </div>
          </Dropdown>
        </div>
      </div>

      {/* Main content area */}
      <div style={{ flex: 1, display: 'flex', flexDirection: 'column', minWidth: 0 }}>
        <Header className="app-header">
          <div className="app-header-inner">
            <span className="app-header-breadcrumb">zcid</span>
          </div>
        </Header>
        <Content className="app-content">{children}</Content>
      </div>
    </div>
  );
}
