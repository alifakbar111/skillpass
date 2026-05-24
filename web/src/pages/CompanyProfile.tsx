import { type FormEvent, useEffect, useState } from 'react';
import { useAuth } from '../hooks/useAuth';
import { api } from '../lib/api';

export function CompanyProfile() {
  const { user } = useAuth();
  const [form, setForm] = useState({ companyName: '', website: '', industry: '', description: '' });
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [industries, setIndustries] = useState<Array<{ id: string; name: string }>>([]);

  useEffect(() => {
    api<Array<{ id: string; name: string }>>('/industries').then(setIndustries);
    api<{ companyName: string; website?: string; industry: string; description?: string }>('/company/profile')
      .then((data) =>
        setForm({
          companyName: data.companyName,
          website: data.website || '',
          industry: data.industry,
          description: data.description || '',
        }),
      )
      .finally(() => setLoading(false));
  }, []);

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setSaving(true);
    await api('/company/profile', { method: 'PUT', body: JSON.stringify(form) });
    setSaving(false);
  };

  if (loading)
    return (
      <div className="flex justify-center p-8" role="status" aria-label="Loading company profile">
        <span className="loading loading-spinner loading-lg" aria-hidden="true" />
      </div>
    );
  if (!user || user.role !== 'company') return <div className="text-center p-8 text-error">Access denied</div>;

  return (
    <div className="max-w-lg mx-auto p-4">
      <h1 className="text-2xl font-bold mb-6">Company Profile</h1>
      <form onSubmit={handleSubmit} className="card bg-base-200 p-4 space-y-4">
        <label className="form-control">
          <span className="label-text">Company Name</span>
          <input
            className="input input-bordered"
            value={form.companyName}
            onChange={(e) => setForm({ ...form, companyName: e.target.value })}
            required
          />
        </label>
        <label className="form-control">
          <span className="label-text">Website</span>
          <input
            className="input input-bordered"
            value={form.website}
            onChange={(e) => setForm({ ...form, website: e.target.value })}
          />
        </label>
        <label className="form-control">
          <span className="label-text">Industry</span>
          <select
            className="select select-bordered"
            value={form.industry}
            onChange={(e) => setForm({ ...form, industry: e.target.value })}
          >
            {industries.map((ind) => (
              <option key={ind.id} value={ind.name}>
                {ind.name}
              </option>
            ))}
          </select>
        </label>
        <label className="form-control">
          <span className="label-text">Description</span>
          <textarea
            className="textarea textarea-bordered h-24"
            value={form.description}
            onChange={(e) => setForm({ ...form, description: e.target.value })}
          />
        </label>
        <button type="submit" className="btn btn-primary" disabled={saving}>
          {saving ? <span className="loading loading-spinner" aria-hidden="true" /> : 'Save'}
        </button>
      </form>
    </div>
  );
}
