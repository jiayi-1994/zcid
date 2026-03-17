import { Button, Result } from '@arco-design/web-react';
import { useNavigate } from 'react-router-dom';
import { AppLayout } from '../../components/layout/AppLayout';

export function NotFoundPage() {
  const navigate = useNavigate();
  return (
    <AppLayout>
      <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: '60vh' }}>
        <Result
          status="404"
          title="404"
          subTitle="页面不存在"
          extra={
            <Button type="primary" onClick={() => navigate('/dashboard')} style={{ borderRadius: 8 }}>
              返回首页
            </Button>
          }
        />
      </div>
    </AppLayout>
  );
}
