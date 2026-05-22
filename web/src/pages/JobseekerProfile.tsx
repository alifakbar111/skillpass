import { useState, useEffect, type FormEvent } from 'react';
import { api } from '../lib/api';
import { useAuth } from '../hooks/useAuth';
import { Plus, Trash2 } from 'lucide-react';

interface Experience {
  id: string; type: string; title: string; organization: string;
  startDate: string; endDate?: string; isCurrent: boolean;
  description?: string; industry?: string; skillsUsed?: string[]; url?: string;
}

interface Profile {
  id: string; headline?: string; about?: string; yearsOfExperience?: number; slug: string;
  experiences: Experience[];
}

export function JobseekerProfile() {
  const { user } = useAuth();
  const [profile, setProfile] = useState<Profile | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [form, setForm] = useState({ headline: '', about: '', yearsOfExperience: 0 });
  const [showExpForm, setShowExpForm] = useState(false);
  const [expForm, setExpForm] = useState({ type: 'employment', title: '', organization: '', startDate: '', endDate: '', isCurrent: false, description: '', industry: '', skills: '' });

  useEffect(() => {
    api<Profile>('/profiles/me').then((data) => {
      setProfile(data);
      setForm({ headline: data.headline || '', about: data.about || '', yearsOfExperience: data.yearsOfExperience || 0 });
    }).finally(() => setLoading(false));
  }, []);

  const saveProfile = async (e: FormEvent) => {
    e.preventDefault();
    setSaving(true);
    const updated = await api<Profile>('/profiles/me', {
      method: 'PUT', body: JSON.stringify(form),
    });
    setProfile(prev => prev ? { ...prev, ...updated } : null);
    setSaving(false);
  };

  const addExperience = async (e: FormEvent) => {
    e.preventDefault();
    const skills = expForm.skills.split(',').map(s => s.trim()).filter(Boolean);
    const added = await api<Experience>('/profiles/me/experience', {
      method: 'POST',
      body: JSON.stringify({ ...expForm, skillsUsed: skills, endDate: expForm.isCurrent ? undefined : expForm.endDate || undefined }),
    });
    setProfile(prev => prev ? { ...prev, experiences: [...prev.experiences, added] } : null);
    setExpForm({ type: 'employment', title: '', organization: '', startDate: '', endDate: '', isCurrent: false, description: '', industry: '', skills: '' });
    setShowExpForm(false);
  };

  const deleteExperience = async (id: string) => {
    await api(`/profiles/me/experience/${id}`, { method: 'DELETE' });
    setProfile(prev => prev ? { ...prev, experiences: prev.experiences.filter(e => e.id !== id) } : null);
  };

  if (loading) return <div className="flex justify-center p-8"><span className="loading loading-spinner loading-lg" /></div>;
  if (!user || user.role !== 'jobseeker') return <div className="text-center p-8 text-error">Access denied</div>;

  return (
    <div className="max-w-2xl mx-auto p-4 space-y-6">
      <h1 className="text-2xl font-bold">My Profile</h1>

      <form onSubmit={saveProfile} className="card bg-base-200 p-4 space-y-4">
        <label className="form-control">
          <span className="label-text">Headline</span>
          <input className="input input-bordered" value={form.headline} onChange={e => setForm({ ...form, headline: e.target.value })} placeholder="e.g. Senior Software Engineer" />
        </label>
        <label className="form-control">
          <span className="label-text">About</span>
          <textarea className="textarea textarea-bordered h-24" value={form.about} onChange={e => setForm({ ...form, about: e.target.value })} placeholder="Tell companies about yourself" />
        </label>
        <label className="form-control">
          <span className="label-text">Years of Experience</span>
          <input type="number" className="input input-bordered" value={form.yearsOfExperience} onChange={e => setForm({ ...form, yearsOfExperience: Number(e.target.value) })} />
        </label>
        <button type="submit" className="btn btn-primary" disabled={saving}>
          {saving ? <span className="loading loading-spinner" /> : 'Save Profile'}
        </button>
      </form>

      <div className="card bg-base-200 p-4">
        <div className="flex justify-between items-center mb-4">
          <h2 className="text-xl font-semibold">Experience</h2>
          <button className="btn btn-primary btn-sm" onClick={() => setShowExpForm(!showExpForm)}>
            <Plus size={16} /> Add
          </button>
        </div>

        {showExpForm && (
          <form onSubmit={addExperience} className="space-y-3 mb-4 p-3 border border-base-300 rounded-box">
            <select className="select select-bordered w-full" value={expForm.type}
              onChange={e => setExpForm({ ...expForm, type: e.target.value })}>
              <option value="employment">Employment</option>
              <option value="gig">Gig / Freelance</option>
              <option value="education">Education</option>
              <option value="certification">Certification</option>
              <option value="project">Project</option>
              <option value="volunteering">Volunteering</option>
            </select>
            <input className="input input-bordered w-full" placeholder="Title / Degree" value={expForm.title}
              onChange={e => setExpForm({ ...expForm, title: e.target.value })} required />
            <input className="input input-bordered w-full" placeholder="Organization / Institution" value={expForm.organization}
              onChange={e => setExpForm({ ...expForm, organization: e.target.value })} required />
            <div className="flex gap-2">
              <input type="date" className="input input-bordered flex-1" value={expForm.startDate}
                onChange={e => setExpForm({ ...expForm, startDate: e.target.value })} required />
              <input type="date" className="input input-bordered flex-1" value={expForm.endDate}
                onChange={e => setExpForm({ ...expForm, endDate: e.target.value })} disabled={expForm.isCurrent} />
            </div>
            <label className="flex items-center gap-2">
              <input type="checkbox" className="checkbox checkbox-sm" checked={expForm.isCurrent}
                onChange={e => setExpForm({ ...expForm, isCurrent: e.target.checked })} />
              <span className="text-sm">I currently work here</span>
            </label>
            <textarea className="textarea textarea-bordered w-full" placeholder="Description" value={expForm.description}
              onChange={e => setExpForm({ ...expForm, description: e.target.value })} />
            <input className="input input-bordered w-full" placeholder="Skills (comma-separated)" value={expForm.skills}
              onChange={e => setExpForm({ ...expForm, skills: e.target.value })} />
            <button type="submit" className="btn btn-primary btn-sm">Add Experience</button>
          </form>
        )}

        <div className="space-y-2">
          {profile?.experiences.map(exp => (
            <div key={exp.id} className="flex justify-between items-start p-3 bg-base-100 rounded-box">
              <div>
                <p className="font-medium">{exp.title}</p>
                <p className="text-sm opacity-70">{exp.organization} · {exp.startDate} {exp.isCurrent ? '- Present' : exp.endDate ? `- ${exp.endDate}` : ''}</p>
                {exp.skillsUsed && exp.skillsUsed.length > 0 && (
                  <div className="flex flex-wrap gap-1 mt-1">
                    {exp.skillsUsed.map(s => <span key={s} className="badge badge-sm">{s}</span>)}
                  </div>
                )}
              </div>
              <button className="btn btn-ghost btn-xs text-error" onClick={() => deleteExperience(exp.id)}>
                <Trash2 size={14} />
              </button>
            </div>
          ))}
          {(!profile?.experiences || profile.experiences.length === 0) && (
            <p className="text-sm opacity-50 text-center py-4">No experience added yet</p>
          )}
        </div>
      </div>
    </div>
  );
}
