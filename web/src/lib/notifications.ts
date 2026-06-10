import { api } from './api';

export interface Notification {
  id: string;
  userId: string;
  type: string;
  title: string;
  body: string;
  link: string;
  readAt: string | null;
  createdAt: string;
}

export interface NotificationList {
  notifications: Notification[];
  unreadCount: number;
}

export async function getNotifications(): Promise<NotificationList> {
  return api<NotificationList>('/notifications/me');
}

export async function markNotificationRead(id: string): Promise<void> {
  await api(`/notifications/${id}/read`, { method: 'PUT' });
}

export async function markAllNotificationsRead(): Promise<void> {
  await api('/notifications/read-all', { method: 'PUT' });
}
