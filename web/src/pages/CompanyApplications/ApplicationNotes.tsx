import { Send } from 'lucide-react';
import { useEffect, useState } from 'react';
import { LoadingSpinner } from '../../components/ui/LoadingFallback';
import { ApiError } from '../../lib/api';
import { addApplicationMessage, type ApplicationMessage, getApplicationMessages } from '../../lib/application';

export function ApplicationNotes({ applicationId }: { applicationId: string }) {
  const [messages, setMessages] = useState<ApplicationMessage[]>([]);
  const [loading, setLoading] = useState(true);
  const [body, setBody] = useState('');
  const [sending, setSending] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;
    getApplicationMessages(applicationId)
      .then((data) => {
        if (!cancelled) setMessages(data);
      })
      .catch((err) => {
        if (!cancelled) setError(err instanceof ApiError ? (err.serverMessage ?? err.message) : 'Failed to load notes');
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });
    return () => {
      cancelled = true;
    };
  }, [applicationId]);

  async function handleSend() {
    const trimmed = body.trim();
    if (!trimmed || sending) return;
    setSending(true);
    setError(null);
    try {
      const msg = await addApplicationMessage(applicationId, trimmed);
      setMessages((prev) => [...prev, msg]);
      setBody('');
    } catch (err) {
      setError(err instanceof ApiError ? (err.serverMessage ?? err.message) : 'Failed to send note');
    } finally {
      setSending(false);
    }
  }

  return (
    <div className="py-2 px-1">
      <h4 className="text-sm font-semibold opacity-80 mb-2">Notes to candidate</h4>

      {loading ? (
        <LoadingSpinner />
      ) : (
        <div className="space-y-2 mb-3">
          {messages.length === 0 && (
            <p className="text-xs opacity-50">No notes yet. Add one below — the candidate will see it.</p>
          )}
          {messages.map((m) => (
            <div key={m.id} className="bg-base-200 rounded-lg p-2">
              <div className="text-sm">{m.body}</div>
              <div className="text-[10px] opacity-50 mt-1">
                {m.senderName} · {m.createdAt.slice(0, 10)}
              </div>
            </div>
          ))}
        </div>
      )}

      {error && <p className="text-error text-xs mb-2">{error}</p>}

      <div className="flex gap-2">
        <input
          type="text"
          className="input input-sm input-bordered flex-1"
          placeholder="Write a note to the candidate…"
          value={body}
          onChange={(e) => setBody(e.target.value)}
          onKeyDown={(e) => {
            if (e.key === 'Enter') handleSend();
          }}
        />
        <button
          type="button"
          className="btn btn-sm btn-primary"
          onClick={handleSend}
          disabled={sending || !body.trim()}
        >
          {sending ? <span className="loading loading-spinner loading-xs" /> : <Send size={14} />}
        </button>
      </div>
    </div>
  );
}
