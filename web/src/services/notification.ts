import { http } from './http';

export interface NotificationRule {
  id: string;
  projectId: string;
  name: string;
  eventType: string;
  webhookUrl: string;
  enabled: boolean;
  createdBy: string;
  createdAt: string;
  updatedAt: string;
}

export interface NotificationRuleList {
  items: NotificationRule[];
  total: number;
  page: number;
  pageSize: number;
}

export interface CreateRuleRequest {
  name: string;
  eventType: string;
  webhookUrl: string;
  enabled?: boolean;
}

export interface UpdateRuleRequest {
  name?: string;
  eventType?: string;
  webhookUrl?: string;
  enabled?: boolean;
}

export async function fetchNotificationRules(projectId: string, page = 1, pageSize = 20) {
  const resp = await http.get<{ code: number; data: NotificationRuleList }>(
    `/projects/${projectId}/notification-rules`,
    { params: { page, pageSize } },
  );
  return resp.data.data;
}

export async function createNotificationRule(projectId: string, data: CreateRuleRequest) {
  const resp = await http.post<{ code: number; data: NotificationRule }>(
    `/projects/${projectId}/notification-rules`,
    data,
  );
  return resp.data.data;
}

export async function updateNotificationRule(projectId: string, ruleId: string, data: UpdateRuleRequest) {
  const resp = await http.put<{ code: number; data: NotificationRule }>(
    `/projects/${projectId}/notification-rules/${ruleId}`,
    data,
  );
  return resp.data.data;
}

export async function deleteNotificationRule(projectId: string, ruleId: string) {
  await http.delete(`/projects/${projectId}/notification-rules/${ruleId}`);
}
