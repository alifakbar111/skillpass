import { zodResolver } from '@hookform/resolvers/zod';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Pencil, Plus, X } from 'lucide-react';
import { useState } from 'react';
import { useForm } from 'react-hook-form';
import { Link } from 'react-router-dom';
import { FormInput, FormSelect, FormTextarea } from '../../components/ui/FormField';
import { LoadingSpinner } from '../../components/ui/LoadingFallback';
import { useIndustries } from '../../hooks/useIndustries';
import { ApiError, api } from '../../lib/api';
import { EXPERIENCE_LEVEL_OPTIONS } from '../../lib/constants';
import { type JobForm, jobSchema } from '../../lib/schemas';
import type { Job } from './type';

export function CompanyJobs() {
  const queryClient = useQueryClient();
  const [showForm, setShowForm] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const { data: industries = [] } = useIndustries();
  const { data: jobs = [] } = useQuery({
    queryKey: ['jobs', 'me'],
    queryFn: () => api<Job[]>('/jobs/me'),
  });

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

  const createMutation = useMutation({
    mutationFn: (data: JobForm) => {
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
      return api('/jobs', {
        method: 'POST',
        body: JSON.stringify({ ...data, tags, requiredSkills }),
      });
    },
    onMutate: () => setError(null),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['jobs'] });
      setShowForm(false);
      reset();
    },
    onError: (err) => {
      setError(err instanceof ApiError ? (err.serverMessage ?? err.message) : 'Failed to create job');
    },
  });

  const closeMutation = useMutation({
    mutationFn: (id: string) =>
      api(`/jobs/${encodeURIComponent(id)}`, {
        method: 'PUT',
        body: JSON.stringify({ status: 'closed' }),
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['jobs'] });
    },
    onError: (err) => {
      setError(err instanceof ApiError ? (err.serverMessage ?? err.message) : 'Failed to close job');
    },
  });

  const createJob = (data: JobForm) => createMutation.mutate(data);
  const closeJob = (id: string) => closeMutation.mutate(id);
  return (
    <div className="max-w-3xl mx-auto p-4 space-y-4">
      <div className="flex justify-between items-center">
        <h1 className="text-2xl font-bold">My Job Postings</h1>
        <button type="button" className="btn btn-primary btn-sm" onClick={() => setShowForm(!showForm)}>
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
            <button type="submit" className="btn btn-primary" disabled={createMutation.isPending}>
              {createMutation.isPending ? <LoadingSpinner /> : 'Post Job'}
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
