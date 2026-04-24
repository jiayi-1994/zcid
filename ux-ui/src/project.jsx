// Project subpages: Environments, Deployments (list + detail), Services, Members, Variables, Notifications
const { I, Btn, Input, Badge, Avatar, Card, StatusBadge, Switch, Select, PageHeader, Metric } = window;

function EnvironmentListPage() {
  const healthTone = { env_healthy:'green', env_syncing:'blue', env_degraded:'amber', env_down:'red' };
  return (
    <>
      <PageHeader crumb="Environments" title="Environment Health" sub="Deployment status & rollback. 监控和管理部署环境状态。"
        actions={<Btn size="sm" variant="primary" icon={<I.plus size={13}/>}>New Environment</Btn>} />
      <div style={{padding:24,display:'flex',flexDirection:'column',gap:18}}>
        <div style={{display:'grid',gridTemplateColumns:'repeat(3,1fr)',gap:14}}>
          <Metric label="TOTAL ENVIRONMENTS" value={ENVIRONMENTS.length} icon={<I.layers size={14}/>} iconBg="var(--z-100)" iconColor="var(--z-700)" />
          <Metric label="ACTIVE SERVICES" value={ENVIRONMENTS.reduce((s,e)=>s+e.services,0)} icon={<I.server size={14}/>} iconBg="var(--accent-soft)" iconColor="var(--accent-ink)" />
          <Metric label="HEALTH RATE" value="83.7%" icon={<I.heart size={14}/>} iconBg="var(--green-soft)" iconColor="var(--green-ink)" trend="4 of 5 healthy" trendTone="green" />
        </div>
        <div style={{display:'grid',gridTemplateColumns:'repeat(auto-fill, minmax(300px, 1fr))', gap:14}}>
          {ENVIRONMENTS.map(e=>{
            const tone = healthTone[e.health];
            return (
              <div key={e.name} className="card" style={{ padding:0, borderColor: `color-mix(in oklch, var(--${tone}), white 60%)` }}>
                <div style={{padding:'14px 14px 8px', display:'flex', alignItems:'flex-start', justifyContent:'space-between', gap:10}}>
                  <div>
                    <div style={{fontSize:14, fontWeight:600, display:'flex', alignItems:'center', gap:8}}>
                      {e.name}
                      <StatusBadge status={e.health} />
                    </div>
                    <div className="sub" style={{fontSize:11.5, marginTop:2}}>{e.desc}</div>
                  </div>
                  <Btn size="xs" variant="ghost" icon={<I.trash size={12}/>} />
                </div>
                <div style={{padding:'8px 14px 14px', display:'flex', flexDirection:'column', gap:6, fontSize:11.5}}>
                  <div style={{display:'flex',justifyContent:'space-between'}}><span style={{color:'var(--z-500)'}}>Namespace</span><span className="tag">{e.ns}</span></div>
                  <div style={{display:'flex',justifyContent:'space-between'}}><span style={{color:'var(--z-500)'}}>Services</span><b style={{fontFamily:'var(--font-mono)', fontWeight:500}}>{e.services}</b></div>
                  <div style={{display:'flex',justifyContent:'space-between'}}><span style={{color:'var(--z-500)'}}>Created</span><span className="mono" style={{color:'var(--z-700)'}}>{e.created}</span></div>
                </div>
              </div>
            );
          })}
        </div>
      </div>
    </>
  );
}

function DeploymentListPage({ onOpen }) {
  return (
    <>
      <PageHeader crumb="Project · Delivery" title="部署管理" sub="触发与追踪 ArgoCD 部署同步状态。"
        actions={<Btn size="sm" variant="primary" icon={<I.rocket size={13}/>}>触发部署</Btn>} />
      <div style={{padding:24}}>
        <Card padding={false}>
          <table className="table">
            <thead><tr><th>镜像</th><th>环境</th><th>状态</th><th>同步状态</th><th>健康状态</th><th>部署人</th><th>时间</th><th style={{textAlign:'right'}}>详情</th></tr></thead>
            <tbody>
              {DEPLOYMENTS.map(d=>(
                <tr key={d.id}>
                  <td><span className="code">{d.image}</span></td>
                  <td><span className="tag">{d.env}</span></td>
                  <td><StatusBadge status={d.status}/></td>
                  <td>{d.sync === 'Synced' ? <Badge tone="green" dot>{d.sync}</Badge> : d.sync === 'Syncing' ? <Badge tone="blue" dot pulse>{d.sync}</Badge> : <Badge tone="amber" dot>{d.sync}</Badge>}</td>
                  <td>{d.health === 'Healthy' ? <Badge tone="green" dot>{d.health}</Badge> : d.health === 'Progressing' ? <Badge tone="blue" dot pulse>{d.health}</Badge> : d.health === 'Missing' ? <Badge tone="red" dot>{d.health}</Badge> : <Badge tone="amber" dot>{d.health}</Badge>}</td>
                  <td><span className="mono" style={{fontSize:11.5}}>{d.user}</span></td>
                  <td><span className="sub">{d.time}</span></td>
                  <td style={{textAlign:'right'}}><Btn size="xs" variant="outline" onClick={()=>onOpen && onOpen(d)}>详情 →</Btn></td>
                </tr>
              ))}
            </tbody>
          </table>
        </Card>
      </div>
    </>
  );
}

