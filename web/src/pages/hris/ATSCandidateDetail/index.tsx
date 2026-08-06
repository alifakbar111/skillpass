import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { ArrowLeft, CalendarPlus, Copy, FileSignature, Star } from 'lucide-react';
import { useState } from 'react';
import { Link, useParams } from 'react-router-dom';
import { usePermissions } from '@/hooks/usePermissions';
import {
  type AtsOffer,
  addScorecard,
  createOffer,
  getCandidate,
  listInterviews,
  listOffers,
  listOfferTemplates,
  listScorecards,
  scheduleInterview,
  sendOffer,
  setCandidateStatus,
} from '@/lib/hris/ats';

export default function ATSCandidateDetail() {
  const { id = '' } = useParams();
  const qc = useQueryClient();
  const { hasPermission } = usePermissions();
  const canManage = hasPermission('ats.manage');
  const canScore = hasPermission('ats.scorecard');
  const canInterview = hasPermission('ats.interview');
  const canOffer = hasPermission('ats.offer');

  const invalidate = () => qc.invalidateQueries({ queryKey: ['hris', 'ats'] });

  const { data: candidate, isLoading } = useQuery({
    queryKey: ['hris', 'ats', 'candidate', id],
    queryFn: () => getCandidate(id),
  });

  const statusMut = useMutation({
    mutationFn: (status: 'rejected' | 'withdrawn' | 'active') => setCandidateStatus(id, status),
    onSuccess: invalidate,
  });

  if (isLoading || !candidate)
    return (
      <div className="flex justify-center p-8">
        <span className="loading loading-spinner loading-lg" />
      </div>
    );

  return (
    <div className="max-w-4xl">
      <Link to="/hris/ats" className="btn btn-ghost btn-sm mb-4">
        <ArrowLeft className="h-4 w-4" /> Back to pipeline
      </Link>

      <div className="mb-6 rounded-xl border border-base-300 bg-base-100 p-5">
        <div className="flex flex-wrap items-start justify-between gap-3">
          <div>
            <h1 className="text-2xl font-semibold tracking-tight">{candidate.candidateName}</h1>
            <p className="text-sm text-base-content/60">{candidate.candidateEmail}</p>
            <div className="mt-2 flex flex-wrap gap-2">
              {candidate.jobTitle && <span className="badge badge-ghost">{candidate.jobTitle}</span>}
              {candidate.currentStageName && <span className="badge badge-primary">{candidate.currentStageName}</span>}
              <StatusBadge status={candidate.status} />
            </div>
          </div>
          {canManage && candidate.status === 'active' && (
            <div className="flex gap-2">
              <button
                type="button"
                className="btn btn-error btn-outline btn-sm"
                onClick={() => statusMut.mutate('rejected')}
              >
                Reject
              </button>
              <button type="button" className="btn btn-ghost btn-sm" onClick={() => statusMut.mutate('withdrawn')}>
                Withdraw
              </button>
            </div>
          )}
        </div>
      </div>

      <ScorecardSection candidateId={id} canScore={canScore} />
      <InterviewSection candidateId={id} canInterview={canInterview} />
      <OfferSection candidateId={id} canOffer={canOffer} />
    </div>
  );
}

function StatusBadge({ status }: { status: string }) {
  const cls =
    status === 'hired'
      ? 'badge-success'
      : status === 'rejected'
        ? 'badge-error'
        : status === 'withdrawn'
          ? 'badge-ghost'
          : 'badge-info';
  return <span className={`badge ${cls}`}>{status}</span>;
}

function Section({ title, action, children }: { title: string; action?: React.ReactNode; children: React.ReactNode }) {
  return (
    <section className="mb-6 rounded-xl border border-base-300 bg-base-100 p-5">
      <div className="mb-3 flex items-center justify-between">
        <h2 className="text-lg font-semibold">{title}</h2>
        {action}
      </div>
      {children}
    </section>
  );
}

