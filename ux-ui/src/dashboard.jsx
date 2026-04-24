// Dashboard (3 variations) + Project List
const { I, Btn, Input, Badge, Avatar, Card, StatusBadge, PageHeader, Metric } = window;

function greet() {
  const h = new Date().getHours();
  if (h < 6) return '深夜好';
  if (h < 12) return '早上好';
  if (h < 18) return '下午好';
  return '晚上好';
}

// Shared metrics
const METRICS = [
  { label: 'TOTAL PROJECTS', value: 6, icon: <I.folder size={14} />, trend: '+2 this month', iconBg: 'color-mix(in oklch, var(--accent-1), white 85%)', iconColor: 'var(--accent-ink)' },
  { label: 'TOTAL PIPELINES', value: 36, icon: <I.zap size={14} />, trend: '24 active', iconBg: 'var(--z-100)', iconColor: 'var(--z-700)' },
  { label: 'RECENT SUCCESS', value: 128, icon: <I.check size={14} />, trend: '94.3% success rate', trendTone: 'green', iconBg: 'var(--green-soft)', iconColor: 'var(--green-ink)' },
  { label: 'RECENT FAILURES', value: 4, icon: <I.x size={14} />, trend: 'Needs attention', trendTone: 'red', iconBg: 'var(--red-soft)', iconColor: 'var(--red-ink)' },
];

