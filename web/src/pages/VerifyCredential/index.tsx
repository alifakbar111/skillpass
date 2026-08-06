import { useMutation } from '@tanstack/react-query';
import { BadgeCheck, ShieldAlert, ShieldCheck } from 'lucide-react';
import { useEffect, useState } from 'react';
import { useSearchParams } from 'react-router-dom';
import { verifyCredential } from '@/lib/verify';

/**
 * Public credential verifier (Phase 2 · Sprint 6). Paste a credential id (or
 * arrive via ?id=…) and the page re-checks the Ed25519 signature against the
 * issuer's published keys (JWKS). This replaces "view on-chain": authenticity
 * comes from the signature, not a ledger.
 */
export default function VerifyCredential() {
  const [params] = useSearchParams();
  const [id, setId] = useState(params.get('id') ?? '');

  const verifyMut = useMutation({ mutationFn: (credId: string) => verifyCredential(credId) });

  // Auto-verify once when arriving with ?id=… (intentional mount-only effect).
  const initialId = params.get('id');
  const verify = verifyMut.mutate;
  useEffect(() => {
    if (initialId) verify(initialId);
  }, [initialId, verify]);

  const result = verifyMut.data;

  return (
    <div className="mx-auto max-w-xl p-6">
      <div className="mb-6 flex items-center gap-2">
        <ShieldCheck className="h-6 w-6 text-primary" />
        <h1 className="text-2xl font-semibold tracking-tight">Verify a credential</h1>
      </div>

      <form
        className="flex gap-2"
        onSubmit={(e) => {
          e.preventDefault();
          if (id.trim()) verifyMut.mutate(id.trim());
        }}
      >
        <input
          value={id}
          onChange={(e) => setId(e.target.value)}
          placeholder="Credential ID"
          className="input input-bordered flex-1"
        />
        <button type="submit" className="btn btn-primary" disabled={verifyMut.isPending || !id.trim()}>
          Verify
        </button>
      </form>

      {verifyMut.isError && (
        <div className="mt-6 flex items-center gap-2 rounded-lg bg-error/10 p-4 text-error">
          <ShieldAlert className="h-5 w-5" /> Credential not found.
        </div>
      )}

      {result && (
        <div
          className={[
            'mt-6 rounded-2xl border p-6',
            result.verified ? 'border-success/40 bg-success/5' : 'border-error/40 bg-error/5',
          ].join(' ')}
        >
          <div className="flex items-center gap-2">
            {result.verified ? (
              <BadgeCheck className="h-6 w-6 text-success" />
            ) : (
              <ShieldAlert className="h-6 w-6 text-error" />
            )}
            <span className="text-lg font-semibold">
              {result.verified ? 'Signature valid' : result.revoked ? 'Revoked' : 'Signature invalid'}
            </span>
          </div>
          <dl className="mt-4 grid grid-cols-3 gap-2 text-sm">
            <dt className="text-base-content/50">Skill</dt>
            <dd className="col-span-2 font-medium">{result.skillName}</dd>
            <dt className="text-base-content/50">Score</dt>
            <dd className="col-span-2 font-medium">{result.score}</dd>
            <dt className="text-base-content/50">Issuer</dt>
            <dd className="col-span-2 break-all font-mono text-xs">{result.issuerDid}</dd>
            <dt className="text-base-content/50">Issued</dt>
            <dd className="col-span-2">{new Date(result.issuedAt).toLocaleString()}</dd>
          </dl>
        </div>
      )}
    </div>
  );
}
