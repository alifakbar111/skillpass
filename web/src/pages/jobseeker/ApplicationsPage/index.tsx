import { useQuery } from '@tanstack/react-query';
import { Briefcase, Search } from 'lucide-react';
import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { ApplicationKanban } from '@/components/jobseeker/ApplicationKanban';
import { LoadingFallback, LoadingSpinner } from '@/components/ui/LoadingFallback';
import { useAuth } from '@/hooks/useAuth';
import { getJobseekerAnalytics } from '@/lib/analytics';
import { api } from '@/lib/api';
import { getMyApplications } from '@/lib/application';

interface JobResult {
  id: string;
  title: string;
  companyName: string;
  location: string | null;
  salaryRange: string | null;
  industry: string;
}

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

  const [jobSearchOpen, setJobSearchOpen] = useState(false);
  const [jobQuery, setJobQuery] = useState('');
  const [jobResults, setJobResults] = useState<JobResult[]>([]);
  const [searching, setSearching] = useState(false);
  const navigate = useNavigate();

  const searchJobs = async (q: string) => {
    if (!q.trim()) return;
    setSearching(true);
    try {
      const results = await api<JobResult[]>(`/jobs?q=${encodeURIComponent(q)}`);
      setJobResults(results.slice(0, 5));
    } catch {
      setJobResults([]);
    } finally {
      setSearching(false);
    }
  };

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

      <details
        className="collapse collapse-arrow bg-base-200 rounded-box"
        open={jobSearchOpen}
        onToggle={(e) => setJobSearchOpen((e.target as HTMLDetailsElement).open)}
      >
        <summary className="collapse-title font-semibold flex items-center gap-2">
          <Briefcase size={18} /> Browse Open Jobs
        </summary>
        <div className="collapse-content space-y-3">
          <div className="flex gap-2">
            <input
              type="text"
              className="input input-bordered input-sm flex-1"
              placeholder="Search jobs by title, skill, or company..."
              value={jobQuery}
              onChange={(e) => setJobQuery(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === 'Enter') searchJobs(jobQuery);
              }}
            />
            <button
              type="button"
              className="btn btn-primary btn-sm"
              onClick={() => searchJobs(jobQuery)}
              disabled={searching}
            >
              {searching ? <LoadingSpinner size="sm" /> : <Search size={16} />}
            </button>
          </div>
          {jobResults.length > 0 && (
            <div className="space-y-2">
              {jobResults.map((job) => (
                <div key={job.id} className="p-3 bg-base-100 rounded-box flex justify-between items-center">
                  <div>
                    <p className="font-medium text-sm">{job.title}</p>
                    <p className="text-xs text-muted">
                      {job.companyName} &middot; {job.industry}
                      {job.location ? ` &middot; ${job.location}` : ''}
                    </p>
                    {job.salaryRange && <p className="text-xs text-muted">{job.salaryRange}</p>}
                  </div>
                  <button type="button" className="btn btn-primary btn-xs" onClick={() => navigate(`/jobs/${job.id}`)}>
                    Apply
                  </button>
                </div>
              ))}
              <div className="text-center">
                <a href="/jobs" className="link link-primary text-sm">
                  View all jobs →
                </a>
              </div>
            </div>
          )}
          {jobQuery && !searching && jobResults.length === 0 && (
            <p className="text-sm text-muted text-center py-2">No jobs found for "{jobQuery}"</p>
          )}
        </div>
      </details>

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
