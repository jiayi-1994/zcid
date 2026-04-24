import { useNavigate } from 'react-router-dom';
import { AppLayout } from '../../components/layout/AppLayout';
import { Btn } from '../../components/ui/Btn';
import { IArrL } from '../../components/ui/icons';

export function ForbiddenPage() {
  const navigate = useNavigate();
  return (
    <AppLayout>
      <div style={{ padding: '80px 40px', display: 'flex', flexDirection: 'column', alignItems: 'center', textAlign: 'center' }}>
        <div style={{
          fontSize: 120, fontWeight: 700, letterSpacing: -0.06, lineHeight: 1,
          background: 'linear-gradient(135deg, var(--accent-1), var(--accent-2))',
          WebkitBackgroundClip: 'text', WebkitTextFillColor: 'transparent',
        }}>403</div>
        <h1 style={{ fontSize: 22, marginTop: 24 }}>无权限访问</h1>
        <p className="sub" style={{ marginTop: 8, maxWidth: 420 }}>你没有访问当前页面的权限，请联系管理员获取相应权限。</p>
        <div style={{ marginTop: 24 }}>
          <Btn variant="primary" icon={<IArrL size={13} />} onClick={() => navigate('/dashboard', { replace: true })}>
            返回 Dashboard
          </Btn>
        </div>
      </div>
    </AppLayout>
  );
}
