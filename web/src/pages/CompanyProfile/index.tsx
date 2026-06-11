import { zodResolver } from '@hookform/resolvers/zod';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { X } from 'lucide-react';
import { useEffect, useState } from 'react';
import { useForm } from 'react-hook-form';
import { FormInput, FormSelect, FormTextarea } from '../../components/ui/FormField';
import { LoadingFallback, LoadingSpinner } from '../../components/ui/LoadingFallback';
import { useIndustries } from '../../hooks/useIndustries';
import { ApiError, api } from '../../lib/api';
import { type CompanyProfileForm, companyProfileSchema } from '../../lib/schemas';

type CompanyProfileData = { companyName: string; website?: string; industry: string; description?: string };

export function CompanyProfile() {
  const queryClient = useQueryClient();
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState(false);

  const { data: industries = [] } = useIndustries();

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<CompanyProfileForm>({
    resolver: zodResolver(companyProfileSchema),
  });

  const { data: companyProfile, isLoading: loading } = useQuery({
    queryKey: ['company', 'profile'],
    queryFn: () => api<CompanyProfileData>('/company/profile'),
  });

  // Seed the form once the profile loads (react-hook-form reset moved out of .then()).
  useEffect(() => {
    if (!companyProfile) return;
    reset({
      companyName: companyProfile.companyName,
      website: companyProfile.website || '',
      industry: companyProfile.industry,
      description: companyProfile.description || '',
    });
  }, [companyProfile, reset]);

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

      <form onSubmit={handleSubmit(onSubmit)} className="card bg-base-200 p-4 space-y-4">
        <FormInput label="Company Name" registration={register('companyName')} error={errors.companyName} />
        <FormInput
          label="Website"
          registration={register('website')}
          error={errors.website}
          placeholder="https://example.com"
        />
        <FormSelect
          label="Industry"
          registration={register('industry')}
          error={errors.industry}
          options={industries.map((ind) => ({ value: ind.name, label: ind.name }))}
        />
        <FormTextarea label="Description" registration={register('description')} error={errors.description} rows={4} />
        <button type="submit" className="btn btn-primary" disabled={saveMutation.isPending}>
          {saveMutation.isPending ? <LoadingSpinner /> : 'Save'}
        </button>
      </form>
    </div>
  );
}
