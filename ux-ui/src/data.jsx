// Mock data for zcid
const PROJECTS = [
  { id: 'web-console', name: 'web-console', desc: '运营后台 · Next.js · Node 20', status: 'active', created: '2024-09-12', pipelines: 8, runs: 1284, lastStatus: 'succeeded', firstLetter: 'W', hue: 220 },
  { id: 'payment-svc', name: 'payment-svc', desc: '支付核心服务 · Go 1.22 · gRPC', status: 'active', created: '2024-07-03', pipelines: 12, runs: 3921, lastStatus: 'running', firstLetter: 'P', hue: 160 },
  { id: 'ml-inference', name: 'ml-inference', desc: 'LLM 推理网关 · Python 3.11 · CUDA', status: 'active', created: '2025-01-20', pipelines: 5, runs: 442, lastStatus: 'failed', firstLetter: 'M', hue: 280 },
  { id: 'iam-gateway', name: 'iam-gateway', desc: '统一身份网关 · Java 21 · Spring', status: 'active', created: '2024-03-18', pipelines: 6, runs: 2108, lastStatus: 'succeeded', firstLetter: 'I', hue: 30 },
  { id: 'docs-site', name: 'docs-site', desc: '开发者文档站点 · Astro', status: 'active', created: '2025-02-14', pipelines: 2, runs: 89, lastStatus: 'succeeded', firstLetter: 'D', hue: 340 },
  { id: 'legacy-admin', name: 'legacy-admin', desc: '旧版管理后台 · 即将归档', status: 'disabled', created: '2023-06-01', pipelines: 3, runs: 612, lastStatus: 'cancelled', firstLetter: 'L', hue: 60 },
];

const PIPELINES = [
  { id: 'build-web', name: 'build-web-console', desc: 'Next.js build + image push to harbor', trigger: 'Webhook', status: 'active', runs: 128, updated: '5 分钟前', lastRun: 'succeeded' },
  { id: 'deploy-payment-prod', name: 'deploy-payment-prod', desc: 'ArgoCD sync to prod cluster · 人工审批', trigger: 'Manual', status: 'active', runs: 34, updated: '1 小时前', lastRun: 'running' },
  { id: 'ml-batch-train', name: 'ml-batch-train', desc: '每日凌晨批量训练 + 指标回传', trigger: 'Cron · 0 2 * * *', status: 'active', runs: 421, updated: '2 小时前', lastRun: 'succeeded' },
  { id: 'iam-test', name: 'iam-integration-test', desc: '契约测试 + postgres-17 集成用例', trigger: 'Webhook', status: 'draft', runs: 0, updated: '昨天', lastRun: null },
  { id: 'docs-deploy', name: 'docs-deploy', desc: 'Astro build → CDN 推送', trigger: 'Webhook', status: 'active', runs: 23, updated: '3 天前', lastRun: 'succeeded' },
  { id: 'nightly-scan', name: 'nightly-security-scan', desc: 'Trivy + Grype 双扫 · SBOM 归档', trigger: 'Cron · 0 4 * * *', status: 'disabled', runs: 89, updated: '一周前', lastRun: 'failed' },
  { id: 'payment-perf', name: 'payment-perf-bench', desc: 'k6 压测 + 回归对比', trigger: 'Manual', status: 'draft', runs: 0, updated: '2 天前', lastRun: null },
];

const RUNS = [
  { n: 1284, status: 'running',   trigger: 'Webhook', user: 'zhao.wei',   branch: 'feat/cart-v2',   duration: 147, time: '刚刚',       commit: '8f3c1a2' },
  { n: 1283, status: 'succeeded', trigger: 'Webhook', user: 'li.qiang',   branch: 'main',           duration: 312, time: '12 分钟前',  commit: 'b4a5d81' },
  { n: 1282, status: 'succeeded', trigger: 'Manual',  user: 'admin',      branch: 'release/2.4.0',  duration: 289, time: '1 小时前',   commit: '2e0fb7c' },
  { n: 1281, status: 'failed',    trigger: 'Webhook', user: 'wang.ming',  branch: 'fix/login-csrf', duration: 84,  time: '2 小时前',   commit: '91acd0e' },
  { n: 1280, status: 'cancelled', trigger: 'Manual',  user: 'chen.hao',   branch: 'wip/experiment', duration: 42,  time: '3 小时前',   commit: 'f0d3499' },
  { n: 1279, status: 'succeeded', trigger: 'Cron',    user: 'system',     branch: 'main',           duration: 298, time: '昨天 02:00',  commit: 'a17f22b' },
  { n: 1278, status: 'succeeded', trigger: 'Webhook', user: 'zhao.wei',   branch: 'main',           duration: 305, time: '昨天 11:24',  commit: '33e8c9f' },
  { n: 1277, status: 'queued',    trigger: 'Webhook', user: 'li.qiang',   branch: 'feat/ui-revamp', duration: 0,   time: '排队中',       commit: '7b2a104' },
];

