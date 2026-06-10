import { zodResolver } from '@hookform/resolvers/zod';
import { Plus, Trash2, X } from 'lucide-react';
import { useEffect, useState } from 'react';
import { useForm } from 'react-hook-form';
import { AvatarUploader } from '../../components/jobseeker/AvatarUploader';
import { JobseekerOnboarding } from '../../components/onboarding/JobseekerOnboarding';
import { FormInput, FormSelect, FormTextarea } from '../../components/ui/FormField';
import { LoadingFallback, LoadingSpinner } from '../../components/ui/LoadingFallback';
import { ApiError, api } from '../../lib/api';
import { type ExperienceForm, experienceSchema, type ProfileForm, profileSchema } from '../../lib/schemas';
import { ResumeImport } from './ResumeImport';
import type { Experience, Profile } from './type';

const EXPERIENCE_TYPES = [
  { value: 'employment', label: 'Employment' },
  { value: 'gig', label: 'Gig' },
  { value: 'education', label: 'Education' },
  { value: 'certification', label: 'Certification' },
  { value: 'project', label: 'Project' },
  { value: 'volunteering', label: 'Volunteering' },
];

export function JobseekerProfile() {
  const [profile, setProfile] = useState<Profile | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [showExpForm, setShowExpForm] = useState(false);
  const [importOpen, setImportOpen] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const profileForm = useForm<ProfileForm>({
    resolver: zodResolver(profileSchema),
    defaultValues: { headline: '', about: '', yearsOfExperience: undefined },
  });

  const expForm = useForm<ExperienceForm>({
    resolver: zodResolver(experienceSchema),
    defaultValues: {
      type: 'employment',
      title: '',
      organization: '',
      startDate: '',
      endDate: '',
      isCurrent: false,
      description: '',
      industry: '',
      skills: '',
      url: '',
    },
  });

  useEffect(() => {
    let cancelled = false;
    api<Profile>('/profiles/me')
      .then((data) => {
        if (cancelled) return;
        setProfile(data);
        // Surface the AI importer as step one for brand-new profiles.
        if (data.experiences.length === 0) setImportOpen(true);
        profileForm.reset({
          headline: data.headline || '',
          about: data.about || '',
          yearsOfExperience: data.yearsOfExperience || undefined,
        });
      })
      .catch((err) => {
        if (!cancelled) {
          setError(err instanceof ApiError ? (err.serverMessage ?? err.message) : 'Failed to load profile');
        }
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });
    return () => {
      cancelled = true;
    };
  }, [profileForm]);

  const saveProfile = async (data: ProfileForm) => {
    setSaving(true);
    setError(null);
    try {
      const updated = await api<Profile>('/profiles/me', {
        method: 'PUT',
        body: JSON.stringify(data),
      });
      setProfile((prev) => (prev ? { ...prev, ...updated } : null));
    } catch (err) {
      setError(err instanceof ApiError ? (err.serverMessage ?? err.message) : 'Failed to save profile');
    } finally {
      setSaving(false);
    }
  };

  const addExperience = async (data: ExperienceForm) => {
    setError(null);
    const skills = data.skills
      ? data.skills
          .split(',')
          .map((s) => s.trim())
          .filter(Boolean)
      : [];
    try {
      const added = await api<Experience>('/profiles/me/experience', {
        method: 'POST',
        body: JSON.stringify({
          ...data,
          skillsUsed: skills,
          endDate: data.isCurrent ? undefined : data.endDate || undefined,
          url: data.url || undefined,
        }),
      });
      setProfile((prev) => (prev ? { ...prev, experiences: [...prev.experiences, added] } : null));
      setShowExpForm(false);
      expForm.reset({
        type: 'employment',
        title: '',
        organization: '',
        startDate: '',
        endDate: '',
        isCurrent: false,
        description: '',
        industry: '',
        skills: '',
        url: '',
      });
    } catch (err) {
      setError(err instanceof ApiError ? (err.serverMessage ?? err.message) : 'Failed to add experience');
    }
  };

  const deleteExperience = async (id: string) => {
    setError(null);
    try {
      await api(`/profiles/me/experience/${encodeURIComponent(id)}`, { method: 'DELETE' });
      setProfile((prev) => (prev ? { ...prev, experiences: prev.experiences.filter((e) => e.id !== id) } : null));
    } catch (err) {
      setError(err instanceof ApiError ? (err.serverMessage ?? err.message) : 'Failed to delete experience');
    }
  };

  if (loading) return <LoadingFallback text="Loading profile" />;

  return (
    <div className="max-w-2xl mx-auto p-4 space-y-6">
      <h1 className="text-2xl font-bold">My Profile</h1>

      {error && (
        <div className="alert alert-error">
          <span>{error}</span>
          <button type="button" title="close" className="btn btn-ghost btn-xs" onClick={() => setError(null)}>
            <X size={14} />
          </button>
        </div>
      )}

      {profile && (
        <JobseekerOnboarding
          hasHeadline={Boolean(profile.headline)}
          experienceCount={profile.experiences.length}
          onAddExperience={() => setImportOpen(true)}
        />
      )}

      <form onSubmit={profileForm.handleSubmit(saveProfile)} className="card bg-base-200 p-6 space-y-4">
        <div className="flex items-center justify-between">
          <h2 className="font-semibold">Profile Details</h2>
          {profile && (
            <AvatarUploader
              name={profile.name ?? ''}
              avatarUrl={profile.avatarUrl}
              onUploaded={(url) => setProfile((prev) => (prev ? { ...prev, avatarUrl: url } : prev))}
            />
          )}
        </div>
        <FormInput
          label="Headline"
          registration={profileForm.register('headline')}
          error={profileForm.formState.errors.headline}
          placeholder="e.g. Senior Full-Stack Developer"
        />
        <FormTextarea
          label="About"
          registration={profileForm.register('about')}
          error={profileForm.formState.errors.about}
          rows={4}
        />
        <FormInput
          label="Years of Experience"
          registration={profileForm.register('yearsOfExperience', { valueAsNumber: true })}
          error={profileForm.formState.errors.yearsOfExperience}
          type="number"
          min={0}
        />
        <button type="submit" className="btn btn-primary" disabled={saving}>
          {saving ? <LoadingSpinner size="sm" /> : 'Save Profile'}
        </button>
      </form>

      <ResumeImport
        open={importOpen}
        onToggle={setImportOpen}
        onExperienceAdded={(added) =>
          setProfile((prev) => (prev ? { ...prev, experiences: [...prev.experiences, added] } : prev))
        }
      />

      <div className="card bg-base-200 p-4">
        <div className="flex justify-between items-center mb-3">
          <h2 className="font-semibold">Experience</h2>
          <button type="button" className="btn btn-outline btn-sm gap-1" onClick={() => setShowExpForm(!showExpForm)}>
            <Plus size={16} aria-hidden="true" /> Add
          </button>
        </div>

        {showExpForm && (
          <form onSubmit={expForm.handleSubmit(addExperience)} className="space-y-3 mb-4 p-3 bg-base-100 rounded-box">
            <FormSelect
              label="Type"
              registration={expForm.register('type')}
              error={expForm.formState.errors.type}
              options={EXPERIENCE_TYPES}
            />
            <FormInput label="Title" registration={expForm.register('title')} error={expForm.formState.errors.title} />
            <FormInput
              label="Organization"
              registration={expForm.register('organization')}
              error={expForm.formState.errors.organization}
            />
            <div className="grid grid-cols-2 gap-2">
              <FormInput
                label="Start Date"
                registration={expForm.register('startDate')}
                error={expForm.formState.errors.startDate}
                placeholder="2020-01"
              />
              <FormInput
                label="End Date"
                registration={expForm.register('endDate')}
                error={expForm.formState.errors.endDate}
                placeholder="2023-12"
                disabled={expForm.watch('isCurrent')}
              />
            </div>
            <label className="flex items-center gap-2 cursor-pointer">
              <input type="checkbox" className="checkbox checkbox-sm" {...expForm.register('isCurrent')} />
              <span className="label-text">I currently work here</span>
            </label>
            <FormTextarea
              label="Description"
              registration={expForm.register('description')}
              error={expForm.formState.errors.description}
              rows={3}
            />
            <FormInput
              label="Skills (comma separated)"
              registration={expForm.register('skills')}
              error={expForm.formState.errors.skills}
              placeholder="React, TypeScript, Node.js"
            />
            <FormInput
              label="Evidence URL (optional)"
              registration={expForm.register('url')}
              error={expForm.formState.errors.url}
              placeholder="https://github.com/you/project or certificate link"
            />
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
