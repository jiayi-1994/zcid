// Main app — routes, design canvas with all artboards, tweaks panel
const {
  DesignCanvas, DCSection, DCArtboard,
  ChromeWindow,
  TweaksPanel, useTweaks, TweakSection, TweakSlider, TweakColor,
  AppShell, ProjectShell,
  I, Btn,
  // public
  LoginSplit, LoginMinimal, ForbiddenPage, NotFoundPage,
  // dashboards & admin
  DashboardA, DashboardB, DashboardC, ProjectListPage,
  AdminUsersPage, AdminVariablePage, IntegrationsPage, AuditLogPage, SystemSettingsPage,
  // project subpages
  EnvironmentListPage, DeploymentListPage, DeploymentDetailPage, ServiceListPage,
  MemberListPage, VariableListPage, NotificationRulesPage,
  // pipeline
  PipelineListRow, PipelineListCard, TemplateSelectPage, PipelineEditorPage,
  PipelineRunListPage, PipelineRunDetailA, PipelineRunDetailB,
  PROJECTS,
} = window;

const TWEAK_DEFAULTS = /*EDITMODE-BEGIN*/{
  "accentHue": 220
}/*EDITMODE-END*/;

// ───────── Wrappers with stable AppShell around each page ─────────
const AppWrap = (active, crumbs, node) => (
  <AppShell active={active} crumbs={crumbs}>{node}</AppShell>
);
const ProjWrap = (projActive, crumbs, node, projectId = 'web-console') => {
  const project = PROJECTS.find(p => p.id === projectId) || PROJECTS[0];
  return <ProjectShell projectActive={projActive} project={project} crumbs={crumbs}>{node}</ProjectShell>;
};

// ───────── Interactive "live demo" host with real routing ─────────
function LiveDemo() {
  const [route, setRoute] = React.useState({ app: 'dashboard', project: null, projectRoute: null, pipelineRun: null, showEditor: false });

  const go = (app) => setRoute({ app, project: null, projectRoute: null, pipelineRun: null, showEditor: false });
  const openProject = (id, projectRoute = 'pipelines') =>
    setRoute({ app: 'projects', project: id, projectRoute, pipelineRun: null, showEditor: false });
  const openProjectRoute = (projectRoute) =>
    setRoute(r => ({ ...r, projectRoute, pipelineRun: null, showEditor: false }));

  // Project shell when inside a project
  if (route.project) {
    const project = PROJECTS.find(p => p.id === route.project) || PROJECTS[0];
    let body, crumbs = [];
    if (route.projectRoute === 'pipelines' && !route.pipelineRun && !route.showEditor) {
      body = <PipelineListRow />; crumbs = ['流水线'];
    } else if (route.projectRoute === 'pipelines' && route.pipelineRun === 'list') {
      body = <PipelineRunListPage onOpen={(r) => setRoute(x => ({ ...x, pipelineRun: r }))} />;
      crumbs = ['流水线', '运行历史'];
    } else if (route.projectRoute === 'pipelines' && route.pipelineRun && typeof route.pipelineRun === 'object') {
      body = <PipelineRunDetailA run={route.pipelineRun} onBack={() => setRoute(x => ({ ...x, pipelineRun: 'list' }))} />;
      crumbs = ['流水线', `#${route.pipelineRun.n}`];
    } else if (route.projectRoute === 'environments') { body = <EnvironmentListPage />; crumbs = ['环境']; }
    else if (route.projectRoute === 'deployments') { body = <DeploymentListPage />; crumbs = ['部署']; }
    else if (route.projectRoute === 'services') { body = <ServiceListPage />; crumbs = ['服务']; }
    else if (route.projectRoute === 'members') { body = <MemberListPage />; crumbs = ['成员']; }
    else if (route.projectRoute === 'variables') { body = <VariableListPage />; crumbs = ['变量']; }
    else if (route.projectRoute === 'notifications') { body = <NotificationRulesPage />; crumbs = ['通知']; }
    else { body = <PipelineListRow />; crumbs = ['流水线']; }

    return (
      <>
        <ProjectShell
          projectActive={route.projectRoute}
          project={project}
          crumbs={crumbs}
          onNavigate={(id) => { if (id === 'projects') go('projects'); else go(id); }}
          onProjectNavigate={openProjectRoute}
        >{body}</ProjectShell>
        {route.showEditor && <PipelineEditorPage onClose={() => setRoute(x => ({ ...x, showEditor: false }))} />}
      </>
    );
  }

  // App-level pages
  let body, crumbs = [];
  const a = route.app;
  if (a === 'dashboard') { body = <DashboardA onOpenProject={openProject} />; crumbs = ['Dashboard']; }
  else if (a === 'projects') {
    body = <ProjectListPage onOpen={(id) => openProject(id, 'pipelines')} />; crumbs = ['项目管理'];
  }
  else if (a === 'users') { body = <AdminUsersPage />; crumbs = ['用户管理']; }
  else if (a === 'globalvars') { body = <AdminVariablePage />; crumbs = ['全局变量']; }
  else if (a === 'integrations') { body = <IntegrationsPage />; crumbs = ['集成管理']; }
  else if (a === 'audit') { body = <AuditLogPage />; crumbs = ['审计日志']; }
  else if (a === 'systemset') { body = <SystemSettingsPage />; crumbs = ['系统设置']; }
  else { body = <DashboardA onOpenProject={openProject} />; crumbs = ['Dashboard']; }

  return <AppShell active={a} crumbs={crumbs} onNavigate={go}>{body}</AppShell>;
}

