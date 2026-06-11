import { Copy, Plus, Trash2, Webhook as WebhookIcon } from 'lucide-react';
import { useEffect, useState } from 'react';
import { ApiError } from '../../lib/api';
import { createWebhook, deleteWebhook, getWebhooks, type Webhook } from '../../lib/webhooks';

export function WebhooksSection() {
  const [webhooks, setWebhooks] = useState<Webhook[]>([]);
  const [url, setUrl] = useState('');
  const [newSecret, setNewSecret] = useState<string | null>(null);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    getWebhooks()
      .then(setWebhooks)
      .catch(() => {
        // Section is non-critical; company may not be verified yet.
      });
  }, []);

  async function handleAdd() {
    const trimmed = url.trim();
    if (!trimmed || saving) return;
    setSaving(true);
    setError(null);
    setNewSecret(null);
    try {
      const created = await createWebhook(trimmed);
      setWebhooks((prev) => [created, ...prev]);
      setNewSecret(created.secret ?? null);
      setUrl('');
    } catch (err) {
      setError(err instanceof ApiError ? (err.serverMessage ?? err.message) : 'Failed to add webhook');
    } finally {
      setSaving(false);
    }
  }

  async function handleDelete(id: string) {
    setError(null);
    try {
      await deleteWebhook(id);
      setWebhooks((prev) => prev.filter((w) => w.id !== id));
    } catch (err) {
      setError(err instanceof ApiError ? (err.serverMessage ?? err.message) : 'Failed to delete webhook');
    }
  }

  return (
    <div className="card bg-base-200 p-4 mt-4">
      <div className="flex items-center gap-2 mb-2">
        <WebhookIcon size={18} className="text-primary" aria-hidden="true" />
        <h2 className="font-semibold">Webhooks</h2>
      </div>
      <p className="text-sm opacity-70 mb-3">
        Get notified in your own tools when candidates apply. Events are POSTed as JSON, signed with HMAC-SHA256 in the{' '}
        <code className="text-xs">X-SkillPass-Signature</code> header.
      </p>

      {error && <p className="text-error text-sm mb-2">{error}</p>}

      {newSecret && (
        <div className="alert alert-warning mb-3 text-sm">
          <div>
            <p className="font-semibold">Save this signing secret — it won't be shown again:</p>
            <div className="flex items-center gap-2 mt-1">
              <code className="text-xs break-all">{newSecret}</code>
              <button
                type="button"
                className="btn btn-ghost btn-xs"
                title="Copy secret"
                onClick={() => navigator.clipboard?.writeText(newSecret)}
              >
                <Copy size={12} />
              </button>
            </div>
          </div>
        </div>
      )}

      <div className="flex gap-2 mb-3">
        <input
          type="url"
          className="input input-sm input-bordered flex-1"
          placeholder="https://your-server.com/webhooks/skillpass"
          value={url}
          onChange={(e) => setUrl(e.target.value)}
        />
        <button type="button" className="btn btn-sm btn-primary" onClick={handleAdd} disabled={saving || !url.trim()}>
          {saving ? <span className="loading loading-spinner loading-xs" /> : <Plus size={14} />} Add
        </button>
      </div>

      {webhooks.length === 0 ? (
        <p className="text-xs opacity-50">No webhooks registered.</p>
      ) : (
        <ul className="space-y-2">
          {webhooks.map((w) => (
            <li key={w.id} className="flex justify-between items-center bg-base-100 rounded-lg p-2">
              <div className="min-w-0">
                <code className="text-xs break-all">{w.url}</code>
                <div className="text-[10px] opacity-50">{w.createdAt.slice(0, 10)}</div>
              </div>
              <button
                type="button"
                className="btn btn-ghost btn-xs text-error shrink-0"
                onClick={() => handleDelete(w.id)}
                aria-label={`Delete webhook ${w.url}`}
              >
                <Trash2 size={14} />
              </button>
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}
