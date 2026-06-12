import { zodResolver } from '@hookform/resolvers/zod';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Plus, Trash2, X } from 'lucide-react';
import { useEffect, useState } from 'react';
import { useForm } from 'react-hook-form';
import type { Experience, Profile } from '@/lib/api-types';
import { AvatarUploader } from '../../components/jobseeker/AvatarUploader';
import { JobseekerOnboarding } from '../../components/onboarding/JobseekerOnboarding';
import { Form } from '../../components/ui/Form';
import { FormInput } from '../../components/ui/FormInput';
import { FormNumberInput } from '../../components/ui/FormNumberInput';
import { FormSelect } from '../../components/ui/FormSelect';
import { FormTextarea } from '../../components/ui/FormTextarea';
import { LoadingFallback, LoadingSpinner } from '../../components/ui/LoadingFallback';
import { useAuth } from '../../hooks/useAuth';
import { ApiError, api } from '../../lib/api';
import { type ExperienceForm, experienceSchema, type ProfileForm, profileSchema } from '../../lib/schemas';
import { ResumeImport } from './ResumeImport';

const EXPERIENCE_TYPES = [
  { value: 'employment', label: 'Employment' },
  { value: 'gig', label: 'Gig' },
  { value: 'education', label: 'Education' },
  { value: 'certification', label: 'Certification' },
  { value: 'project', label: 'Project' },
  { value: 'volunteering', label: 'Volunteering' },
];

