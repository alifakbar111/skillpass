import { zodResolver } from '@hookform/resolvers/zod';
import { X } from 'lucide-react';
import { useEffect, useState } from 'react';
import { useForm } from 'react-hook-form';
import { FormInput, FormSelect, FormTextarea } from '../../components/ui/FormField';
import { LoadingFallback, LoadingSpinner } from '../../components/ui/LoadingFallback';
import { ApiError, api } from '../../lib/api';
import { type CompanyProfileForm, companyProfileSchema } from '../../lib/schemas';

export function CompanyProfile() {
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [industries, setIndustries] = useState<Array<{ id: string; name: string }>>([]);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState(false);

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<CompanyProfileForm>({
    resolver: zodResolver(companyProfileSchema),
  });

  useEffect(() => {
    let cancelled = false;
    api<Array<{ id: string; name: string }>>('/industries')
      .then((data) => {
        if (!cancelled) setIndustries(data);
      })
      .catch(() => {});
    api<{ companyName: string; website?: string; industry: string; description?: string }>('/company/profile')
      .then((data) => {
        if (cancelled) return;
        reset({
          companyName: data.companyName,
          website: data.website || '',
          industry: data.industry,
          description: data.description || '',
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
  }, [reset]);

  const onSubmit = async (data: CompanyProfileForm) => {
    setSaving(true);
    setError(null);
    setSuccess(false);
    try {
      await api('/company/profile', { method: 'PUT', body: JSON.stringify(data) });
      setSuccess(true);
    } catch (err) {
      setError(err instanceof ApiError ? (err.serverMessage ?? err.message) : 'Failed to save profile');
    } finally {
      setSaving(false);
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
        <button type="submit" className="btn btn-primary" disabled={saving}>
          {saving ? <LoadingSpinner /> : 'Save'}
        </button>
      </form>
    </div>
  );
}