function ScorecardSection({ candidateId, canScore }: { candidateId: string; canScore: boolean }) {
  const qc = useQueryClient();
  const [open, setOpen] = useState(false);
  const { data: scorecards } = useQuery({
    queryKey: ['hris', 'ats', 'scorecards', candidateId],
    queryFn: () => listScorecards(candidateId),
  });
  const addMut = useMutation({
    mutationFn: (body: { overallRating: number; recommendation: string; notes: string }) =>
      addScorecard(candidateId, body),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['hris', 'ats', 'scorecards', candidateId] });
      setOpen(false);
    },
  });

  return (
    <Section
      title="Scorecards"
      action={
        canScore && (
          <button type="button" className="btn btn-ghost btn-sm" onClick={() => setOpen((v) => !v)}>
            <Star className="h-4 w-4" /> Add
          </button>
        )
      }
    >
      {open && (
        <form
          className="mb-4 space-y-2 rounded-lg bg-base-200/50 p-3"
          onSubmit={(e) => {
            e.preventDefault();
            const fd = new FormData(e.currentTarget);
            addMut.mutate({
              overallRating: Number(fd.get('overallRating')),
              recommendation: fd.get('recommendation') as string,
              notes: fd.get('notes') as string,
            });
          }}
        >
          <div className="flex gap-2">
            <select name="overallRating" className="select select-bordered select-sm" defaultValue="3">
              {[1, 2, 3, 4, 5].map((n) => (
                <option key={n} value={n}>
                  {n} ★
                </option>
              ))}
            </select>
            <select name="recommendation" className="select select-bordered select-sm flex-1" defaultValue="yes">
              <option value="strong_yes">Strong yes</option>
              <option value="yes">Yes</option>
              <option value="no">No</option>
              <option value="strong_no">Strong no</option>
            </select>
          </div>
          <textarea name="notes" placeholder="Notes" className="textarea textarea-bordered w-full" rows={2} />
          <button type="submit" className="btn btn-primary btn-sm" disabled={addMut.isPending}>
            Save scorecard
          </button>
        </form>
      )}
      {scorecards?.length ? (
        <ul className="space-y-2">
          {scorecards.map((s) => (
            <li key={s.id} className="rounded-lg border border-base-200 p-3">
              <div className="flex items-center justify-between">
                <span className="font-medium">{s.evaluatorName || 'Evaluator'}</span>
                <span className="text-warning">{'★'.repeat(s.overallRating ?? 0)}</span>
              </div>
              {s.recommendation && <span className="badge badge-ghost badge-sm mt-1">{s.recommendation}</span>}
              {s.notes && <p className="mt-1 text-sm text-base-content/70">{s.notes}</p>}
            </li>
          ))}
        </ul>
      ) : (
        <p className="text-sm text-base-content/50">No scorecards yet.</p>
      )}
    </Section>
  );
}

function InterviewSection({ candidateId, canInterview }: { candidateId: string; canInterview: boolean }) {
  const qc = useQueryClient();
  const [open, setOpen] = useState(false);
  const { data: interviews } = useQuery({
    queryKey: ['hris', 'ats', 'interviews', candidateId],
    queryFn: () => listInterviews(candidateId),
  });
  const addMut = useMutation({
    mutationFn: (body: {
      scheduledAt: string;
      mode: string;
      location: string;
      meetingLink: string;
      interviewer: string;
      notes: string;
    }) => scheduleInterview(candidateId, body),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['hris', 'ats', 'interviews', candidateId] });
      setOpen(false);
    },
  });

  return (
    <Section
      title="Interviews"
      action={
        canInterview && (
          <button type="button" className="btn btn-ghost btn-sm" onClick={() => setOpen((v) => !v)}>
            <CalendarPlus className="h-4 w-4" /> Schedule
          </button>
        )
      }
    >
      {open && (
        <form
          className="mb-4 space-y-2 rounded-lg bg-base-200/50 p-3"
          onSubmit={(e) => {
            e.preventDefault();
            const fd = new FormData(e.currentTarget);
            const local = fd.get('scheduledAt') as string;
            addMut.mutate({
              scheduledAt: new Date(local).toISOString(),
              mode: fd.get('mode') as string,
              location: fd.get('location') as string,
              meetingLink: fd.get('meetingLink') as string,
              interviewer: fd.get('interviewer') as string,
              notes: fd.get('notes') as string,
            });
          }}
        >
          <div className="flex flex-wrap gap-2">
            <input name="scheduledAt" type="datetime-local" className="input input-bordered input-sm" required />
            <select name="mode" className="select select-bordered select-sm" defaultValue="onsite">
              <option value="onsite">Onsite</option>
              <option value="online">Online</option>
            </select>
            <input name="interviewer" placeholder="Interviewer" className="input input-bordered input-sm flex-1" />
          </div>
          <div className="flex flex-wrap gap-2">
            <input name="location" placeholder="Location" className="input input-bordered input-sm flex-1" />
            <input name="meetingLink" placeholder="Meeting link" className="input input-bordered input-sm flex-1" />
          </div>
          <textarea name="notes" placeholder="Notes" className="textarea textarea-bordered w-full" rows={2} />
          <button type="submit" className="btn btn-primary btn-sm" disabled={addMut.isPending}>
            Schedule interview
          </button>
        </form>
      )}
      {interviews?.length ? (
        <ul className="space-y-2">
          {interviews.map((iv) => (
            <li key={iv.id} className="rounded-lg border border-base-200 p-3">
              <div className="flex items-center justify-between">
                <span className="font-medium">{new Date(iv.scheduledAt).toLocaleString()}</span>
                <span className="badge badge-ghost badge-sm">{iv.mode}</span>
              </div>
              <p className="text-sm text-base-content/60">
                {iv.interviewer && `with ${iv.interviewer} · `}
                {iv.location || iv.meetingLink || ''}
              </p>
              <span className="badge badge-outline badge-sm mt-1">{iv.status}</span>
            </li>
          ))}
        </ul>
      ) : (
        <p className="text-sm text-base-content/50">No interviews scheduled.</p>
      )}
    </Section>
  );
}

