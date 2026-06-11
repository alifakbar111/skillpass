import { useQuery } from '@tanstack/react-query';
import { ApplicationKanban } from '../../../components/jobseeker/ApplicationKanban';
import { LoadingFallback } from '../../../components/ui/LoadingFallback';
import { useAuth } from '../../../hooks/useAuth';
import { getJobseekerAnalytics, type JobseekerAnalytics } from '../../../lib/analytics';
import type { Application } from '../../../lib/application';
import { getMyApplications } from '../../../lib/application';

export function ApplicationsPage() {
  const { user } = useAuth();

  const {
    data: applications = [],
    error,
    isLoading,
  } = useQuery({
    queryKey: ['applications', 'me'],
    enabled: !!user,
    queryFn: getMyApplications,
  });

  const { data: stats } = useQuery({
    queryKey: ['jobseeker', 'analytics'],
    enabled: !!user,
    queryFn: () => getJobseekerAnalytics(),
  });

  if (!user) {
    return (
      <div className="max-w-4xl mx-auto p-4">
        <p>Please log in to view your applications.</p>
      </div>
    );
  }

  if (isLoading) return <LoadingFallback text="Loading applications" />;

  if (error) {
    return (
      <div className="max-w-4xl mx-auto p-4">
        <div className="alert alert-error">
          {error instanceof Error ? error.message : 'Failed to load applications'}
        </div>
      </div>
    );
  }

  return (
    <div className="max-w-5xl mx-auto p-4 space-y-4">
      <h1 className="text-2xl font-bold">My Applications</h1>

      {stats && (
        <div className="grid grid-cols-3 gap-4">
          <div className="stat bg-base-200 rounded-box p-4">
            <div className="stat-title text-xs">Total Applications</div>
            <div className="stat-value text-xl">{stats.totalApplications}</div>
          </div>
          <div className="stat bg-base-200 rounded-box p-4">
            <div className="stat-title text-xs">Response Rate</div>
            <div className="stat-value text-xl">
              {stats.responseRate !== null ? `${Math.round(stats.responseRate)}%` : '—'}
            </div>
          </div>
          <div className="stat bg-base-200 rounded-box p-4">
            <div className="stat-title text-xs">Passport Views</div>
            <div className="stat-value text-xl">{stats.passportViews}</div>
          </div>
        </div>
      )}

      <ApplicationKanban applications={applications} />
    </div>
  );
}
