import { zodResolver } from '@hookform/resolvers/zod';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Pencil, Plus, Users, X } from 'lucide-react';
import { useState } from 'react';
import { useForm } from 'react-hook-form';
import type { Job } from '@/lib/api-types';
import { JobMatches } from '../../components/company/JobMatches';
import { Form } from '../../components/ui/Form';
import { FormInput } from '../../components/ui/FormInput';
import { FormSelect } from '../../components/ui/FormSelect';
import { FormTextarea } from '../../components/ui/FormTextarea';
import { LoadingSpinner } from '../../components/ui/LoadingFallback';
import { useIndustries } from '../../hooks/useIndustries';
import { ApiError, api } from '../../lib/api';
import { EXPERIENCE_LEVEL_OPTIONS } from '../../lib/constants';
import { type JobForm, jobSchema } from '../../lib/schemas';

function parseFormData(data: JobForm) {
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
  return { ...data, tags, requiredSkills };
}

export function CompanyJobs() {
  const queryClient = useQueryClient();
  const [editingJobId, setEditingJobId] = useState<string | null>(null);
  const [showForm, setShowForm] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [matchesJobId, setMatchesJobId] = useState<string | null>(null);

  const { data: industries = [] } = useIndustries();
  const { data: jobs = [] } = useQuery({
    queryKey: ['jobs', 'me'],
    queryFn: () => api<Job[]>('/jobs/me'),
  });

  const methods = useForm<JobForm>({
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
    mutationFn: (data: JobForm) =>
      api('/jobs', {
        method: 'POST',
        body: JSON.stringify(parseFormData(data)),
      }),
    onMutate: () => setError(null),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['jobs'] });
      setShowForm(false);
      setEditingJobId(null);
      methods.reset();
    },
    onError: (err) => {
      setError(err instanceof ApiError ? (err.serverMessage ?? err.message) : 'Failed to create job');
    },
  });

  const updateMutation = useMutation({
    mutationFn: (data: JobForm & { id: string }) =>
      api(`/jobs/${encodeURIComponent(data.id)}`, {
        method: 'PUT',
        body: JSON.stringify(parseFormData(data)),
      }),
    onMutate: () => setError(null),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['jobs'] });
      setShowForm(false);
      setEditingJobId(null);
      methods.reset();
    },
    onError: (err) => {
      setError(err instanceof ApiError ? (err.serverMessage ?? err.message) : 'Failed to update job');
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

  const isSaving = createMutation.isPending || updateMutation.isPending;

  function openCreateForm() {
    methods.reset({
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
    setShowForm(true);
  }

  function openEditForm(job: Job) {
    methods.reset({
      title: job.title,
      description: job.description,
      industry: job.industry,
      tags: job.tags?.join(', ') ?? '',
      requiredSkills: job.requiredSkills?.join(', ') ?? '',
      experienceLevel: (job.experienceLevel as 'entry' | 'mid' | 'senior' | 'lead') ?? 'mid',
      location: job.location ?? '',
      salaryRange: job.salaryRange ?? '',
    });
    setEditingJobId(job.id ?? null);
    setShowForm(true);
  }

  function closeForm() {
    setShowForm(false);
    setEditingJobId(null);
    methods.reset();
  }

  function onSubmit(data: JobForm) {
    if (editingJobId) {
      updateMutation.mutate({ ...data, id: editingJobId });
    } else {
      createMutation.mutate(data);
    }
  }

  const closeJob = (id: string) => closeMutation.mutate(id);

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

      {showForm && (
        <Form methods={methods} onSubmit={onSubmit} className="card bg-base-200 p-4 space-y-3">
          <h2 className="font-semibold text-lg">{editingJobId ? 'Edit Job' : 'New Job'}</h2>
          <FormInput label="Job Title" name="title" placeholder="Job Title" />
          <FormTextarea label="Job Description" name="description" placeholder="Job Description" rows={4} />
          <FormSelect
            label="Industry"
            name="industry"
            options={industries
              .filter((ind): ind is typeof ind & { name: string } => ind.name != null)
              .map((ind) => ({ value: ind.name, label: ind.name }))}
          />
          <FormSelect label="Experience Level" name="experienceLevel" options={EXPERIENCE_LEVEL_OPTIONS} />
          <FormInput label="Tags (comma-separated)" name="tags" placeholder="e.g. remote, full-time" />
          <FormInput
            label="Required Skills (comma-separated)"
            name="requiredSkills"
            placeholder="e.g. React, TypeScript"
          />
          <div className="flex gap-2">
            <FormInput label="Location" name="location" placeholder="Location" />
            <FormInput label="Salary Range" name="salaryRange" placeholder="e.g. $80k-$120k" />
          </div>
          <div className="flex gap-2">
            <button type="submit" className="btn btn-primary" disabled={isSaving}>
              {isSaving ? <LoadingSpinner /> : editingJobId ? 'Save Changes' : 'Post Job'}
            </button>
            <button type="button" className="btn" onClick={closeForm}>
              Cancel
            </button>
          </div>
        </Form>
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
                  onClick={() => setMatchesJobId((prev) => (prev === job.id ? null : (job.id ?? null)))}
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
                    onClick={() => job.id && closeJob(job.id)}
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
        {jobs.length === 0 && !showForm && (
          <p className="text-center opacity-50 py-8">No job postings yet. Create your first one!</p>
        )}
      </div>
    </div>
  );
}
