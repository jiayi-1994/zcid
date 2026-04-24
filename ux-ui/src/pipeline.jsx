// Pipeline module: List (2 variants), Template Select, Editor, Run List, Run Detail (2 variants)
const { I, Btn, Input, Badge, Card, StatusBadge, Select, Segmented, PageHeader, Metric, Modal, Field } = window;

// ───────── Pipeline List — Row variant ─────────
function PipelineListRow() {
  const [filter, setFilter] = React.useState('all');
  const items = PIPELINES.filter(p => filter==='all' || p.status===filter);
  const counts = {
    all: PIPELINES.length,
    active: PIPELINES.filter(p=>p.status==='active').length,
    draft: PIPELINES.filter(p=>p.status==='draft').length,
    disabled: PIPELINES.filter(p=>p.status==='disabled').length,
  };
  return (
    <>
      <PageHeader crumb="Pipelines" title="Automated Pipelines" sub="Real-time status of your deployment infrastructure. 自动化流水线管理。"
        actions={<Btn size="sm" variant="primary" icon={<I.plus size={13}/>}>Create Pipeline</Btn>} />
      <div style={{padding:24,display:'flex',flexDirection:'column',gap:16}}>
        <div style={{display:'grid',gridTemplateColumns:'repeat(4,1fr)',gap:14}}>
          {[
            ['all','TOTAL PIPELINES', counts.all, <I.zap size={14}/>, 'var(--z-100)','var(--z-700)'],
            ['active','ACTIVE', counts.active, <I.check size={14}/>, 'var(--green-soft)','var(--green-ink)'],
            ['draft','DRAFT', counts.draft, <I.edit size={14}/>, 'var(--blue-soft)','var(--blue-ink)'],
            ['disabled','DISABLED', counts.disabled, <I.pause size={14}/>, 'var(--amber-soft)','var(--amber-ink)'],
          ].map(([k,l,v,ic,bg,fg])=>(
            <Metric key={k} label={l} value={v} icon={ic} iconBg={bg} iconColor={fg} active={filter===k} onClick={()=>setFilter(k)} />
          ))}
        </div>
        <div style={{display:'flex',gap:10,alignItems:'center',padding:'8px 10px',background:'var(--z-25)',border:'1px solid var(--z-150)',borderRadius:10}}>
          <div className="input-wrap" style={{flex:1,maxWidth:380}}>
            <I.search size={13}/>
            <input className="input input--with-icon" placeholder="搜索流水线名称或描述..." />
          </div>
          <div className="seg">
            {[['all','全部'],['active','已启用'],['draft','草稿'],['disabled','已停用']].map(([k,l])=>(
              <button key={k} className={filter===k?'is-on':''} onClick={()=>setFilter(k)}>{l}</button>
            ))}
          </div>
        </div>
        <div style={{display:'flex',flexDirection:'column',gap:8}}>
          {items.map(p=>{
            const bigTone = p.status==='active'?'green':p.status==='draft'?'blue':'amber';
            const bigIcon = p.status==='active'?<I.check size={15}/>:p.status==='draft'?<I.edit size={14}/>:<I.pause size={14}/>;
            return (
              <div key={p.id} className="card" style={{padding:'12px 14px', display:'flex', alignItems:'center', gap:14}}>
                <div style={{width:32,height:32,borderRadius:8,background:`var(--${bigTone}-soft)`, color:`var(--${bigTone}-ink)`, display:'flex', alignItems:'center', justifyContent:'center'}}>{bigIcon}</div>
                <div style={{flex:1,minWidth:0}}>
                  <div style={{display:'flex',alignItems:'center',gap:8}}>
                    <b style={{fontSize:13.5,fontWeight:600}}>{p.name}</b>
                    <StatusBadge status={p.status}/>
                  </div>
                  <div className="sub" style={{fontSize:11.5,marginTop:2}}>
                    <span className="tag" style={{marginRight:6}}>{p.trigger}</span>
                    {p.desc} · 更新 {p.updated}
                  </div>
                </div>
                <div style={{display:'flex',gap:4}}>
                  <Btn size="sm" variant="primary" icon={<I.play size={11}/>}>Run</Btn>
                  <Btn size="sm" variant="outline" icon={<I.clock size={12}/>}>运行历史</Btn>
                  <Btn size="sm" variant="ghost" icon={<I.more size={13}/>} iconOnly />
                  <Btn size="sm" variant="ghost" icon={<I.trash size={12}/>} iconOnly />
                </div>
              </div>
            );
          })}
        </div>
      </div>
    </>
  );
}

// ───────── Pipeline List — Card Grid variant ─────────
function PipelineListCard() {
  const [filter, setFilter] = React.useState('all');
  const items = PIPELINES.filter(p => filter==='all' || p.status===filter);
  return (
    <>
      <PageHeader crumb="Pipelines" title="Automated Pipelines" sub="Card grid · 更视觉化的浏览体验。"
        actions={<Btn size="sm" variant="primary" icon={<I.plus size={13}/>}>Create Pipeline</Btn>} />
      <div style={{padding:24,display:'flex',flexDirection:'column',gap:14}}>
        <div style={{display:'flex',gap:10,alignItems:'center',justifyContent:'space-between'}}>
          <div className="input-wrap" style={{width:320}}>
            <I.search size={13}/>
            <input className="input input--with-icon" placeholder="搜索流水线..." />
          </div>
          <div className="seg">
            {[['all','全部'],['active','已启用'],['draft','草稿'],['disabled','已停用']].map(([k,l])=>(
              <button key={k} className={filter===k?'is-on':''} onClick={()=>setFilter(k)}>{l}</button>
            ))}
          </div>
        </div>
        <div style={{display:'grid',gridTemplateColumns:'repeat(auto-fill, minmax(280px, 1fr))',gap:14}}>
          {items.map(p=>(
            <div key={p.id} className="card" style={{padding:14,display:'flex',flexDirection:'column',gap:10}}>
              <div style={{display:'flex', alignItems:'center', gap:8}}>
                <div style={{width:30,height:30,borderRadius:8,background:'var(--z-100)', color:'var(--z-700)',display:'flex', alignItems:'center', justifyContent:'center'}}><I.zap size={14}/></div>
                <div style={{flex:1,minWidth:0}}>
                  <div style={{fontSize:13,fontWeight:600,overflow:'hidden',textOverflow:'ellipsis',whiteSpace:'nowrap'}}>{p.name}</div>
                  <div className="sub" style={{fontSize:11}}>{p.trigger}</div>
                </div>
                <StatusBadge status={p.status}/>
              </div>
              <div className="sub" style={{fontSize:11.5, minHeight:32, overflow:'hidden', display:'-webkit-box', WebkitLineClamp:2, WebkitBoxOrient:'vertical'}}>{p.desc}</div>
              <div style={{display:'flex',alignItems:'center',justifyContent:'space-between',borderTop:'1px solid var(--z-150)',paddingTop:10}}>
                <span className="sub" style={{fontSize:11}}>{p.runs} runs · {p.updated}</span>
                <div style={{display:'flex',gap:4}}>
                  <Btn size="xs" variant="ghost" icon={<I.play size={11}/>}/>
                  <Btn size="xs" variant="ghost" icon={<I.edit size={11}/>}/>
                  <Btn size="xs" variant="ghost" icon={<I.more size={12}/>}/>
                </div>
              </div>
            </div>
          ))}
        </div>
      </div>
    </>
  );
}

