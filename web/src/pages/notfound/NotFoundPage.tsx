import { useNavigate } from 'react-router-dom';
import { AppLayout } from '../../components/layout/AppLayout';
import { Btn } from '../../components/ui/Btn';
import { IHome } from '../../components/ui/icons';

export function NotFoundPage() {
  const navigate = useNavigate();
  return (
    <AppLayout>
      <div style={{ padding: '80px 40px', display: 'flex', flexDirection: 'column', alignItems: 'center', textAlign: 'center' }}>
        <div style={{
          fontSize: 120, fontWeight: 700, letterSpacing: -0.06, lineHeight: 1,
          background: 'linear-gradient(135deg, var(--accent-1), var(--accent-2))',
          WebkitBackgroundClip: 'text', WebkitTextFillColor: 'transparent',
        }}>404</div>
        <h1 style={{ fontSize: 22, marginTop: 24 }}>页面不存在</h1>
        <p className="sub" style={{ marginTop: 8, maxWidth: 420 }}>你访问的页面不存在或已被移除。</p>
        <div style={{ marginTop: 24 }}>
          <Btn variant="primary" icon={<IHome size={13} />} onClick={() => navigate('/dashboard')}>
            返回首页
          </Btn>
        </div>
      </div>
    </AppLayout>
  );
}
