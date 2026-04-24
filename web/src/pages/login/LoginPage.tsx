import { Message } from '@arco-design/web-react';
import { type FormEvent, useState } from 'react';
import { Navigate, useNavigate } from 'react-router-dom';
import { login } from '../../services/auth';
import { useAuthStore } from '../../stores/auth';
import { IZap, IRocket } from '../../components/ui/icons';

export function LoginPage() {
  const navigate = useNavigate();
  const setSession = useAuthStore((s) => s.setSession);
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated());

  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [loading, setLoading] = useState(false);

  if (isAuthenticated) return <Navigate to="/dashboard" replace />;

  const submit = async (e: FormEvent) => {
    e.preventDefault();
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
    <div className="zc" style={{ display: 'grid', gridTemplateColumns: '1fr 1fr' }}>
      {/* Left brand panel */}
      <div style={{
        position: 'relative',
        background: 'linear-gradient(135deg, oklch(0.38 0.18 var(--accent-h)), oklch(0.28 0.16 calc(var(--accent-h) + 40)))',
        color: '#fff',
        padding: '48px 56px',
        display: 'flex', flexDirection: 'column', justifyContent: 'space-between',
        overflow: 'hidden',
      }}>
        <div style={{ position: 'absolute', inset: 0, background: 'radial-gradient(ellipse at 80% 20%, rgba(255,255,255,0.15), transparent 60%)', pointerEvents: 'none' }} />
        <div style={{ position: 'relative' }}>
          <div className="zmark" style={{ width: 56, height: 56, borderRadius: 14, fontSize: 28, marginBottom: 24 }}>Z</div>
          <div style={{ fontSize: 34, fontWeight: 700, letterSpacing: -0.04, lineHeight: 1.1 }}>zcid</div>
          <div style={{ fontSize: 14, opacity: 0.75, marginTop: 8, maxWidth: 340 }}>
            云原生 CI/CD 平台 · 基于 Tekton + ArgoCD 构建，让部署像 commit 一样简单。
          </div>
        </div>
        <div style={{ position: 'relative', display: 'flex', flexDirection: 'column', gap: 12, maxWidth: 360 }}>
          {[
            ['⚡', '可视化流水线编辑', '拖拽节点即可构建 DAG'],
            ['🐳', 'Kaniko / BuildKit 容器化构建', '免 Docker Daemon，安全隔离'],
            ['🚀', 'ArgoCD GitOps 自动部署', '声明式同步，一键回滚'],
            ['🔐', 'AES-256 加密变量 + RBAC', '敏感数据加密传输与存储'],
          ].map(([e, t, s]) => (
            <div key={t} style={{ display: 'flex', gap: 11, alignItems: 'flex-start' }}>
              <span style={{ fontSize: 16 }}>{e}</span>
              <div>
                <div style={{ fontSize: 13, fontWeight: 500 }}>{t}</div>
                <div style={{ fontSize: 11.5, opacity: 0.6 }}>{s}</div>
              </div>
            </div>
          ))}
        </div>
        <div style={{ position: 'relative', fontSize: 11, opacity: 0.5 }}>© 2025 zcid — v2.4.0</div>
      </div>

      {/* Right form panel */}
      <div style={{ background: 'var(--z-0)', display: 'flex', alignItems: 'center', justifyContent: 'center', padding: 32 }}>
        <div style={{ width: 360, maxWidth: '100%' }}>
          <h1 style={{ fontSize: 22, marginBottom: 6 }}>欢迎回来</h1>
          <p className="sub" style={{ marginBottom: 28 }}>使用你的账号密码登录平台</p>
          <form onSubmit={submit} style={{ display: 'flex', flexDirection: 'column', gap: 14 }}>
            <div>
              <div className="field-label">用户名</div>
              <input
                className="input"
                style={{ height: 40 }}
                placeholder="admin"
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                autoComplete="username"
              />
            </div>
            <div>
              <div className="field-label" style={{ display: 'flex', justifyContent: 'space-between' }}>
                <span>密码</span>
                <a style={{ fontSize: 11, color: 'var(--accent-ink)', cursor: 'pointer' }}>忘记密码?</a>
              </div>
              <input
                className="input"
                style={{ height: 40 }}
                type="password"
                placeholder="••••••••"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                autoComplete="current-password"
              />
            </div>
            <button
              type="submit"
              className="btn btn--primary"
              disabled={loading}
              style={{ height: 40, justifyContent: 'center', fontSize: 13.5 }}
            >
              {loading ? '登录中...' : '登录'}
            </button>
          </form>
          <div style={{ marginTop: 32, fontSize: 10.5, color: 'var(--z-400)', textAlign: 'center', display: 'flex', alignItems: 'center', justifyContent: 'center', gap: 8 }}>
            <span>Powered by</span>
            <span className="tag"><IZap size={10} />Tekton</span>
            <span>+</span>
            <span className="tag"><IRocket size={10} />ArgoCD</span>
          </div>
        </div>
      </div>
    </div>
  );
}
