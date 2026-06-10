import { zodResolver } from '@hookform/resolvers/zod';
import { Pencil, Plus, Users, X } from 'lucide-react';
import { useEffect, useState } from 'react';
import { useForm } from 'react-hook-form';
import { JobMatches } from '../../components/company/JobMatches';
import { FormInput, FormSelect, FormTextarea } from '../../components/ui/FormField';
import { LoadingSpinner } from '../../components/ui/LoadingFallback';
import { ApiError, api } from '../../lib/api';
import { EXPERIENCE_LEVEL_OPTIONS } from '../../lib/constants';
import { type JobForm, jobSchema } from '../../lib/schemas';
import type { Industry, Job } from './type';

export function CompanyJobs() {
  const [jobs, setJobs] = useState<Job[]>([]);
  const [formMode, setFormMode] = useState<'hidden' | 'create' | 'edit'>('hidden');
  const [editingJobId, setEditingJobId] = useState<string | null>(null);
  const [industries, setIndustries] = useState<Industry[]>([]);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [matchesJobId, setMatchesJobId] = useState<string | null>(null);

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
    let cancelled = false;
    api<Industry[]>('/industries')
      .then((data) => {
        if (!cancelled) setIndustries(data);
      })
      .catch(() => {});
    api<Job[]>('/jobs/me')
      .then((data) => {
        if (!cancelled) setJobs(data);
      })
      .catch(() => {});
    return () => {
      cancelled = true;
    };
  }, []);

  function openCreateForm() {
    reset({
      title: '',
      description: '',
      industry: 'Technology',
      tags: '',
      requiredSkills: '',
      experienceLevel: 'mid',
      location: '',
      salaryRange: '',
    });
    setEditingJobId(null);
    setFormMode('create');
  }

  function openEditForm(job: Job) {
    reset({
      title: job.title,
      description: job.description,
      industry: job.industry,
      tags: job.tags?.join(', ') ?? '',
      requiredSkills: job.requiredSkills?.join(', ') ?? '',
      experienceLevel: (job.experienceLevel as 'entry' | 'mid' | 'senior' | 'lead') ?? 'mid',
      location: job.location ?? '',
      salaryRange: job.salaryRange ?? '',
    });
    setEditingJobId(job.id);
    setFormMode('edit');
  }

  function closeForm() {
    setFormMode('hidden');
    setEditingJobId(null);
    reset();
  }

  function parseArrayField(value?: string): string[] {
    return value
      ? value
          .split(',')
          .map((s) => s.trim())
          .filter(Boolean)
      : [];
  }

  const saveJob = async (data: JobForm) => {
    setSaving(true);
    setError(null);
    const payload = {
      ...data,
      tags: parseArrayField(data.tags),
      requiredSkills: parseArrayField(data.requiredSkills),
    };
    try {
      if (formMode === 'edit' && editingJobId) {
        await api(`/jobs/${encodeURIComponent(editingJobId)}`, {
          method: 'PUT',
          body: JSON.stringify(payload),
        });
      } else {
        await api('/jobs', {
          method: 'POST',
          body: JSON.stringify(payload),
        });
      }
      const updated = await api<Job[]>('/jobs/me');
      setJobs(updated);
      closeForm();
    } catch (err) {
      setError(err instanceof ApiError ? (err.serverMessage ?? err.message) : 'Failed to save job');
    } finally {
      setSaving(false);
    }
  };

  const closeJob = async (id: string) => {
    try {
      await api(`/jobs/${encodeURIComponent(id)}`, {
        method: 'PUT',
        body: JSON.stringify({ status: 'closed' }),
      });
      setJobs((prev) => prev.map((j) => (j.id === id ? { ...j, status: 'closed' } : j)));
    } catch (err) {
      setError(err instanceof ApiError ? (err.serverMessage ?? err.message) : 'Failed to close job');
    }
  };
  return (
    <div className="max-w-3xl mx-auto p-4 space-y-4">
      <div className="flex justify-between items-center">
        <h1 className="text-2xl font-bold">My Job Postings</h1>
        <button type="button" className="btn btn-primary btn-sm" onClick={openCreateForm}>
          <Plus size={16} aria-hidden="true" /> New Job
        </button>
      </div>

      {error && (
        <div className="alert alert-error">
          <span>{error}</span>
          <button type="button" title="close" className="btn btn-ghost btn-xs" onClick={() => setError(null)}>
            <X size={14} />
          </button>
        </div>
      )}

      {formMode !== 'hidden' && (
        <form onSubmit={handleSubmit(saveJob)} className="card bg-base-200 p-4 space-y-3">
          <h2 className="font-semibold text-lg">{formMode === 'edit' ? 'Edit Job' : 'New Job'}</h2>
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
            options={industries.map((ind) => ({ value: ind.Name, label: ind.Name }))}
          />
          <FormSelect
            label="Experience Level"
            registration={register('experienceLevel')}
            error={errors.experienceLevel}
            options={EXPERIENCE_LEVEL_OPTIONS}
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
              {saving ? <LoadingSpinner /> : formMode === 'edit' ? 'Save Changes' : 'Post Job'}
            </button>
            <button type="button" className="btn" onClick={closeForm}>
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
                <button
                  type="button"
                  className="btn btn-ghost btn-xs"
                  aria-label={`View matches for ${job.title}`}
                  onClick={() => setMatchesJobId((prev) => (prev === job.id ? null : job.id))}
                >
                  <Users size={14} aria-hidden="true" />
                </button>
                {job.status === 'open' && (
                  <button
                    type="button"
                    className="btn btn-ghost btn-xs"
                    aria-label={`Edit ${job.title}`}
                    onClick={() => openEditForm(job)}
                  >
                    <Pencil size={14} aria-hidden="true" />
                  </button>
                )}
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
            {matchesJobId === job.id && <JobMatches jobId={job.id} />}
          </div>
        ))}
        {jobs.length === 0 && formMode === 'hidden' && (
          <p className="text-center opacity-50 py-8">No job postings yet. Create your first one!</p>
        )}
      </div>
    </div>
  );
}
