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

  const handleLogout = async () => {
    try {
      if (refreshToken) {
        await logout(refreshToken);
      }
    } catch {
      // ignore
    } finally {
      clearSession();
      navigate('/login', { replace: true });
    }
  };

  return (
    <Layout className="app-root">
      <Sider className="app-sider" width={collapsed ? 60 : 232} collapsed={collapsed}>
        <div className="sider-logo">
          <div className="sider-logo-icon">Z</div>
          {!collapsed && <span className="sider-logo-text">zcid</span>}
        </div>
        <Menu
          selectedKeys={[location.pathname]}
          onClickMenuItem={(key) => navigate(key)}
        >
          {canViewDashboard && (
            <MenuItem key="/dashboard">
              <IconDashboard />
              {!collapsed && 'Dashboard'}
            </MenuItem>
          )}
          <MenuItem key="/projects">
            <IconApps />
            {!collapsed && '项目管理'}
          </MenuItem>
          {canViewAdminUsers && (
            <MenuItem key="/admin/users">
              <IconUser />
              {!collapsed && '用户管理'}
            </MenuItem>
          )}
          {canViewAdminVariables && (
            <MenuItem key="/admin/variables">
              <IconLock />
              {!collapsed && '全局变量'}
            </MenuItem>
          )}
          {canViewAdminIntegrations && (
            <MenuItem key="/admin/integrations">
              <IconLink />
              {!collapsed && '集成管理'}
            </MenuItem>
          )}
          {canViewAuditLogs && (
            <MenuItem key="/admin/audit-logs">
              <IconFile />
              {!collapsed && '审计日志'}
            </MenuItem>
          )}
          {canViewSystemSettings && (
            <MenuItem key="/admin/settings">
              <IconSettings />
              {!collapsed && '系统设置'}
            </MenuItem>
          )}
        </Menu>
      </Sider>
      <Layout>
        <Header className="app-header">
          <div className="app-header-inner">
            <span className="app-header-breadcrumb">zcid</span>
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
          </div>
        </Header>
        <Content className="app-content">{children}</Content>
      </Layout>
    </Layout>
  );
}