const DEPLOYMENTS = [
  { id: 1, image: 'harbor.local/web/console:2.4.0', env: 'prod',    status: 'healthy',  sync: 'Synced',     health: 'Healthy',     user: 'admin',     time: '1 小时前' },
  { id: 2, image: 'harbor.local/web/console:2.4.0-rc.3', env: 'staging', status: 'syncing',  sync: 'Syncing', health: 'Progressing', user: 'zhao.wei',  time: '刚刚' },
  { id: 3, image: 'harbor.local/payment/core:1.18.2', env: 'prod',    status: 'degraded', sync: 'Synced',   health: 'Degraded',    user: 'li.qiang',  time: '2 小时前' },
  { id: 4, image: 'harbor.local/ml/gateway:0.9.7', env: 'staging', status: 'sync_failed', sync: 'OutOfSync', health: 'Missing', user: 'wang.ming', time: '4 小时前' },
  { id: 5, image: 'harbor.local/iam/gateway:3.0.1',   env: 'prod',    status: 'healthy',  sync: 'Synced',    health: 'Healthy',   user: 'admin',     time: '昨天' },
  { id: 6, image: 'harbor.local/web/console:2.3.9',   env: 'prod',    status: 'rolled_back', sync: 'Synced', health: 'Healthy', user: 'admin',     time: '昨天' },
];

const ENVIRONMENTS = [
  { name: 'prod',    ns: 'zcid-prod',    desc: '生产集群 · cn-hangzhou-1',   health: 'env_healthy', services: 18, created: '2024-03-01' },
  { name: 'staging', ns: 'zcid-staging', desc: '预发环境 · 与 prod 同构', health: 'env_syncing',  services: 18, created: '2024-03-01' },
  { name: 'dev',     ns: 'zcid-dev',     desc: '日常开发 · 自动部署 main',  health: 'env_degraded', services: 21, created: '2024-03-01' },
  { name: 'sandbox', ns: 'zcid-sandbox', desc: '临时验证 · 非永久',          health: 'env_healthy', services: 6,  created: '2025-01-12' },
  { name: 'canary',  ns: 'zcid-canary',  desc: '灰度 · 5% 流量',              health: 'env_down',    services: 3,  created: '2024-11-20' },
];

const USERS = [
  { name: 'admin',      role: 'admin',         status: 'enabled',      created: '2024-01-01' },
  { name: 'zhao.wei',   role: 'project_admin', status: 'enabled',      created: '2024-03-18' },
  { name: 'li.qiang',   role: 'project_admin', status: 'enabled',      created: '2024-05-02' },
  { name: 'wang.ming',  role: 'member',        status: 'enabled',      created: '2024-08-11' },
  { name: 'chen.hao',   role: 'member',        status: 'enabled',      created: '2024-09-30' },
  { name: 'ex.contractor', role: 'member',     status: 'userDisabled', created: '2024-02-05' },
];

const VARIABLES_GLOBAL = [
  { key: 'HARBOR_REGISTRY', value: 'harbor.zcid.local',  type: 'Variable', desc: '镜像仓库地址',             created: '2024-03-01' },
  { key: 'HARBOR_PASSWORD', value: '••••••••••••',       type: 'Secret',   desc: 'Harbor 推送账户密码',     created: '2024-03-01' },
  { key: 'ARGOCD_TOKEN',    value: '••••••••••••',       type: 'Secret',   desc: 'ArgoCD API token',       created: '2024-03-12' },
  { key: 'SLACK_WEBHOOK',   value: '••••••••••••',       type: 'Secret',   desc: '通知 Webhook URL',        created: '2024-06-21' },
  { key: 'KUBE_VERSION',    value: '1.30.4',              type: 'Variable', desc: 'K8s 目标版本',             created: '2024-07-14' },
  { key: 'SONAR_TOKEN',     value: '••••••••••••',       type: 'Secret',   desc: 'SonarQube 扫描 token',    created: '2024-09-01' },
];

