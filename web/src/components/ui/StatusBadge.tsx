import { Badge } from './Badge';

type Tone = 'green' | 'red' | 'amber' | 'blue' | 'cyan' | 'grey';

const STATUS_MAP: Record<string, [Tone, string]> = {
  active:           ['green',  '已启用'],
  draft:            ['grey',   '草稿'],
  disabled:         ['amber',  '已停用'],
  pending:          ['grey',   '待执行'],
  queued:           ['blue',   '排队中'],
  running:          ['cyan',   '运行中'],
  succeeded:        ['green',  '成功'],
  failed:           ['red',    '失败'],
  cancelled:        ['amber',  '已取消'],
  not_deployed:     ['grey',   '待部署'],
  syncing:          ['blue',   '同步中'],
  healthy:          ['green',  '健康'],
  degraded:         ['amber',  '异常'],
  sync_failed:      ['red',    '失败'],
  rolled_back:      ['grey',   '已回滚'],
  Synced:           ['green',  'Synced'],
  Syncing:          ['blue',   'Syncing'],
  OutOfSync:        ['amber',  'OutOfSync'],
  Missing:          ['red',    'Missing'],
  Healthy:          ['green',  'Healthy'],
  Degraded:         ['amber',  'Degraded'],
  env_healthy:      ['green',  'Healthy'],
  env_syncing:      ['blue',   'Syncing'],
  env_degraded:     ['amber',  'Degraded'],
  env_down:         ['red',    'Down'],
  connected:        ['green',  'Connected'],
  token_expired:    ['amber',  'Token Expired'],
  disconnected:     ['red',    'Disconnected'],
  enabled:          ['green',  '启用'],
  personal:         ['blue',   '个人令牌'],
  project:          ['cyan',   '项目令牌'],
  revoked:          ['grey',   '已撤销'],
  userDisabled:     ['grey',   '禁用'],
  build_success:    ['green',  '构建成功'],
  build_failed:     ['red',    '构建失败'],
  deploy_success:   ['blue',   '部署成功'],
  deploy_failed:    ['amber',  '部署失败'],
  GET:              ['blue',   'GET'],
  POST:             ['green',  'POST'],
  PUT:              ['amber',  'PUT'],
  DELETE:           ['red',    'DELETE'],
  ok:               ['green',  '成功'],
  err:              ['red',    '失败'],
  admin:            ['blue',   '管理员'],
  project_admin:    ['green',  '项目管理员'],
  member:           ['grey',   '普通成员'],
};

export function StatusBadge({ status, labels }: { status: string; labels?: Record<string, [Tone, string]> }) {
  const map = labels ?? STATUS_MAP;
  const [tone, label] = map[status] ?? ['grey', status];
  return <Badge tone={tone} dot>{label}</Badge>;
}
