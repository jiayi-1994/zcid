import { useCallback, useEffect, useState } from 'react';
import { AppLayout } from '../../../components/layout/AppLayout';
import { fetchAuditLogs, type AuditLog, type AuditLogFilters } from '../../../services/audit';
import { PageHeader } from '../../../components/ui/PageHeader';
import { Card } from '../../../components/ui/Card';
import { Btn } from '../../../components/ui/Btn';
import { ZSelect } from '../../../components/ui/ZSelect';
import { StatusBadge } from '../../../components/ui/StatusBadge';
import { IUser, ICalendar, IFilter, IChevL, IChevR } from '../../../components/ui/icons';

function extractMethod(action: string): string {
  if (action.startsWith('auth.')) return 'AUTH';
  const m = action.match(/^(GET|POST|PUT|DELETE|PATCH)/);
  return m ? m[1] : 'GET';
}

function extractPath(action: string): string {
  if (action.startsWith('auth.')) return action;
  return action.replace(/^(GET|POST|PUT|DELETE|PATCH)\s+/, '').replace(/\/api\/v1\//, '');
}

function detailReason(detail?: string): string {
  if (!detail) return '';
  try {
    const parsed = JSON.parse(detail) as { reason?: string };
    return parsed.reason ?? '';
  } catch {
    return '';
  }
}

export function AuditLogPage() {
  const [logs, setLogs] = useState<AuditLog[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(false);
  const [filters, setFilters] = useState<AuditLogFilters>({ page: 1, pageSize: 20 });

  const load = useCallback(async () => {
    setLoading(true);
    try { const data = await fetchAuditLogs(filters); setLogs(data.items || []); setTotal(data.total || 0); }
    catch { setLogs([]); }
    finally { setLoading(false); }
  }, [filters]);

  useEffect(() => { load(); }, [load]);

  const updateFilter = (key: string, value: string | undefined) =>
    setFilters((prev) => ({ ...prev, [key]: value || undefined, page: 1 }));

  const totalPages = Math.ceil(total / (filters.pageSize ?? 20));

  return (
    <AppLayout>
      <PageHeader crumb="System · Compliance" title="审计日志" sub="全量 API 操作记录与合规追溯。" />
      <div style={{ padding: 24, display: 'flex', flexDirection: 'column', gap: 14 }}>
        <div style={{ display: 'flex', gap: 8, flexWrap: 'wrap', alignItems: 'center' }}>
          <ZSelect
            width={130}
            value={filters.action ?? ''}
            options={[{ value: '', label: '全部方法' }, 'GET', 'POST', 'PUT', 'DELETE']}
            onChange={(v) => updateFilter('action', v)}
          />
          <ZSelect
            width={150}
            value={filters.category ?? ''}
            options={[{ value: '', label: '全部类别' }, { value: 'auth_security', label: '认证与访问' }]}
            onChange={(v) => updateFilter('category', v)}
          />
          <div className="input-wrap">
            <IUser size={13} />
            <input className="input input--with-icon" style={{ width: 160 }} placeholder="用户名..." onChange={(e) => updateFilter('userId', e.target.value)} />
          </div>
          <div style={{ flex: 1 }} />
          <span className="sub" style={{ fontSize: 11.5 }}>{total} 条记录</span>
        </div>

        <Card padding={false}>
          {loading ? (
            <div style={{ padding: '40px 0', textAlign: 'center', color: 'var(--z-400)' }}>加载中...</div>
          ) : (
            <table className="table">
              <thead>
                <tr>
                  <th>时间</th><th>用户</th><th>方法</th><th>接口</th><th>资源类型</th><th>资源 ID</th><th>结果</th><th>IP</th>
                </tr>
              </thead>
              <tbody>
                {logs.map((a, i) => {
                  const method = extractMethod(a.action);
                  const path = extractPath(a.action);
                  return (
                    <tr key={i}>
                      <td><span className="sub mono" style={{ fontSize: 11 }}>{new Date(a.createdAt).toLocaleString()}</span></td>
                      <td><span className="mono" style={{ fontSize: 11.5 }}>{a.userId?.replace('admin-bootstrap-', '#') || '-'}</span></td>
                      <td><StatusBadge status={method} /></td>
                      <td><span className="code" title={detailReason(a.detail)} style={{ maxWidth: 240, display: 'inline-block', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap', verticalAlign: 'middle' }}>{path}</span></td>
                      <td><span className="sub">{a.resourceType || '-'}</span></td>
                      <td><span className="mono" style={{ fontSize: 11 }}>{a.resourceId ? a.resourceId.substring(0, 8) + '…' : '-'}</span></td>
                      <td><StatusBadge status={a.result === 'success' ? 'ok' : 'err'} /></td>
                      <td><span className="mono sub" style={{ fontSize: 11 }}>{a.ip || '-'}</span></td>
                    </tr>
                  );
                })}
                {logs.length === 0 && !loading && (
                  <tr><td colSpan={8} style={{ textAlign: 'center', padding: '40px 0', color: 'var(--z-400)' }}>暂无审计记录</td></tr>
                )}
              </tbody>
            </table>
          )}
        </Card>

        {totalPages > 1 && (
          <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: 11.5, color: 'var(--z-500)' }}>
            <span>共 {total} 条 · 第 {filters.page} / {totalPages} 页</span>
            <div style={{ display: 'flex', gap: 4 }}>
              <Btn size="xs" variant="ghost" iconOnly icon={<IChevL size={12} />} disabled={(filters.page ?? 1) <= 1} onClick={() => setFilters((p) => ({ ...p, page: (p.page ?? 1) - 1 }))} />
              <Btn size="xs" variant="outline">{filters.page}</Btn>
              <Btn size="xs" variant="ghost" iconOnly icon={<IChevR size={12} />} disabled={(filters.page ?? 1) >= totalPages} onClick={() => setFilters((p) => ({ ...p, page: (p.page ?? 1) + 1 }))} />
            </div>
          </div>
        )}
      </div>
    </AppLayout>
  );
}

export default AuditLogPage;
