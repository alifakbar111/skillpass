import { useQuery } from '@tanstack/react-query';
import { Briefcase } from 'lucide-react';
import { useState } from 'react';
import { Link } from 'react-router-dom';
import { useIndustries } from '../../hooks/useIndustries';
import { ApiError, api } from '../../lib/api';
import type { Job } from '@/lib/api-types';

export function PublicJobs() {
  const [industry, setIndustry] = useState('');

  const { data: industries = [] } = useIndustries();

  const { data: jobs = [], error } = useQuery({
    queryKey: ['jobs', { industry }],
    queryFn: () => {
      const params = new URLSearchParams();
      if (industry) params.set('industry', industry);
      const qs = params.toString();
      return api<Job[]>(`/jobs${qs ? `?${qs}` : ''}`);
    },
  });

  const errorMessage = error
    ? error instanceof ApiError
      ? (error.serverMessage ?? error.message)
      : 'Failed to load jobs'
    : null;

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

      {errorMessage && <div className="alert alert-error">{errorMessage}</div>}

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
