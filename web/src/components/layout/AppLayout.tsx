import { Dropdown, Layout, Menu } from '@arco-design/web-react';
import { IconDashboard, IconDown, IconUser, IconApps, IconLock, IconLink, IconFile, IconSettings } from '@arco-design/web-react/icon';
import { type ReactNode } from 'react';
import { useLocation, useNavigate } from 'react-router-dom';
import { useSidebarCollapsed } from '../../hooks/useSidebarCollapsed';
import { logout } from '../../services/auth';
import { type SystemRole, useAuthStore } from '../../stores/auth';

const { Sider, Header, Content } = Layout;
const MenuItem = Menu.Item;

const ROLE_LABELS: Record<SystemRole, string> = {
  admin: '管理员',
  project_admin: '项目管理员',
  member: '普通成员',
};

interface AppLayoutProps {
  children?: ReactNode;
}

export function AppLayout({ children }: AppLayoutProps) {
  const collapsed = useSidebarCollapsed();
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

  return (
    <Layout className="app-root">
      <Sider className="app-sider" width={collapsed ? 60 : 232} collapsed={collapsed}>
        <div className="sider-logo">
          <div className="sider-logo-icon">Z</div>
          {!collapsed && <span className="sider-logo-text">zcid</span>}
        </div>
        <Menu selectedKeys={[location.pathname]} onClickMenuItem={(key) => navigate(key)}>
          {!collapsed && <div className="sider-section-label">工作台</div>}
          {canViewDashboard && (
            <MenuItem key="/dashboard"><IconDashboard />{!collapsed && 'Dashboard'}</MenuItem>
          )}
          <MenuItem key="/projects"><IconApps />{!collapsed && '项目管理'}</MenuItem>

          {hasAdminSection && !collapsed && <div className="sider-section-label" style={{ marginTop: 8 }}>系统管理</div>}
          {canViewAdminUsers && (
            <MenuItem key="/admin/users"><IconUser />{!collapsed && '用户管理'}</MenuItem>
          )}
          {canViewAdminVariables && (
            <MenuItem key="/admin/variables"><IconLock />{!collapsed && '全局变量'}</MenuItem>
          )}
          {canViewAdminIntegrations && (
            <MenuItem key="/admin/integrations"><IconLink />{!collapsed && '集成管理'}</MenuItem>
          )}
          {canViewAuditLogs && (
            <MenuItem key="/admin/audit-logs"><IconFile />{!collapsed && '审计日志'}</MenuItem>
          )}
          {canViewSystemSettings && (
            <MenuItem key="/admin/settings"><IconSettings />{!collapsed && '系统设置'}</MenuItem>
          )}
        </Menu>

        {/* Bottom user section */}
        {!collapsed && (
          <div style={{ padding: '12px 16px', borderTop: '1px solid rgba(255,255,255,0.08)', flexShrink: 0 }}>
            <Dropdown
              trigger="click"
              droplist={(
                <Menu onClickMenuItem={(key) => key === 'logout' && void handleLogout()}>
                  <MenuItem key="logout">退出登录</MenuItem>
                </Menu>
              )}
            >
              <div style={{
                display: 'flex', alignItems: 'center', gap: 10,
                padding: '8px 10px', borderRadius: 8,
                cursor: 'pointer', transition: 'background 0.15s',
              }}
                onMouseEnter={(e) => { (e.currentTarget as HTMLElement).style.background = 'rgba(255,255,255,0.06)'; }}
                onMouseLeave={(e) => { (e.currentTarget as HTMLElement).style.background = 'transparent'; }}
              >
                <div style={{
                  width: 32, height: 32, borderRadius: 8,
                  background: 'linear-gradient(135deg, #3B82F6 0%, #06B6D4 100%)',
                  color: '#fff', display: 'flex', alignItems: 'center', justifyContent: 'center',
                  fontSize: 13, fontWeight: 700, flexShrink: 0,
                }}>
                  {userInitial}
                </div>
                <div style={{ flex: 1, minWidth: 0 }}>
                  <div style={{ fontSize: 13, fontWeight: 600, color: 'rgba(255,255,255,0.9)', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                    {user?.username}
                  </div>
                  <div style={{ fontSize: 11, color: 'rgba(255,255,255,0.4)' }}>{roleLabel}</div>
                </div>
                <IconDown style={{ fontSize: 10, color: 'rgba(255,255,255,0.3)' }} />
              </div>
            </Dropdown>
          </div>
        )}
      </Sider>
      <Layout>
        <Header className="app-header">
          <div className="app-header-inner">
            <span className="app-header-breadcrumb">zcid</span>
            {collapsed && (
              <Dropdown
                trigger="click"
                droplist={(
                  <Menu onClickMenuItem={(key) => key === 'logout' && void handleLogout()}>
                    <MenuItem key="logout">退出登录</MenuItem>
                  </Menu>
                )}
              >
                <div className="user-entry">
                  <div className="user-avatar">{userInitial}</div>
                  <div className="user-info">
                    <span className="user-name">{user?.username}</span>
                    <span className="user-role">{roleLabel}</span>
                  </div>
                  <IconDown style={{ fontSize: 12, color: 'var(--zcid-text-3)' }} />
                </div>
              </Dropdown>
            )}
          </div>
        </Header>
        <Content className="app-content">{children}</Content>
      </Layout>
    </Layout>
  );
}