// ───────── Template Select Wizard ─────────
function TemplateSelectPage() {
  const [step, setStep] = React.useState(1);
  const [tpl, setTpl] = React.useState('go');
  const templates = [
    { id:'custom', name:'Custom', desc:'从零开始配置', icon:'＋', custom:true },
    { id:'go', name:'Go', desc:'编译 · 测试 · 容器镜像', icon:'🔵', tint:220 },
    { id:'java', name:'Java Maven', desc:'mvn clean package', icon:'☕', tint:30 },
    { id:'node', name:'Node.js', desc:'pnpm + Next.js', icon:'🟢', tint:140 },
    { id:'docker', name:'Docker', desc:'Dockerfile 构建', icon:'🐳', tint:200 },
    { id:'jar', name:'Java JAR', desc:'现成产物打包', icon:'☕', tint:340 },
  ];
  return (
    <>
      <div style={{padding:'16px 24px 0'}}>
        <a style={{fontSize:12, color:'var(--z-500)', cursor:'pointer', display:'inline-flex', alignItems:'center', gap:4}}><I.arrL size={12}/>返回</a>
      </div>
      <PageHeader crumb="Create Pipeline" title="Architect New Pipeline" sub="Real-time topology preview. 从模板快速创建流水线，或从空白开始。" />
      <div style={{padding:24, maxWidth:960, margin:'0 auto'}}>
        {/* Stepper */}
        <div style={{display:'flex',alignItems:'center',justifyContent:'center',gap:14,marginBottom:28}}>
          {[[1,'Template Selection'],[2,'Configuration']].map(([n,l],i)=>(
            <React.Fragment key={n}>
              <div style={{display:'flex',alignItems:'center',gap:8}}>
                <div style={{width:26,height:26,borderRadius:13, display:'flex', alignItems:'center', justifyContent:'center',
                  background: step>=n ? 'linear-gradient(180deg, var(--accent-1), var(--accent-2))' : 'var(--z-100)',
                  color: step>=n?'#fff':'var(--z-500)', fontSize:12, fontWeight:600,
                  boxShadow: step===n ? '0 0 0 4px color-mix(in oklch, var(--accent-1), white 85%)' : 'none'}}>
                  {step>n ? <I.check size={12}/> : n}
                </div>
                <span style={{fontSize:12.5,fontWeight: step===n?600:500, color: step>=n?'var(--z-900)':'var(--z-500)'}}>{l}</span>
              </div>
              {i===0 && <div style={{flex:'0 0 64px', height:1, background:'var(--z-200)'}}/>}
            </React.Fragment>
          ))}
        </div>

        {step===1 && (
          <div style={{display:'grid',gridTemplateColumns:'repeat(4,1fr)',gap:14}}>
            {templates.map(t=>(
              <div key={t.id} onClick={()=>{ setTpl(t.id); if(!t.custom) setStep(2); }}
                className="card" style={{ padding:16, height:180, display:'flex', flexDirection:'column', gap:8, cursor:'pointer',
                  borderColor: tpl===t.id?'var(--accent-1)':'var(--z-200)', boxShadow: tpl===t.id?'0 0 0 3px color-mix(in oklch, var(--accent-1), white 85%)':'var(--shadow-xs)' }}>
                <div style={{fontSize:26}}>{t.icon}</div>
                <div style={{fontSize:14,fontWeight:600}}>{t.name}</div>
                <div className="sub" style={{fontSize:11.5}}>{t.desc}</div>
                {t.custom && <div style={{marginTop:'auto',fontSize:11, color:'var(--accent-ink)'}}>跳过模板 →</div>}
              </div>
            ))}
          </div>
        )}

        {step===2 && (
          <div style={{display:'grid',gridTemplateColumns:'14fr 10fr',gap:16}}>
            <Card title="Configuration Parameters">
              <div style={{display:'flex',flexDirection:'column',gap:12}}>
                <Field label="流水线名称" required><input className="input" defaultValue="build-web-console"/></Field>
                <Field label="描述"><input className="input" defaultValue="Next.js build + image push to harbor"/></Field>
                <div style={{borderTop:'1px solid var(--z-150)',margin:'4px 0'}}/>
                <div style={{fontSize:11, fontWeight:600, color:'var(--z-500)', textTransform:'uppercase', letterSpacing:'.06em'}}>Template Parameters</div>
                <Field label="repoUrl" required><input className="input mono" defaultValue="git@github.com:zcid/web-console.git"/></Field>
                <div style={{display:'grid',gridTemplateColumns:'1fr 1fr', gap:10}}>
                  <Field label="branch"><input className="input mono" defaultValue="main"/></Field>
                  <Field label="nodeVersion"><input className="input mono" defaultValue="20-alpine"/></Field>
                  <Field label="registry"><input className="input mono" defaultValue="harbor.zcid.local"/></Field>
                  <Field label="imageName"><input className="input mono" defaultValue="web/console"/></Field>
                </div>
                <div style={{display:'flex',gap:8, marginTop:4}}>
                  <Btn variant="primary">创建流水线</Btn>
                  <Btn onClick={()=>setStep(1)}>返回</Btn>
                </div>
              </div>
            </Card>
            <Card title="Real-time Stage Preview">
              <div style={{display:'flex',alignItems:'center',gap:6, marginBottom:14, overflow:'auto'}}>
                {[['1','Source','1 step'],['2','Build','3 steps'],['3','Package','1 step'],['4','Deploy','1 step']].map(([n,l,s],i,a)=>(
                  <React.Fragment key={n}>
                    <div style={{display:'flex',flexDirection:'column',alignItems:'center',gap:4, minWidth:72}}>
                      <div style={{width:28,height:28,borderRadius:14, background:'var(--z-100)', display:'flex', alignItems:'center', justifyContent:'center', fontSize:12, fontWeight:600, color:'var(--z-700)'}}>{n}</div>
                      <div style={{fontSize:11,fontWeight:500}}>{l}</div>
                      <div className="sub" style={{fontSize:10}}>{s}</div>
                    </div>
                    {i<a.length-1 && <div style={{flex:1, height:1, background:'var(--z-200)'}}/>}
                  </React.Fragment>
                ))}
              </div>
              <div className="codeblock">{`{
  "name": "build-web-console",
  "stages": [
    { "name": "Source", "steps": [
      { "type": "git-clone", "url": "\${repoUrl}" }
    ]},
    { "name": "Build", "steps": [
      { "type": "shell", "run": "pnpm install" },
      { "type": "shell", "run": "pnpm test" },
      { "type": "shell", "run": "pnpm build" }
    ]},
    { "name": "Package", "steps": [
      { "type": "kaniko-build", "tag": "\${registry}/\${imageName}:\${SHA}" }
    ]},
    { "name": "Deploy", "steps": [
      { "type": "shell", "run": "argocd app sync \${APP}" }
    ]}
  ]
}`}</div>
            </Card>
          </div>
        )}
      </div>
    </>
  );
}

