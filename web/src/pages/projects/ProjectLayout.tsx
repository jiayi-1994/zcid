import { Layout, Menu, Skeleton } from '@arco-design/web-react';
import { IconCloud, IconApps, IconUserGroup, IconLock, IconThunderbolt, IconSend, IconNotification } from '@arco-design/web-react/icon';
import { useEffect, useState } from 'react';
import { Outlet, useLocation, useNavigate, useParams } from 'react-router-dom';
import { AppLayout } from '../../components/layout/AppLayout';
import { fetchProject, type Project } from '../../services/project';

const { Sider, Content } = Layout;
const MenuItem = Menu.Item;

export function ProjectLayout() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const location = useLocation();
  const [project, setProject] = useState<Project | null>(null);

  useEffect(() => {
    if (id) fetchProject(id).then(setProject).catch(() => setProject(null));
  }, [id]);

  const basePath = `/projects/${id}`;
  const currentKey = location.pathname.replace(basePath, '').split('/')[1] || 'environments';

  return (
    <AppLayout>
      <Layout style={{ height: '100%' }}>
        <Sider width={200} style={{
          background: '#fff',
          borderRight: '1px solid var(--zcid-border)',
        }}>
          {/* Project header */}
          <div style={{
            padding: '16px 16px 12px',
            borderBottom: '1px solid var(--zcid-border)',
          }}>
            {project ? (
              <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
                <div style={{
                  width: 32, height: 32, borderRadius: 8, flexShrink: 0,
                  background: 'linear-gradient(135deg, #E8F3FF 0%, #D6E4FF 100%)',
                  display: 'flex', alignItems: 'center', justifyContent: 'center',
                  fontSize: 14, fontWeight: 700, color: '#165DFF',
                }}>
                  {project.name.charAt(0).toUpperCase()}
                </div>
                <div style={{ minWidth: 0 }}>
                  <div style={{
                    fontSize: 14, fontWeight: 600, color: 'var(--zcid-text-1)',
                    overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap',
                  }}>
                    {project.name}
                  </div>
                </div>
              </div>
            ) : (
              <Skeleton text={{ rows: 1 }} animation />
            )}
          </div>
          <Menu
            selectedKeys={[currentKey]}
            onClickMenuItem={(key) => navigate(`${basePath}/${key}`)}
            style={{ borderRight: 'none', padding: '8px' }}
          >
            <MenuItem key="pipelines" style={{ borderRadius: 8, margin: '2px 0', height: 36, lineHeight: '36px' }}><IconThunderbolt /> 流水线</MenuItem>
            <MenuItem key="environments" style={{ borderRadius: 8, margin: '2px 0', height: 36, lineHeight: '36px' }}><IconCloud /> 环境</MenuItem>
            <MenuItem key="deployments" style={{ borderRadius: 8, margin: '2px 0', height: 36, lineHeight: '36px' }}><IconSend /> 部署</MenuItem>
            <MenuItem key="services" style={{ borderRadius: 8, margin: '2px 0', height: 36, lineHeight: '36px' }}><IconApps /> 服务</MenuItem>
            <MenuItem key="members" style={{ borderRadius: 8, margin: '2px 0', height: 36, lineHeight: '36px' }}><IconUserGroup /> 成员</MenuItem>
            <MenuItem key="variables" style={{ borderRadius: 8, margin: '2px 0', height: 36, lineHeight: '36px' }}><IconLock /> 变量</MenuItem>
            <MenuItem key="notifications" style={{ borderRadius: 8, margin: '2px 0', height: 36, lineHeight: '36px' }}><IconNotification /> 通知</MenuItem>
          </Menu>
        </Sider>
        <Content className="page-container" style={{ overflow: 'auto' }}>
          <Outlet />
        </Content>
      </Layout>
    </AppLayout>
  );
}
