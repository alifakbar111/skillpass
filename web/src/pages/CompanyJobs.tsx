import { Pencil, Plus, X } from 'lucide-react';
import { type FormEvent, useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { api } from '../lib/api';

interface Job {
  id: string;
  title: string;
  industry: string;
  location?: string;
  experienceLevel?: string;
  status: string;
  createdAt: string;
}

export function CompanyJobs() {
  const [jobs, setJobs] = useState<Job[]>([]);
  const [showForm, setShowForm] = useState(false);
  const [form, setForm] = useState({
    title: '',
    description: '',
    industry: 'Technology',
    tags: '',
    requiredSkills: '',
    experienceLevel: 'mid',
    location: '',
    salaryRange: '',
  });
  const [industries, setIndustries] = useState<Array<{ id: string; name: string }>>([]);
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    api<Array<{ id: string; name: string }>>('/industries').then(setIndustries);
    api<Job[]>('/jobs/me').then(setJobs);
  }, []);

  const createJob = async (e: FormEvent) => {
    e.preventDefault();
    setSaving(true);
    const tags = form.tags
      .split(',')
      .map((t) => t.trim())
      .filter(Boolean);
    const requiredSkills = form.requiredSkills
      .split(',')
      .map((s) => s.trim())
      .filter(Boolean);
    await api('/jobs', {
      method: 'POST',
      body: JSON.stringify({ ...form, tags, requiredSkills }),
    });
    const updated = await api<Job[]>('/jobs/me');
    setJobs(updated);
    setShowForm(false);
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
        <form onSubmit={createJob} className="card bg-base-200 p-4 space-y-3">
          <input
            className="input input-bordered"
            placeholder="Job Title"
            value={form.title}
            onChange={(e) => setForm({ ...form, title: e.target.value })}
            required
          />
          <textarea
            className="textarea textarea-bordered h-24"
            placeholder="Job Description"
            value={form.description}
            onChange={(e) => setForm({ ...form, description: e.target.value })}
            required
          />
          <select
            className="select select-bordered"
            value={form.industry}
            onChange={(e) => setForm({ ...form, industry: e.target.value })}
            aria-label="Industry"
          >
            {industries.map((ind) => (
              <option key={ind.id} value={ind.name}>
                {ind.name}
              </option>
            ))}
          </select>
          <select
            className="select select-bordered"
            value={form.experienceLevel}
            onChange={(e) => setForm({ ...form, experienceLevel: e.target.value })}
            aria-label="Experience level"
          >
            <option value="entry">Entry</option>
            <option value="mid">Mid</option>
            <option value="senior">Senior</option>
            <option value="lead">Lead</option>
          </select>
          <input
            className="input input-bordered"
            placeholder="Tags (comma-separated)"
            value={form.tags}
            onChange={(e) => setForm({ ...form, tags: e.target.value })}
          />
          <input
            className="input input-bordered"
            placeholder="Required Skills (comma-separated)"
            value={form.requiredSkills}
            onChange={(e) => setForm({ ...form, requiredSkills: e.target.value })}
          />
          <div className="flex gap-2">
            <input
              className="input input-bordered flex-1"
              placeholder="Location"
              value={form.location}
              onChange={(e) => setForm({ ...form, location: e.target.value })}
            />
            <input
              className="input input-bordered flex-1"
              placeholder="Salary Range"
              value={form.salaryRange}
              onChange={(e) => setForm({ ...form, salaryRange: e.target.value })}
            />
          </div>
          <div className="flex gap-2">
            <button type="submit" className="btn btn-primary" disabled={saving}>
              {saving ? <span className="loading loading-spinner" aria-hidden="true" /> : 'Post Job'}
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
                  <button type="button" className="btn btn-ghost btn-xs text-error" onClick={() => closeJob(job.id)} aria-label={`Close ${job.title}`}>
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
