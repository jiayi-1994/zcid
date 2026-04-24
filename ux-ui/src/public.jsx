// Public pages: Login (2 variants), 403, 404
const { I, Btn, Input, AppShell } = window;

// ──── Login A — split-screen brand ────
function LoginSplit() {
  return (
    <div className="zc" style={{ display: 'grid', gridTemplateColumns: '1fr 1fr' }}>
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
          <div style={{ fontSize: 14, opacity: 0.75, marginTop: 8, maxWidth: 340 }}>云原生 CI/CD 平台 · 基于 Tekton + ArgoCD 构建，让部署像 commit 一样简单。</div>
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

      <div style={{ background: 'var(--z-0)', display: 'flex', alignItems: 'center', justifyContent: 'center', padding: 32 }}>
        <div style={{ width: 360, maxWidth: '100%' }}>
          <h1 style={{ fontSize: 22, marginBottom: 6 }}>欢迎回来</h1>
          <p className="sub" style={{ marginBottom: 28 }}>使用你的账号密码登录平台</p>
          <div style={{ display: 'flex', flexDirection: 'column', gap: 14 }}>
            <div>
              <div className="field-label">用户名</div>
              <input className="input" style={{ height: 40 }} defaultValue="admin" />
            </div>
            <div>
              <div className="field-label" style={{ display: 'flex', justifyContent: 'space-between' }}>
                <span>密码</span>
                <a style={{ fontSize: 11, color: 'var(--accent-ink)' }}>忘记密码?</a>
              </div>
              <input className="input" style={{ height: 40 }} type="password" defaultValue="••••••••" />
            </div>
            <Btn variant="primary" style={{ height: 40, justifyContent: 'center', fontSize: 13.5 }}>登录</Btn>
          </div>
          <div style={{ marginTop: 32, fontSize: 10.5, color: 'var(--z-400)', textAlign: 'center', display: 'flex', alignItems: 'center', justifyContent: 'center', gap: 8 }}>
            <span>Powered by</span>
            <span className="tag"><I.zap size={10} />Tekton</span>
            <span>+</span>
            <span className="tag"><I.rocket size={10} />ArgoCD</span>
          </div>
        </div>
      </div>
    </div>
  );
}

// ──── Login B — centered minimal ────
function LoginMinimal() {
  return (
    <div className="zc" style={{ alignItems: 'center', justifyContent: 'center', background: 'var(--z-25)' }}>
      <div style={{
        background: 'var(--z-0)', border: '1px solid var(--z-200)', borderRadius: 14,
        padding: 36, width: 380, boxShadow: '0 1px 2px rgba(0,0,0,.04), 0 16px 48px rgba(0,0,0,.06)'
      }}>
        <div className="zmark" style={{ width: 40, height: 40, borderRadius: 10, fontSize: 20, marginBottom: 20 }}>Z</div>
        <h1 style={{ fontSize: 20, marginBottom: 4 }}>登录 zcid</h1>
        <p className="sub" style={{ marginBottom: 22 }}>云原生 CI/CD 平台</p>
        <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
          <div>
            <div className="field-label">用户名</div>
            <input className="input" style={{ height: 38 }} placeholder="admin" />
          </div>
          <div>
            <div className="field-label">密码</div>
            <input className="input" style={{ height: 38 }} type="password" placeholder="••••••••" />
          </div>
          <Btn variant="primary" style={{ height: 38, justifyContent: 'center', marginTop: 4 }}>登录</Btn>
        </div>
        <div style={{ borderTop: '1px solid var(--z-150)', margin: '20px 0 14px' }} />
        <div style={{ fontSize: 11, color: 'var(--z-500)', textAlign: 'center' }}>Powered by Tekton + ArgoCD</div>
      </div>
    </div>
  );
}

// ──── 403 ────
function ForbiddenPage() {
  return (
    <AppShell active="dashboard" crumbs={['403']}>
      <div style={{ padding: '80px 40px', display: 'flex', flexDirection: 'column', alignItems: 'center', textAlign: 'center' }}>
        <div style={{
          fontSize: 120, fontWeight: 700, letterSpacing: -0.06, lineHeight: 1,
          background: 'linear-gradient(135deg, var(--accent-1), var(--accent-2))',
          WebkitBackgroundClip: 'text', WebkitTextFillColor: 'transparent',
        }}>403</div>
        <h1 style={{ fontSize: 22, marginTop: 24 }}>无权限访问</h1>
        <p className="sub" style={{ marginTop: 8, maxWidth: 420 }}>你没有访问当前页面的权限，请联系管理员获取相应权限。</p>
        <div style={{ marginTop: 24 }}>
          <Btn variant="primary" icon={<I.arrL size={13} />}>返回 Dashboard</Btn>
        </div>
      </div>
    </AppShell>
  );
}

// ──── 404 ────
function NotFoundPage() {
  return (
    <AppShell active="dashboard" crumbs={['404']}>
      <div style={{ padding: '80px 40px', display: 'flex', flexDirection: 'column', alignItems: 'center', textAlign: 'center' }}>
        <div style={{
          fontSize: 120, fontWeight: 700, letterSpacing: -0.06, lineHeight: 1,
          background: 'linear-gradient(135deg, var(--accent-1), var(--accent-2))',
          WebkitBackgroundClip: 'text', WebkitTextFillColor: 'transparent',
        }}>404</div>
        <h1 style={{ fontSize: 22, marginTop: 24 }}>页面不存在</h1>
        <p className="sub" style={{ marginTop: 8, maxWidth: 420 }}>你访问的页面不存在或已被移除。</p>
        <div style={{ marginTop: 24 }}>
          <Btn variant="primary" icon={<I.home size={13} />}>返回首页</Btn>
        </div>
      </div>
    </AppShell>
  );
}

Object.assign(window, { LoginSplit, LoginMinimal, ForbiddenPage, NotFoundPage });
