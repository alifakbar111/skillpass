import { Briefcase, Calendar, CheckCircle, DollarSign, MapPin, Send } from 'lucide-react';
import { useEffect, useState } from 'react';
import { Link, useParams } from 'react-router-dom';
import { LoadingFallback } from '../../components/ui/LoadingFallback';
import { useAuth } from '../../hooks/useAuth';
import { ApiError, api } from '../../lib/api';
import { applyToJob } from '../../lib/application';
import type { Job } from './type';

export function JobDetail() {
  const { id } = useParams();
  const { user } = useAuth();
  const [job, setJob] = useState<Job | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [applyState, setApplyState] = useState<'idle' | 'loading' | 'applied' | 'duplicate' | 'error'>('idle');
  const [applyError, setApplyError] = useState<string | null>(null);

  useEffect(() => {
    if (!id) return;
    const safe = encodeURIComponent(id);
    let cancelled = false;
    api<Job>(`/jobs/${safe}`)
      .then((j) => {
        if (!cancelled) setJob(j);
      })
      .catch((err) => {
        if (!cancelled) {
          setError(err instanceof ApiError ? (err.serverMessage ?? err.message) : 'Failed to load job');
        }
      });
    return () => {
      cancelled = true;
    };
  }, [id]);

  async function handleApply() {
    if (!id || applyState === 'loading') return;
    setApplyState('loading');
    setApplyError(null);
    try {
      await applyToJob(id);
      setApplyState('applied');
    } catch (err) {
      if (err instanceof ApiError && err.status === 409) {
        setApplyState('duplicate');
      } else {
        setApplyState('error');
        setApplyError(err instanceof ApiError ? (err.serverMessage ?? err.message) : 'Failed to apply');
      }
    }
  }

  if (error) return <p className="text-center p-8 text-error">{error}</p>;
  if (!job) return <LoadingFallback text="Loading job details" />;

  return (
    <div className="max-w-2xl mx-auto p-4">
      <div className="card bg-base-200 p-6">
        <h1 className="text-2xl font-bold mb-2">{job.title}</h1>
        <div className="flex flex-wrap gap-3 text-sm opacity-70 mb-4">
          <span className="flex items-center gap-1">
            <Briefcase size={14} aria-hidden="true" /> {job.industry}
          </span>
          {job.location && (
            <span className="flex items-center gap-1">
              <MapPin size={14} aria-hidden="true" /> {job.location}
            </span>
          )}
          {job.salaryRange && (
            <span className="flex items-center gap-1">
              <DollarSign size={14} aria-hidden="true" /> {job.salaryRange}
            </span>
          )}
          <span className="flex items-center gap-1">
            <Calendar size={14} aria-hidden="true" /> {job.createdAt?.slice(0, 10)}
          </span>
        </div>

        {job.experienceLevel && <span className="badge mb-4">{job.experienceLevel}</span>}

        <p className="mb-4 whitespace-pre-wrap">{job.description}</p>

        {job.requiredSkills && job.requiredSkills.length > 0 && (
          <div className="mb-4">
            <h3 className="font-semibold mb-2">Required Skills</h3>
            <div className="flex flex-wrap gap-1">
              {job.requiredSkills.map((s) => (
                <span key={s} className="badge badge-primary">
                  {s}
                </span>
              ))}
            </div>
          </div>
        )}

        {job.tags && job.tags.length > 0 && (
          <div>
            <h3 className="font-semibold mb-2">Tags</h3>
            <div className="flex flex-wrap gap-1">
              {job.tags.map((t) => (
                <span key={t} className="badge badge-ghost">
                  {t}
                </span>
              ))}
            </div>
          </div>
        )}

        {job.status === 'open' && (
          <div className="mt-6 pt-4 border-t border-base-300">
            {!user && (
              <Link to="/auth/login" className="btn btn-primary w-full">
                Login to Apply
              </Link>
            )}

            {user?.role === 'jobseeker' && applyState === 'idle' && (
              <button type="button" className="btn btn-primary w-full" onClick={handleApply}>
                <Send size={16} /> Apply Now
              </button>
            )}

            {user?.role === 'jobseeker' && applyState === 'loading' && (
              <button type="button" className="btn btn-primary w-full" disabled>
                <span className="loading loading-spinner loading-sm" /> Applying...
              </button>
            )}

            {user?.role === 'jobseeker' && applyState === 'applied' && (
              <div className="alert alert-success">
                <CheckCircle size={16} />
                <span>Application submitted! Track it in your applications dashboard.</span>
              </div>
            )}

            {user?.role === 'jobseeker' && applyState === 'duplicate' && (
              <div className="alert alert-info">
                <CheckCircle size={16} />
                <span>You've already applied to this job.</span>
              </div>
            )}

            {user?.role === 'jobseeker' && applyState === 'error' && (
              <div className="alert alert-error">
                <span>{applyError ?? 'Something went wrong. Please try again.'}</span>
                <button type="button" className="btn btn-sm btn-ghost" onClick={handleApply}>
                  Retry
                </button>
              </div>
            )}
          </div>
        )}
      </div>
    </div>
  );
}