function OfferSection({ candidateId, canOffer }: { candidateId: string; canOffer: boolean }) {
  const qc = useQueryClient();
  const [open, setOpen] = useState(false);
  const [copied, setCopied] = useState('');
  const { data: offers } = useQuery({
    queryKey: ['hris', 'ats', 'offers', candidateId],
    queryFn: () => listOffers(candidateId),
  });
  const { data: templates } = useQuery({
    queryKey: ['hris', 'ats', 'offer-templates'],
    queryFn: listOfferTemplates,
    enabled: canOffer,
  });
  const invalidate = () => qc.invalidateQueries({ queryKey: ['hris', 'ats', 'offers', candidateId] });

  const createMut = useMutation({
    mutationFn: (body: {
      positionTitle: string;
      salary: string;
      startDate: string;
      templateId?: string;
      body: string;
    }) => createOffer(candidateId, body),
    onSuccess: () => {
      invalidate();
      setOpen(false);
    },
  });
  const sendMut = useMutation({ mutationFn: (offerId: string) => sendOffer(offerId), onSuccess: invalidate });

  function copyLink(offer: AtsOffer) {
    if (!offer.acceptToken) return;
    const url = `${window.location.origin}/offer/${offer.acceptToken}`;
    navigator.clipboard.writeText(url);
    setCopied(offer.id);
    setTimeout(() => setCopied(''), 1500);
  }

  return (
    <Section
      title="Offers"
      action={
        canOffer && (
          <button type="button" className="btn btn-ghost btn-sm" onClick={() => setOpen((v) => !v)}>
            <FileSignature className="h-4 w-4" /> New offer
          </button>
        )
      }
    >
      {open && (
        <form
          className="mb-4 space-y-2 rounded-lg bg-base-200/50 p-3"
          onSubmit={(e) => {
            e.preventDefault();
            const fd = new FormData(e.currentTarget);
            createMut.mutate({
              positionTitle: fd.get('positionTitle') as string,
              salary: fd.get('salary') as string,
              startDate: fd.get('startDate') as string,
              templateId: (fd.get('templateId') as string) || undefined,
              body: fd.get('body') as string,
            });
          }}
        >
          <div className="flex flex-wrap gap-2">
            <input
              name="positionTitle"
              placeholder="Position title"
              className="input input-bordered input-sm flex-1"
              required
            />
            <input name="salary" placeholder="Salary" className="input input-bordered input-sm" />
            <input name="startDate" type="date" className="input input-bordered input-sm" />
          </div>
          {!!templates?.length && (
            <select name="templateId" className="select select-bordered select-sm w-full" defaultValue="">
              <option value="">No template (write below)</option>
              {templates.map((t) => (
                <option key={t.id} value={t.id}>
                  {t.name}
                </option>
              ))}
            </select>
          )}
          <textarea
            name="body"
            placeholder="Offer letter body (leave blank to render from the selected template)"
            className="textarea textarea-bordered w-full"
            rows={3}
          />
          <button type="submit" className="btn btn-primary btn-sm" disabled={createMut.isPending}>
            Create draft
          </button>
        </form>
      )}
      {offers?.length ? (
        <ul className="space-y-2">
          {offers.map((o) => (
            <li key={o.id} className="rounded-lg border border-base-200 p-3">
              <div className="flex flex-wrap items-center justify-between gap-2">
                <span className="font-medium">{o.positionTitle}</span>
                <span className="badge badge-outline badge-sm">{o.status}</span>
              </div>
              {o.salary && <p className="text-sm text-base-content/60">Salary: {o.salary}</p>}
              <div className="mt-2 flex flex-wrap gap-2">
                {o.status === 'draft' && canOffer && (
                  <button type="button" className="btn btn-primary btn-xs" onClick={() => sendMut.mutate(o.id)}>
                    Send
                  </button>
                )}
                {o.acceptToken && (o.status === 'sent' || o.status === 'draft') && (
                  <button type="button" className="btn btn-ghost btn-xs" onClick={() => copyLink(o)}>
                    <Copy className="h-3 w-3" /> {copied === o.id ? 'Copied!' : 'Copy accept link'}
                  </button>
                )}
                {o.status === 'accepted' && o.signatureName && (
                  <span className="text-xs text-success">Signed by {o.signatureName}</span>
                )}
              </div>
            </li>
          ))}
        </ul>
      ) : (
        <p className="text-sm text-base-content/50">No offers yet.</p>
      )}
    </Section>
  );
}
