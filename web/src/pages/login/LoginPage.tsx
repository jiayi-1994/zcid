import { Button, Card, Input, Space, Typography, Message } from '@arco-design/web-react';
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
    <div
      style={{
        minHeight: '100vh',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        padding: 16,
        background: 'linear-gradient(180deg, var(--zcid-color-bg-secondary) 0%, var(--zcid-color-bg-tertiary) 50%, var(--zcid-color-bg-secondary) 100%)',
      }}
    >
      <Card
        className="zcid-card"
        bodyStyle={{ padding: 24 }}
        style={{
          maxWidth: 420,
          width: '100%',
          boxShadow: 'var(--zcid-shadow-lg)',
          animation: 'fadeSlideUp 0.4s ease-out',
        }}
      >
        <Space direction="vertical" size="large" style={{ width: '100%' }}>
          <Typography.Title heading={4} style={{ margin: 0, letterSpacing: '-0.5px' }}>
            登录 zcid
          </Typography.Title>
          <Typography.Text type="secondary" style={{ fontSize: 14 }}>
            使用你的账号密码登录平台
          </Typography.Text>
          <form onSubmit={submit}>
            <Space direction="vertical" size="medium" style={{ width: '100%' }}>
              <Input
                placeholder="用户名"
                value={username}
                onChange={setUsername}
                autoComplete="username"
              />
              <Input.Password
                placeholder="密码"
                value={password}
                onChange={setPassword}
                autoComplete="current-password"
              />
              <Button type="primary" htmlType="submit" long loading={loading}>
                登录
              </Button>
            </Space>
          </form>
        </Space>
      </Card>
    </div>
  );
}