// ───────── BrowserWindow wrapper for live demo ─────────
function LiveFrame() {
  return (
    <ChromeWindow
      width="100%" height="100%"
      url="app.zcid.local/dashboard"
      tabs={[{ title: 'zcid · Cloud-Native CI/CD' }, { title: 'ArgoCD' }, { title: 'Grafana' }]}
      activeIndex={0}
    >
      <LiveDemo />
    </ChromeWindow>
  );
}

// ───────── Static artboards ─────────
const Screen = ({ children }) => (
  <div style={{ width: '100%', height: '100%', background: 'var(--z-0)' }}>{children}</div>
);

function App() {
  const [tw, setTw, removeTw] = useTweaks(TWEAK_DEFAULTS);

  // Apply accent hue globally
  React.useEffect(() => {
    document.documentElement.style.setProperty('--accent-h', tw.accentHue);
  }, [tw.accentHue]);

  // Dummy project for shell-only frames
  const demoProject = PROJECTS[0];

  return (
    <>
      <DesignCanvas>
        {/* ───── Public & Shell ───── */}
        <DCSection id="public" title="Public · Auth & Errors">
          <DCArtboard id="login-split" label="01 · Login (Split brand panel)" width={1280} height={820}>
            <Screen><LoginSplit /></Screen>
          </DCArtboard>
          <DCArtboard id="login-minimal" label="01b · Login (Minimal centered)" width={1280} height={820}>
            <Screen><LoginMinimal /></Screen>
          </DCArtboard>
          <DCArtboard id="forbidden" label="04 · 403 Forbidden" width={1280} height={820}>
            <Screen><ForbiddenPage /></Screen>
          </DCArtboard>
          <DCArtboard id="not-found" label="05 · 404 Not Found" width={1280} height={820}>
            <Screen><NotFoundPage /></Screen>
          </DCArtboard>
        </DCSection>

        {/* ───── Live interactive demo ───── */}
        <DCSection id="live" title="Live · Interactive prototype (clickable nav)">
          <DCArtboard id="live-demo" label="Live app — sidebar nav works, modals, filters, log streaming" width={1440} height={900}>
            <LiveFrame />
          </DCArtboard>
        </DCSection>

        {/* ───── Dashboard variations ───── */}
        <DCSection id="dashboard" title="Dashboard · 3 variations">
          <DCArtboard id="dash-a" label="06a · Dashboard — Metric-heavy (primary)" width={1440} height={920}>
            <Screen><AppShell active="dashboard" crumbs={['Dashboard']}><DashboardA /></AppShell></Screen>
          </DCArtboard>
          <DCArtboard id="dash-b" label="06b · Dashboard — Activity feed forward" width={1440} height={920}>
            <Screen><AppShell active="dashboard" crumbs={['Dashboard']}><DashboardB /></AppShell></Screen>
          </DCArtboard>
          <DCArtboard id="dash-c" label="06c · Dashboard — Projects-first" width={1440} height={920}>
            <Screen><AppShell active="dashboard" crumbs={['Dashboard']}><DashboardC /></AppShell></Screen>
          </DCArtboard>
        </DCSection>

        {/* ───── Projects & Admin ───── */}
        <DCSection id="admin" title="Workspace · Projects + System pages">
          <DCArtboard id="project-list" label="07 · Project List" width={1440} height={920}>
            <Screen><AppShell active="projects" crumbs={['项目管理']}><ProjectListPage /></AppShell></Screen>
          </DCArtboard>
          <DCArtboard id="admin-users" label="08 · Admin — Users" width={1440} height={900}>
            <Screen><AppShell active="users" crumbs={['用户管理']}><AdminUsersPage /></AppShell></Screen>
          </DCArtboard>
          <DCArtboard id="admin-vars" label="09 · Admin — Global Variables" width={1440} height={900}>
            <Screen><AppShell active="globalvars" crumbs={['全局变量']}><AdminVariablePage /></AppShell></Screen>
          </DCArtboard>
          <DCArtboard id="admin-integrations" label="10 · Admin — Integrations" width={1440} height={920}>
            <Screen><AppShell active="integrations" crumbs={['集成管理']}><IntegrationsPage /></AppShell></Screen>
          </DCArtboard>
          <DCArtboard id="admin-audit" label="11 · Admin — Audit Log" width={1440} height={900}>
            <Screen><AppShell active="audit" crumbs={['审计日志']}><AuditLogPage /></AppShell></Screen>
          </DCArtboard>
          <DCArtboard id="admin-settings" label="12 · Admin — System Settings" width={1440} height={900}>
            <Screen><AppShell active="systemset" crumbs={['系统设置']}><SystemSettingsPage /></AppShell></Screen>
          </DCArtboard>
        </DCSection>

        {/* ───── Project subpages ───── */}
        <DCSection id="project" title="Project · Environments / Deployments / Services / Members / Variables / Notifications">
          <DCArtboard id="env-list" label="13 · Environments" width={1440} height={880}>
            <Screen><ProjectShell projectActive="environments" project={demoProject} crumbs={['环境']}><EnvironmentListPage /></ProjectShell></Screen>
          </DCArtboard>
          <DCArtboard id="deploy-list" label="14 · Deployments" width={1440} height={880}>
            <Screen><ProjectShell projectActive="deployments" project={demoProject} crumbs={['部署']}><DeploymentListPage /></ProjectShell></Screen>
          </DCArtboard>
          <DCArtboard id="deploy-detail" label="15 · Deployment Detail" width={1440} height={900}>
            <Screen><ProjectShell projectActive="deployments" project={demoProject} crumbs={['部署', '详情']}><DeploymentDetailPage /></ProjectShell></Screen>
          </DCArtboard>
          <DCArtboard id="svc-list" label="16 · Services" width={1440} height={820}>
            <Screen><ProjectShell projectActive="services" project={demoProject} crumbs={['服务']}><ServiceListPage /></ProjectShell></Screen>
          </DCArtboard>
          <DCArtboard id="mem-list" label="17 · Members" width={1440} height={820}>
            <Screen><ProjectShell projectActive="members" project={demoProject} crumbs={['成员']}><MemberListPage /></ProjectShell></Screen>
          </DCArtboard>
          <DCArtboard id="var-list" label="18 · Project Variables" width={1440} height={820}>
            <Screen><ProjectShell projectActive="variables" project={demoProject} crumbs={['变量']}><VariableListPage /></ProjectShell></Screen>
          </DCArtboard>
          <DCArtboard id="notif-rules" label="19 · Notification Rules" width={1440} height={820}>
            <Screen><ProjectShell projectActive="notifications" project={demoProject} crumbs={['通知']}><NotificationRulesPage /></ProjectShell></Screen>
          </DCArtboard>
        </DCSection>

        {/* ───── Pipeline ───── */}
        <DCSection id="pipelines" title="Pipeline · List / Template wizard / Editor / Runs">
          <DCArtboard id="pipe-list-row" label="20a · Pipeline List (row)" width={1440} height={900}>
            <Screen><ProjectShell projectActive="pipelines" project={demoProject} crumbs={['流水线']}><PipelineListRow /></ProjectShell></Screen>
          </DCArtboard>
          <DCArtboard id="pipe-list-card" label="20b · Pipeline List (card grid)" width={1440} height={900}>
            <Screen><ProjectShell projectActive="pipelines" project={demoProject} crumbs={['流水线']}><PipelineListCard /></ProjectShell></Screen>
          </DCArtboard>
          <DCArtboard id="pipe-tpl" label="21 · Create Pipeline — Template Select Wizard" width={1440} height={920}>
            <Screen><ProjectShell projectActive="pipelines" project={demoProject} crumbs={['流水线', '创建']}><TemplateSelectPage /></ProjectShell></Screen>
          </DCArtboard>
          <DCArtboard id="pipe-editor" label="22 · Pipeline Editor — Visual DAG (drag + edit)" width={1440} height={900}>
            <Screen>
              <div style={{ position: 'relative', width: '100%', height: '100%' }}>
                <PipelineEditorPage />
              </div>
            </Screen>
          </DCArtboard>
          <DCArtboard id="pipe-runs" label="23 · Run History" width={1440} height={900}>
            <Screen><ProjectShell projectActive="pipelines" project={demoProject} crumbs={['流水线', '运行历史']}><PipelineRunListPage /></ProjectShell></Screen>
          </DCArtboard>
          <DCArtboard id="pipe-run-a" label="24a · Run Detail — Horizontal stage bar + streaming logs" width={1440} height={1100}>
            <Screen><ProjectShell projectActive="pipelines" project={demoProject} crumbs={['流水线', 'Build #1284']}><PipelineRunDetailA /></ProjectShell></Screen>
          </DCArtboard>
          <DCArtboard id="pipe-run-b" label="24b · Run Detail — Vertical timeline" width={1440} height={1000}>
            <Screen><ProjectShell projectActive="pipelines" project={demoProject} crumbs={['流水线', 'Build #1284']}><PipelineRunDetailB /></ProjectShell></Screen>
          </DCArtboard>
        </DCSection>
      </DesignCanvas>

      <TweaksPanel title="Tweaks · zcid">
        <TweakSection title="Accent">
          <TweakSlider label="Brand hue" value={tw.accentHue} onChange={v => setTw('accentHue', v)} min={0} max={360} step={1} suffix="°" />
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(6,1fr)', gap: 6, marginTop: 8 }}>
            {[220, 250, 200, 160, 280, 30].map(h => (
              <button key={h} onClick={() => setTw('accentHue', h)} title={`hue ${h}`}
                style={{
                  height: 26, borderRadius: 5, border: tw.accentHue === h ? '2px solid #111' : '1px solid #e3e3e6',
                  background: `linear-gradient(135deg, oklch(0.62 0.20 ${h}), oklch(0.55 0.22 ${(h + 20) % 360}))`,
                  cursor: 'pointer', padding: 0,
                }} />
            ))}
          </div>
        </TweakSection>
      </TweaksPanel>
    </>
  );
}

ReactDOM.createRoot(document.getElementById('root')).render(<App />);
