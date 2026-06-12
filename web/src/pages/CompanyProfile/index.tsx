import { zodResolver } from '@hookform/resolvers/zod';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { X } from 'lucide-react';
import { useEffect, useState } from 'react';
import { useForm } from 'react-hook-form';
import { Form } from '../../components/ui/Form';
import { FormInput } from '../../components/ui/FormInput';
import { FormSelect } from '../../components/ui/FormSelect';
import { FormTextarea } from '../../components/ui/FormTextarea';
import { LoadingFallback, LoadingSpinner } from '../../components/ui/LoadingFallback';
import { useIndustries } from '../../hooks/useIndustries';
import { ApiError, api } from '../../lib/api';
import { type CompanyProfileForm, companyProfileSchema } from '../../lib/schemas';
import { WebhooksSection } from './WebhooksSection';

type CompanyProfileData = { companyName: string; website?: string; industry: string; description?: string };

export function CompanyProfile() {
  const queryClient = useQueryClient();
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState(false);
  const [blindMode, setBlindMode] = useState(false);
  const [blindSaving, setBlindSaving] = useState(false);

  const { data: industries = [] } = useIndustries();

  const methods = useForm<CompanyProfileForm>({
    resolver: zodResolver(companyProfileSchema),
  });

  const { data: companyProfile, isLoading: loading } = useQuery({
    queryKey: ['company', 'profile'],
    queryFn: () => api<CompanyProfileData>('/company/profile'),
  });

  // Seed the form once the profile loads (react-hook-form reset moved out of .then()).
  useEffect(() => {
    if (!companyProfile) return;
    methods.reset({
      companyName: companyProfile.companyName,
      website: companyProfile.website || '',
      industry: companyProfile.industry,
      description: companyProfile.description || '',
    });
    setBlindMode((companyProfile as Record<string, unknown>).blindMode === true);
  }, [companyProfile, methods]);

  const saveMutation = useMutation({
    mutationFn: (data: CompanyProfileForm) => api('/company/profile', { method: 'PUT', body: JSON.stringify(data) }),
    onMutate: () => {
      setError(null);
      setSuccess(false);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['company', 'profile'] });
      setSuccess(true);
    },
    onError: (err) => {
      setError(err instanceof ApiError ? (err.serverMessage ?? err.message) : 'Failed to save profile');
    },
  });

  const onSubmit = (data: CompanyProfileForm) => saveMutation.mutate(data);

  const toggleBlindMode = async (next: boolean) => {
    setBlindSaving(true);
    setError(null);
    const prev = blindMode;
    setBlindMode(next);
    try {
      await api('/company/profile', { method: 'PUT', body: JSON.stringify({ blindMode: next }) });
    } catch (err) {
      setBlindMode(prev);
      setError(err instanceof ApiError ? (err.serverMessage ?? err.message) : 'Failed to update blind mode');
    } finally {
      setBlindSaving(false);
    }
  };

  if (loading) return <LoadingFallback text="Loading company profile" />;

  return (
    <div className="max-w-lg mx-auto p-4">
      <h1 className="text-2xl font-bold mb-6">Company Profile</h1>

      {error && (
        <div className="alert alert-error mb-4">
          <span>{error}</span>
          <button type="button" title="close" className="btn btn-ghost btn-xs" onClick={() => setError(null)}>
            <X size={14} />
          </button>
        </div>
      )}

      {success && (
        <div className="alert alert-success mb-4" role="status">
          <span>Profile saved</span>
        </div>
      )}

      <Form methods={methods} onSubmit={onSubmit} className="card bg-base-200 p-4 space-y-4">
        <FormInput label="Company Name" name="companyName" />
        <FormInput label="Website" name="website" placeholder="https://example.com" />
        <FormSelect
          label="Industry"
          name="industry"
          options={industries
            .filter((ind): ind is typeof ind & { name: string } => ind.name != null)
            .map((ind) => ({ value: ind.name, label: ind.name }))}
        />
        <FormTextarea label="Description" name="description" rows={4} />
        <button type="submit" className="btn btn-primary" disabled={saveMutation.isPending}>
          {saveMutation.isPending ? <LoadingSpinner /> : 'Save'}
        </button>
      </Form>

      <div className="card bg-base-200 p-4 mt-4">
        <label className="flex items-start justify-between gap-4 cursor-pointer">
          <div>
            <span className="font-semibold">Blind screening</span>
            <p className="text-sm opacity-70 mt-1">
              Hide candidate names, photos, and bios in search and matches to reduce bias. Identities stay masked until
              you move a candidate forward.
            </p>
          </div>
          <input
            type="checkbox"
            className="toggle toggle-primary"
            checked={blindMode}
            disabled={blindSaving}
            onChange={(e) => toggleBlindMode(e.target.checked)}
          />
        </label>
      </div>

      <WebhooksSection />
    </div>
  );
}