function DeploymentDetailPage({ deployment, onBack }) {
  const d = deployment || DEPLOYMENTS[0];
  return (
    <>
      <div style={{padding:'16px 24px 0'}}>
        <a onClick={onBack} style={{fontSize:12, color:'var(--z-500)', cursor:'pointer', display:'inline-flex', alignItems:'center', gap:4}}><I.arrL size={12}/>返回列表</a>
      </div>
      <PageHeader crumb="Project · Delivery" title="部署详情" sub={<><span className="code">{d.image}</span></>}
        actions={<>
          <Btn size="sm" icon={<I.refresh size={13}/>}>刷新状态</Btn>
          <Btn size="sm" icon={<I.sync size={13}/>}>重新同步</Btn>
          <Btn size="sm" variant="danger" icon={<I.arrL size={13}/>}>回滚</Btn>
        </>} />
      <div style={{padding:24}}>
        <Card title={<div style={{display:'flex',alignItems:'center',gap:10}}><h2>基本信息</h2></div>} extra={<StatusBadge status={d.status}/>}>
          <div style={{display:'grid',gridTemplateColumns:'140px 1fr',rowGap:10,columnGap:16,fontSize:12.5}}>
            {[
              ['ID', <span className="mono" style={{fontSize:11.5}}>d-4a19b2cf38</span>],
              ['镜像', <span className="code">{d.image}</span>],
              ['环境 ID', <span className="tag">{d.env}</span>],
              ['同步状态', d.sync === 'Synced' ? <Badge tone="green" dot>{d.sync}</Badge> : <Badge tone="blue" dot pulse>{d.sync}</Badge>],
              ['健康状态', <Badge tone={d.health==='Healthy'?'green':d.health==='Missing'?'red':'amber'} dot>{d.health}</Badge>],
              ['ArgoCD 应用', <span className="mono" style={{fontSize:11.5}}>web-console-{d.env}</span>],
              ['部署人', <span className="mono" style={{fontSize:11.5}}>{d.user}</span>],
              ['开始时间', <span className="mono sub" style={{fontSize:11.5}}>2025-04-24 09:38:11</span>],
              ['完成时间', d.status==='syncing' ? <span className="sub">—</span> : <span className="mono sub" style={{fontSize:11.5}}>2025-04-24 09:41:02</span>],
              ['创建时间', <span className="mono sub" style={{fontSize:11.5}}>2025-04-24 09:38:09</span>],
            ].map(([k,v],i)=>(
              <React.Fragment key={i}>
                <div style={{color:'var(--z-500)'}}>{k}</div>
                <div>{v}</div>
              </React.Fragment>
            ))}
            {d.status==='sync_failed' && (
              <>
                <div style={{color:'var(--z-500)'}}>错误信息</div>
                <div style={{background:'var(--red-soft)', color:'var(--red-ink)', borderRadius:6, padding:'8px 10px', fontFamily:'var(--font-mono)', fontSize:11.5}}>
                  Error: ComparisonError · Deployment/ml-gateway failed health probe on port 9000 — readiness: 0/3
                </div>
              </>
            )}
          </div>
        </Card>
      </div>
    </>
  );
}

function ServiceListPage() {
  return (
    <>
      <PageHeader crumb="Project · Services" title="服务管理" sub="管理项目下的微服务与源码仓库绑定。"
        actions={<Btn size="sm" variant="primary" icon={<I.plus size={13}/>}>新建服务</Btn>} />
      <div style={{padding:24}}>
        <Card padding={false}>
          <table className="table">
            <thead><tr><th>服务名</th><th>描述</th><th>仓库地址</th><th>创建时间</th><th style={{textAlign:'right'}}>操作</th></tr></thead>
            <tbody>
              {SERVICES.map(s=>(
                <tr key={s.name}>
                  <td><span style={{display:'flex',alignItems:'center',gap:8}}><span className="avatar avatar--sm" style={{background:'var(--z-100)',color:'var(--z-700)'}}>{s.name[0].toUpperCase()}</span><b style={{fontWeight:500}}>{s.name}</b></span></td>
                  <td><span className="sub">{s.desc}</span></td>
                  <td><span className="code">{s.repo}</span></td>
                  <td><span className="sub mono" style={{fontSize:11.5}}>{s.created}</span></td>
                  <td style={{textAlign:'right'}}><Btn size="xs" variant="ghost" icon={<I.trash size={12}/>}/></td>
                </tr>
              ))}
            </tbody>
          </table>
        </Card>
      </div>
    </>
  );
}

