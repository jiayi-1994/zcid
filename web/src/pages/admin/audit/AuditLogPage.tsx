import { DatePicker, Input, Select, Space, Table, Tag, Tooltip } from '@arco-design/web-react';
import { useCallback, useEffect, useState } from 'react';
import { AppLayout } from '../../../components/layout/AppLayout';
import { fetchAuditLogs, type AuditLog, type AuditLogFilters } from '../../../services/audit';

const RESULT_CONFIG: Record<string, { color: string; text: string }> = {
  success: { color: 'green', text: '成功' },
  failure: { color: 'red', text: '失败' },
};

const ACTION_OPTIONS = [
  { label: '全部', value: '' },
  { label: 'POST', value: 'POST' },
  { label: 'PUT', value: 'PUT' },
  { label: 'DELETE', value: 'DELETE' },
];

function extractResourceLabel(path: string): string {
  if (!path) return '-';
  const parts = path.replace(/^(GET|POST|PUT|DELETE|PATCH)\s+/i, '').split('/').filter(Boolean);
  const meaningful = parts.filter(p => !p.startsWith(':') && p !== 'api' && p !== 'v1' && !p.match(/^[0-9a-f-]{20,}$/));
  return meaningful.slice(-2).join('/') || path;
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
        <span style={{ fontSize: 12, color: 'var(--muted-foreground)', whiteSpace: 'nowrap' }}>
          {new Date(v).toLocaleString()}
        </span>
      ),
    },
    {
      title: '用户',
      dataIndex: 'userId',
      width: 100,
      render: (v: string) => (
        <span style={{ fontSize: 13 }}>{v?.replace('admin-bootstrap-', 'admin#') || '-'}</span>
      ),
    },
    {
      title: '操作',
      dataIndex: 'action',
      width: 70,
      render: (val: string) => (
        <Tag size="small" style={{ borderRadius: 999, fontSize: 11 }}>{val}</Tag>
      ),
    },
    {
      title: '资源',
      dataIndex: 'resourceType',
      width: 150,
      render: (v: string) => (
        <Tooltip content={v} position="top" mini>
          <span style={{
            fontSize: 12, fontFamily: 'var(--font-mono)',
            display: 'inline-block', maxWidth: 140,
            overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap',
          }}>
            {extractResourceLabel(v)}
          </span>
        </Tooltip>
      ),
    },
    {
      title: '资源 ID',
      dataIndex: 'resourceId',
      width: 100,
      render: (v: string) => (
        <Tooltip content={v} position="top" mini>
          <code style={{ fontSize: 11, color: 'var(--muted-foreground)' }}>
            {shortenId(v)}
          </code>
        </Tooltip>
      ),
    },
    {
      title: '结果',
      dataIndex: 'result',
      width: 60,
      render: (val: string) => {
        const cfg = RESULT_CONFIG[val] || { color: 'gray', text: val };
        return <Tag size="small" color={cfg.color} style={{ borderRadius: 999, fontSize: 11 }}>{cfg.text}</Tag>;
      },
    },
    {
      title: 'IP',
      dataIndex: 'ip',
      width: 80,
      render: (v: string) => (
        <span style={{ fontSize: 12, fontFamily: 'var(--font-mono)' }}>{v || '-'}</span>
      ),
    },
  ];

  return (
    <AppLayout>
      <div className="page-container">
        <div className="page-header">
          <h3 className="page-title">审计日志</h3>
          <Space wrap size={8}>
            <Select
              placeholder="操作类型"
              options={ACTION_OPTIONS}
              style={{ width: 110 }}
              allowClear
              size="small"
              onChange={(v) => updateFilter('action', v)}
            />
            <Input
              placeholder="资源类型"
              style={{ width: 120 }}
              allowClear
              size="small"
              onChange={(v) => updateFilter('resourceType', v)}
            />
            <Input
              placeholder="用户 ID"
              style={{ width: 120 }}
              allowClear
              size="small"
              onChange={(v) => updateFilter('userId', v)}
            />
            <DatePicker.RangePicker
              style={{ width: 280 }}
              size="small"
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
        <Table
          columns={columns}
          data={logs}
          rowKey="id"
          loading={loading}
          border={false}
          stripe
          style={{ tableLayout: 'fixed' }}
          scroll={{ x: 720 }}
          pagination={{
            total,
            current: filters.page,
            pageSize: filters.pageSize,
            showTotal: true,
            onChange: (page, pageSize) => setFilters((prev) => ({ ...prev, page, pageSize })),
          }}
        />
      </div>
    </AppLayout>
  );
}

export default AuditLogPage;
