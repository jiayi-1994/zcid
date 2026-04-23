import { Button } from '@arco-design/web-react';
import { IconArrowLeft } from '@arco-design/web-react/icon';
import { useNavigate } from 'react-router-dom';
import { AppLayout } from '../../components/layout/AppLayout';

export function ForbiddenPage() {
  const navigate = useNavigate();

  return (
    <AppLayout>
      <div className="forbidden-page">
        <div>
          <div className="forbidden-code">403</div>
          <div className="forbidden-title">无权限访问</div>
          <div className="forbidden-desc">
            你没有访问当前页面的权限，请联系管理员获取相应权限。
          </div>
          <Button
            type="primary"
            icon={<IconArrowLeft />}
            onClick={() => navigate('/dashboard', { replace: true })}
          >
            返回 Dashboard
          </Button>
        </div>
      </div>
    </AppLayout>
  );
}