function MemberListPage() {
  return (
    <>
      <PageHeader crumb="Project · Access" title="成员管理" sub="分配项目成员与角色权限。"
        actions={<Btn size="sm" variant="primary" icon={<I.plus size={13}/>}>添加成员</Btn>} />
      <div style={{padding:24}}>
        <Card padding={false}>
          <table className="table">
            <thead><tr><th>用户名</th><th>角色</th><th>加入时间</th><th style={{textAlign:'right'}}>操作</th></tr></thead>
            <tbody>
              {USERS.slice(0,5).map(u=>(
                <tr key={u.name}>
                  <td><span style={{display:'flex',alignItems:'center',gap:8}}><Avatar name={u.name} size="sm" round /><b style={{fontWeight:500}}>{u.name}</b></span></td>
                  <td>
                    <select className="input" style={{height:26, fontSize:11.5, width:140}} defaultValue={u.role}>
                      <option value="admin">管理员</option>
                      <option value="project_admin">项目管理员</option>
                      <option value="member">普通成员</option>
                    </select>
                  </td>
                  <td><span className="sub mono" style={{fontSize:11.5}}>{u.created}</span></td>
                  <td style={{textAlign:'right'}}><Btn size="xs" variant="ghost">移除</Btn></td>
                </tr>
              ))}
            </tbody>
          </table>
        </Card>
      </div>
    </>
  );
}

function VariableListPage() {
  return (
    <>
      <PageHeader crumb="Project › Variables" title="Project Variables" sub="Secure variable management for the project. 项目级环境变量与密钥管理。"
        actions={<Btn size="sm" variant="primary" icon={<I.plus size={13}/>}>Add Variable</Btn>} />
      <div style={{padding:24}}>
        <Card padding={false}>
          <table className="table">
            <thead><tr><th>变量名</th><th>值</th><th>类型</th><th>描述</th><th>创建时间</th><th style={{textAlign:'right'}}>操作</th></tr></thead>
            <tbody>
              {PROJECT_VARIABLES.map(v=>(
                <tr key={v.key}>
                  <td><span className="code">{v.key}</span></td>
                  <td><span className="mono" style={{color: v.type==='Secret'?'var(--z-400)':'var(--z-800)'}}>{v.value}</span></td>
                  <td><Badge tone={v.type==='Secret'?'red':'blue'}>{v.type}</Badge></td>
                  <td><span className="sub">{v.desc}</span></td>
                  <td><span className="sub mono" style={{fontSize:11.5}}>{v.created}</span></td>
                  <td style={{textAlign:'right'}}>
                    <div style={{display:'inline-flex',gap:4}}>
                      <Btn size="xs" variant="ghost" icon={<I.edit size={12}/>}/>
                      <Btn size="xs" variant="ghost" icon={<I.trash size={12}/>}/>
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

function NotificationRulesPage() {
  const [rules, setRules] = React.useState(NOTIFICATION_RULES);
  return (
    <>
      <PageHeader crumb="Project · Signals" title="通知规则" sub="配置构建与部署事件的 Webhook 推送。"
        actions={<Btn size="sm" variant="primary" icon={<I.plus size={13}/>}>创建规则</Btn>} />
      <div style={{padding:24}}>
        <Card padding={false}>
          <table className="table">
            <thead><tr><th>名称</th><th>事件类型</th><th>Webhook URL</th><th>启用</th><th>创建时间</th><th style={{textAlign:'right'}}>操作</th></tr></thead>
            <tbody>
              {rules.map((r,i)=>(
                <tr key={r.name}>
                  <td><b style={{fontWeight:500}}>{r.name}</b></td>
                  <td><StatusBadge status={r.event}/></td>
                  <td><span className="code" style={{maxWidth:280, display:'inline-block', overflow:'hidden', textOverflow:'ellipsis', whiteSpace:'nowrap', verticalAlign:'middle'}}>{r.url}</span></td>
                  <td><Switch on={r.enabled} onChange={v=>{ const n=[...rules]; n[i]={...n[i],enabled:v}; setRules(n); }} /></td>
                  <td><span className="sub mono" style={{fontSize:11.5}}>{r.created}</span></td>
                  <td style={{textAlign:'right'}}>
                    <div style={{display:'inline-flex',gap:4}}>
                      <Btn size="xs" variant="ghost" icon={<I.edit size={12}/>}/>
                      <Btn size="xs" variant="ghost" icon={<I.trash size={12}/>}/>
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

Object.assign(window, { EnvironmentListPage, DeploymentListPage, DeploymentDetailPage, ServiceListPage, MemberListPage, VariableListPage, NotificationRulesPage });