// ─── Variant A: metric-heavy + projects + quick actions (spec default) ───
function DashboardA({ onNavigate }) {
  return (
    <>
      <PageHeader
        crumb="Cloud-Native Overview"
        title={`${greet()}，admin`}
        sub="Global infrastructure health and deployment telemetry. 当前 CI/CD 工作台概览。"
        actions={
          <>
            <Btn size="sm" icon={<I.refresh size={13} />}>Refresh</Btn>
            <Btn size="sm" variant="primary" icon={<I.plus size={13} />}>New Pipeline</Btn>
          </>
        }
      />
      <div style={{ padding: 24, display: 'flex', flexDirection: 'column', gap: 20 }}>
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(4, 1fr)', gap: 14 }}>
          {METRICS.map((m) => <Metric key={m.label} {...m} />)}
        </div>
        <div style={{ display: 'grid', gridTemplateColumns: '2fr 1fr', gap: 20 }}>
          <Card
            padding={false}
            title={<div style={{display:'flex', alignItems:'baseline', gap:8}}><h2>Projects</h2><span className="sub">最近活跃</span></div>}
            extra={<a style={{ fontSize: 12, color: 'var(--accent-ink)', cursor: 'pointer' }} onClick={() => onNavigate && onNavigate('projects')}>查看全部 →</a>}
          >
            <div style={{ padding: 8 }}>
              {PROJECTS.slice(0, 5).map((p) => (
                <div key={p.id} className="lrow">
                  <Avatar name={p.name} tone={p.hue} round size="sm" />
                  <div style={{ flex: 1, minWidth: 0 }}>
                    <div style={{ fontSize: 13, fontWeight: 500, color: 'var(--z-900)' }}>{p.name}</div>
                    <div className="sub" style={{ fontSize: 11.5, whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis' }}>{p.desc}</div>
                  </div>
                  <StatusBadge status={p.lastStatus} />
                  <I.chevR size={14} style={{ color: 'var(--z-400)' }} />
                </div>
              ))}
            </div>
          </Card>
          <Card title="快速操作">
            <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
              {[
                { i: <I.folder size={15} />, t: '新建项目', s: '配置 git、变量和环境' },
                { i: <I.zap size={15} />, t: '创建流水线', s: '从模板或空白开始' },
                { i: <I.plug size={15} />, t: '集成管理', s: '连接 GitHub / GitLab' },
                { i: <I.book size={15} />, t: '查看文档', s: '产品指南 & API' },
              ].map((x) => (
                <div key={x.t} className="lrow" style={{ padding: '9px 10px' }}>
                  <div style={{ width: 30, height: 30, borderRadius: 8, background: 'var(--z-100)', color: 'var(--z-700)', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>{x.i}</div>
                  <div style={{ flex: 1, minWidth: 0 }}>
                    <div style={{ fontSize: 12.5, fontWeight: 500 }}>{x.t}</div>
                    <div className="sub" style={{ fontSize: 11 }}>{x.s}</div>
                  </div>
                  <I.arrR size={13} style={{ color: 'var(--z-400)' }} />
                </div>
              ))}
            </div>
          </Card>
        </div>
      </div>
    </>
  );
}

// ─── Variant B: feed-heavy (activity stream) ───
function DashboardB() {
  const feed = [
    { who: 'zhao.wei', did: '触发了流水线', on: 'build-web-console', at: '刚刚', status: 'running', icon: <I.zap size={12} /> },
    { who: 'admin', did: '部署到 prod', on: 'payment-svc@1.18.2', at: '12 分钟前', status: 'succeeded', icon: <I.rocket size={12} /> },
    { who: 'wang.ming', did: '流水线失败', on: 'iam-integration-test', at: '2 小时前', status: 'failed', icon: <I.zap size={12} /> },
    { who: 'system', did: '定时触发', on: 'nightly-security-scan', at: '昨天 02:00', status: 'succeeded', icon: <I.clock size={12} /> },
    { who: 'li.qiang', did: '创建了流水线', on: 'payment-perf-bench', at: '2 天前', status: 'draft', icon: <I.plus size={12} /> },
    { who: 'admin', did: '回滚部署', on: 'web-console@2.3.9', at: '3 天前', status: 'rolled_back', icon: <I.sync size={12} /> },
  ];
  return (
    <>
      <PageHeader
        crumb="Cloud-Native Overview"
        title={`${greet()}，admin`}
        sub="Global infrastructure health and deployment telemetry."
        actions={<><Btn size="sm" icon={<I.refresh size={13} />}>Refresh</Btn><Btn size="sm" variant="primary" icon={<I.plus size={13} />}>New Pipeline</Btn></>}
      />
      <div style={{ padding: 24, display: 'flex', flexDirection: 'column', gap: 20 }}>
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(4, 1fr)', gap: 14 }}>
          {METRICS.map((m) => <Metric key={m.label} {...m} />)}
        </div>
        <div style={{ display: 'grid', gridTemplateColumns: '3fr 2fr', gap: 20 }}>
          <Card title="Activity Feed" extra={<span className="sub" style={{ fontSize: 11 }}>全局操作流</span>}>
            <div style={{ display: 'flex', flexDirection: 'column' }}>
              {feed.map((f, i) => (
                <div key={i} style={{ display: 'flex', gap: 11, padding: '9px 0', borderBottom: i < feed.length - 1 ? '1px solid var(--z-100)' : 'none', alignItems: 'center' }}>
                  <Avatar name={f.who} size="sm" round />
                  <div style={{ flex: 1, minWidth: 0, fontSize: 12.5 }}>
                    <b style={{ fontWeight: 500 }}>{f.who}</b>
                    <span style={{ color: 'var(--z-500)' }}> {f.did} </span>
                    <span className="code" style={{ padding: '1px 5px' }}>{f.on}</span>
                  </div>
                  <StatusBadge status={f.status} />
                  <span className="sub" style={{ fontSize: 11 }}>{f.at}</span>
                </div>
              ))}
            </div>
          </Card>
          <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
            <Card title="部署频率">
              <BuildSparkline />
            </Card>
            <Card title="运行时间 · 本周">
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'baseline' }}>
                <div>
                  <div style={{ fontSize: 26, fontWeight: 600, letterSpacing: -0.02 }}>99.94%</div>
                  <div className="sub" style={{ fontSize: 11.5 }}>平台可用率 · 过去 7 天</div>
                </div>
                <Badge tone="green" dot>all up</Badge>
              </div>
              <div className="pbar" style={{ marginTop: 12 }}><i style={{ width: '99.94%' }} /></div>
            </Card>
          </div>
        </div>
      </div>
    </>
  );
}

function BuildSparkline() {
  // purely static bars
  const data = [4, 6, 3, 8, 5, 9, 12, 7, 10, 14, 11, 8, 13, 15, 12];
  const max = Math.max(...data);
  return (
    <div>
      <div style={{ display: 'flex', alignItems: 'baseline', gap: 10 }}>
        <div style={{ fontSize: 26, fontWeight: 600, letterSpacing: -0.02 }}>128</div>
        <span className="sub" style={{ fontSize: 11.5 }}>deploys / 7d</span>
        <span className="badge badge--green" style={{ marginLeft: 'auto' }}>+18%</span>
      </div>
      <div style={{ display: 'flex', alignItems: 'flex-end', gap: 3, height: 52, marginTop: 10 }}>
        {data.map((v, i) => (
          <div key={i} style={{
            width: 8, height: `${(v / max) * 100}%`,
            background: i === data.length - 1 ? 'linear-gradient(180deg, var(--accent-1), var(--accent-2))' : 'var(--z-200)',
            borderRadius: 2,
          }} />
        ))}
      </div>
    </div>
  );
}

// ─── Variant C: projects-heavy grid ───
function DashboardC() {
  return (
    <>
      <PageHeader
        crumb="Cloud-Native Overview"
        title={`${greet()}，admin`}
        sub="所有项目一览 · 从这里直接深入任何一个。"
        actions={<><Btn size="sm" icon={<I.refresh size={13} />}>Refresh</Btn><Btn size="sm" variant="primary" icon={<I.plus size={13} />}>New Pipeline</Btn></>}
      />
      <div style={{ padding: 24, display: 'flex', flexDirection: 'column', gap: 20 }}>
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(4, 1fr)', gap: 14 }}>
          {METRICS.map((m) => <Metric key={m.label} {...m} />)}
        </div>
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: 14 }}>
          {PROJECTS.slice(0, 6).map((p) => (
            <div key={p.id} className="card" style={{ padding: 14, display: 'flex', flexDirection: 'column', gap: 10, cursor: 'pointer' }}>
              <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
                <Avatar name={p.name} tone={p.hue} />
                <div style={{ flex: 1, minWidth: 0 }}>
                  <div style={{ fontSize: 13, fontWeight: 600 }}>{p.name}</div>
                  <div className="sub" style={{ fontSize: 11 }}>{p.desc}</div>
                </div>
                <StatusBadge status={p.lastStatus} />
              </div>
              <div style={{ display: 'flex', gap: 14, fontSize: 11.5, color: 'var(--z-500)' }}>
                <span><b style={{ color: 'var(--z-800)', fontFamily: 'var(--font-mono)' }}>{p.pipelines}</b> 流水线</span>
                <span><b style={{ color: 'var(--z-800)', fontFamily: 'var(--font-mono)' }}>{p.runs.toLocaleString()}</b> runs</span>
                <span style={{ marginLeft: 'auto' }}>{p.created}</span>
              </div>
            </div>
          ))}
        </div>
      </div>
    </>
  );
}