// ───────── Pipeline Editor (Full-screen DAG) ─────────
const STEP_ICONS = {
  'git-clone': <I.branch size={12}/>,
  'shell': <I.terminal size={12}/>,
  'kaniko-build': <I.cube size={12}/>,
  'buildkit-build': <I.layers size={12}/>,
};
const STEP_TONES = {
  'git-clone': 'blue',
  'shell': 'grey',
  'kaniko-build': 'accent',
  'buildkit-build': 'cyan',
};

const STEP_PALETTE = [
  { type:'git-clone',     label:'Git Clone',     desc:'从 Git 仓库拉取源码' },
  { type:'shell',         label:'Shell',         desc:'运行 bash / sh 脚本' },
  { type:'kaniko-build',  label:'Kaniko Build',  desc:'无守护容器构建镜像' },
  { type:'buildkit-build',label:'BuildKit',      desc:'并发容器构建镜像' },
];

function PipelineEditorPage({ onClose }) {
  const [dag, setDag] = React.useState(INITIAL_DAG);
  const [mode, setMode] = React.useState('visual');
  const [selectedStep, setSelectedStep] = React.useState({ stageId: INITIAL_DAG.stages[1].id, stepId: INITIAL_DAG.stages[1].steps[2].id });
  const [name, setName] = React.useState(dag.name);
  const [zoom, setZoom] = React.useState(1);
  const [dragOverStage, setDragOverStage] = React.useState(null);
  const [dragStep, setDragStep] = React.useState(null); // {stageId, stepId}
  const [paletteDrag, setPaletteDrag] = React.useState(null); // type
  const [consoleOpen, setConsoleOpen] = React.useState(true);

  const stepCount = dag.stages.reduce((a,s)=>a+s.steps.length, 0);

  const addStage = () => {
    setDag(d => ({ ...d, stages: [...d.stages, { id:'s'+Date.now(), name: `Stage ${d.stages.length+1}`, steps: [] }] }));
  };
  const addStep = (stageId, type='shell') => {
    const id = 'st'+Date.now();
    setDag(d => ({ ...d, stages: d.stages.map(s => s.id===stageId ? { ...s, steps: [...s.steps, { id, name: `new-${type}`, type, config:{script:''} }] } : s) }));
    setSelectedStep({ stageId, stepId: id });
  };
  const removeStep = (stageId, stepId) => {
    setDag(d => ({ ...d, stages: d.stages.map(s => s.id===stageId ? { ...s, steps: s.steps.filter(x=>x.id!==stepId) } : s) }));
    setSelectedStep(cur => cur && cur.stepId===stepId ? null : cur);
  };
  const removeStage = (stageId) => {
    setDag(d => ({ ...d, stages: d.stages.filter(s=>s.id!==stageId) }));
  };
  const duplicateStep = (stageId, stepId) => {
    setDag(d => ({ ...d, stages: d.stages.map(s => {
      if (s.id !== stageId) return s;
      const idx = s.steps.findIndex(x=>x.id===stepId);
      if (idx<0) return s;
      const orig = s.steps[idx];
      const copy = { ...orig, id:'st'+Date.now(), name: orig.name+'-copy' };
      return { ...s, steps: [...s.steps.slice(0,idx+1), copy, ...s.steps.slice(idx+1)] };
    }) }));
  };
  const moveStep = (fromStageId, fromStepId, toStageId, toIndex) => {
    setDag(d => {
      let moved = null;
      const stages = d.stages.map(s => {
        if (s.id !== fromStageId) return s;
        const i = s.steps.findIndex(x=>x.id===fromStepId);
        if (i<0) return s;
        moved = s.steps[i];
        return { ...s, steps: [...s.steps.slice(0,i), ...s.steps.slice(i+1)] };
      });
      if (!moved) return d;
      return { ...d, stages: stages.map(s => s.id===toStageId ? { ...s, steps: [...s.steps.slice(0,toIndex), moved, ...s.steps.slice(toIndex)] } : s) };
    });
  };

  return (
    <div style={{position:'absolute', inset:0, background:'var(--z-0)', display:'flex', flexDirection:'column'}}>
      {/* Header */}
      <div style={{height:56, flex:'none', display:'flex', alignItems:'center', gap:12, padding:'0 16px',
        background:'rgba(255,255,255,.72)', backdropFilter:'blur(10px)', borderBottom:'1px solid var(--z-150)', zIndex:5}}>
        <Btn size="sm" variant="ghost" icon={<I.arrL size={13}/>} onClick={onClose} iconOnly/>
        <div style={{width:1,height:22,background:'var(--z-200)'}}/>
        <div style={{width:32,height:32,borderRadius:8,background:'linear-gradient(135deg,var(--accent-1),var(--accent-2))',color:'#fff',display:'flex',alignItems:'center',justifyContent:'center'}}>
          <I.code size={14}/>
        </div>
        <div style={{display:'flex',flexDirection:'column',gap:1, minWidth:0}}>
          <input value={name} onChange={e=>setName(e.target.value)} style={{border:0, outline:'none', background:'transparent', font:'600 15px / 1 var(--font-sans)', letterSpacing:'-0.01em', color:'var(--z-900)', width:'auto', minWidth:200, padding:0}}/>
          <span className="sub mono" style={{fontSize:10.5}}>Webhook · {dag.stages.length} stages · {stepCount} steps · <span style={{color:'var(--amber-ink)'}}>● 未保存</span></span>
        </div>
        <div style={{flex:1}}/>
        <Segmented value={mode} onChange={setMode} options={[{value:'visual',label:'可视化'},{value:'json',label:'JSON'}]}/>
        <div style={{width:1,height:22,background:'var(--z-200)'}}/>
        <Btn size="sm" variant="ghost" icon={<I.play size={12}/>}>试运行</Btn>
        <Btn size="sm" variant="ghost" icon={<I.settings size={13}/>} iconOnly/>
        <Btn size="sm" variant="outline">取消</Btn>
        <Btn size="sm" variant="primary" icon={<I.check size={13}/>}>保存</Btn>
      </div>

      {/* Main area */}
      <div style={{flex:1, display:'flex', minHeight:0}}>
        {/* Left palette */}
        {mode==='visual' && (
          <div style={{width:208, flex:'none', borderRight:'1px solid var(--z-150)', background:'var(--z-25)', display:'flex', flexDirection:'column'}}>
            <div style={{padding:'12px 14px 8px', borderBottom:'1px solid var(--z-150)'}}>
              <div style={{fontSize:10.5, fontWeight:600, letterSpacing:'.08em', color:'var(--z-500)', textTransform:'uppercase'}}>Step Palette</div>
              <div className="sub" style={{fontSize:11, marginTop:3}}>拖拽到 Stage 即可添加</div>
            </div>
            <div style={{padding:10, display:'flex', flexDirection:'column', gap:6, flex:1, overflow:'auto'}}>
              {STEP_PALETTE.map(p => {
                const tone = STEP_TONES[p.type];
                return (
                  <div key={p.type} draggable onDragStart={()=>setPaletteDrag(p.type)} onDragEnd={()=>setPaletteDrag(null)}
                    style={{padding:'8px 10px', border:'1px solid var(--z-200)', background:'var(--z-0)', borderRadius:7, cursor:'grab', boxShadow:'var(--shadow-xs)'}}>
                    <div style={{display:'flex',alignItems:'center',gap:8}}>
                      <div style={{width:22,height:22,borderRadius:5, background:`var(--${tone==='grey'?'z-100':tone+'-soft'})`, color:`var(--${tone==='grey'?'z-700':tone+'-ink'})`, display:'flex',alignItems:'center',justifyContent:'center'}}>
                        {STEP_ICONS[p.type]}
                      </div>
                      <div style={{fontSize:12, fontWeight:500}}>{p.label}</div>
                      <I.drag size={11} style={{marginLeft:'auto', color:'var(--z-400)'}}/>
                    </div>
                    <div className="sub" style={{fontSize:10.5, marginTop:3}}>{p.desc}</div>
                  </div>
                );
              })}
            </div>
            <div style={{padding:10, borderTop:'1px solid var(--z-150)'}}>
              <div style={{fontSize:10.5, fontWeight:600, letterSpacing:'.08em', color:'var(--z-500)', textTransform:'uppercase', marginBottom:6}}>Variables</div>
              <div style={{display:'flex',flexDirection:'column',gap:4}}>
                {['${BRANCH}', '${SHA}', '${REGISTRY}', '${IMAGE}'].map(v=>(
                  <span key={v} className="tag" style={{justifyContent:'flex-start', fontSize:10.5}}>{v}</span>
                ))}
              </div>
            </div>
          </div>
        )}

        {/* Canvas */}
        <div style={{flex:1, display:'flex', flexDirection:'column', minWidth:0, position:'relative'}}>
          <div style={{flex:1, overflow:'auto', position:'relative',
            background:`var(--z-25) url("data:image/svg+xml;utf8,<svg xmlns='http://www.w3.org/2000/svg' width='20' height='20'><circle cx='1' cy='1' r='1' fill='%23d3d3d7'/></svg>")`}}>
            {mode==='visual' ? (
              <div style={{padding:'44px 44px 100px', transform:`scale(${zoom})`, transformOrigin:'0 0', minWidth:'max-content', minHeight:'100%'}}>
                <div style={{display:'flex', gap:18, alignItems:'flex-start', position:'relative'}}>
                  {dag.stages.map((stage, si)=>(
                    <React.Fragment key={stage.id}>
                      <div
                        onDragOver={e=>{ e.preventDefault(); setDragOverStage(stage.id); }}
                        onDragLeave={()=>setDragOverStage(cur => cur===stage.id?null:cur)}
                        onDrop={e=>{
                          e.preventDefault();
                          if (paletteDrag) { addStep(stage.id, paletteDrag); setPaletteDrag(null); }
                          else if (dragStep) { moveStep(dragStep.stageId, dragStep.stepId, stage.id, stage.steps.length); setDragStep(null); }
                          setDragOverStage(null);
                        }}
                        style={{display:'flex', flexDirection:'column', gap:8, width:234,
                          padding:10, borderRadius:10,
                          background: dragOverStage===stage.id?'color-mix(in oklch, var(--accent-1), white 90%)':'rgba(255,255,255,.4)',
                          border: `1.5px ${dragOverStage===stage.id?'dashed':'solid'} ${dragOverStage===stage.id?'var(--accent-1)':'var(--z-200)'}`,
                          transition:'background .12s'}}>
                        {/* Stage header */}
                        <div style={{display:'flex',alignItems:'center',gap:7, padding:'4px 4px'}}>
                          <div style={{width:20,height:20,borderRadius:10,background:'linear-gradient(135deg,var(--accent-1),var(--accent-2))',color:'#fff',display:'flex',alignItems:'center',justifyContent:'center', fontSize:10.5, fontWeight:600, boxShadow:'0 1px 3px rgba(0,0,0,.12)'}}>{si+1}</div>
                          <input defaultValue={stage.name} style={{border:0,outline:'none',background:'transparent',font:'600 13px var(--font-sans)', color:'var(--z-900)', flex:1, minWidth:0, padding:0}}/>
                          <span className="tag" style={{fontSize:10}}>{stage.steps.length} step{stage.steps.length!==1?'s':''}</span>
                          <Btn size="xs" variant="ghost" icon={<I.trash size={10}/>} iconOnly onClick={()=>removeStage(stage.id)}/>
                        </div>

                        {/* Steps */}
                        {stage.steps.length===0 && (
                          <div style={{padding:'18px 10px', textAlign:'center', border:'1.5px dashed var(--z-200)', borderRadius:7, color:'var(--z-400)', fontSize:11}}>
                            拖拽步骤至此
                          </div>
                        )}
                        {stage.steps.map((step)=>{
                          const sel = selectedStep && selectedStep.stageId===stage.id && selectedStep.stepId===step.id;
                          const tone = STEP_TONES[step.type] || 'grey';
                          return (
                            <div key={step.id}
                              draggable
                              onDragStart={()=>setDragStep({stageId:stage.id, stepId:step.id})}
                              onDragEnd={()=>setDragStep(null)}
                              onClick={()=>setSelectedStep({stageId:stage.id, stepId:step.id})}
                              className="card"
                              style={{ padding:'10px 11px', cursor:'pointer', display:'flex', alignItems:'flex-start', gap:9, background:'var(--z-0)',
                                borderColor: sel?'var(--accent-1)':'var(--z-200)',
                                boxShadow: sel?'0 0 0 3px color-mix(in oklch, var(--accent-1), white 85%), 0 2px 6px rgba(0,0,0,.04)':'var(--shadow-xs)',
                                transform: dragStep && dragStep.stepId===step.id ? 'scale(0.98)' : 'none',
                                opacity: dragStep && dragStep.stepId===step.id ? 0.5 : 1,
                                transition:'box-shadow .12s, border-color .12s'}}>
                              <div style={{width:26,height:26,borderRadius:6, background:`var(--${tone==='grey'?'z-100':tone+'-soft'})`, color:`var(--${tone==='grey'?'z-700':tone+'-ink'})`, display:'flex', alignItems:'center', justifyContent:'center', flex:'none'}}>
                                {STEP_ICONS[step.type]}
                              </div>
                              <div style={{flex:1, minWidth:0}}>
                                <div style={{display:'flex', alignItems:'center', gap:4}}>
                                  <div style={{fontSize:12, fontWeight:500, overflow:'hidden', textOverflow:'ellipsis', whiteSpace:'nowrap'}}>{step.name}</div>
                                </div>
                                <div className="sub mono" style={{fontSize:10, marginTop:1}}>{step.type}</div>
                                {step.type==='shell' && step.config.script && (
                                  <div className="mono" style={{fontSize:10, color:'var(--z-500)', marginTop:4, overflow:'hidden', textOverflow:'ellipsis', whiteSpace:'nowrap', background:'var(--z-50)', padding:'2px 5px', borderRadius:3}}>
                                    $ {step.config.script}
                                  </div>
                                )}
                              </div>
                              <div style={{display:'flex',flexDirection:'column',gap:2}}>
                                <Btn size="xs" variant="ghost" icon={<I.copy size={10}/>} iconOnly onClick={(e)=>{e.stopPropagation(); duplicateStep(stage.id, step.id);}}/>
                                <Btn size="xs" variant="ghost" icon={<I.x size={10}/>} iconOnly onClick={(e)=>{e.stopPropagation(); removeStep(stage.id, step.id);}}/>
                              </div>
                            </div>
                          );
                        })}
                        <button className="btn btn--outline btn--xs" style={{justifyContent:'center', borderStyle:'dashed', color:'var(--z-500)', height:28}}
                          onClick={()=>addStep(stage.id)}>
                          <I.plus size={11}/> Add Step
                        </button>
                      </div>
                      {si<dag.stages.length-1 && (
                        <div style={{alignSelf:'stretch', display:'flex', alignItems:'center', minWidth:32}}>
                          <svg width="32" height="80" viewBox="0 0 32 80" style={{overflow:'visible'}}>
                            <defs>
                              <marker id={`arr-${si}`} viewBox="0 0 10 10" refX="8" refY="5" markerWidth="6" markerHeight="6" orient="auto">
                                <path d="M 0 0 L 10 5 L 0 10 Z" fill="var(--z-400)"/>
                              </marker>
                            </defs>
                            <path d="M 0 40 C 10 40, 22 40, 30 40" stroke="var(--z-300)" strokeWidth="1.5" fill="none" markerEnd={`url(#arr-${si})`}/>
                          </svg>
                        </div>
                      )}
                    </React.Fragment>
                  ))}
                  <div style={{alignSelf:'flex-start', paddingTop:8}}>
                    <button className="btn btn--outline btn--sm" style={{borderStyle:'dashed', color:'var(--z-500)', width:148, justifyContent:'center', height:40, borderRadius:10}} onClick={addStage}>
                      <I.plus size={12}/> Add Stage
                    </button>
                  </div>
                </div>
              </div>
            ) : (
              <div style={{padding:24}}>
                <div style={{background:'#0b0b0d', color:'#e4e4e7', borderRadius:10, overflow:'hidden', boxShadow:'var(--shadow-md)'}}>
                  <div style={{display:'flex',alignItems:'center',gap:8, padding:'8px 12px', borderBottom:'1px solid #1f1f23', fontSize:11.5, color:'#a1a1aa', fontFamily:'var(--font-mono)'}}>
                    <I.code size={12}/> pipeline.json
                    <span style={{marginLeft:'auto', color:'#52525b'}}>YAML / JSON · Monaco</span>
                  </div>
                  <div style={{display:'flex', fontFamily:'var(--font-mono)', fontSize:12.5, lineHeight:1.65}}>
                    <div style={{padding:'14px 0 14px 14px', color:'#3f3f46', textAlign:'right', userSelect:'none', borderRight:'1px solid #1f1f23'}}>
                      {Array.from({length: JSON.stringify(dag,null,2).split('\n').length}).map((_,i)=>(<div key={i} style={{padding:'0 10px'}}>{i+1}</div>))}
                    </div>
                    <pre style={{margin:0, padding:14, flex:1, whiteSpace:'pre', overflow:'auto'}}>{JSON.stringify(dag, null, 2)}</pre>
                  </div>
                </div>
              </div>
            )}
          </div>

          {/* Zoom + minimap controls */}
          {mode==='visual' && (
            <>
              <div style={{position:'absolute', right:16, top:16, display:'flex', flexDirection:'column', gap:4, background:'var(--z-0)', border:'1px solid var(--z-200)', borderRadius:7, padding:3, boxShadow:'var(--shadow-sm)'}}>
                <Btn size="xs" variant="ghost" icon={<I.plus size={12}/>} iconOnly onClick={()=>setZoom(z=>Math.min(1.6, z+0.1))}/>
                <div style={{fontSize:10, textAlign:'center', color:'var(--z-500)', fontVariantNumeric:'tabular-nums'}}>{Math.round(zoom*100)}%</div>
                <Btn size="xs" variant="ghost" icon={<I.minus size={12}/>} iconOnly onClick={()=>setZoom(z=>Math.max(0.5, z-0.1))}/>
                <div style={{height:1, background:'var(--z-150)', margin:'2px 0'}}/>
                <Btn size="xs" variant="ghost" icon={<I.target size={11}/>} iconOnly onClick={()=>setZoom(1)}/>
              </div>

            </>
          )}

          {/* Bottom console */}
          {mode==='visual' && (
            <div style={{flex:'none', borderTop:'1px solid var(--z-150)', background:'var(--z-0)'}}>
              <div style={{display:'flex', alignItems:'center', gap:10, padding:'7px 14px', fontSize:11.5, cursor:'pointer'}} onClick={()=>setConsoleOpen(v=>!v)}>
                <I.terminal size={12}/>
                <b style={{fontWeight:600}}>Validation</b>
                <span className="badge badge--green" style={{height:17}}><span className="dot" style={{background:'var(--green)'}}/>DAG OK</span>
                <span className="badge badge--amber" style={{height:17}}>2 warnings</span>
                <div style={{flex:1}}/>
                <span className="sub" style={{fontSize:11}}>lint · pipeline.schema v2</span>
                <I.chevD size={12} style={{transform: consoleOpen?'rotate(180deg)':'none', transition:'transform .12s'}}/>
              </div>
              {consoleOpen && (
                <div style={{padding:'4px 14px 12px', display:'flex', flexDirection:'column', gap:4, fontFamily:'var(--font-mono)', fontSize:11.5}}>
                  <div style={{display:'flex',gap:10, color:'var(--amber-ink)'}}><span>⚠</span><span>Build → run-tests</span><span style={{color:'var(--z-500)'}}>超时未设置，建议 ≤ 10m</span></div>
                  <div style={{display:'flex',gap:10, color:'var(--amber-ink)'}}><span>⚠</span><span>Package → kaniko-build</span><span style={{color:'var(--z-500)'}}>未显式指定 context，默认使用 repo root</span></div>
                  <div style={{display:'flex',gap:10, color:'var(--z-500)'}}><span>ℹ</span><span>Deploy → argocd-sync</span><span>依赖 Package 输出镜像 tag</span></div>
                </div>
              )}
            </div>
          )}
        </div>

        {/* Right config panel */}
        {mode==='visual' && selectedStep && (() => {
          const stage = dag.stages.find(s=>s.id===selectedStep.stageId);
          const step = stage && stage.steps.find(st=>st.id===selectedStep.stepId);
          if (!step) return null;
          return (
            <div style={{width:320, flex:'none', borderLeft:'1px solid var(--z-150)', background:'var(--z-0)', display:'flex', flexDirection:'column'}}>
              <div style={{padding:'12px 14px', borderBottom:'1px solid var(--z-150)', display:'flex', alignItems:'center', gap:8}}>
                <div style={{width:22,height:22,borderRadius:5,background:'var(--z-100)',color:'var(--z-700)',display:'flex',alignItems:'center',justifyContent:'center'}}>{STEP_ICONS[step.type]}</div>
                <b style={{fontSize:13}}>{step.name}</b>
                <span className="tag" style={{marginLeft:'auto'}}>{step.type}</span>
              </div>
              <div style={{padding:14, display:'flex', flexDirection:'column', gap:12, flex:1, overflow:'auto'}}>
                <Field label="Step Name"><input className="input" defaultValue={step.name}/></Field>
                <Field label="Type">
                  <select className="input" defaultValue={step.type}>
                    <option>shell</option><option>git-clone</option><option>kaniko-build</option><option>buildkit-build</option>
                  </select>
                </Field>
                {step.type==='shell' && (
                  <Field label="Script">
                    <div style={{background:'#0b0b0d', color:'#e4e4e7', borderRadius:7, padding:10, fontFamily:'var(--font-mono)', fontSize:12, lineHeight:1.5, minHeight:140}}>
                      <div><span style={{color:'#71717a'}}>#!/usr/bin/env bash</span></div>
                      <div><span style={{color:'#a78bfa'}}>set</span> -euo pipefail</div>
                      <div><span style={{color:'#60a5fa'}}>echo</span> <span style={{color:'#fbbf24'}}>"⚡ running: {step.name}"</span></div>
                      <div>{step.config.script || 'pnpm install --frozen-lockfile'}</div>
                      <div style={{color:'#71717a'}}># exit $?</div>
                    </div>
                  </Field>
                )}
                {step.type==='git-clone' && (
                  <>
                    <Field label="Repository URL"><input className="input mono" defaultValue={step.config.url || 'github.com/zcid/web-console'}/></Field>
                    <Field label="Branch / Ref"><input className="input mono" defaultValue={step.config.branch || 'main'}/></Field>
                  </>
                )}
                {step.type==='kaniko-build' && (
                  <>
                    <Field label="Dockerfile"><input className="input mono" defaultValue={step.config.dockerfile || './Dockerfile'}/></Field>
                    <Field label="Image Tag"><input className="input mono" defaultValue={step.config.tag || '${REGISTRY}/app:${SHA}'}/></Field>
                  </>
                )}
                <Field label="Timeout"><input className="input" defaultValue="10m"/></Field>
              </div>
              <div style={{padding:'10px 14px', borderTop:'1px solid var(--z-150)', display:'flex', gap:8}}>
                <Btn size="sm" variant="danger" icon={<I.trash size={12}/>}>删除</Btn>
                <div style={{flex:1}}/>
                <Btn size="sm" variant="outline" onClick={()=>setSelectedStep(null)}>关闭</Btn>
                <Btn size="sm" variant="primary">应用</Btn>
              </div>
            </div>
          );
        })()}
      </div>
    </div>
  );
}