const INTEGRATIONS = [
  { name: 'zcid / platform', provider: 'GitHub', url: 'github.com', token: 'ghp_****z8', desc: '主业务代码托管', created: '2024-03-01', status: 'connected', icon: '🐙' },
  { name: 'zcid / infra',    provider: 'GitLab', url: 'gitlab.zcid.local', token: 'glpat-****3a', desc: '基础设施 IaC', created: '2024-04-17', status: 'token_expired', icon: '🦊' },
  { name: 'zcid / ml-ops',   provider: 'GitHub', url: 'github.com', token: 'ghp_****1c', desc: '模型 + 实验仓库', created: '2024-11-02', status: 'connected', icon: '🐙' },
  { name: 'legacy / tools',  provider: 'GitLab', url: 'gitlab.corp.old', token: 'glpat-****7f', desc: '历史归档', created: '2023-01-10', status: 'disconnected', icon: '🦊' },
];

const AUDIT = [
  { time: '2025-04-24 10:42:11', user: 'admin#u1',      method: 'POST',   path: '/api/v1/projects',                        rt: 'project',     rid: '8f3c1a29', result: 'ok', ip: '10.1.3.42' },
  { time: '2025-04-24 10:40:55', user: 'zhao.wei#u3',    method: 'PUT',    path: '/api/v1/pipelines/build-web',             rt: 'pipeline',    rid: 'build-we', result: 'ok', ip: '10.1.4.11' },
  { time: '2025-04-24 10:39:02', user: 'li.qiang#u4',    method: 'DELETE', path: '/api/v1/variables/SLACK_OLD',             rt: 'variable',    rid: 'SLACK_OL', result: 'ok', ip: '10.1.5.18' },
  { time: '2025-04-24 10:37:48', user: 'wang.ming#u5',   method: 'GET',    path: '/api/v1/runs/1283',                       rt: 'run',         rid: '00001283', result: 'ok', ip: '10.1.5.92' },
  { time: '2025-04-24 10:32:14', user: 'system',         method: 'POST',   path: '/api/v1/webhooks/github',                 rt: 'webhook',     rid: 'gh-12042', result: 'err', ip: '140.82.114.6' },
  { time: '2025-04-24 10:28:07', user: 'chen.hao#u6',    method: 'POST',   path: '/api/v1/deployments',                      rt: 'deployment', rid: 'd-4a19b2', result: 'ok', ip: '10.1.5.33' },
  { time: '2025-04-24 10:15:33', user: 'admin#u1',      method: 'PUT',    path: '/api/v1/users/chen.hao/role',              rt: 'user',       rid: 'u6',       result: 'ok', ip: '10.1.3.42' },
  { time: '2025-04-24 10:02:51', user: 'zhao.wei#u3',    method: 'POST',   path: '/api/v1/runs',                             rt: 'run',        rid: '00001284', result: 'ok', ip: '10.1.4.11' },
];

const BUILD_LOG = [
  { lvl: 'info',  text: '[git-clone] cloning github.com/zcid/web-console#feat/cart-v2' },
  { lvl: 'info',  text: '[git-clone] head at 8f3c1a29 · "feat(cart): unified discount pipeline"' },
  { lvl: 'info',  text: '[git-clone] done in 2.3s' },
  { lvl: 'info',  text: '[install] pnpm install --frozen-lockfile' },
  { lvl: 'info',  text: '[install] Packages: +1284 -0' },
  { lvl: 'info',  text: '[install] Progress: resolved 1284, reused 1284, downloaded 0 — 4.1s' },
  { lvl: 'info',  text: '[build] next build' },
  { lvl: 'info',  text: '[build]  ▲ Next.js 15.0.3  ·  Environments: .env.production' },
  { lvl: 'info',  text: '[build]  ✓ Creating an optimized production build' },
  { lvl: 'info',  text: '[build]  ✓ Compiled successfully in 28.4s' },
  { lvl: 'warn',  text: '[build]   ⚠  components/Tooltip.tsx:42  prefer const' },
  { lvl: 'info',  text: '[build]  ✓ Linting and checking validity of types' },
  { lvl: 'info',  text: '[build]  ✓ Collecting page data · Generating static pages (118/118)' },
  { lvl: 'info',  text: '[kaniko] building image harbor.local/web/console:2.4.0-rc.4' },
  { lvl: 'info',  text: '[kaniko] COPY . /app — layer size 42.1 MB' },
  { lvl: 'info',  text: '[kaniko] RUN pnpm build:static — cached' },
  { lvl: 'info',  text: '[kaniko] pushing 8 layers → harbor.local' },
  { lvl: 'info',  text: '[kaniko] digest sha256:b81f4ca…9e23 — pushed in 12.8s' },
  { lvl: 'info',  text: '[argocd] triggering sync on web-console-staging' },
  { lvl: 'info',  text: '[argocd] wave 0/3 · deploying ReplicaSet' },
];

