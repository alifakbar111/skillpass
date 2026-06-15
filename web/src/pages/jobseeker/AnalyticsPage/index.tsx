import { useQuery } from '@tanstack/react-query';
import { LoadingFallback } from '../../../components/ui/LoadingFallback';
import { useAuth } from '../../../hooks/useAuth';
import { getJobseekerAnalytics } from '../../../lib/analytics';
import type { ProfileView } from '../../../lib/profile-views';
import { getMyProfileViews } from '../../../lib/profile-views';

export function AnalyticsPage() {
  const { user } = useAuth();
  const { data: views = [], isLoading: loadingViews } = useQuery({
    queryKey: ['profile-views'],
    enabled: !!user,
    queryFn: getMyProfileViews,
  });
  const { data: stats, isLoading: loadingStats } = useQuery({
    queryKey: ['jobseeker', 'analytics'],
    enabled: !!user,
    queryFn: () => getJobseekerAnalytics(),
  });
  if (!user)
    return (
      <div className="max-w-4xl mx-auto p-4">
        <p>Please log in to view analytics.</p>
      </div>
    );
  if (loadingViews || loadingStats) return <LoadingFallback text="Loading analytics" />;
  return (
    <div className="max-w-5xl mx-auto p-4 space-y-6">
      <h1 className="text-2xl font-bold">Analytics</h1>
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
            <div className="stat-title text-xs">Profile Views</div>
            <div className="stat-value text-xl">{views.length}</div>
          </div>
        </div>
      )}
      <div className="card bg-base-200 shadow-md">
        <div className="card-body">
          <h2 className="card-title">Profile Views</h2>
          {views.length === 0 ? (
            <p className="text-sm text-base-content/60">No profile views yet.</p>
          ) : (
            <div className="overflow-x-auto">
              <table className="table table-zebra">
                <thead>
                  <tr>
                    <th>Company</th>
                    <th>Viewed At</th>
                  </tr>
                </thead>
                <tbody>
                  {views.map((view: ProfileView) => (
                    <tr key={view.id}>
                      <td>{view.companyId}</td>
                      <td>{new Date(view.viewedAt).toLocaleString()}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
