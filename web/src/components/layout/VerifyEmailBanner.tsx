import { MailWarning, X } from 'lucide-react';
import { useState } from 'react';
import { useAuth } from '../../hooks/useAuth';
import { api } from '../../lib/api';

// Soft prompt for unverified accounts. Verification is not enforced (yet) —
// this only nudges; nothing is gated on it.
export function VerifyEmailBanner() {
  const { user } = useAuth();
  const [dismissed, setDismissed] = useState(false);
  const [sent, setSent] = useState(false);
  const [sending, setSending] = useState(false);

  if (!user || user.isVerified || user.role === 'admin' || dismissed) return null;

  async function resend() {
    if (sending || sent) return;
    setSending(true);
    try {
      await api('/auth/resend-verification', { method: 'POST' });
      setSent(true);
    } catch {
      // Soft feature — swallow errors, user can retry.
    } finally {
      setSending(false);
    }
  }

  return (
    <div className="bg-warning/15 text-sm px-4 py-2 flex items-center justify-center gap-3 flex-wrap" role="status">
      <MailWarning size={16} className="text-warning shrink-0" aria-hidden="true" />
      <span>
        Please verify your email address — check your inbox for the confirmation link.
        {sent && <span className="text-success font-medium"> Sent!</span>}
      </span>
      {!sent && (
        <button type="button" className="btn btn-xs btn-warning btn-outline" onClick={resend} disabled={sending}>
          {sending ? 'Sending…' : 'Resend email'}
        </button>
      )}
      <button
        type="button"
        className="btn btn-ghost btn-xs"
        onClick={() => setDismissed(true)}
        aria-label="Dismiss verification reminder"
      >
        <X size={14} />
      </button>
    </div>
  );
}