const SERVICES = [
  { name: 'web', desc: '运营后台 Next.js 前端', repo: 'git@github.com:zcid/web-console.git', created: '2024-03-01' },
  { name: 'api', desc: 'BFF · tRPC + fastify', repo: 'git@github.com:zcid/web-api.git', created: '2024-03-01' },
  { name: 'worker', desc: '后台任务 · BullMQ', repo: 'git@github.com:zcid/web-worker.git', created: '2024-05-12' },
  { name: 'edge', desc: 'CDN 边缘函数 · Deno', repo: 'git@github.com:zcid/web-edge.git', created: '2024-09-04' },
];

const NOTIFICATION_RULES = [
  { name: 'Slack · 构建失败警报', event: 'build_failed', url: 'https://hooks.slack.com/services/T***/B***/xxx', enabled: true, created: '2024-06-21' },
  { name: 'Feishu · 生产部署', event: 'deploy_success', url: 'https://open.feishu.cn/open-apis/bot/v2/hook/abc', enabled: true, created: '2024-08-02' },
  { name: 'Teams · 构建成功通知', event: 'build_success', url: 'https://zcid.webhook.office.com/webhookb2/xxx', enabled: false, created: '2024-09-10' },
  { name: 'PagerDuty · 部署失败', event: 'deploy_failed', url: 'https://events.pagerduty.com/v2/enqueue', enabled: true, created: '2024-10-05' },
];

const PROJECT_VARIABLES = [
  { key: 'NEXT_PUBLIC_API_URL', value: 'https://api.zcid.local', type: 'Variable', desc: '前端 API 基址', created: '2024-03-12' },
  { key: 'DATABASE_URL', value: '••••••••••••', type: 'Secret', desc: 'Postgres 连接串', created: '2024-03-12' },
  { key: 'REDIS_URL', value: 'redis://redis.zcid.svc:6379', type: 'Variable', desc: 'Redis 地址', created: '2024-04-02' },
  { key: 'OAUTH_CLIENT_SECRET', value: '••••••••••••', type: 'Secret', desc: 'OAuth 客户端密钥', created: '2024-07-18' },
];

// Initial DAG for pipeline editor
const INITIAL_DAG = {
  name: 'build-web-console',
  stages: [
    { id: 's1', name: 'Source', steps: [
      { id: 'st1', name: 'git-clone', type: 'git-clone', config: { url: 'github.com/zcid/web-console', branch: '${BRANCH}' } },
    ]},
    { id: 's2', name: 'Build', steps: [
      { id: 'st2', name: 'install-deps', type: 'shell', config: { script: 'pnpm install --frozen-lockfile' } },
      { id: 'st3', name: 'run-tests', type: 'shell', config: { script: 'pnpm test -- --coverage' } },
      { id: 'st4', name: 'next-build', type: 'shell', config: { script: 'pnpm build' } },
    ]},
    { id: 's3', name: 'Package', steps: [
      { id: 'st5', name: 'kaniko-build', type: 'kaniko-build', config: { dockerfile: './Dockerfile', tag: '${REGISTRY}/web/console:${SHA}' } },
    ]},
    { id: 's4', name: 'Deploy', steps: [
      { id: 'st6', name: 'argocd-sync', type: 'shell', config: { script: 'argocd app sync web-console-staging' } },
    ]},
  ],
};

Object.assign(window, {
  PROJECTS, PIPELINES, RUNS, DEPLOYMENTS, ENVIRONMENTS, USERS,
  VARIABLES_GLOBAL, INTEGRATIONS, AUDIT, BUILD_LOG, SERVICES,
  NOTIFICATION_RULES, PROJECT_VARIABLES, INITIAL_DAG,
});
