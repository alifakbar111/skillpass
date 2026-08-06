import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { BadgeCheck, Building2 } from 'lucide-react';
import { useState } from 'react';
import { useParams } from 'react-router-dom';
import { acceptOffer, declineOffer, getPublicOffer } from '@/lib/verify';

/**
 * Public offer-acceptance page (Phase 2 · Sprint 5). Reached via a tokenized
 * link — no login required. Accepting captures a typed signature; on the server
 * this bridges the candidate into the company's HRIS, linking their existing
 * SkillPass login if one matches their email.
 */
export default function OfferAccept() {
  const { token = '' } = useParams();
  const qc = useQueryClient();
  const [signature, setSignature] = useState('');

  const {
    data: offer,
    isLoading,
    error,
  } = useQuery({
    queryKey: ['public-offer', token],
    queryFn: () => getPublicOffer(token),
    retry: false,
  });

  const acceptMut = useMutation({
    mutationFn: () => acceptOffer(token, signature),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['public-offer', token] }),
  });
  const declineMut = useMutation({
    mutationFn: () => declineOffer(token),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['public-offer', token] }),
  });

  if (isLoading)
    return (
      <div className="flex min-h-[60vh] items-center justify-center">
        <span className="loading loading-spinner loading-lg" />
      </div>
    );

  if (error || !offer)
    return (
      <div className="mx-auto max-w-lg p-8 text-center">
        <h1 className="text-xl font-semibold">Offer not found</h1>
        <p className="mt-2 text-base-content/60">This offer link is invalid or has expired.</p>
      </div>
    );

  const settled = offer.status === 'accepted' || offer.status === 'declined';

  return (
    <div className="mx-auto max-w-2xl p-6">
      <div className="mb-6 flex items-center gap-2 text-base-content/60">
        <Building2 className="h-5 w-5" />
        <span className="font-medium">{offer.companyName}</span>
      </div>

      <div className="rounded-2xl border border-base-300 bg-base-100 p-6 shadow-sm">
        <h1 className="text-2xl font-semibold tracking-tight">Offer: {offer.positionTitle}</h1>
        <p className="mt-1 text-base-content/60">Prepared for {offer.candidateName}</p>
        <div className="mt-3 flex flex-wrap gap-2">
          {offer.salary && <span className="badge badge-ghost">Salary: {offer.salary}</span>}
          {offer.startDate && <span className="badge badge-ghost">Start: {offer.startDate}</span>}
        </div>

        <div className="mt-5 whitespace-pre-wrap rounded-lg bg-base-200/50 p-4 text-sm leading-relaxed">
          {offer.body}
        </div>

        {acceptMut.data ? (
          <div className="mt-6 rounded-lg bg-success/10 p-4 text-success">
            <div className="flex items-center gap-2 font-medium">
              <BadgeCheck className="h-5 w-5" /> {acceptMut.data.message}
            </div>
            {acceptMut.data.employeeLinked && (
              <p className="mt-1 text-sm">Your existing SkillPass login now also unlocks your employee self-service.</p>
            )}
          </div>
        ) : settled ? (
          <div className="mt-6 rounded-lg bg-base-200 p-4 text-center text-base-content/70">
            This offer has already been {offer.status}.
          </div>
        ) : (
          <div className="mt-6 space-y-3">
            <label htmlFor="sig" className="label">
              <span className="label-text font-medium">Type your full name to sign &amp; accept</span>
            </label>
            <input
              id="sig"
              value={signature}
              onChange={(e) => setSignature(e.target.value)}
              placeholder="Your full legal name"
              className="input input-bordered w-full"
            />
            {acceptMut.error && (
              <div className="alert alert-error text-sm">Could not accept the offer. Check your signature.</div>
            )}
            <div className="flex gap-2">
              <button
                type="button"
                className="btn btn-primary"
                disabled={signature.trim().length < 2 || acceptMut.isPending}
                onClick={() => acceptMut.mutate()}
              >
                Sign &amp; Accept
              </button>
              <button
                type="button"
                className="btn btn-ghost"
                disabled={declineMut.isPending}
                onClick={() => declineMut.mutate()}
              >
                Decline
              </button>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
