import { api, getAccessToken } from '@/lib/api';

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

export async function clearAllNotifications(): Promise<void> {
  await api('/notifications', { method: 'DELETE' });
}

/**
 * Subscribe to real-time notification updates via SSE (Server-Sent Events).
 *
 * The server sends `notification` events with JSON data in two shapes:
 *
 *   {"type":"init","data":{"notifications":[...],"unreadCount":5}}
 *   {"type":"refresh","data":null}
 *
 * Returns an EventSource that the caller must `.close()` on cleanup.
 * Reconnection is handled automatically by the browser.
 */
export function subscribeToNotifications(onEvent: (event: MessageEvent) => void, onError?: () => void): EventSource {
  const token = getAccessToken();
  const params = token ? `?token=${encodeURIComponent(token)}` : '';
  const url = `/api/v1/notifications/stream${params}`;
  const es = new EventSource(url);
  es.addEventListener('notification', onEvent);
  if (onError) es.onerror = onError;
  return es;
}
