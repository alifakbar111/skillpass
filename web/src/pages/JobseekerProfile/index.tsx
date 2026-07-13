import { zodResolver } from '@hookform/resolvers/zod';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { CheckCircle, X } from 'lucide-react';
import { useEffect, useState } from 'react';
import { useForm } from 'react-hook-form';
import { AIEvaluationSection } from '@/components/jobseeker/AIEvaluationSection';
import { AvatarUploader } from '@/components/jobseeker/AvatarUploader';
import { CertificationSection } from '@/components/jobseeker/CertificationSection';
import { EducationSection } from '@/components/jobseeker/EducationSection';
import { ProjectSection } from '@/components/jobseeker/ProjectSection';
import { VolunteeringSection } from '@/components/jobseeker/VolunteeringSection';
import { WorkHistorySection } from '@/components/jobseeker/WorkHistorySection';
import { JobseekerOnboarding } from '@/components/onboarding/JobseekerOnboarding';
import { Form } from '@/components/ui/Form';
import { FormDialog } from '@/components/ui/FormDialog';
import { FormInput } from '@/components/ui/FormInput';
import { FormNumberInput } from '@/components/ui/FormNumberInput';
import { FormTextarea } from '@/components/ui/FormTextarea';
import { LoadingFallback, LoadingSpinner } from '@/components/ui/LoadingFallback';
import { SkillsAutocomplete } from '@/components/ui/SkillsAutocomplete';
import { useAuth } from '@/hooks/useAuth';
import { ApiError, api } from '@/lib/api';
import type { Experience, Profile } from '@/lib/api-types';
import { type ExperienceForm, experienceSchema, type ProfileForm, profileSchema } from '@/lib/schemas';
import { ResumeImport } from '@/pages/JobseekerProfile/ResumeImport';

