import { useQuery } from '@tanstack/react-query';
import { BadgeCheck, ShieldCheck } from 'lucide-react';
import { useParams } from 'react-router-dom';
import { VerifiedBadge } from '@/components/VerifiedBadge';
import { getPublicPassport } from '@/lib/verify';

/**
 * Public verifiable Skill Passport (Phase 2 · Sprint 6). Renders an employee's
 * signature-verified skill badges. Each badge links to the public verifier so
 * anyone can independently confirm the Ed25519 signature. Distinct from the
 * marketplace passport at /profiles/:username.
 */
export default function VerifiablePassport() {
  const { slug = '' } = useParams();
  const { data, isLoading, error } = useQuery({
    queryKey: ['verifiable-passport', slug],
    queryFn: () => getPublicPassport(slug),
    retry: false,
  });

  if (isLoading)
    return (
      <div className="flex min-h-[60vh] items-center justify-center">
        <span className="loading loading-spinner loading-lg" />
      </div>
    );

  if (error || !data)
    return (
      <div className="mx-auto max-w-lg p-8 text-center">
        <h1 className="text-xl font-semibold">Passport not available</h1>
        <p className="mt-2 text-base-content/60">This Skill Passport is private or does not exist.</p>
      </div>
    );

  return (
    <div className="mx-auto max-w-2xl p-6">
      <div className="rounded-2xl border border-base-300 bg-base-100 p-6 shadow-sm">
        <div className="flex items-start justify-between">
          <div>
            <h1 className="text-2xl font-semibold tracking-tight">{data.name}</h1>
            {data.companyName && <p className="text-base-content/60">{data.companyName}</p>}
          </div>
          {data.identityVerified && (
            <span className="inline-flex items-center gap-1 rounded-full bg-success/15 px-2.5 py-1 text-sm font-medium text-success">
              <ShieldCheck className="h-4 w-4" /> Identity verified
            </span>
          )}
        </div>

        <h2 className="mb-3 mt-6 flex items-center gap-2 text-sm font-semibold uppercase tracking-wider text-base-content/50">
          <BadgeCheck className="h-4 w-4" /> Verified skills
        </h2>
        {data.skills?.length ? (
          <ul className="space-y-2">
            {data.skills.map((s) => (
              <li
                key={s.attestationId}
                className="flex items-center justify-between rounded-lg border border-base-200 p-3"
              >
                <div>
                  <span className="font-medium">{s.skillName}</span>
                  <span className="ml-2 text-base-content/50">{s.score}/100</span>
                </div>
                <VerifiedBadge verified={s.verified} credentialId={s.attestationId} />
              </li>
            ))}
          </ul>
        ) : (
          <p className="text-sm text-base-content/50">No verified skills published yet.</p>
        )}

        <p className="mt-6 border-t border-base-200 pt-4 text-xs text-base-content/40">
          Credentials are signed by {data.issuerDid} (Ed25519). Verify any badge independently against the issuer's
          published keys.
        </p>
      </div>
    </div>
  );
}
