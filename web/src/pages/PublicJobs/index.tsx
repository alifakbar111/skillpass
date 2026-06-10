import { Briefcase } from 'lucide-react';
import { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { ApiError, api } from '../../lib/api';
import type { Job } from './type';

export function PublicJobs() {
  const [jobs, setJobs] = useState<Job[]>([]);
  const [industry, setIndustry] = useState('');
  const [industries, setIndustries] = useState<Array<{ id: string; name: string }>>([]);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const controller = new AbortController();
    setError(null);
    api<Array<{ id: string; name: string }>>('/industries')
      .then(setIndustries)
      .catch(() => {});
    const params = new URLSearchParams();
    if (industry) params.set('industry', industry);
    const qs = params.toString();
    api<Job[]>(`/jobs${qs ? `?${qs}` : ''}`)
      .then(setJobs)
      .catch((err) => {
        if (err.name === 'AbortError') return;
        setError(err instanceof ApiError ? (err.serverMessage ?? err.message) : 'Failed to load jobs');
      });
    return () => controller.abort();
  }, [industry]);

  return (
    <div className="max-w-3xl mx-auto p-4 space-y-4">
      <h1 className="text-2xl font-bold">Job Openings</h1>
      <select
        className="select select-bordered w-full max-w-xs"
        value={industry}
        onChange={(e) => setIndustry(e.target.value)}
        aria-label="Filter by industry"
      >
        <option value="">All Industries</option>
        {industries.map((ind) => (
          <option key={ind.id} value={ind.name}>
            {ind.name}
          </option>
        ))}
      </select>

      {error && <div className="alert alert-error">{error}</div>}

      <div className="space-y-2">
        {jobs.map((job) => (
          <Link
            key={job.id}
            to={`/jobs/${job.id}`}
            className="card bg-base-200 p-4 block hover:bg-base-300 transition-colors"
          >
            <div className="flex items-start gap-3">
              <Briefcase className="mt-1 opacity-50" size={20} aria-hidden="true" />
              <div>
                <h3 className="font-semibold">{job.title}</h3>
                <p className="text-sm opacity-70">
                  {job.industry} {job.location ? `· ${job.location}` : ''}
                </p>
                <div className="flex gap-2 mt-1">
                  {job.experienceLevel && <span className="badge badge-sm">{job.experienceLevel}</span>}
                  {job.salaryRange && <span className="badge badge-sm badge-outline">{job.salaryRange}</span>}
                </div>
              </div>
            </div>
          </Link>
        ))}
        {jobs.length === 0 && <p className="text-center opacity-50 py-8">No jobs found</p>}
      </div>
    </div>
  );
}
