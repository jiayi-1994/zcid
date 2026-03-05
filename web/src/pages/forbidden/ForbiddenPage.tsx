import { Button, Space, Typography } from '@arco-design/web-react';
import { useNavigate } from 'react-router-dom';
import { AppLayout } from '../../components/layout/AppLayout';

export function ForbiddenPage() {
  const navigate = useNavigate();

  return (
    <AppLayout>
      <div style={{ padding: 24 }}>
        <Space direction="vertical" size="large">
          <Typography.Title heading={4} style={{ margin: 0 }}>
            403 无权限访问
          </Typography.Title>
          <Typography.Paragraph style={{ margin: 0 }}>
            你没有访问当前页面的权限。
          </Typography.Paragraph>
          <Button type="primary" onClick={() => navigate('/dashboard', { replace: true })}>
            返回 Dashboard
          </Button>
        </Space>
      </div>
    </AppLayout>
  );
}
