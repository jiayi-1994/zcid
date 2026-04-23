import { Button } from '@arco-design/web-react';
import { IconHome } from '@arco-design/web-react/icon';
import { useNavigate } from 'react-router-dom';
import { AppLayout } from '../../components/layout/AppLayout';

export function NotFoundPage() {
  const navigate = useNavigate();
  return (
    <AppLayout>
      <div className="forbidden-page">
        <div>
          <div className="forbidden-code">404</div>
          <div className="forbidden-title">页面不存在</div>
          <div className="forbidden-desc">
            你访问的页面不存在或已被移除。请检查地址或返回首页。
          </div>
          <Button
            type="primary"
            icon={<IconHome />}
            onClick={() => navigate('/dashboard')}
          >
            返回首页
          </Button>
        </div>
      </div>
    </AppLayout>
  );
}
