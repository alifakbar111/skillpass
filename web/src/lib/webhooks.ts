import { api } from './api';

export interface Webhook {
  id: string;
  url: string;
  secret?: string;
  active: boolean;
  createdAt: string;
}

export async function getWebhooks(): Promise<Webhook[]> {
  return api<Webhook[]>('/company/webhooks');
}

export async function createWebhook(url: string): Promise<Webhook> {
  return api<Webhook>('/company/webhooks', {
    method: 'POST',
    body: JSON.stringify({ url }),
  });
}

export async function deleteWebhook(id: string): Promise<void> {
  await api(`/company/webhooks/${id}`, { method: 'DELETE' });
}
