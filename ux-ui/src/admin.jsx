// Admin pages: Users, Global Variables, Integrations, Audit Log, System Settings
const { I, Btn, Input, Badge, Card, StatusBadge, Select, PageHeader, Metric, Switch } = window;

function AdminUsersPage() {
  return (
    <>
      <PageHeader crumb="System · Access Control" title="用户管理" sub="管理系统用户账号与角色。"
        actions={<Btn size="sm" variant="primary" icon={<I.plus size={13} />}>新建用户</Btn>} />
      <div style={{ padding: 24 }}>
        <Card padding={false}>
          <table className="table">
            <thead><tr><th>用户名</th><th>角色</th><th>状态</th><th>创建时间</th><th style={{textAlign:'right'}}>操作</th></tr></thead>
            <tbody>
              {USERS.map((u) => (
                <tr key={u.name}>
                  <td><div style={{display:'flex',alignItems:'center',gap:8}}><span className="avatar avatar--sm avatar--round">{u.name[0].toUpperCase()}</span><b style={{fontWeight:500}}>{u.name}</b></div></td>
                  <td><StatusBadge status={u.role} /></td>
                  <td><StatusBadge status={u.status} /></td>
                  <td><span className="sub mono" style={{fontSize:11.5}}>{u.created}</span></td>
                  <td style={{textAlign:'right'}}>
                    <div style={{display:'inline-flex',gap:4}}>
                      <Btn size="xs" variant="ghost" icon={<I.edit size={12} />} />
                      <Btn size="xs" variant="ghost">{u.status === 'enabled' ? '禁用' : '启用'}</Btn>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </Card>
      </div>
    </>
  );
}

function AdminVariablePage() {
  return (
    <>
      <PageHeader crumb="System › Variables" title="Global Variables" sub="System-wide variable management. 管理跨项目共享的全局变量和密钥。"
        actions={<Btn size="sm" variant="primary" icon={<I.plus size={13} />}>Add Variable</Btn>} />
      <div style={{ padding: 24 }}>
        <Card padding={false}>
          <table className="table">
            <thead><tr><th>变量名</th><th>值</th><th>类型</th><th>描述</th><th>创建时间</th><th style={{textAlign:'right'}}>操作</th></tr></thead>
            <tbody>
              {VARIABLES_GLOBAL.map((v) => (
                <tr key={v.key}>
                  <td><span className="code">{v.key}</span></td>
                  <td><span className="mono" style={{color: v.type==='Secret'?'var(--z-400)':'var(--z-800)'}}>{v.value}</span></td>
                  <td><Badge tone={v.type==='Secret'?'red':'blue'}>{v.type}</Badge></td>
                  <td><span className="sub">{v.desc}</span></td>
                  <td><span className="sub mono" style={{fontSize:11.5}}>{v.created}</span></td>
                  <td style={{textAlign:'right'}}>
                    <div style={{display:'inline-flex',gap:4}}>
                      <Btn size="xs" variant="ghost" icon={<I.edit size={12} />} />
                      <Btn size="xs" variant="ghost" icon={<I.trash size={12} />} />
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </Card>
      </div>
    </>
  );
}

function IntegrationsPage() {
  return (
    <>
      <PageHeader crumb="Settings › Integrations" title="Integration Management" sub="Connect and manage your external CI/CD toolchain. 连接并管理 Git 源、镜像仓库、通知渠道。"
        actions={<><Btn size="sm" icon={<I.refresh size={13} />}>Refresh</Btn><Btn size="sm" variant="primary" icon={<I.plus size={13} />}>Connect New Service</Btn></>} />
      <div style={{ padding: 24, display:'flex', flexDirection:'column', gap: 18 }}>
        <div style={{display:'grid',gridTemplateColumns:'repeat(3,1fr)',gap:14}}>
          <Metric label="TOTAL INTEGRATIONS" value={INTEGRATIONS.length} icon={<I.plug size={14}/>} iconBg="var(--z-100)" iconColor="var(--z-700)" />
          <Metric label="CONNECTED" value={INTEGRATIONS.filter(i=>i.status==='connected').length} icon={<I.check size={14}/>} iconBg="var(--green-soft)" iconColor="var(--green-ink)" trend="all healthy" trendTone="green" />
          <Metric label="NEEDS ATTENTION" value={INTEGRATIONS.filter(i=>i.status!=='connected').length} icon={<I.alert size={14}/>} iconBg="var(--amber-soft)" iconColor="var(--amber-ink)" trend="review tokens" trendTone="amber" />
        </div>
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(320px, 1fr))', gap: 14 }}>
          {INTEGRATIONS.map((it) => (
            <div key={it.name} className="card" style={{ padding: 0 }}>
              <div style={{padding:'14px 14px 10px', display:'flex', gap:11, alignItems:'flex-start'}}>
                <div style={{width:36,height:36,borderRadius:8,background:'var(--z-100)',display:'flex',alignItems:'center',justifyContent:'center',fontSize:18}}>{it.icon}</div>
                <div style={{flex:1,minWidth:0}}>
                  <div style={{fontSize:13.5, fontWeight:600}}>{it.name}</div>
                  <div className="sub" style={{fontSize:11.5}}>{it.provider} · {it.url}</div>
                </div>
              </div>
              <div style={{padding:'6px 14px 12px', display:'flex', flexDirection:'column', gap:6, fontSize:11.5}}>
                <div style={{display:'flex',justifyContent:'space-between',gap:10}}><span style={{color:'var(--z-500)'}}>Token</span><span className="mono" style={{color:'var(--z-700)'}}>{it.token}</span></div>
                <div style={{display:'flex',justifyContent:'space-between',gap:10}}><span style={{color:'var(--z-500)'}}>Description</span><span style={{color:'var(--z-800)', textAlign:'right',overflow:'hidden',textOverflow:'ellipsis',whiteSpace:'nowrap',maxWidth:180}}>{it.desc}</span></div>
                <div style={{display:'flex',justifyContent:'space-between'}}><span style={{color:'var(--z-500)'}}>Created</span><span className="mono" style={{color:'var(--z-700)'}}>{it.created}</span></div>
              </div>
              <div style={{padding:'10px 14px', borderTop:'1px solid var(--z-150)', background:'var(--z-25)', display:'flex', alignItems:'center', justifyContent:'space-between'}}>
                <StatusBadge status={it.status} />
                <div style={{display:'inline-flex',gap:4}}>
                  <Btn size="xs" variant="ghost" icon={<I.play size={11}/>} title="Test" />
                  <Btn size="xs" variant="ghost" icon={<I.copy size={12}/>} title="Copy webhook secret" />
                  <Btn size="xs" variant="ghost" icon={<I.edit size={12}/>} />
                  <Btn size="xs" variant="ghost" icon={<I.trash size={12}/>} />
                </div>
              </div>
            </div>
          ))}
        </div>
      </div>
    </>
  );
}

function AuditLogPage() {
  return (
    <>
      <PageHeader crumb="System · Compliance" title="审计日志" sub="全量 API 操作记录与合规追溯。" />
      <div style={{ padding: 24, display:'flex', flexDirection:'column', gap: 14 }}>
        <div style={{ display: 'flex', gap: 8, flexWrap: 'wrap', alignItems: 'center' }}>
          <Select width={120} value="all" options={[{value:'all',label:'全部方法'},'GET','POST','PUT','DELETE']} />
          <Input placeholder="用户名..." style={{ width: 160 }} icon={<I.user size={13} />} />
          <Input placeholder="2025-04-01 → 2025-04-24" style={{ width: 240 }} icon={<I.calendar size={13} />} />
          <Btn size="sm" variant="ghost" icon={<I.filter size={13} />}>更多过滤</Btn>
          <div style={{ flex: 1 }} />
          <span className="sub" style={{ fontSize: 11.5 }}>{AUDIT.length} 条 / 共 2,184 条</span>
        </div>
        <Card padding={false}>
          <table className="table">
            <thead><tr><th>时间</th><th>用户</th><th>方法</th><th>接口</th><th>资源类型</th><th>资源 ID</th><th>结果</th><th>IP</th></tr></thead>
            <tbody>
              {AUDIT.map((a,i) => (
                <tr key={i}>
                  <td><span className="sub mono" style={{fontSize:11}}>{a.time}</span></td>
                  <td><span className="mono" style={{fontSize:11.5}}>{a.user}</span></td>
                  <td><StatusBadge status={a.method} /></td>
                  <td><span className="code" style={{maxWidth: 240, display:'inline-block', overflow:'hidden', textOverflow:'ellipsis', whiteSpace:'nowrap', verticalAlign:'middle'}}>{a.path}</span></td>
                  <td><span className="sub">{a.rt}</span></td>
                  <td><span className="mono" style={{fontSize:11}}>{a.rid}</span></td>
                  <td><StatusBadge status={a.result} /></td>
                  <td><span className="mono sub" style={{fontSize:11}}>{a.ip}</span></td>
                </tr>
              ))}
            </tbody>
          </table>
        </Card>
        <div style={{display:'flex',justifyContent:'space-between',fontSize:11.5,color:'var(--z-500)'}}>
          <span>共 2,184 条 · 当前 1 - 8</span>
          <div style={{display:'flex',gap:4}}><Btn size="xs" variant="ghost" icon={<I.chevL size={12}/>} /><Btn size="xs" variant="outline">1</Btn><Btn size="xs" variant="ghost">2</Btn><Btn size="xs" variant="ghost">3</Btn><span style={{padding:'0 4px'}}>…</span><Btn size="xs" variant="ghost">273</Btn><Btn size="xs" variant="ghost" icon={<I.chevR size={12}/>} /></div>
        </div>
      </div>
    </>
  );
}

function SystemSettingsPage() {
  const [cfg, setCfg] = React.useState({ k8s:'https://k8s.zcid.local:6443', registry:'harbor.zcid.local', argocd:'https://argocd.zcid.local' });
  return (
    <>
      <PageHeader crumb="System › Settings" title="System Settings" sub="Platform configuration & health. 平台级配置、健康状态与集成监控。" />
      <div style={{ padding: '24px 24px 48px', display:'flex', flexDirection:'column', gap: 18, maxWidth: 900 }}>
        <Card title="平台配置" extra={<Btn size="sm" variant="primary">保存配置</Btn>}>
          <div style={{ display: 'flex', flexDirection: 'column', gap: 14 }}>
            <Field label="K8s API Server 地址" required><input className="input" value={cfg.k8s} onChange={e=>setCfg({...cfg,k8s:e.target.value})} /></Field>
            <Field label="默认镜像仓库" required><input className="input" value={cfg.registry} onChange={e=>setCfg({...cfg,registry:e.target.value})} /></Field>
            <Field label="ArgoCD 地址" required><input className="input" value={cfg.argocd} onChange={e=>setCfg({...cfg,argocd:e.target.value})} /></Field>
          </div>
        </Card>
        <Card title={<div style={{display:'flex',alignItems:'center',gap:10}}><h2>健康状态</h2><Badge tone="green" dot>ok · 全部健康</Badge></div>} extra={<Btn size="sm" icon={<I.refresh size={13}/>}>刷新</Btn>}>
          <div style={{display:'grid',gridTemplateColumns:'repeat(2,1fr)',rowGap:10,columnGap:20, fontSize:12.5}}>
            {[
              ['Tekton Pipelines Controller','ok','running · v0.56.0'],
              ['ArgoCD Application Controller','ok','synced · v2.11.3'],
              ['Postgres (meta store)','ok','conn 3 / 20'],
              ['Redis (cache + queue)','ok','mem 182 MB'],
              ['Object Storage (logs)','ok','free 72%'],
              ['K8s cluster (api)','ok','4 nodes ready'],
              ['Webhook relay','ok','last ping 12s'],
              ['Notification dispatcher','ok','0 backlog'],
            ].map(([k,st,d]) => (
              <div key={k} style={{display:'flex',justifyContent:'space-between', padding:'8px 0', borderBottom:'1px dashed var(--z-150)'}}>
                <span>{k}</span>
                <span style={{display:'inline-flex',gap:8,alignItems:'center'}}>
                  <span className="mono sub" style={{fontSize:11}}>{d}</span>
                  <Badge tone="green" dot>{st}</Badge>
                </span>
              </div>
            ))}
          </div>
        </Card>
        <Card title="集成状态">
          <div style={{display:'flex',flexDirection:'column',gap:10}}>
            {INTEGRATIONS.map((it)=> (
              <div key={it.name} style={{display:'flex',alignItems:'center',justifyContent:'space-between',padding:'8px 0',borderBottom:'1px dashed var(--z-150)'}}>
                <div>
                  <div style={{fontSize:12.5,fontWeight:500}}>{it.name}</div>
                  <div className="sub mono" style={{fontSize:11}}>{it.url} · last sync ok</div>
                </div>
                <StatusBadge status={it.status} />
              </div>
            ))}
          </div>
        </Card>
      </div>
    </>
  );
}

Object.assign(window, { AdminUsersPage, AdminVariablePage, IntegrationsPage, AuditLogPage, SystemSettingsPage });
