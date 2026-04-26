import { http } from './http';

export interface AuditLog {
  id: string;
  userId?: string;
  action: string;
  resourceType: string;
  resourceId?: string;
  result: string;
  ip?: string;
  detail?: string;
  createdAt: string;
}

export interface AuditLogList {
  items: AuditLog[];
  total: number;
  page: number;
  pageSize: number;
}

export interface AuditLogFilters {
  page?: number;
  pageSize?: number;
  userId?: string;
  action?: string;
  resourceType?: string;
  category?: string;
  startTime?: string;
  endTime?: string;
}

export async function fetchAuditLogs(filters: AuditLogFilters = {}) {
  const resp = await http.get<{ code: number; data: AuditLogList }>(
    '/admin/audit-logs',
    { params: filters },
  );
  return resp.data.data;
}
