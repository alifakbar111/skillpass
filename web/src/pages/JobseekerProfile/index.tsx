import { closestCenter, DndContext, type DragEndEvent } from '@dnd-kit/core';
import { SortableContext, useSortable, verticalListSortingStrategy } from '@dnd-kit/sortable';
import { CSS } from '@dnd-kit/utilities';
import { zodResolver } from '@hookform/resolvers/zod';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { CheckCircle, GripVertical, Pencil, Plus, Trash2, X } from 'lucide-react';
import { useEffect, useState } from 'react';
import { useForm } from 'react-hook-form';
import { AIEvaluationSection } from '@/components/jobseeker/AIEvaluationSection';
import { AvatarUploader } from '@/components/jobseeker/AvatarUploader';
import { JobseekerOnboarding } from '@/components/onboarding/JobseekerOnboarding';
import { Form } from '@/components/ui/Form';
import { FormInput } from '@/components/ui/FormInput';
import { FormNumberInput } from '@/components/ui/FormNumberInput';
import { FormSelect } from '@/components/ui/FormSelect';
import { FormTextarea } from '@/components/ui/FormTextarea';
import { LoadingFallback, LoadingSpinner } from '@/components/ui/LoadingFallback';
import { SkillsAutocomplete } from '@/components/ui/SkillsAutocomplete';
import { useAuth } from '@/hooks/useAuth';
import { ApiError, api } from '@/lib/api';
import type { Experience, Profile } from '@/lib/api-types';
import { type ExperienceForm, experienceSchema, type ProfileForm, profileSchema } from '@/lib/schemas';
import { ResumeImport } from '@/pages/JobseekerProfile/ResumeImport';

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

  const [editingId, setEditingId] = useState<string | null>(null);

  const [sortedExperiences, setSortedExperiences] = useState<Experience[]>([]);

  useEffect(() => {
    if (profile?.experiences) {
      setSortedExperiences(profile.experiences);
    }
  }, [profile?.experiences]);

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
      setError(err instanceof ApiError ? (err.serverMessage ?? err.message) : 'Failed to update experience');
    },
  });

  const reorderMutation = useMutation({
    mutationFn: (experiences: { id: string; sortOrder: number }[]) =>
      api('/profiles/me/experience/reorder', {
        method: 'PUT',
        body: { experiences },
      }),
    onError: () => {
      if (profile?.experiences) {
        setSortedExperiences(profile.experiences);
      }
      setError('Failed to save new order. Please try again.');
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

  const handleDragEnd = (event: DragEndEvent) => {
    const { active, over } = event;
    if (!over || active.id === over.id) return;

    const oldIndex = sortedExperiences.findIndex((e) => e.id === active.id);
    const newIndex = sortedExperiences.findIndex((e) => e.id === over.id);
    if (oldIndex === -1 || newIndex === -1) return;

    const newOrder = [...sortedExperiences];
    const [moved] = newOrder.splice(oldIndex, 1);
    newOrder.splice(newIndex, 0, moved);
    setSortedExperiences(newOrder);

    const reorderPayload = newOrder.map((exp, i) => ({ id: exp.id ?? '', sortOrder: i }));
    reorderMutation.mutate(reorderPayload);
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
          <Form methods={expForm} onSubmit={submitExperience} className="space-y-3 mb-4 p-3 bg-base-100 rounded-box">
            <FormSelect label="Type" name="type" options={EXPERIENCE_TYPES} />
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
            <SkillsAutocomplete
              value={expForm.watch('skills') ?? ''}
              onChange={(val) => expForm.setValue('skills', val, { shouldDirty: false })}
              label="Skills"
              placeholder="Type a skill and press Enter"
            />
            <FormInput
              label="Evidence URL (optional)"
              name="url"
              placeholder="https://github.com/you/project or certificate link"
            />
            <div className="flex gap-2">
              <button type="submit" className="btn btn-primary btn-sm">
                {editingId ? 'Update Experience' : 'Add Experience'}
              </button>
              <button
                type="button"
                className="btn btn-ghost btn-sm"
                onClick={() => {
                  setShowExpForm(false);
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
                }}
              >
                Cancel
              </button>
            </div>
          </Form>
        )}

        <DndContext collisionDetection={closestCenter} onDragEnd={handleDragEnd}>
          <SortableContext items={sortedExperiences.map((e) => e.id ?? '')} strategy={verticalListSortingStrategy}>
            <div className="space-y-2">
              {sortedExperiences.length === 0 && (
                <p className="text-sm text-muted py-4 text-center">
                  No experience added yet. Click "Add" to get started.
                </p>
              )}
              {sortedExperiences.map((exp) => (
                <SortableExperienceItem
                  key={exp.id}
                  exp={exp}
                  onEdit={() => {
                    if (!exp.id) return;
                    setEditingId(exp.id);
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
                    setShowExpForm(true);
                    window.scrollTo({ top: 0, behavior: 'smooth' });
                  }}
                  onDelete={() => exp.id && deleteExperience(exp.id)}
                />
              ))}
            </div>
          </SortableContext>
        </DndContext>
      </div>

      <AIEvaluationSection />
    </div>
  );
}

function SortableExperienceItem({
  exp,
  onEdit,
  onDelete,
}: {
  exp: Experience;
  onEdit: () => void;
  onDelete: () => void;
}) {
  const sortableId = exp.id ?? '';
  const { attributes, listeners, setNodeRef, transform, transition, isDragging } = useSortable({
    id: sortableId,
  });
  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
    opacity: isDragging ? 0.5 : 1,
  };

  return (
    <div
      ref={setNodeRef}
      style={style}
      className={`collapse collapse-arrow bg-base-100 rounded-box ${isDragging ? 'shadow-lg' : ''}`}
    >
      <summary className="collapse-title flex items-center gap-2">
        <button
          className="cursor-grab touch-none btn btn-ghost btn-xs px-1"
          {...attributes}
          {...listeners}
          aria-label="Drag to reorder"
        >
          <GripVertical size={16} className="text-muted" />
        </button>
        <div className="flex-1">
          <p className="font-medium">{exp.title}</p>
          <p className="text-sm text-muted">
            {exp.organization} &middot; {exp.startDate}
            {exp.isCurrent ? ' - Present' : exp.endDate ? ` - ${exp.endDate}` : ''}
          </p>
        </div>
        <div className="flex gap-1">
          <button
            type="button"
            className="btn btn-ghost btn-xs"
            onClick={(e) => {
              e.stopPropagation();
              onEdit();
            }}
            aria-label={`Edit ${exp.title}`}
          >
            <Pencil size={16} aria-hidden="true" />
          </button>
          <button
            type="button"
            className="btn btn-ghost btn-xs text-error"
            onClick={(e) => {
              e.stopPropagation();
              onDelete();
            }}
            aria-label={`Delete ${exp.title}`}
          >
            <Trash2 size={16} aria-hidden="true" />
          </button>
        </div>
      </summary>
      <div className="collapse-content space-y-2">
        {exp.description && <p className="text-sm">{exp.description}</p>}
        {exp.skillsUsed && exp.skillsUsed.length > 0 && (
          <div className="flex flex-wrap gap-1">
            {exp.skillsUsed.map((s) => (
              <span key={s} className="badge badge-primary badge-sm">
                {s}
              </span>
            ))}
          </div>
        )}
        {exp.industry && <p className="text-xs text-muted">Industry: {exp.industry}</p>}
        {exp.url && (
          <a href={exp.url} target="_blank" rel="noopener noreferrer" className="text-xs link link-primary">
            Evidence URL
          </a>
        )}
      </div>
    </div>
  );
}
