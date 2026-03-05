import { Layout, Menu, Skeleton, Typography } from '@arco-design/web-react';
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
    if (id) {
      fetchProject(id).then(setProject).catch(() => setProject(null));
    }
  }, [id]);

  const basePath = `/projects/${id}`;
  const currentKey = location.pathname.replace(basePath, '').split('/')[1] || 'environments';

  return (
    <AppLayout>
      <Layout style={{ height: '100%' }}>
        <Sider width={180} style={{ background: 'var(--zcid-color-bg-primary)', borderRight: '1px solid var(--zcid-color-border)' }}>
          <div style={{ padding: '16px 14px 10px', borderBottom: '1px solid var(--zcid-color-border)' }}>
            {project ? (
              <Typography.Text bold ellipsis style={{ fontSize: 14 }}>
                {project.name}
              </Typography.Text>
            ) : (
              <Skeleton text={{ rows: 1 }} animation />
            )}
          </div>
          <Menu
            selectedKeys={[currentKey]}
            onClickMenuItem={(key) => navigate(`${basePath}/${key}`)}
            style={{ borderRight: 'none' }}
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
