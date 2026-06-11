import { useQuery } from '@tanstack/react-query';
import { ExternalLink, Eye, Sparkles, X } from 'lucide-react';
import { useRef, useState } from 'react';
import { Link } from 'react-router-dom';
import type { PublicProfile } from '@/lib/api-types';
import { EvaluationScoreBadge } from '../../components/jobseeker/EvaluationScoreBadge';
import { SharePassport } from '../../components/passport/SharePassport';
import { LoadingFallback } from '../../components/ui/LoadingFallback';
import { useAuth } from '../../hooks/useAuth';
import { ApiError, api } from '../../lib/api';
import { getLatestEvaluation } from '../../lib/evaluation';

export function JobseekerPassport() {
  const { user } = useAuth();
  const [errorDismissed, setErrorDismissed] = useState(false);
  const printRef = useRef<HTMLDivElement>(null);

  const {
    data,
    error: passportError,
    isLoading,
  } = useQuery({
    queryKey: ['passport', user?.username],
    enabled: !!user?.username,
    queryFn: () => api<PublicProfile>(`/profiles/${encodeURIComponent(user?.username as string)}`),
  });

  const { data: evaluation } = useQuery({
    queryKey: ['evaluation', 'latest'],
    enabled: !!user,
    queryFn: getLatestEvaluation,
    retry: (count, err) => count < 1 && !(err instanceof ApiError && err.status === 404),
  });

  const error =
    passportError && !errorDismissed
      ? passportError instanceof ApiError
        ? (passportError.serverMessage ?? passportError.message)
        : 'Failed to load passport'
      : null;

  if (error) {
    return (
      <div className="max-w-2xl mx-auto p-4">
        <div className="alert alert-error">
          <span>{error}</span>
          <button type="button" title="close" className="btn btn-ghost btn-xs" onClick={() => setErrorDismissed(true)}>
            <X size={14} />
          </button>
        </div>
      </div>
    );
  }

  if (isLoading || !data) return <LoadingFallback text="Loading passport" />;

  return (
    <div className="max-w-2xl mx-auto p-4 space-y-4">
      <div className="flex justify-between items-center flex-wrap gap-2">
        <h1 className="text-2xl font-bold">My Passport</h1>
        <div className="flex items-center gap-2 flex-wrap">
          {evaluation && <EvaluationScoreBadge overallScore={evaluation.overallScore ?? 0} />}
          {user?.username && <SharePassport slug={user.username} name={data.name ?? ''} printRef={printRef} />}
          <Link to={`/profiles/${user?.username}`} className="btn btn-outline btn-sm gap-2" target="_blank">
            <ExternalLink size={14} aria-hidden="true" /> View Public
          </Link>
        </div>
      </div>

      <div ref={printRef} className="space-y-4">
        <div className="card bg-base-200 p-6">
          <div className="flex items-center gap-4 mb-4">
            <div className="avatar placeholder">
              {data.avatarUrl ? (
                <div className="w-16 rounded-full">
                  <img src={data.avatarUrl} alt={`${data.name} avatar`} />
                </div>
              ) : (
                <div className="bg-neutral text-neutral-content rounded-full w-16">
                  <span className="text-xl">{data.name?.charAt(0)}</span>
                </div>
              )}
            </div>
            <div>
              <h2 className="text-xl font-bold">{data.name}</h2>
              {data.headline && <p className="text-muted-strong">{data.headline}</p>}
              {data.yearsOfExperience !== undefined && (
                <p className="text-sm text-muted">{data.yearsOfExperience} years of experience</p>
              )}
              {data.viewCount !== undefined && (
                <p className="text-xs text-muted flex items-center gap-1 mt-1">
                  <Eye size={12} aria-hidden="true" /> {data.viewCount} passport{' '}
                  {data.viewCount === 1 ? 'view' : 'views'}
                </p>
              )}
            </div>
          </div>
          {data.about && <p className="text-muted-strong mb-4">{data.about}</p>}
        </div>

        {evaluation && (
          <div className="card bg-base-200 p-4">
            <div className="flex justify-between items-center mb-3">
              <h3 className="font-semibold">AI Evaluation</h3>
              <Link to="/jobseeker/evaluation" className="btn btn-ghost btn-sm gap-1">
                <Sparkles size={14} aria-hidden="true" /> Details
              </Link>
            </div>
            {(evaluation.strengths ?? []).length > 0 && (
              <div className="mb-3">
                <p className="text-sm font-semibold text-success mb-1">Top Strengths</p>
                <div className="flex flex-wrap gap-1">
                  {(evaluation.strengths ?? []).slice(0, 5).map((s) => (
                    <span key={s.skill} className="badge badge-success badge-sm">
                      {s.skill} ({s.score})
                    </span>
                  ))}
                </div>
              </div>
            )}
            {(evaluation.skillScores ?? []).length > 0 && (
              <div>
                <p className="text-sm font-semibold mb-1">Skill Scores</p>
                <div className="flex flex-wrap gap-1">
                  {(evaluation.skillScores ?? []).slice(0, 8).map((s) => (
                    <span key={s.skill} className="badge badge-ghost badge-sm">
                      {s.skill}: {s.score}
                    </span>
                  ))}
                </div>
              </div>
            )}
          </div>
        )}

        <div className="card bg-base-200 p-4">
          <h3 className="font-semibold mb-3">Experience</h3>
          <div className="space-y-2">
            {(data.experiences ?? []).map((exp, i) => (
              // biome-ignore lint/suspicious/noArrayIndexKey: experiences array has no stable id in this view
              <div key={i} className="p-3 bg-base-100 rounded-box">
                <p className="font-medium">{exp.title}</p>
                <p className="text-sm opacity-70">
                  {exp.organization} · {exp.startDate}{' '}
                  {exp.isCurrent ? '- Present' : exp.endDate ? `- ${exp.endDate}` : ''}
                </p>
                {exp.description && <p className="text-sm mt-1 opacity-60">{exp.description}</p>}
                {exp.skillsUsed && exp.skillsUsed.length > 0 && (
                  <div className="flex flex-wrap gap-1 mt-1">
                    {exp.skillsUsed.map((s) => (
                      <span key={s} className="badge badge-sm">
                        {s}
                      </span>
                    ))}
                  </div>
                )}
                {exp.url && (
                  <a
                    href={exp.url}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="link link-primary text-xs inline-flex items-center gap-1 mt-2"
                  >
                    <ExternalLink size={12} aria-hidden="true" /> View evidence
                  </a>
                )}
              </div>
            ))}
          </div>
        </div>
      </div>
    </div>
  );
}
