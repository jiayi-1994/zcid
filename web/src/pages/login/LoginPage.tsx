import { Button, Input, Message } from '@arco-design/web-react';
import { type FormEvent, useState } from 'react';
import { Navigate, useNavigate } from 'react-router-dom';
import { login } from '../../services/auth';
import { useAuthStore } from '../../stores/auth';

export function LoginPage() {
  const navigate = useNavigate();
  const setSession = useAuthStore((state) => state.setSession);
  const isAuthenticated = useAuthStore((state) => state.isAuthenticated());

  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [loading, setLoading] = useState(false);

  if (isAuthenticated) {
    return <Navigate to="/dashboard" replace />;
  }

  const submit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (!username.trim() || !password.trim()) { Message.error('请输入账号和密码'); return; }
    setLoading(true);
    try {
      const result = await login(username.trim(), password);
      setSession({ accessToken: result.accessToken, refreshToken: result.refreshToken });
      navigate('/dashboard', { replace: true });
    } catch { Message.error('登录失败，请检查账号密码'); }
    finally { setLoading(false); }
  };

  return (
    <div className="login-page">
      {/* Left brand panel */}
      <div className="login-brand-panel">
        <div className="login-brand-content">
          <div className="login-brand-logo">Z</div>
          <h1 className="login-brand-title">zcid</h1>
          <p className="login-brand-desc">
            云原生 CI/CD 平台<br />
            基于 Tekton + ArgoCD 构建
          </p>
          <div className="login-brand-features">
            <div className="login-brand-feature">
              <div className="login-brand-feature-icon">⚡</div>
              <span>可视化流水线编辑，拖拽式 DAG 编排</span>
            </div>
            <div className="login-brand-feature">
              <div className="login-brand-feature-icon">🐳</div>
              <span>Kaniko / BuildKit 容器化构建</span>
            </div>
            <div className="login-brand-feature">
              <div className="login-brand-feature-icon">🚀</div>
              <span>ArgoCD GitOps 自动部署与回滚</span>
            </div>
            <div className="login-brand-feature">
              <div className="login-brand-feature-icon">🔐</div>
              <span>AES-256 加密变量管理 + RBAC 权限控制</span>
            </div>
          </div>
        </div>
      </div>

      {/* Right form panel */}
      <div className="login-form-panel">
        <div style={{ width: '100%', maxWidth: 360 }}>
          <h1 className="login-heading">欢迎回来</h1>
          <p className="login-desc">使用你的账号密码登录平台</p>
          <form onSubmit={submit}>
            <div style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>
              <div>
                <label htmlFor="login-username" style={{ display: 'block', fontSize: 13, fontWeight: 600, color: 'var(--zcid-text-2)', marginBottom: 8 }}>
                  用户名
                </label>
                <Input id="login-username" placeholder="请输入用户名" value={username} onChange={setUsername} autoComplete="username" size="large" style={{ borderRadius: 8 }} />
              </div>
              <div>
                <label htmlFor="login-password" style={{ display: 'block', fontSize: 13, fontWeight: 600, color: 'var(--zcid-text-2)', marginBottom: 8 }}>
                  密码
                </label>
                <Input.Password id="login-password" placeholder="请输入密码" value={password} onChange={setPassword} autoComplete="current-password" size="large" style={{ borderRadius: 8 }} />
              </div>
              <Button type="primary" htmlType="submit" long loading={loading} size="large"
                style={{ marginTop: 4, height: 46, borderRadius: 10, fontWeight: 600, fontSize: 15, background: 'linear-gradient(135deg, #3B82F6 0%, #2563EB 100%)', border: 'none' }}
              >
                登录
              </Button>
            </div>
          </form>
          <div style={{ textAlign: 'center', marginTop: 32, fontSize: 12, color: 'var(--zcid-text-4)' }}>
            Powered by Tekton + ArgoCD
          </div>
        </div>
      </div>
    </div>
  );
}