export function JobseekerProfile() {
  const queryClient = useQueryClient();
  const { user } = useAuth();

  // Profile edits also affect the public passport view (/profiles/:username).
  const invalidateProfileViews = () => {
    queryClient.invalidateQueries({ queryKey: ['profile', 'me'] });
    queryClient.invalidateQueries({ queryKey: ['passport', user?.username] });
  };
  const [showFormType, setShowFormType] = useState<string | null>(null);
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
    mutationFn: (data: ProfileForm) => api<Profile>('/profiles/me', { method: 'PUT', body: data }),
    onMutate: () => setError(null),
    onSuccess: () => {
      invalidateProfileViews();
      profileForm.reset(profileForm.getValues());
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
        body: {
          ...data,
          skillsUsed: skills,
          endDate: data.isCurrent ? undefined : data.endDate || undefined,
          url: data.url || undefined,
        },
      });
    },
    onMutate: () => setError(null),
    onSuccess: () => {
      invalidateProfileViews();
      setShowFormType(null);
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

  const [editingId, setEditingId] = useState<string | null>(null);

  const updateExperienceMutation = useMutation({
    mutationFn: ({ id, data }: { id: string; data: ExperienceForm }) => {
      const skills = data.skills
        ? data.skills
            .split(',')
            .map((s) => s.trim())
            .filter(Boolean)
        : [];
      return api<Experience>(`/profiles/me/experience/${encodeURIComponent(id)}`, {
        method: 'PUT',
        body: {
          ...data,
          skillsUsed: skills,
          endDate: data.isCurrent ? undefined : data.endDate || undefined,
          url: data.url || undefined,
        },
      });
    },
    onMutate: () => setError(null),
    onSuccess: () => {
      invalidateProfileViews();
      setEditingId(null);
      setShowFormType(null);
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
      setError(err instanceof ApiError ? (err.serverMessage ?? err.message) : 'Failed to update experience');
    },
  });

  const saveProfile = (data: ProfileForm) => saveProfileMutation.mutate(data);
  const submitExperience = (data: ExperienceForm) => {
    if (editingId) {
      updateExperienceMutation.mutate({ id: editingId, data });
    } else {
      addExperienceMutation.mutate(data);
    }
  };
  const deleteExperience = (id: string) => deleteExperienceMutation.mutate(id);

  const cancelForm = () => {
    setShowFormType(null);
    setEditingId(null);
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
  };

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
        {profileForm.formState.isDirty ? (
          <button type="submit" className="btn btn-primary" disabled={saveProfileMutation.isPending}>
            {saveProfileMutation.isPending ? <LoadingSpinner size="sm" /> : 'Save Profile'}
          </button>
        ) : (
          <span className="text-sm text-muted flex items-center gap-1">
            <CheckCircle size={14} className="text-success" /> No unsaved changes
          </span>
        )}
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

      {/* Work History — employment + gig */}
      <WorkHistorySection
        experiences={(profile?.experiences ?? []).filter((e) => e.type === 'employment' || e.type === 'gig')}
        onAdd={() => {
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
          setShowFormType('work');
        }}
        onEdit={(id) => {
          const exp = (profile?.experiences ?? []).find((e) => e.id === id);
          if (!exp) return;
          setEditingId(id);
          expForm.reset({
            type: exp.type as ExperienceForm['type'],
            title: exp.title,
            organization: exp.organization,
            startDate: exp.startDate,
            endDate: exp.endDate ?? '',
            isCurrent: exp.isCurrent ?? false,
            description: exp.description ?? '',
            industry: exp.industry ?? '',
            skills: (exp.skillsUsed ?? []).join(', '),
            url: exp.url ?? '',
          });
          setShowFormType('work');
        }}
        onDelete={(id) => deleteExperience(id)}
      />
      <FormDialog
        open={showFormType === 'work'}
        title={editingId ? 'Edit Work' : 'Add Work'}
        onClose={cancelForm}
        actions={
          <button type="submit" form="work-form" className="btn btn-primary btn-sm">
            {editingId ? 'Update Work' : 'Add Work'}
          </button>
        }
      >
        <Form id="work-form" methods={expForm} onSubmit={submitExperience} className="space-y-3">
          <FormInput label="Title" name="title" />
          <FormInput label="Organization" name="organization" />
          <div className="grid grid-cols-2 gap-2">
            <FormInput label="Start Date" name="startDate" type="month" />
            <FormInput label="End Date" name="endDate" type="month" disabled={expForm.watch('isCurrent')} />
          </div>
          <label className="flex items-center gap-2 cursor-pointer">
            <input type="checkbox" className="checkbox checkbox-sm" {...expForm.register('isCurrent')} />
            <span className="label-text">I currently work here</span>
          </label>
          <FormTextarea label="Description" name="description" rows={3} />
          <FormInput label="Industry" name="industry" placeholder="e.g. Technology" />
          <SkillsAutocomplete
            value={expForm.watch('skills') ?? ''}
            onChange={(val) => expForm.setValue('skills', val, { shouldDirty: false })}
            label="Skills"
            placeholder="Type a skill and press Enter"
          />
          <FormInput label="Evidence URL (optional)" name="url" placeholder="https://..." />
        </Form>
      </FormDialog>

      {/* Education */}
      <EducationSection
        experiences={(profile?.experiences ?? []).filter((e) => e.type === 'education')}
        onAdd={() => {
          expForm.reset({
            type: 'education',
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
          setShowFormType('education');
        }}
        onEdit={(id) => {
          const exp = (profile?.experiences ?? []).find((e) => e.id === id);
          if (!exp) return;
          setEditingId(id);
          expForm.reset({
            type: 'education',
            title: exp.title,
            organization: exp.organization,
            startDate: exp.startDate,
            endDate: exp.endDate ?? '',
            isCurrent: exp.isCurrent ?? false,
            description: exp.description ?? '',
            industry: exp.industry ?? '',
            skills: (exp.skillsUsed ?? []).join(', '),
            url: exp.url ?? '',
          });
          setShowFormType('education');
        }}
        onDelete={(id) => deleteExperience(id)}
      />
      <FormDialog
        open={showFormType === 'education'}
        title={editingId ? 'Edit Education' : 'Add Education'}
        onClose={cancelForm}
        actions={
          <button type="submit" form="education-form" className="btn btn-primary btn-sm">
            {editingId ? 'Update Education' : 'Add Education'}
          </button>
        }
      >
        <Form id="education-form" methods={expForm} onSubmit={submitExperience} className="space-y-3">
          <FormInput label="Degree/Diploma" name="title" />
          <FormInput label="Organization" name="organization" />
          <div className="grid grid-cols-2 gap-2">
            <FormInput label="Start Date" name="startDate" type="month" />
            <FormInput label="End Date" name="endDate" type="month" />
          </div>
          <FormTextarea label="Description" name="description" rows={3} />
          <SkillsAutocomplete
            value={expForm.watch('skills') ?? ''}
            onChange={(val) => expForm.setValue('skills', val, { shouldDirty: false })}
            label="Skills"
            placeholder="Type a skill and press Enter"
          />
          <FormInput label="URL (optional)" name="url" placeholder="https://..." />
        </Form>
      </FormDialog>

      {/* Certifications & Licenses */}
      <CertificationSection
        experiences={(profile?.experiences ?? []).filter((e) => e.type === 'certification')}
        onAdd={() => {
          expForm.reset({
            type: 'certification',
            title: '',
            organization: '',
            startDate: '',
            endDate: '',
            description: '',
            industry: '',
            skills: '',
            url: '',
          });
          setShowFormType('certification');
        }}
        onEdit={(id) => {
          const exp = (profile?.experiences ?? []).find((e) => e.id === id);
          if (!exp) return;
          setEditingId(id);
          expForm.reset({
            type: 'certification',
            title: exp.title,
            organization: exp.organization,
            startDate: exp.startDate,
            endDate: exp.endDate ?? '',
            isCurrent: exp.isCurrent ?? false,
            description: exp.description ?? '',
            industry: exp.industry ?? '',
            skills: (exp.skillsUsed ?? []).join(', '),
            url: exp.url ?? '',
          });
          setShowFormType('certification');
        }}
        onDelete={(id) => deleteExperience(id)}
      />
      <FormDialog
        open={showFormType === 'certification'}
        title={editingId ? 'Edit Certification' : 'Add Certification'}
        onClose={cancelForm}
        actions={
          <button type="submit" form="certification-form" className="btn btn-primary btn-sm">
            {editingId ? 'Update Certification' : 'Add Certification'}
          </button>
        }
      >
        <Form id="certification-form" methods={expForm} onSubmit={submitExperience} className="space-y-3">
          <FormInput label="Title" name="title" />
          <FormInput label="Issuer" name="organization" />
          <div className="grid grid-cols-2 gap-2">
            <FormInput label="Issue Date" name="startDate" type="month" />
            <FormInput label="Expiry Date" name="endDate" type="month" />
          </div>
          <FormTextarea label="Description" name="description" rows={3} />
          <SkillsAutocomplete
            value={expForm.watch('skills') ?? ''}
            onChange={(val) => expForm.setValue('skills', val, { shouldDirty: false })}
            label="Skills"
            placeholder="Type a skill and press Enter"
          />
          <FormInput label="Verification URL (optional)" name="url" placeholder="https://..." />
        </Form>
      </FormDialog>

      {/* Projects & Portfolio */}
      <ProjectSection
        experiences={(profile?.experiences ?? []).filter((e) => e.type === 'project')}
        onAdd={() => {
          expForm.reset({
            type: 'project',
            title: '',
            organization: '',
            startDate: '',
            endDate: '',
            description: '',
            industry: '',
            skills: '',
            url: '',
          });
          setShowFormType('project');
        }}
        onEdit={(id) => {
          const exp = (profile?.experiences ?? []).find((e) => e.id === id);
          if (!exp) return;
          setEditingId(id);
          expForm.reset({
            type: 'project',
            title: exp.title,
            organization: exp.organization ?? '',
            startDate: exp.startDate,
            endDate: exp.endDate ?? '',
            isCurrent: exp.isCurrent ?? false,
            description: exp.description ?? '',
            industry: exp.industry ?? '',
            skills: (exp.skillsUsed ?? []).join(', '),
            url: exp.url ?? '',
          });
          setShowFormType('project');
        }}
        onDelete={(id) => deleteExperience(id)}
      />
      <FormDialog
        open={showFormType === 'project'}
        title={editingId ? 'Edit Project' : 'Add Project'}
        onClose={cancelForm}
        actions={
          <button type="submit" form="project-form" className="btn btn-primary btn-sm">
            {editingId ? 'Update Project' : 'Add Project'}
          </button>
        }
      >
        <Form id="project-form" methods={expForm} onSubmit={submitExperience} className="space-y-3">
          <FormInput label="Title" name="title" />
          <FormInput label="Organization (optional)" name="organization" />
          <div className="grid grid-cols-2 gap-2">
            <FormInput label="Start Date" name="startDate" type="month" />
            <FormInput label="End Date" name="endDate" type="month" />
          </div>
          <FormTextarea label="Description" name="description" rows={3} />
          <SkillsAutocomplete
            value={expForm.watch('skills') ?? ''}
            onChange={(val) => expForm.setValue('skills', val, { shouldDirty: false })}
            label="Skills"
            placeholder="Type a skill and press Enter"
          />
          <FormInput label="Project URL (optional)" name="url" placeholder="https://..." />
        </Form>
      </FormDialog>

      {/* Volunteering */}
      <VolunteeringSection
        experiences={(profile?.experiences ?? []).filter((e) => e.type === 'volunteering')}
        onAdd={() => {
          expForm.reset({
            type: 'volunteering',
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
          setShowFormType('volunteering');
        }}
        onEdit={(id) => {
          const exp = (profile?.experiences ?? []).find((e) => e.id === id);
          if (!exp) return;
          setEditingId(id);
          expForm.reset({
            type: 'volunteering',
            title: exp.title,
            organization: exp.organization,
            startDate: exp.startDate,
            endDate: exp.endDate ?? '',
            isCurrent: exp.isCurrent ?? false,
            description: exp.description ?? '',
            industry: exp.industry ?? '',
            skills: (exp.skillsUsed ?? []).join(', '),
            url: exp.url ?? '',
          });
          setShowFormType('volunteering');
        }}
        onDelete={(id) => deleteExperience(id)}
      />
      <FormDialog
        open={showFormType === 'volunteering'}
        title={editingId ? 'Edit Volunteering' : 'Add Volunteering'}
        onClose={cancelForm}
        actions={
          <button type="submit" form="volunteering-form" className="btn btn-primary btn-sm">
            {editingId ? 'Update Volunteering' : 'Add Volunteering'}
          </button>
        }
      >
        <Form id="volunteering-form" methods={expForm} onSubmit={submitExperience} className="space-y-3">
          <FormInput label="Title" name="title" />
          <FormInput label="Organization" name="organization" />
          <div className="grid grid-cols-2 gap-2">
            <FormInput label="Start Date" name="startDate" type="month" />
            <FormInput label="End Date" name="endDate" type="month" disabled={expForm.watch('isCurrent')} />
          </div>
          <label className="flex items-center gap-2 cursor-pointer">
            <input type="checkbox" className="checkbox checkbox-sm" {...expForm.register('isCurrent')} />
            <span className="label-text">I currently volunteer here</span>
          </label>
          <FormTextarea label="Description" name="description" rows={3} />
          <SkillsAutocomplete
            value={expForm.watch('skills') ?? ''}
            onChange={(val) => expForm.setValue('skills', val, { shouldDirty: false })}
            label="Skills"
            placeholder="Type a skill and press Enter"
          />
          <FormInput label="URL (optional)" name="url" placeholder="https://..." />
        </Form>
      </FormDialog>

      <AIEvaluationSection />
    </div>
  );
}
