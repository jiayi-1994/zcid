import { DatePicker, Input, Select, Space, Table, Tag } from '@arco-design/web-react';
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
      width: 170,
      render: (v: string) => <span style={{ fontSize: 12, color: 'var(--muted-foreground)', whiteSpace: 'nowrap' }}>{new Date(v).toLocaleString()}</span>,
    },
    {
      title: '用户',
      dataIndex: 'userId',
      width: 130,
      ellipsis: true,
      render: (v: string) => <span style={{ fontSize: 13 }}>{v?.replace('admin-bootstrap-', 'admin#') || '-'}</span>,
    },
    {
      title: '操作',
      dataIndex: 'action',
      width: 80,
      render: (val: string) => <Tag size="small" style={{ borderRadius: 999 }}>{val}</Tag>,
    },
    {
      title: '资源类型',
      dataIndex: 'resourceType',
      width: 140,
      render: (v: string) => <span style={{ fontSize: 13 }}>{v || '-'}</span>,
    },
    {
      title: '资源 ID',
      dataIndex: 'resourceId',
      width: 180,
      ellipsis: true,
      render: (v: string) => <span style={{ fontSize: 12, fontFamily: 'var(--font-mono)', color: 'var(--muted-foreground)' }}>{v ? v.substring(0, 16) + '...' : '-'}</span>,
    },
    {
      title: '结果',
      dataIndex: 'result',
      width: 70,
      render: (val: string) => {
        const cfg = RESULT_CONFIG[val] || { color: 'gray', text: val };
        return <Tag size="small" color={cfg.color} style={{ borderRadius: 999 }}>{cfg.text}</Tag>;
      },
    },
    {
      title: 'IP',
      dataIndex: 'ip',
      width: 110,
      render: (v: string) => <span style={{ fontSize: 12, fontFamily: 'var(--font-mono)' }}>{v || '-'}</span>,
    },
  ];

  return (
    <AppLayout>
      <div className="page-container">
        <div className="page-header">
          <h3 className="page-title">审计日志</h3>
          <Space wrap>
          <Select
            placeholder="操作类型"
            options={ACTION_OPTIONS}
            style={{ width: 120 }}
            allowClear
            onChange={(v) => updateFilter('action', v)}
          />
          <Input
            placeholder="资源类型"
            style={{ width: 140 }}
            allowClear
            onChange={(v) => updateFilter('resourceType', v)}
          />
          <Input
            placeholder="用户 ID"
            style={{ width: 200 }}
            allowClear
            onChange={(v) => updateFilter('userId', v)}
          />
          <DatePicker.RangePicker
            style={{ width: 320 }}
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
          scroll={{ x: 900 }}
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
