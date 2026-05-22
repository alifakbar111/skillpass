import { useEffect, useState } from 'react';
import { useParams } from 'react-router-dom';
import { api } from '../lib/api';
import { Calendar, MapPin, DollarSign, Briefcase } from 'lucide-react';

interface Job {
  id: string; title: string; description: string; industry: string;
  tags?: string[]; requiredSkills?: string[]; experienceLevel?: string;
  location?: string; salaryRange?: string; status: string; createdAt: string;
}

export function JobDetail() {
  const { id } = useParams();
  const [job, setJob] = useState<Job | null>(null);

  useEffect(() => {
    if (id) api<Job>(`/jobs/${id}`).then(setJob);
  }, [id]);

  if (!job) return <div className="flex justify-center p-8"><span className="loading loading-spinner loading-lg" /></div>;

  return (
    <div className="max-w-2xl mx-auto p-4">
      <div className="card bg-base-200 p-6">
        <h1 className="text-2xl font-bold mb-2">{job.title}</h1>
        <div className="flex flex-wrap gap-3 text-sm opacity-70 mb-4">
          <span className="flex items-center gap-1"><Briefcase size={14} /> {job.industry}</span>
          {job.location && <span className="flex items-center gap-1"><MapPin size={14} /> {job.location}</span>}
          {job.salaryRange && <span className="flex items-center gap-1"><DollarSign size={14} /> {job.salaryRange}</span>}
          <span className="flex items-center gap-1"><Calendar size={14} /> {job.createdAt?.slice(0, 10)}</span>
        </div>

        {job.experienceLevel && <span className="badge mb-4">{job.experienceLevel}</span>}

        <p className="mb-4 whitespace-pre-wrap">{job.description}</p>

        {job.requiredSkills && job.requiredSkills.length > 0 && (
          <div className="mb-4">
            <h3 className="font-semibold mb-2">Required Skills</h3>
            <div className="flex flex-wrap gap-1">
              {job.requiredSkills.map(s => <span key={s} className="badge badge-primary">{s}</span>)}
            </div>
          </div>
        )}

        {job.tags && job.tags.length > 0 && (
          <div>
            <h3 className="font-semibold mb-2">Tags</h3>
            <div className="flex flex-wrap gap-1">
              {job.tags.map(t => <span key={t} className="badge badge-ghost">{t}</span>)}
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
