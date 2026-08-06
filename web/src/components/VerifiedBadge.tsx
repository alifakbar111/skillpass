import { BadgeCheck, ShieldAlert } from 'lucide-react';

interface Props {
  verified: boolean;
  /** Attestation/credential id — when set, the badge links to /verify. */
  credentialId?: string;
  label?: string;
  size?: 'sm' | 'md';
}

/**
 * VerifiedBadge shows a "Signed & verified" chip backed by an Ed25519 signature
 * check (Phase 2 · Sprint 6). It replaces the old "view on Polygon" tx link:
 * clicking opens the public /verify page for the credential.
 */
export function VerifiedBadge({ verified, credentialId, label, size = 'sm' }: Props) {
  const text = label ?? (verified ? 'Signed & verified' : 'Unverified');
  const badge = (
    <span
      className={[
        'inline-flex items-center gap-1 rounded-full font-medium',
        size === 'sm' ? 'px-2 py-0.5 text-xs' : 'px-2.5 py-1 text-sm',
        verified ? 'bg-success/15 text-success' : 'bg-warning/15 text-warning',
      ].join(' ')}
    >
      {verified ? <BadgeCheck className="h-3.5 w-3.5" /> : <ShieldAlert className="h-3.5 w-3.5" />}
      {text}
    </span>
  );

  if (credentialId) {
    return (
      <a
        href={`/verify?id=${encodeURIComponent(credentialId)}`}
        target="_blank"
        rel="noreferrer"
        className="hover:opacity-80"
        title="Verify this credential's signature"
      >
        {badge}
      </a>
    );
  }
  return badge;
}
