import { zodResolver } from '@hookform/resolvers/zod';
import { Pencil, Plus, X } from 'lucide-react';
import { useEffect, useState } from 'react';
import { useForm } from 'react-hook-form';
import { Link } from 'react-router-dom';
import { FormInput, FormSelect, FormTextarea } from '../components/ui/FormField';
import { LoadingSpinner } from '../components/ui/LoadingFallback';
import { api } from '../lib/api';
import { type JobForm, jobSchema } from '../lib/schemas';

interface Job {
  id: string;
  title: string;
  industry: string;
  location?: string;
  experienceLevel?: string;
  status: string;
  createdAt: string;
}

const EXPERIENCE_LEVELS = [
  { value: 'entry', label: 'Entry' },
  { value: 'mid', label: 'Mid' },
  { value: 'senior', label: 'Senior' },
  { value: 'lead', label: 'Lead' },
];

export function CompanyJobs() {
  const [jobs, setJobs] = useState<Job[]>([]);
  const [showForm, setShowForm] = useState(false);
  const [industries, setIndustries] = useState<Array<{ id: string; name: string }>>([]);
  const [saving, setSaving] = useState(false);

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<JobForm>({
    resolver: zodResolver(jobSchema),
    defaultValues: {
      title: '',
      description: '',
      industry: 'Technology',
      tags: '',
      requiredSkills: '',
      experienceLevel: 'mid',
      location: '',
      salaryRange: '',
    },
  });

  useEffect(() => {
    api<Array<{ id: string; name: string }>>('/industries').then(setIndustries);
    api<Job[]>('/jobs/me').then(setJobs);
  }, []);

  const createJob = async (data: JobForm) => {
    setSaving(true);
    const tags = data.tags
      ? data.tags
          .split(',')
          .map((t) => t.trim())
          .filter(Boolean)
      : [];
    const requiredSkills = data.requiredSkills
      ? data.requiredSkills
          .split(',')
          .map((s) => s.trim())
          .filter(Boolean)
      : [];
    await api('/jobs', {
      method: 'POST',
      body: JSON.stringify({ ...data, tags, requiredSkills }),
    });
    const updated = await api<Job[]>('/jobs/me');
    setJobs(updated);
    setShowForm(false);
    reset();
    setSaving(false);
  };

  const closeJob = async (id: string) => {
    await api(`/jobs/${id}`, { method: 'PUT', body: JSON.stringify({ status: 'closed' }) });
    setJobs((prev) => prev.map((j) => (j.id === id ? { ...j, status: 'closed' } : j)));
  };

  return (
    <div className="max-w-3xl mx-auto p-4 space-y-4">
      <div className="flex justify-between items-center">
        <h1 className="text-2xl font-bold">My Job Postings</h1>
        <button type="button" className="btn btn-primary btn-sm" onClick={() => setShowForm(!showForm)}>
          <Plus size={16} aria-hidden="true" /> New Job
        </button>
      </div>

      {showForm && (
        <form onSubmit={handleSubmit(createJob)} className="card bg-base-200 p-4 space-y-3">
          <FormInput label="Job Title" registration={register('title')} error={errors.title} placeholder="Job Title" />
          <FormTextarea
            label="Job Description"
            registration={register('description')}
            error={errors.description}
            placeholder="Job Description"
            rows={4}
          />
          <FormSelect
            label="Industry"
            registration={register('industry')}
            error={errors.industry}
            options={industries.map((ind) => ({ value: ind.name, label: ind.name }))}
          />
          <FormSelect
            label="Experience Level"
            registration={register('experienceLevel')}
            error={errors.experienceLevel}
            options={EXPERIENCE_LEVELS}
          />
          <FormInput
            label="Tags (comma-separated)"
            registration={register('tags')}
            error={errors.tags}
            placeholder="e.g. remote, full-time"
          />
          <FormInput
            label="Required Skills (comma-separated)"
            registration={register('requiredSkills')}
            error={errors.requiredSkills}
            placeholder="e.g. React, TypeScript"
          />
          <div className="flex gap-2">
            <FormInput
              label="Location"
              registration={register('location')}
              error={errors.location}
              placeholder="Location"
            />
            <FormInput
              label="Salary Range"
              registration={register('salaryRange')}
              error={errors.salaryRange}
              placeholder="e.g. $80k-$120k"
            />
          </div>
          <div className="flex gap-2">
            <button type="submit" className="btn btn-primary" disabled={saving}>
              {saving ? <LoadingSpinner /> : 'Post Job'}
            </button>
            <button type="button" className="btn" onClick={() => setShowForm(false)}>
              Cancel
            </button>
          </div>
        </form>
      )}

      <div className="space-y-2">
        {jobs.map((job) => (
          <div key={job.id} className="card bg-base-200 p-4">
            <div className="flex justify-between items-start">
              <div>
                <h3 className="font-semibold">{job.title}</h3>
                <p className="text-sm opacity-70">
                  {job.industry} {job.location ? `· ${job.location}` : ''}
                </p>
                <div className="flex gap-2 mt-1">
                  <span className="badge badge-sm">{job.experienceLevel}</span>
                  <span className={`badge badge-sm ${job.status === 'open' ? 'badge-success' : 'badge-ghost'}`}>
                    {job.status}
                  </span>
                </div>
              </div>
              <div className="flex gap-1">
                <Link to={`/jobs/${job.id}`} className="btn btn-ghost btn-xs" aria-label={`Edit ${job.title}`}>
                  <Pencil size={14} aria-hidden="true" />
                </Link>
                {job.status === 'open' && (
                  <button
                    type="button"
                    className="btn btn-ghost btn-xs text-error"
                    onClick={() => closeJob(job.id)}
                    aria-label={`Close ${job.title}`}
                  >
                    <X size={14} aria-hidden="true" />
                  </button>
                )}
              </div>
            </div>
          </div>
        ))}
        {jobs.length === 0 && !showForm && (
          <p className="text-center opacity-50 py-8">No job postings yet. Create your first one!</p>
        )}
      </div>
    </div>
  );
}
