import { Layout, Menu, Skeleton } from '@arco-design/web-react';
import {
  IconCloud,
  IconApps,
  IconUserGroup,
  IconLock,
  IconThunderbolt,
  IconSend,
  IconNotification,
} from '@arco-design/web-react/icon';
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
        <Sider width={220} className="project-sider">
          <div className="project-sider-header">
            {project ? (
              <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
                <div
                  className="stat-card-icon stat-card-icon--primary"
                  style={{
                    width: 32,
                    height: 32,
                    borderRadius: 'var(--radius-md)',
                    fontSize: 14,
                    fontWeight: 700,
                  }}
                >
                  {project.name.charAt(0).toUpperCase()}
                </div>
                <div style={{ minWidth: 0 }}>
                  <div className="project-sider-name">{project.name}</div>
                </div>
              </div>
            ) : (
              <Skeleton text={{ rows: 1 }} animation />
            )}
          </div>
          <Menu
            selectedKeys={[currentKey]}
            onClickMenuItem={(key) => navigate(`${basePath}/${key}`)}
          >
            <MenuItem key="pipelines"><IconThunderbolt /> 流水线</MenuItem>
            <MenuItem key="environments"><IconCloud /> 环境</MenuItem>
            <MenuItem key="deployments"><IconSend /> 部署</MenuItem>
            <MenuItem key="services"><IconApps /> 服务</MenuItem>
            <MenuItem key="members"><IconUserGroup /> 成员</MenuItem>
            <MenuItem key="variables"><IconLock /> 变量</MenuItem>
            <MenuItem key="notifications"><IconNotification /> 通知</MenuItem>
          </Menu>
        </Sider>
        <Content className="page-container" style={{ overflow: 'auto' }}>
          <Outlet />
        </Content>
      </Layout>
    </AppLayout>
  );
}
