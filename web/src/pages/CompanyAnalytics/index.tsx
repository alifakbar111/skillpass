import { BarChart3, Briefcase, Clock, Users } from 'lucide-react';
import { useEffect, useState } from 'react';
import { LoadingFallback } from '../../components/ui/LoadingFallback';
import { type CompanyAnalytics as Analytics, getCompanyAnalytics } from '../../lib/analytics';
import { ApiError } from '../../lib/api';

const STATUS_ORDER = ['applied', 'reviewed', 'interviewed', 'offered', 'rejected'];

const STATUS_COLOR: Record<string, string> = {
  applied: 'bg-info',
  reviewed: 'bg-warning',
  interviewed: 'bg-accent',
  offered: 'bg-success',
  rejected: 'bg-error',
};

function sortByStatus(counts: { status: string; count: number }[]) {
  return [...counts].sort((a, b) => STATUS_ORDER.indexOf(a.status) - STATUS_ORDER.indexOf(b.status));
}

export function CompanyAnalytics() {
  const [data, setData] = useState<Analytics | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    getCompanyAnalytics()
      .then(setData)
      .catch((err) => {
        setError(err instanceof ApiError ? (err.serverMessage ?? err.message) : 'Failed to load analytics');
      });
  }, []);

  if (error) return <p className="text-center p-8 text-error">{error}</p>;
  if (!data) return <LoadingFallback text="Loading analytics" />;

  return (
    <div className="max-w-4xl mx-auto p-4 space-y-6">
      <div className="flex items-center gap-2">
        <BarChart3 size={22} className="text-primary" aria-hidden="true" />
        <h1 className="text-2xl font-bold">Hiring Analytics</h1>
      </div>

      <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
        <div className="stat bg-base-200 rounded-box">
          <div className="stat-figure text-primary">
            <Briefcase size={24} aria-hidden="true" />
          </div>
          <div className="stat-title">Open Jobs</div>
          <div className="stat-value text-2xl">{data.openJobs}</div>
          <div className="stat-desc">{data.totalJobs} total</div>
        </div>
        <div className="stat bg-base-200 rounded-box">
          <div className="stat-figure text-primary">
            <Users size={24} aria-hidden="true" />
          </div>
          <div className="stat-title">Applications</div>
          <div className="stat-value text-2xl">{data.totalApplications}</div>
        </div>
        <div className="stat bg-base-200 rounded-box">
          <div className="stat-figure text-primary">
            <Clock size={24} aria-hidden="true" />
          </div>
          <div className="stat-title">Avg. Time to Decision</div>
          <div className="stat-value text-2xl">
            {data.avgDaysToDecision !== null ? `${data.avgDaysToDecision.toFixed(1)}d` : '—'}
          </div>
          <div className="stat-desc">applied → offer/reject</div>
        </div>
        <div className="stat bg-base-200 rounded-box">
          <div className="stat-title">Offers Made</div>
          <div className="stat-value text-2xl">
            {data.applicationsByStatus.find((s) => s.status === 'offered')?.count ?? 0}
          </div>
        </div>
      </div>

      {data.totalApplications > 0 && (
        <div className="card bg-base-200 p-4">
          <h2 className="font-semibold mb-3">Pipeline Overview</h2>
          <div className="flex w-full h-6 rounded-full overflow-hidden">
            {sortByStatus(data.applicationsByStatus).map((s) => (
              <div
                key={s.status}
                className={STATUS_COLOR[s.status] ?? 'bg-neutral'}
                style={{ width: `${(s.count / data.totalApplications) * 100}%` }}
                title={`${s.status}: ${s.count}`}
              />
            ))}
          </div>
          <div className="flex flex-wrap gap-3 mt-2 text-xs">
            {sortByStatus(data.applicationsByStatus).map((s) => (
              <span key={s.status} className="flex items-center gap-1">
                <span className={`w-2 h-2 rounded-full inline-block ${STATUS_COLOR[s.status] ?? 'bg-neutral'}`} />
                {s.status} ({s.count})
              </span>
            ))}
          </div>
        </div>
      )}

      <div className="card bg-base-200 p-4">
        <h2 className="font-semibold mb-3">Per-Job Funnels</h2>
        {data.jobs.length === 0 && <p className="text-sm opacity-60">No job postings yet.</p>}
        <div className="space-y-3">
          {data.jobs.map((job) => (
            <div key={job.jobPostingId} className="bg-base-100 rounded-box p-3">
              <div className="flex justify-between items-center">
                <div className="font-medium text-sm">{job.title}</div>
                <div className="flex items-center gap-2">
                  <span className={`badge badge-xs ${job.status === 'open' ? 'badge-success' : 'badge-ghost'}`}>
                    {job.status}
                  </span>
                  <span className="text-xs opacity-60">{job.total} applications</span>
                </div>
              </div>
              {job.total > 0 && (
                <div className="flex flex-wrap gap-2 mt-2 text-xs">
                  {sortByStatus(job.byStatus).map((s) => (
                    <span key={s.status} className="badge badge-sm badge-ghost">
                      {s.status}: {s.count}
                    </span>
                  ))}
                </div>
              )}
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}
