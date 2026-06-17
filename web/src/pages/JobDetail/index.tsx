import { useMutation, useQuery } from '@tanstack/react-query';
import { Briefcase, Calendar, CheckCircle, DollarSign, MapPin, Send } from 'lucide-react';
import { Link, useParams } from 'react-router-dom';
import { LoadingFallback } from '@/components/ui/LoadingFallback';
import { useAuth } from '@/hooks/useAuth';
import { ApiError, apiWithSchema } from '@/lib/api';
import { JobSchema } from '@/lib/schemas/job';
import { applyToJob } from '@/lib/application';
import { SkillsGapPanel } from '@/pages/JobDetail/SkillsGapPanel';

export function JobDetail() {
  const { id } = useParams();
  const { user } = useAuth();

  const {
    data: job,
    error,
    isLoading,
  } = useQuery({
    queryKey: ['job', id],
    enabled: !!id,
    queryFn: () => apiWithSchema(JobSchema, `/jobs/${encodeURIComponent(id as string)}`),
  });

  const applyMutation = useMutation({
    mutationFn: () => applyToJob(id as string),
    onError: (err) => {
      if (err instanceof ApiError && err.status === 409) {
        // Handled as 'duplicate' via error path
      }
    },
  });

  if (error) {
    const message = error instanceof ApiError ? (error.serverMessage ?? error.message) : 'Failed to load job';
    return <p className="text-center p-8 text-error">{message}</p>;
  }
  if (isLoading || !job) return <LoadingFallback text="Loading job details" />;

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

        {user?.role === 'jobseeker' && id && <SkillsGapPanel jobId={id} />}

        {job.status === 'open' && (!user || user.role === 'jobseeker') && (
          <div className="mt-6 pt-4 border-t border-base-300">
            {!user && (
              <Link to="/auth/login" className="btn btn-primary w-full">
                Login to Apply
              </Link>
            )}

            {user?.role === 'jobseeker' && applyMutation.isIdle && (
              <button type="button" className="btn btn-primary w-full" onClick={() => applyMutation.mutate()}>
                <Send size={16} /> Apply Now
              </button>
            )}

            {user?.role === 'jobseeker' && applyMutation.isPending && (
              <button type="button" className="btn btn-primary w-full" disabled>
                <span className="loading loading-spinner loading-sm" /> Applying...
              </button>
            )}

            {user?.role === 'jobseeker' && applyMutation.isSuccess && (
              <div className="alert alert-success">
                <CheckCircle size={16} />
                <span>Application submitted! Track it in your applications dashboard.</span>
              </div>
            )}

            {user?.role === 'jobseeker' &&
              applyMutation.isError &&
              applyMutation.error instanceof ApiError &&
              applyMutation.error.status === 409 && (
                <div className="alert alert-info">
                  <CheckCircle size={16} />
                  <span>You've already applied to this job.</span>
                </div>
              )}

            {user?.role === 'jobseeker' &&
              applyMutation.isError &&
              !(applyMutation.error instanceof ApiError && applyMutation.error.status === 409) && (
                <div className="alert alert-error">
                  <span>
                    {applyMutation.error instanceof ApiError
                      ? (applyMutation.error.serverMessage ?? applyMutation.error.message)
                      : 'Something went wrong. Please try again.'}
                  </span>
                  <button type="button" className="btn btn-sm btn-ghost" onClick={() => applyMutation.mutate()}>
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
