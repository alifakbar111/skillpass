import { api } from '@/lib/api';

export interface ActivityLog {
  id: string;
  companyId: string;
  actorId: string;
  actorName: string;
  action: string;
  entityType: string;
  entityId?: string;
  metadata?: Record<string, unknown>;
  createdAt: string;
}

export interface ActivityLogResponse {
  logs: ActivityLog[];
  total: number;
  limit: number;
  offset: number;
}

export function listActivityLogs(limit = 50, offset = 0) {
  return api<ActivityLogResponse>(`/hris/activity-logs?limit=${limit}&offset=${offset}`);
}
