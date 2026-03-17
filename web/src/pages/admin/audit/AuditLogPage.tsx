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

function extractMethod(action: string): string {
  const m = action.match(/^(GET|POST|PUT|DELETE|PATCH)/);
  return m ? m[1] : '-';
}

function extractPath(action: string): string {
  return action.replace(/^(GET|POST|PUT|DELETE|PATCH)\s+/, '').replace(/\/api\/v1\//, '').replace(/:id/g, '*');
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
      width: 155,
      render: (v: string) => (
        <span style={{ fontSize: 12, color: 'var(--muted-foreground)', whiteSpace: 'nowrap' }}>
          {new Date(v).toLocaleString()}
        </span>
      ),
    },
    {
      title: '用户',
      dataIndex: 'userId',
      width: 90,
      render: (v: string) => (
        <span style={{ fontSize: 13 }}>{v?.replace('admin-bootstrap-', '#') || '-'}</span>
      ),
    },
    {
      title: '方法',
      dataIndex: 'action',
      width: 65,
      render: (val: string) => {
        const method = extractMethod(val);
        const colors: Record<string, string> = { POST: 'green', PUT: 'orangered', DELETE: 'red', GET: 'blue' };
        return <Tag size="small" color={colors[method] || 'gray'} style={{ borderRadius: 4, fontSize: 11 }}>{method}</Tag>;
      },
    },
    {
      title: '接口路径',
      dataIndex: 'action',
      width: 200,
      render: (val: string) => {
        const path = extractPath(val);
        return (
          <Tooltip content={val} mini>
            <code style={{ fontSize: 11, color: 'var(--muted-foreground)', whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis', display: 'block', maxWidth: 190 }}>
              {path}
            </code>
          </Tooltip>
        );
      },
    },
    {
      title: '资源',
      dataIndex: 'resourceType',
      width: 80,
      render: (v: string) => <span style={{ fontSize: 12 }}>{v || '-'}</span>,
    },
    {
      title: 'ID',
      dataIndex: 'resourceId',
      width: 85,
      render: (v: string) => (
        <Tooltip content={v} mini>
          <code style={{ fontSize: 11, color: 'var(--muted-foreground)' }}>{shortenId(v)}</code>
        </Tooltip>
      ),
    },
    {
      title: '结果',
      dataIndex: 'result',
      width: 55,
      render: (val: string) => {
        const cfg = RESULT_CONFIG[val] || { color: 'gray', text: val };
        return <Tag size="small" color={cfg.color} style={{ borderRadius: 999, fontSize: 11 }}>{cfg.text}</Tag>;
      },
    },
    {
      title: 'IP',
      dataIndex: 'ip',
      width: 70,
      render: (v: string) => <span style={{ fontSize: 11, fontFamily: 'var(--font-mono)' }}>{v || '-'}</span>,
    },
  ];

  return (
    <AppLayout>
      <div className="page-container">
        <div className="page-header">
          <h3 className="page-title">审计日志</h3>
          <Space wrap size={8}>
            <Select placeholder="操作" options={ACTION_OPTIONS} style={{ width: 90 }} allowClear size="small" onChange={(v) => updateFilter('action', v)} />
            <Input placeholder="用户" style={{ width: 100 }} allowClear size="small" onChange={(v) => updateFilter('userId', v)} />
            <DatePicker.RangePicker
              style={{ width: 260 }}
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
          scroll={{ x: 800 }}
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