// ─── Project List Page ───
function ProjectListPage({ onOpen }) {
  return (
    <>
      <PageHeader
        crumb="Project Directory"
        title="项目管理"
        sub="管理所有项目及其 CI/CD 配置。"
        actions={
          <>
            <Input icon={<I.search size={13} />} placeholder="搜索项目..." style={{ width: 220 }} />
            <Btn size="sm" variant="primary" icon={<I.plus size={13} />}>新建项目</Btn>
          </>
        }
      />
      <div style={{ padding: 24, display: 'flex', flexDirection: 'column', gap: 18 }}>
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(2, 1fr)', gap: 14, maxWidth: 520 }}>
          <Metric label="TOTAL PROJECTS" value={PROJECTS.length} icon={<I.folder size={14} />} iconBg="var(--accent-soft)" iconColor="var(--accent-ink)" />
          <Metric label="ACTIVE" value={PROJECTS.filter((p) => p.status === 'active').length} icon={<I.check size={14} />} iconBg="var(--green-soft)" iconColor="var(--green-ink)" />
        </div>
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(280px, 1fr))', gap: 14 }}>
          {PROJECTS.map((p) => (
            <div key={p.id} className="card" style={{ padding: 0, overflow: 'hidden', cursor: 'pointer' }} onClick={() => onOpen && onOpen(p)}>
              <div style={{ padding: '14px 14px 10px', display: 'flex', gap: 10, alignItems: 'flex-start' }}>
                <div style={{ width: 36, height: 36, borderRadius: 8,
                  background: `linear-gradient(135deg, oklch(0.64 0.17 ${p.hue}), oklch(0.52 0.19 ${(p.hue + 24) % 360}))`,
                  color: '#fff', display: 'flex', alignItems: 'center', justifyContent: 'center',
                  fontWeight: 700, fontSize: 15 }}>{p.firstLetter}</div>
                <div style={{ flex: 1, minWidth: 0 }}>
                  <div style={{ fontSize: 13.5, fontWeight: 600 }}>{p.name}</div>
                  <StatusBadge status={p.status === 'disabled' ? 'userDisabled' : 'active'} />
                </div>
              </div>
              <div style={{ padding: '0 14px 12px', fontSize: 12, color: 'var(--z-600)', minHeight: 34 }}>{p.desc}</div>
              <div style={{ padding: '10px 14px', borderTop: '1px solid var(--z-150)', background: 'var(--z-25)',
                display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
                <span className="sub" style={{ fontSize: 11 }}>{p.created}</span>
                <div style={{ display: 'flex', gap: 6 }}>
                  <Btn size="xs" variant="ghost" icon={<I.trash size={12} />} />
                  <Btn size="xs" variant="outline">进入 <I.arrR size={11} /></Btn>
                </div>
              </div>
            </div>
          ))}
        </div>
      </div>
    </>
  );
}

Object.assign(window, { DashboardA, DashboardB, DashboardC, ProjectListPage });
