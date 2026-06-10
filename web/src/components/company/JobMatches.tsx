import { useEffect, useState } from 'react';
import { ApiError } from '../../lib/api';
import { type CandidateMatch, getCandidateMatches } from '../../lib/matching';
import { LoadingSpinner } from '../ui/LoadingFallback';

function scoreBadgeClass(score: number): string {
  if (score >= 80) return 'badge-success';
  if (score >= 60) return 'badge-info';
  if (score >= 40) return 'badge-warning';
  return 'badge-ghost';
}

export function JobMatches({ jobId }: { jobId: string }) {
  const [matches, setMatches] = useState<CandidateMatch[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;
    getCandidateMatches(jobId)
      .then((data) => {
        if (!cancelled) setMatches(data);
      })
      .catch((err) => {
        if (!cancelled) {
          setError(err instanceof ApiError ? (err.serverMessage ?? err.message) : 'Failed to load matches');
        }
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });
    return () => {
      cancelled = true;
    };
  }, [jobId]);

  if (loading)
    return (
      <div className="py-3">
        <LoadingSpinner />
      </div>
    );

  if (error) return <p className="text-error text-sm py-2">{error}</p>;

  if (matches.length === 0)
    return (
      <p className="text-sm opacity-60 py-2">
        No matching candidates yet. Matches require candidates with AI evaluations whose skills overlap the required
        skills.
      </p>
    );

  return (
    <div className="mt-3 space-y-2 border-t border-base-300 pt-3">
      <h4 className="text-sm font-semibold opacity-80">Recommended Candidates</h4>
      {matches.map((c) => (
        <div key={c.profileId} className="flex justify-between items-start gap-3 bg-base-100 rounded-lg p-3">
          <div>
            <div className="font-medium">{c.name}</div>
            {c.headline && <div className="text-xs opacity-60">{c.headline}</div>}
            <div className="flex flex-wrap gap-1 mt-1">
              {c.topSkills.map((s) => (
                <span key={s} className="badge badge-xs badge-ghost">
                  {s}
                </span>
              ))}
            </div>
          </div>
          <div className="text-right shrink-0 space-y-1">
            <span className={`badge badge-sm ${scoreBadgeClass(c.matchScore)}`}>{Math.round(c.matchScore)}%</span>
            <div className="text-xs opacity-50">score {c.overallScore}</div>
          </div>
        </div>
      ))}
    </div>
  );
}
