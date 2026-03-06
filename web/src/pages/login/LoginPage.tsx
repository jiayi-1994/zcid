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

    if (!username.trim() || !password.trim()) {
      Message.error('请输入账号和密码');
      return;
    }

    setLoading(true);

    try {
      const result = await login(username.trim(), password);
      setSession({
        accessToken: result.accessToken,
        refreshToken: result.refreshToken,
      });
      navigate('/dashboard', { replace: true });
    } catch {
      Message.error('登录失败，请检查账号密码');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="login-page">
      <div className="login-card">
        <div className="login-logo">
          <div className="login-logo-icon">Z</div>
          <span className="login-logo-text">zcid</span>
        </div>
        <h1 className="login-heading">欢迎回来</h1>
        <p className="login-desc">使用你的账号密码登录平台</p>
        <form onSubmit={submit}>
          <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
            <div>
              <label
                htmlFor="login-username"
                style={{ display: 'block', fontSize: 13, fontWeight: 500, color: 'var(--zcid-text-2)', marginBottom: 6 }}
              >
                用户名
              </label>
              <Input
                id="login-username"
                placeholder="请输入用户名"
                value={username}
                onChange={setUsername}
                autoComplete="username"
                size="large"
              />
            </div>
            <div>
              <label
                htmlFor="login-password"
                style={{ display: 'block', fontSize: 13, fontWeight: 500, color: 'var(--zcid-text-2)', marginBottom: 6 }}
              >
                密码
              </label>
              <Input.Password
                id="login-password"
                placeholder="请输入密码"
                value={password}
                onChange={setPassword}
                autoComplete="current-password"
                size="large"
              />
            </div>
            <Button
              type="primary"
              htmlType="submit"
              long
              loading={loading}
              size="large"
              style={{ marginTop: 8, height: 44, borderRadius: 'var(--zcid-radius-md)', fontWeight: 600, fontSize: 15 }}
            >
              登录
            </Button>
          </div>
        </form>
        <div style={{ textAlign: 'center', marginTop: 24, fontSize: 12, color: 'var(--zcid-text-4)' }}>
          Powered by Tekton + ArgoCD
        </div>
      </div>
    </div>
  );
}
