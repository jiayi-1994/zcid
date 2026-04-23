import { DatePicker, Input, Select, Space, Table, Tooltip } from '@arco-design/web-react';
import { useCallback, useEffect, useState } from 'react';
import { AppLayout } from '../../../components/layout/AppLayout';
import { fetchAuditLogs, type AuditLog, type AuditLogFilters } from '../../../services/audit';

const RESULT_CLS: Record<string, string> = {
  success: 'pipeline-status-badge--success',
  failure: 'pipeline-status-badge--failed',
};
const RESULT_LABEL: Record<string, string> = {
  success: '成功',
  failure: '失败',
};

const METHOD_CLS: Record<string, string> = {
  GET: 'pipeline-status-badge--running',
  POST: 'pipeline-status-badge--success',
  PUT: 'pipeline-status-badge--cancelled',
  DELETE: 'pipeline-status-badge--failed',
  PATCH: 'pipeline-status-badge--cancelled',
};

const ACTION_OPTIONS = [
  { label: '全部', value: '' },
  { label: 'POST', value: 'POST' },
  { label: 'PUT', value: 'PUT' },
  { label: 'DELETE', value: 'DELETE' },
];

function extractMethod(action: string): string {
  const m = action.match(/^(GET|POST|PUT|DELETE|PATCH)/);
  return m ? m[1] : '-';
}

function extractPath(action: string): string {
  return action
    .replace(/^(GET|POST|PUT|DELETE|PATCH)\s+/, '')
    .replace(/\/api\/v1\//, '')
    .replace(/:id/g, '*');
}

function shortenId(id: string): string {
  if (!id) return '-';
  if (id.length > 12) return id.substring(0, 8) + '…';
  return id;
}

export function AuditLogPage() {
  const [logs, setLogs] = useState<AuditLog[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(false);
  const [filters, setFilters] = useState<AuditLogFilters>({ page: 1, pageSize: 20 });

  const load = useCallback(async () => {
    setLoading(true);
    try {
      const data = await fetchAuditLogs(filters);
      setLogs(data.items || []);
      setTotal(data.total || 0);
    } catch {
      setLogs([]);
    } finally {
      setLoading(false);
    }
  }, [filters]);

  useEffect(() => { load(); }, [load]);

  const updateFilter = (key: string, value: string | undefined) => {
    setFilters((prev) => ({ ...prev, [key]: value || undefined, page: 1 }));
  };

  const columns = [
    {
      title: '时间',
      dataIndex: 'createdAt',
      width: 160,
      render: (v: string) => (
        <span style={{ fontSize: 12, color: 'var(--on-surface-variant)', whiteSpace: 'nowrap' }}>
          {new Date(v).toLocaleString()}
        </span>
      ),
    },
    {
      title: '用户',
      dataIndex: 'userId',
      width: 100,
      render: (v: string) => (
        <span style={{ fontSize: 13 }}>{v?.replace('admin-bootstrap-', '#') || '-'}</span>
      ),
    },
    {
      title: '方法',
      dataIndex: 'action',
      width: 80,
      render: (val: string) => {
        const method = extractMethod(val);
        return (
          <span className={`pipeline-status-badge ${METHOD_CLS[method] || 'pipeline-status-badge--pending'}`}>
            {method}
          </span>
        );
      },
    },
    {
      title: '接口',
      dataIndex: 'action',
      width: 200,
      ellipsis: true,
      render: (val: string) => {
        const path = extractPath(val);
        const short = path.length > 28 ? path.substring(0, 28) + '…' : path;
        return (
          <Tooltip content={val} mini>
            <code className="mono" style={{ color: 'var(--on-surface-variant)' }}>{short}</code>
          </Tooltip>
        );
      },
    },
    {
      title: '资源',
      dataIndex: 'resourceType',
      width: 100,
      render: (v: string) => <span style={{ fontSize: 12 }}>{v || '-'}</span>,
    },
    {
      title: 'ID',
      dataIndex: 'resourceId',
      width: 100,
      render: (v: string) => (
        <Tooltip content={v} mini>
          <code className="mono" style={{ color: 'var(--on-surface-variant)' }}>{shortenId(v)}</code>
        </Tooltip>
      ),
    },
    {
      title: '结果',
      dataIndex: 'result',
      width: 80,
      render: (val: string) => (
        <span className={`pipeline-status-badge ${RESULT_CLS[val] || 'pipeline-status-badge--pending'}`}>
          {RESULT_LABEL[val] || val}
        </span>
      ),
    },
    {
      title: 'IP',
      dataIndex: 'ip',
      width: 120,
      render: (v: string) => (
        <code className="mono" style={{ fontSize: 11, color: 'var(--on-surface-variant)' }}>
          {v || '-'}
        </code>
      ),
    },
  ];

  return (
    <AppLayout>
      <div className="page-container">
        <div className="page-header">
          <div>
            <div className="breadcrumb">System · Compliance</div>
            <h1 className="page-title">审计日志</h1>
            <p className="page-subtitle">全量 API 操作记录与合规追溯。</p>
          </div>
          <Space wrap size={8}>
            <Select
              placeholder="方法"
              options={ACTION_OPTIONS}
              style={{ width: 120 }}
              allowClear
              onChange={(v) => updateFilter('action', v)}
            />
            <Input
              placeholder="用户"
              style={{ width: 140 }}
              allowClear
              onChange={(v) => updateFilter('userId', v)}
            />
            <DatePicker.RangePicker
              style={{ width: 280 }}
              showTime
              onChange={(_v, dates) => {
                setFilters((prev) => ({
                  ...prev,
                  startTime: dates?.[0] ? String(dates[0]) : undefined,
                  endTime: dates?.[1] ? String(dates[1]) : undefined,
                  page: 1,
                }));
              }}
            />
          </Space>
        </div>
        <div className="table-card">
          <Table
            columns={columns}
            data={logs}
            rowKey="id"
            loading={loading}
            border={false}
            scroll={{ x: 960 }}
            pagination={{
              total,
              current: filters.page,
              pageSize: filters.pageSize,
              showTotal: true,
              onChange: (page, pageSize) =>
                setFilters((prev) => ({ ...prev, page, pageSize })),
            }}
            noDataElement={
              <div className="empty-state">
                <div className="empty-state-title">暂无审计记录</div>
                <div className="empty-state-desc">当前筛选条件下没有匹配的日志</div>
              </div>
            }
          />
        </div>
      </div>
    </AppLayout>
  );
}

export default AuditLogPage;