export function JobseekerProfile() {
  const queryClient = useQueryClient();
  const { user } = useAuth();

  // Profile edits also affect the public passport view (/profiles/:username).
  const invalidateProfileViews = () => {
    queryClient.invalidateQueries({ queryKey: ['profile', 'me'] });
    queryClient.invalidateQueries({ queryKey: ['passport', user?.username] });
  };
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

  const {
    data: profile,
    error: loadError,
    isLoading: loading,
  } = useQuery({
    queryKey: ['profile', 'me'],
    queryFn: () => api<Profile>('/profiles/me'),
  });

  // Seed the profile form once data loads (react-hook-form reset moved out of .then()).
  useEffect(() => {
    if (!profile) return;
    // Surface the AI importer as step one for brand-new profiles.
    if ((profile.experiences ?? []).length === 0) setImportOpen(true);
    profileForm.reset({
      headline: profile.headline || '',
      about: profile.about || '',
      yearsOfExperience: profile.yearsOfExperience || undefined,
    });
  }, [profile, profileForm]);

  const saveProfileMutation = useMutation({
    mutationFn: (data: ProfileForm) => api<Profile>('/profiles/me', { method: 'PUT', body: JSON.stringify(data) }),
    onMutate: () => setError(null),
    onSuccess: () => {
      invalidateProfileViews();
    },
    onError: (err) => {
      setError(err instanceof ApiError ? (err.serverMessage ?? err.message) : 'Failed to save profile');
    },
  });

  const addExperienceMutation = useMutation({
    mutationFn: (data: ExperienceForm) => {
      const skills = data.skills
        ? data.skills
            .split(',')
            .map((s) => s.trim())
            .filter(Boolean)
        : [];
      return api<Experience>('/profiles/me/experience', {
        method: 'POST',
        body: JSON.stringify({
          ...data,
          skillsUsed: skills,
          endDate: data.isCurrent ? undefined : data.endDate || undefined,
          url: data.url || undefined,
        }),
      });
    },
    onMutate: () => setError(null),
    onSuccess: () => {
      invalidateProfileViews();
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
    },
    onError: (err) => {
      setError(err instanceof ApiError ? (err.serverMessage ?? err.message) : 'Failed to add experience');
    },
  });

  const deleteExperienceMutation = useMutation({
    mutationFn: (id: string) => api(`/profiles/me/experience/${encodeURIComponent(id)}`, { method: 'DELETE' }),
    onMutate: () => setError(null),
    onSuccess: () => {
      invalidateProfileViews();
    },
    onError: (err) => {
      setError(err instanceof ApiError ? (err.serverMessage ?? err.message) : 'Failed to delete experience');
    },
  });

  const saveProfile = (data: ProfileForm) => saveProfileMutation.mutate(data);
  const addExperience = (data: ExperienceForm) => addExperienceMutation.mutate(data);
  const deleteExperience = (id: string) => deleteExperienceMutation.mutate(id);

  if (loading) return <LoadingFallback text="Loading profile" />;

  if (loadError && !profile) {
    const message =
      loadError instanceof ApiError ? (loadError.serverMessage ?? loadError.message) : 'Failed to load profile';
    return <p className="text-center p-8 text-error">{message}</p>;
  }

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
          experienceCount={(profile.experiences ?? []).length}
          onAddExperience={() => setImportOpen(true)}
        />
      )}

      <Form methods={profileForm} onSubmit={saveProfile} className="card bg-base-200 p-6 space-y-4">
        <div className="flex items-center justify-between">
          <h2 id="profile-details-sections" className="font-semibold">
            Profile Details
          </h2>
          {profile && (
            <AvatarUploader
              name={profile.name ?? ''}
              avatarUrl={profile.avatarUrl}
              onUploaded={(url) => {
                queryClient.setQueryData(['profile', 'me'], { ...profile, avatarUrl: url });
                queryClient.invalidateQueries({ queryKey: ['passport', user?.username] });
              }}
            />
          )}
        </div>
        <FormInput label="Headline" name="headline" placeholder="e.g. Senior Full-Stack Developer" />
        <FormTextarea label="About" name="about" rows={4} />
        <FormNumberInput label="Years of Experience" name="yearsOfExperience" min={0} />
        <button type="submit" className="btn btn-primary" disabled={saveProfileMutation.isPending}>
          {saveProfileMutation.isPending ? <LoadingSpinner size="sm" /> : 'Save Profile'}
        </button>
      </Form>

      <ResumeImport
        open={importOpen}
        onToggle={setImportOpen}
        onExperienceAdded={(added) => {
          if (!profile) return;
          queryClient.setQueryData(['profile', 'me'], {
            ...profile,
            experiences: [...(profile.experiences ?? []), added],
          });
          queryClient.invalidateQueries({ queryKey: ['passport', user?.username] });
        }}
      />

      <div className="card bg-base-200 p-4">
        <div className="flex justify-between items-center mb-3">
          <h2 id="experience-sections" className="font-semibold">
            Experience
          </h2>
          <button type="button" className="btn btn-outline btn-sm gap-1" onClick={() => setShowExpForm(!showExpForm)}>
            <Plus size={16} aria-hidden="true" /> Add
          </button>
        </div>

        {showExpForm && (
          <Form methods={expForm} onSubmit={addExperience} className="space-y-3 mb-4 p-3 bg-base-100 rounded-box">
            <FormSelect label="Type" name="type" options={EXPERIENCE_TYPES} />
            <FormInput label="Title" name="title" />
            <FormInput label="Organization" name="organization" />
            <div className="grid grid-cols-2 gap-2">
              <FormInput label="Start Date" name="startDate" placeholder="2020-01" />
              <FormInput label="End Date" name="endDate" placeholder="2023-12" disabled={expForm.watch('isCurrent')} />
            </div>
            <label className="flex items-center gap-2 cursor-pointer">
              <input type="checkbox" className="checkbox checkbox-sm" {...expForm.register('isCurrent')} />
              <span className="label-text">I currently work here</span>
            </label>
            <FormTextarea label="Description" name="description" rows={3} />
            <FormInput label="Skills (comma separated)" name="skills" placeholder="React, TypeScript, Node.js" />
            <FormInput
              label="Evidence URL (optional)"
              name="url"
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
          </Form>
        )}

        {(profile?.experiences ?? []).length === 0 && (
          <p className="text-sm text-muted py-4 text-center">No experience added yet. Click "Add" to get started.</p>
        )}

        <div className="space-y-2">
          {(profile?.experiences ?? []).map((exp) => (
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
                onClick={() => exp.id && deleteExperience(exp.id)}
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