// ───────── Run List ─────────
function PipelineRunListPage({ onOpen }) {
  return (
    <>
      <PageHeader crumb="Pipelines · build-web-console" title="build-web-console — 运行历史" sub="所有触发的运行记录。"
        actions={<>
          <Input placeholder="Commit SHA..." style={{ width: 160 }} icon={<I.gitCommit size={13}/>} />
          <Select width={110} value="all" options={[{value:'all',label:'所有状态'},'pending','queued','running','succeeded','failed','cancelled']}/>
          <Select width={110} value="all" options={[{value:'all',label:'所有触发'},'Manual','Webhook','Cron']}/>
          <Btn size="sm" variant="primary" icon={<I.play size={13}/>}>触发运行</Btn>
        </>} />
      <div style={{padding:24}}>
        <Card padding={false}>
          <table className="table">
            <thead><tr><th>#</th><th>状态</th><th>触发</th><th>用户</th><th>分支</th><th>耗时</th><th>时间</th><th style={{textAlign:'right'}}>操作</th></tr></thead>
            <tbody>
              {RUNS.map(r=>(
                <tr key={r.n}>
                  <td><b className="mono" style={{fontSize:12, fontWeight:600}}>#{r.n}</b></td>
                  <td><StatusBadge status={r.status}/></td>
                  <td><span className="tag">{r.trigger}</span></td>
                  <td><span className="mono" style={{fontSize:11.5}}>{r.user}</span></td>
                  <td><span style={{display:'inline-flex',alignItems:'center',gap:5}}><I.branch size={11}/><span className="mono" style={{fontSize:11.5}}>{r.branch}</span></span></td>
                  <td><span className="mono sub" style={{fontSize:11.5}}>{r.duration ? dur(r.duration) : '—'}</span></td>
                  <td><span className="sub">{r.time}</span></td>
                  <td style={{textAlign:'right'}}>
                    <div style={{display:'inline-flex',gap:4}}>
                      <Btn size="xs" variant="outline" onClick={()=>onOpen && onOpen(r)}>详情</Btn>
                      {r.status==='running' && <Btn size="xs" variant="danger">取消</Btn>}
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

// ───────── Run Detail — variant A (horizontal stage bar) ─────────
function PipelineRunDetailA({ run, onBack }) {
  const r = run || RUNS[0];
  const isRunning = r.status==='running';
  const stages = [
    { name: 'Source Checkout', state: 'done' },
    { name: 'Build Artifacts', state: 'done' },
    { name: 'Image Push',      state: isRunning ? 'active' : 'done' },
    { name: 'K8s Deployment',  state: isRunning ? 'pending' : 'done' },
  ];
  return (
    <>
      <div style={{padding:'16px 24px 0'}}>
        <a onClick={onBack} style={{fontSize:12, color:'var(--z-500)', cursor:'pointer', display:'inline-flex', alignItems:'center', gap:4}}><I.arrL size={12}/>返回运行历史</a>
      </div>
      <PageHeader crumb="Build Observation" title={`Build #${r.n}`} sub={<span>triggered by <b>{r.user}</b> · duration {dur(r.duration||0)}</span>}
        actions={<>
          {isRunning && <Btn size="sm" variant="danger" icon={<I.x size={12}/>}>取消运行</Btn>}
          <StatusBadge status={r.status}/>
        </>} />
      <div style={{padding:24, maxWidth:1000, display:'flex', flexDirection:'column', gap:16}}>
        <StreamingStageBar stages={stages} isRunning={isRunning}/>
        <div style={{display:'grid', gridTemplateColumns:'repeat(4,1fr)', gap:14}}>
          {[
            ['TRIGGER', r.trigger, <I.zap size={13}/>],
            ['BRANCH', r.branch, <I.branch size={13}/>],
            ['TRIGGERED BY', r.user, <I.user size={13}/>],
            ['COMMIT', r.commit, <I.gitCommit size={13}/>],
          ].map(([l,v,ic])=>(
            <div key={l} style={{padding:'10px 12px', border:'1px solid var(--z-200)', borderRadius:8, background:'var(--z-0)'}}>
              <div style={{display:'flex', alignItems:'center', gap:6, color:'var(--z-500)', fontSize:10.5, fontWeight:600, letterSpacing:'.08em'}}>{ic}<span>{l}</span></div>
              <div className="mono" style={{fontSize:13, fontWeight:500, marginTop:4}}>{v}</div>
            </div>
          ))}
        </div>
        <Card title="Run Details">
          <div style={{display:'grid', gridTemplateColumns:'120px 1fr 120px 1fr', rowGap:10, columnGap:16, fontSize:12.5}}>
            <div style={{color:'var(--z-500)'}}>开始时间</div><div className="mono">2025-04-24 10:02:51</div>
            <div style={{color:'var(--z-500)'}}>结束时间</div><div className="mono">{isRunning?'—':'2025-04-24 10:08:03'}</div>
            <div style={{color:'var(--z-500)'}}>Tekton Name</div><div className="mono">pr-build-web-k9x2m</div>
            <div style={{color:'var(--z-500)'}}>Namespace</div><div className="mono">zcid-runners</div>
          </div>
          <div style={{borderTop:'1px dashed var(--z-200)', margin:'14px 0'}}/>
          <div style={{color:'var(--z-500)',fontSize:11.5, marginBottom:6}}>运行参数</div>
          <div style={{display:'flex',gap:6, flexWrap:'wrap'}}>
            <span className="tag">BRANCH={r.branch}</span>
            <span className="tag">SHA={r.commit}</span>
            <span className="tag">REGISTRY=harbor.zcid.local</span>
            <span className="tag">IMAGE=web/console</span>
            <span className="tag">KUBE_VERSION=1.30.4</span>
          </div>
          {r.status==='failed' && (
            <>
              <div style={{borderTop:'1px dashed var(--z-200)', margin:'14px 0'}}/>
              <div style={{color:'var(--z-500)',fontSize:11.5, marginBottom:6}}>错误信息</div>
              <div style={{background:'var(--red-soft)', color:'var(--red-ink)', borderRadius:6, padding:'10px 12px', fontFamily:'var(--font-mono)', fontSize:11.5}}>
                Error: step <b>run-tests</b> exited with code 1 · 3 of 842 tests failed (see log for details)
              </div>
            </>
          )}
        </Card>
        <StreamingLogBlock isRunning={isRunning} />
        <Card title="Build Artifacts">
          <div style={{display:'flex', gap:8, flexWrap:'wrap'}}>
            <span className="tag" style={{padding:'5px 10px', cursor:'pointer'}}><I.file size={11}/> coverage.xml · 42 KB</span>
            <span className="tag" style={{padding:'5px 10px', cursor:'pointer'}}><I.file size={11}/> sbom.json · 214 KB</span>
            <span className="tag" style={{padding:'5px 10px', cursor:'pointer'}}><I.file size={11}/> next-build.tar.gz · 18.3 MB</span>
          </div>
        </Card>
      </div>
    </>
  );
}

// Horizontal stage progress bar
function StreamingStageBar({ stages, isRunning }) {
  return (
    <div className="card" style={{padding:'16px 18px'}}>
      <div style={{display:'flex',alignItems:'center',gap:8}}>
        {stages.map((s,i,a)=>(
          <React.Fragment key={s.name}>
            <div style={{flex:1, display:'flex', flexDirection:'column', alignItems:'center', gap:6}}>
              <div style={{position:'relative', width:26, height:26, borderRadius:13, display:'flex', alignItems:'center', justifyContent:'center',
                background: s.state==='done' ? 'var(--green-soft)' : s.state==='active' ? 'var(--accent-soft)' : 'var(--z-100)',
                color: s.state==='done' ? 'var(--green-ink)' : s.state==='active' ? 'var(--accent-ink)' : 'var(--z-500)',
                boxShadow: s.state==='active' ? '0 0 0 4px color-mix(in oklch, var(--accent-1), white 85%)' : 'none' }}>
                {s.state==='done' ? <I.check size={13}/> : s.state==='active' ? <span className="st-dot st-dot--blue is-pulse" style={{width:10,height:10}}/> : (i+1)}
              </div>
              <div style={{fontSize:11.5, fontWeight: s.state==='active'?600:500}}>{s.name}</div>
              <div className="sub" style={{fontSize:10.5}}>{s.state==='done'?'Completed': s.state==='active'?'Active':'Pending'}</div>
            </div>
            {i<a.length-1 && <div style={{flex:'0 0 40px', height:2, background: a[i].state==='done' ? 'var(--green)' : 'var(--z-200)', borderRadius:1}}/>}
          </React.Fragment>
        ))}
      </div>
    </div>
  );
}

// Live-streaming log block
function StreamingLogBlock({ isRunning }) {
  const [lines, setLines] = React.useState(() => isRunning ? BUILD_LOG.slice(0,8) : BUILD_LOG);
  React.useEffect(() => {
    if (!isRunning) return;
    const t = setInterval(() => {
      setLines(l => {
        if (l.length >= BUILD_LOG.length) return l;
        const next = BUILD_LOG[l.length];
        return next ? [...l, next] : l;
      });
    }, 700);
    return () => clearInterval(t);
  }, [isRunning]);
  const scrollRef = React.useRef(null);
  React.useEffect(()=>{ if (scrollRef.current) scrollRef.current.scrollTop = scrollRef.current.scrollHeight; }, [lines]);

  return (
    <div className="card" style={{padding:0, overflow:'hidden'}}>
      <div style={{background:'#0b0b0d', color:'#e4e4e7', padding:'9px 14px', display:'flex', alignItems:'center', gap:10, borderBottom:'1px solid #1f1f23'}}>
        <div style={{display:'flex', gap:5}}>
          <span style={{width:10,height:10,borderRadius:5, background:'#ff5f57'}}/>
          <span style={{width:10,height:10,borderRadius:5, background:'#febc2e'}}/>
          <span style={{width:10,height:10,borderRadius:5, background:'#28c840'}}/>
        </div>
        <div style={{fontSize:12, fontWeight:500, fontFamily:'var(--font-mono)'}}>Build Output</div>
        {isRunning && <span style={{fontSize:10.5, color:'#71717a', display:'inline-flex', alignItems:'center', gap:5}}><span className="st-dot st-dot--green is-pulse"/>streaming</span>}
        <div style={{flex:1}}/>
        <button style={{background:'transparent', border:0, color:'#71717a', cursor:'pointer', padding:4}}><I.download size={13}/></button>
      </div>
      <div ref={scrollRef} style={{background:'#0b0b0d', color:'#e4e4e7', padding:'12px 0', fontFamily:'var(--font-mono)', fontSize:11.5, lineHeight:1.65, height:260, overflow:'auto'}}>
        {lines.length===0 && (
          <div style={{display:'flex',alignItems:'center',justifyContent:'center',height:'100%',color:'#52525b',fontSize:13}}>
            <span style={{marginRight:8}}>&gt;_</span> Waiting for build output...
          </div>
        )}
        {lines.map((ln, i)=> ln ? (
          <div key={i} style={{display:'flex', gap:14, padding:'0 14px'}}>
            <span style={{color:'#3f3f46', width:28, textAlign:'right', flex:'none', userSelect:'none'}}>{i+1}</span>
            <span style={{color: ln.lvl==='error'?'#fb7185':ln.lvl==='warn'?'#fbbf24':'#e4e4e7'}}>{ln.text}</span>
          </div>
        ) : null)}
        {isRunning && (
          <div style={{display:'flex', gap:14, padding:'0 14px'}}>
            <span style={{color:'#3f3f46', width:28, textAlign:'right', flex:'none'}}>{lines.length+1}</span>
            <span style={{width:7, height:13, background:'#e4e4e7', display:'inline-block', animation:'blink 1s steps(1) infinite'}}/>
          </div>
        )}
      </div>
      <style>{`@keyframes blink{50%{opacity:0}}`}</style>
    </div>
  );
}

// ───────── Run Detail — variant B (vertical timeline) ─────────
function PipelineRunDetailB({ run, onBack }) {
  const r = run || RUNS[0];
  const isRunning = r.status==='running';
  const stages = [
    { name:'Source Checkout', state:'done',   duration:'2.3s', log:'cloned github.com/zcid/web-console @ 8f3c1a29' },
    { name:'Build Artifacts', state:'done',   duration:'32.5s', log:'pnpm build → 118 pages generated' },
    { name:'Image Push',      state: isRunning ? 'active' : 'done', duration: isRunning ? '…' : '12.8s', log:'kaniko → harbor.local/web/console:2.4.0-rc.4' },
    { name:'K8s Deployment',  state: isRunning ? 'pending' : 'done', duration: isRunning ? '—' : '8.1s', log:'argocd sync web-console-staging' },
  ];
  return (
    <>
      <div style={{padding:'16px 24px 0'}}>
        <a onClick={onBack} style={{fontSize:12, color:'var(--z-500)', cursor:'pointer', display:'inline-flex', alignItems:'center', gap:4}}><I.arrL size={12}/>返回运行历史</a>
      </div>
      <PageHeader crumb="Build Observation · Vertical Timeline" title={`Build #${r.n}`} sub={<span>triggered by <b>{r.user}</b> · commit <span className="mono">{r.commit}</span></span>}
        actions={<><Btn size="sm" icon={<I.refresh size={13}/>}>Refresh</Btn><StatusBadge status={r.status}/></>} />
      <div style={{padding:24, display:'grid', gridTemplateColumns:'320px 1fr', gap:20, maxWidth:1100}}>
        {/* Vertical timeline */}
        <div className="card" style={{padding:16}}>
          <h2 style={{marginBottom:12}}>Stages</h2>
          <div style={{position:'relative', paddingLeft:22}}>
            <div style={{position:'absolute', left:10, top:10, bottom:10, width:2, background:'var(--z-150)', borderRadius:1}}/>
            {stages.map((s,i)=>(
              <div key={s.name} style={{position:'relative', paddingBottom: i<stages.length-1?20:0}}>
                <div style={{position:'absolute', left:-16, top:4, width:14, height:14, borderRadius:7,
                  background: s.state==='done' ? 'var(--green)' : s.state==='active' ? 'var(--accent-1)' : 'var(--z-0)',
                  border: s.state==='pending' ? '2px solid var(--z-200)' : 'none',
                  boxShadow: s.state==='active' ? '0 0 0 4px color-mix(in oklch, var(--accent-1), white 85%)' : 'none',
                  display:'flex', alignItems:'center', justifyContent:'center', color:'#fff'}}>
                  {s.state==='done' && <I.check size={8} strokeWidth={3}/>}
                </div>
                <div style={{display:'flex', justifyContent:'space-between', gap:8}}>
                  <div style={{fontSize:12.5, fontWeight:600}}>{s.name}</div>
                  <div className="sub mono" style={{fontSize:11}}>{s.duration}</div>
                </div>
                <div className="sub" style={{fontSize:11.5, marginTop:2}}>{s.log}</div>
                {s.state==='active' && (
                  <div className="pbar" style={{marginTop:8}}><i style={{width:'42%', animation:'bpulse 2s ease-in-out infinite'}}/></div>
                )}
              </div>
            ))}
          </div>
          <style>{`@keyframes bpulse{0%,100%{width:35%}50%{width:62%}}`}</style>
        </div>

        {/* Right column — metadata + logs */}
        <div style={{display:'flex',flexDirection:'column',gap:16}}>
          <div style={{display:'grid', gridTemplateColumns:'repeat(4,1fr)', gap:14}}>
            {[
              ['TRIGGER', r.trigger],
              ['BRANCH', r.branch],
              ['TRIGGERED BY', r.user],
              ['COMMIT', r.commit],
            ].map(([l,v])=>(
              <div key={l} style={{padding:'10px 12px', border:'1px solid var(--z-200)', borderRadius:8, background:'var(--z-0)'}}>
                <div style={{color:'var(--z-500)', fontSize:10.5, fontWeight:600, letterSpacing:'.08em'}}>{l}</div>
                <div className="mono" style={{fontSize:12.5, fontWeight:500, marginTop:4}}>{v}</div>
              </div>
            ))}
          </div>
          <StreamingLogBlock isRunning={isRunning} />
        </div>
      </div>
    </>
  );
}

Object.assign(window, { PipelineListRow, PipelineListCard, TemplateSelectPage, PipelineEditorPage, PipelineRunListPage, PipelineRunDetailA, PipelineRunDetailB });
