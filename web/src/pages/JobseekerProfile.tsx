import { Plus, Trash2 } from 'lucide-react';
import { type FormEvent, useEffect, useState } from 'react';
import { LoadingFallback, LoadingSpinner } from '../components/ui/LoadingFallback';
import { useAuth } from '../hooks/useAuth';
import { api } from '../lib/api';

interface Experience {
  id: string;
  type: string;
  title: string;
  organization: string;
  startDate: string;
  endDate?: string;
  isCurrent: boolean;
  description?: string;
  industry?: string;
  skillsUsed?: string[];
  url?: string;
}

interface Profile {
  id: string;
  headline?: string;
  about?: string;
  yearsOfExperience?: number;
  slug: string;
  experiences: Experience[];
}

export function JobseekerProfile() {
  const { user } = useAuth();
  const [profile, setProfile] = useState<Profile | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [form, setForm] = useState({ headline: '', about: '', yearsOfExperience: 0 });
  const [showExpForm, setShowExpForm] = useState(false);
  const [expForm, setExpForm] = useState({
    type: 'employment',
    title: '',
    organization: '',
    startDate: '',
    endDate: '',
    isCurrent: false,
    description: '',
    industry: '',
    skills: '',
  });

  useEffect(() => {
    api<Profile>('/profiles/me')
      .then((data) => {
        setProfile(data);
        setForm({
          headline: data.headline || '',
          about: data.about || '',
          yearsOfExperience: data.yearsOfExperience || 0,
        });
      })
      .finally(() => setLoading(false));
  }, []);

  const saveProfile = async (e: FormEvent) => {
    e.preventDefault();
    setSaving(true);
    const updated = await api<Profile>('/profiles/me', {
      method: 'PUT',
      body: JSON.stringify(form),
    });
    setProfile((prev) => (prev ? { ...prev, ...updated } : null));
    setSaving(false);
  };

  const addExperience = async (e: FormEvent) => {
    e.preventDefault();
    const skills = expForm.skills
      .split(',')
      .map((s) => s.trim())
      .filter(Boolean);
    const added = await api<Experience>('/profiles/me/experience', {
      method: 'POST',
      body: JSON.stringify({
        ...expForm,
        skillsUsed: skills,
        endDate: expForm.isCurrent ? undefined : expForm.endDate || undefined,
      }),
    });
    setProfile((prev) => (prev ? { ...prev, experiences: [...prev.experiences, added] } : null));
    setShowExpForm(false);
    setExpForm({
      type: 'employment',
      title: '',
      organization: '',
      startDate: '',
      endDate: '',
      isCurrent: false,
      description: '',
      industry: '',
      skills: '',
    });
  };

  const deleteExperience = async (id: string) => {
    await api(`/profiles/me/experience/${id}`, { method: 'DELETE' });
    setProfile((prev) => (prev ? { ...prev, experiences: prev.experiences.filter((e) => e.id !== id) } : null));
  };

  if (loading) return <LoadingFallback text="Loading profile" />;
  if (!user || user.role !== 'jobseeker') return <div className="text-center p-8 text-error">Access denied</div>;

  return (
    <div className="max-w-2xl mx-auto p-4 space-y-6">
      <h1 className="text-2xl font-bold">My Profile</h1>

      {/* Profile form */}
      <form onSubmit={saveProfile} className="card bg-base-200 p-6 space-y-4">
        <h2 className="font-semibold">Profile Details</h2>
        <div className="form-control">
          <label htmlFor="headline" className="label-text mb-1">
            Headline
          </label>
          <input
            id="headline"
            className="input input-bordered"
            value={form.headline}
            onChange={(e) => setForm({ ...form, headline: e.target.value })}
            placeholder="e.g. Senior Full-Stack Developer"
          />
        </div>
        <div className="form-control">
          <label htmlFor="about" className="label-text mb-1">
            About
          </label>
          <textarea
            id="about"
            className="textarea textarea-bordered"
            rows={4}
            value={form.about}
            onChange={(e) => setForm({ ...form, about: e.target.value })}
          />
        </div>
        <div className="form-control">
          <label htmlFor="yearsOfExperience" className="label-text mb-1">
            Years of Experience
          </label>
          <input
            id="yearsOfExperience"
            type="number"
            className="input input-bordered w-32"
            value={form.yearsOfExperience}
            onChange={(e) => setForm({ ...form, yearsOfExperience: Number(e.target.value) })}
          />
        </div>
        <button type="submit" className="btn btn-primary" disabled={saving}>
          {saving ? <LoadingSpinner size="sm" /> : 'Save Profile'}
        </button>
      </form>

      {/* Experience list */}
      <div className="card bg-base-200 p-4">
        <div className="flex justify-between items-center mb-3">
          <h2 className="font-semibold">Experience</h2>
          <button type="button" className="btn btn-outline btn-sm gap-1" onClick={() => setShowExpForm(!showExpForm)}>
            <Plus size={16} aria-hidden="true" /> Add
          </button>
        </div>

        {showExpForm && (
          <form onSubmit={addExperience} className="space-y-3 mb-4 p-3 bg-base-100 rounded-box">
            <div className="form-control">
              <label htmlFor="exp-type" className="label-text mb-1">
                Type
              </label>
              <select
                id="exp-type"
                className="select select-bordered"
                value={expForm.type}
                onChange={(e) => setExpForm({ ...expForm, type: e.target.value })}
              >
                <option value="employment">Employment</option>
                <option value="gig">Gig</option>
                <option value="education">Education</option>
                <option value="certification">Certification</option>
                <option value="project">Project</option>
                <option value="volunteering">Volunteering</option>
              </select>
            </div>
            <div className="form-control">
              <label htmlFor="exp-title" className="label-text mb-1">
                Title
              </label>
              <input
                id="exp-title"
                className="input input-bordered"
                value={expForm.title}
                onChange={(e) => setExpForm({ ...expForm, title: e.target.value })}
                required
              />
            </div>
            <div className="form-control">
              <label htmlFor="exp-org" className="label-text mb-1">
                Organization
              </label>
              <input
                id="exp-org"
                className="input input-bordered"
                value={expForm.organization}
                onChange={(e) => setExpForm({ ...expForm, organization: e.target.value })}
              />
            </div>
            <div className="grid grid-cols-2 gap-2">
              <div className="form-control">
                <label htmlFor="exp-start" className="label-text mb-1">
                  Start Date
                </label>
                <input
                  id="exp-start"
                  type="text"
                  className="input input-bordered"
                  placeholder="2020-01"
                  value={expForm.startDate}
                  onChange={(e) => setExpForm({ ...expForm, startDate: e.target.value })}
                  required
                />
              </div>
              <div className="form-control">
                <label htmlFor="exp-end" className="label-text mb-1">
                  End Date
                </label>
                <input
                  id="exp-end"
                  type="text"
                  className="input input-bordered"
                  placeholder="2023-12"
                  value={expForm.endDate}
                  onChange={(e) => setExpForm({ ...expForm, endDate: e.target.value })}
                  disabled={expForm.isCurrent}
                />
              </div>
            </div>
            <label className="flex items-center gap-2 cursor-pointer">
              <input
                type="checkbox"
                className="checkbox checkbox-sm"
                checked={expForm.isCurrent}
                onChange={(e) => setExpForm({ ...expForm, isCurrent: e.target.checked })}
              />
              <span className="label-text">I currently work here</span>
            </label>
            <div className="form-control">
              <label htmlFor="exp-desc" className="label-text mb-1">
                Description
              </label>
              <textarea
                id="exp-desc"
                className="textarea textarea-bordered"
                rows={3}
                value={expForm.description}
                onChange={(e) => setExpForm({ ...expForm, description: e.target.value })}
              />
            </div>
            <div className="form-control">
              <label htmlFor="exp-skills" className="label-text mb-1">
                Skills (comma separated)
              </label>
              <input
                id="exp-skills"
                className="input input-bordered"
                value={expForm.skills}
                onChange={(e) => setExpForm({ ...expForm, skills: e.target.value })}
                placeholder="React, TypeScript, Node.js"
              />
            </div>
            <div className="flex gap-2">
              <button type="submit" className="btn btn-primary btn-sm">
                Add Experience
              </button>
              <button type="button" className="btn btn-ghost btn-sm" onClick={() => setShowExpForm(false)}>
                Cancel
              </button>
            </div>
          </form>
        )}

        {profile?.experiences.length === 0 && (
          <p className="text-sm text-muted py-4 text-center">No experience added yet. Click "Add" to get started.</p>
        )}

        <div className="space-y-2">
          {profile?.experiences.map((exp) => (
            <div key={exp.id} className="p-3 bg-base-100 rounded-box flex justify-between items-start">
              <div>
                <p className="font-medium">{exp.title}</p>
                <p className="text-sm text-muted">
                  {exp.organization} &middot; {exp.startDate}
                  {exp.isCurrent ? ' - Present' : exp.endDate ? ` - ${exp.endDate}` : ''}
                </p>
              </div>
              <button
                type="button"
                className="btn btn-ghost btn-xs text-error"
                onClick={() => deleteExperience(exp.id)}
                aria-label={`Delete ${exp.title}`}
              >
                <Trash2 size={16} aria-hidden="true" />
              </button>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}
