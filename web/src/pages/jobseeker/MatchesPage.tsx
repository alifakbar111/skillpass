import { Briefcase, DollarSign, MapPin, Sparkles } from 'lucide-react';
import { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { LoadingFallback } from '../../components/ui/LoadingFallback';
import { ApiError } from '../../lib/api';
import { getJobMatches, type JobMatch } from '../../lib/matching';

function scoreBadgeClass(score: number): string {
  if (score >= 80) return 'badge-success';
  if (score >= 60) return 'badge-info';
  if (score >= 40) return 'badge-warning';
  return 'badge-ghost';
}

export function MatchesPage() {
  const [matches, setMatches] = useState<JobMatch[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    getJobMatches()
      .then(setMatches)
      .catch((err) => {
        setError(err instanceof ApiError ? (err.serverMessage ?? err.message) : 'Failed to load matches');
      })
      .finally(() => setLoading(false));
  }, []);

  if (loading) return <LoadingFallback text="Finding your matches" />;

  return (
    <div className="max-w-3xl mx-auto p-4">
      <div className="flex items-center gap-2 mb-2">
        <Sparkles size={22} className="text-primary" aria-hidden="true" />
        <h1 className="text-2xl font-bold">Recommended Jobs</h1>
      </div>
      <p className="opacity-70 mb-6 text-sm">
        Matched to your skills from your latest AI evaluation. Run an evaluation to improve your matches.
      </p>

      {error && (
        <div className="alert alert-error mb-4">
          <span>{error}</span>
        </div>
      )}

      {!error && matches.length === 0 && (
        <div className="text-center py-12 opacity-70">
          <p className="text-lg">No matches yet.</p>
          <p className="text-sm mt-1">
            Add experiences and run an{' '}
            <Link to="/jobseeker/evaluation" className="link link-primary">
              AI evaluation
            </Link>{' '}
            so we can match you to jobs.
          </p>
        </div>
      )}

      <div className="space-y-3">
        {matches.map((m) => (
          <Link
            key={m.jobPostingId}
            to={`/jobs/${m.jobPostingId}`}
            className="card bg-base-200 p-4 hover:bg-base-300 transition-colors block"
          >
            <div className="flex justify-between items-start gap-3">
              <div>
                <h3 className="font-semibold text-lg">{m.title}</h3>
                <p className="text-sm opacity-70">{m.companyName}</p>
                <div className="flex flex-wrap gap-3 text-xs opacity-70 mt-2">
                  <span className="flex items-center gap-1">
                    <Briefcase size={12} aria-hidden="true" /> {m.industry}
                  </span>
                  {m.location && (
                    <span className="flex items-center gap-1">
                      <MapPin size={12} aria-hidden="true" /> {m.location}
                    </span>
                  )}
                  {m.salaryRange && (
                    <span className="flex items-center gap-1">
                      <DollarSign size={12} aria-hidden="true" /> {m.salaryRange}
                    </span>
                  )}
                </div>
              </div>
              <div className="text-right shrink-0">
                <span className={`badge ${scoreBadgeClass(m.matchScore)}`}>{Math.round(m.matchScore)}% match</span>
              </div>
            </div>
            <p className="text-xs opacity-60 mt-2">{m.matchReason}</p>
          </Link>
        ))}
      </div>
    </div>
  );
}
