import { Briefcase, Calendar, DollarSign, MapPin } from 'lucide-react';
import { useEffect, useState } from 'react';
import { useParams } from 'react-router-dom';
import { LoadingFallback } from '../components/ui/LoadingFallback';
import { api } from '../lib/api';

interface Job {
  id: string;
  title: string;
  description: string;
  industry: string;
  tags?: string[];
  requiredSkills?: string[];
  experienceLevel?: string;
  location?: string;
  salaryRange?: string;
  status: string;
  createdAt: string;
}

export function JobDetail() {
  const { id } = useParams();
  const [job, setJob] = useState<Job | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (id) {
      api<Job>(`/jobs/${id}`)
        .then(setJob)
        .catch((err) => setError(err instanceof Error ? err.message : 'Failed to load job'));
    }
  }, [id]);

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
      </div>
    </div>
  );
}
