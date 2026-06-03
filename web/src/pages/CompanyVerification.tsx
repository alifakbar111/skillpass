import { zodResolver } from '@hookform/resolvers/zod';
import { X } from 'lucide-react';
import { useEffect, useState } from 'react';
import { useForm } from 'react-hook-form';
import { FormInput, FormTextarea } from '../components/ui/FormField';
import { LoadingSpinner } from '../components/ui/LoadingFallback';
import { api } from '../lib/api';
import { type VerificationForm, verificationSchema } from '../lib/schemas';

export function CompanyVerification() {
  const [status, setStatus] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<VerificationForm>({
    resolver: zodResolver(verificationSchema),
  });

  useEffect(() => {
    api<{ verificationStatus: string }>('/company/verification-status')
      .then((data) => setStatus(data.verificationStatus))
      .catch((err) => setError(err instanceof Error ? err.message : 'Failed to load verification status'));
  }, []);

  const onSubmit = async (data: VerificationForm) => {
    setSubmitting(true);
    setError(null);
    try {
      await api('/company/verification', { method: 'POST', body: JSON.stringify(data) });
      setStatus('pending');
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Submission failed');
    } finally {
      setSubmitting(false);
    }
  };

  if (error) {
    return (
      <div className="max-w-lg mx-auto p-4">
        <div className="alert alert-error">
          <span>{error}</span>
          <button type="button" title="close" className="btn btn-ghost btn-xs" onClick={() => setError(null)}>
            <X size={14} />
          </button>
        </div>
      </div>
    );
  }

  if (status === 'verified')
    return (
      <div className="max-w-lg mx-auto p-4 text-center">
        <div className="card bg-base-200 p-6">
          <span className="text-4xl mb-2" aria-hidden="true">
            &#10004;&#65039;
          </span>
          <h2 className="text-xl font-bold">Verified!</h2>
          <p className="text-muted-strong">Your company is verified. You can search candidates and post jobs.</p>
        </div>
      </div>
    );

  if (status === 'pending')
    return (
      <div className="max-w-lg mx-auto p-4 text-center">
        <div className="card bg-base-200 p-6">
          <LoadingSpinner className="mb-2" />
          <h2 className="text-xl font-bold">Verification Pending</h2>
          <p className="text-muted-strong">
            Your verification details were submitted at registration. We're reviewing them now. Check back soon.
          </p>
        </div>
      </div>
    );

  return (
    <div className="max-w-lg mx-auto p-4">
      <h1 className="text-2xl font-bold mb-2">Company Verification</h1>
      <p className="text-muted-strong mb-4">
        {status === 'rejected'
          ? 'Your verification was rejected. Please review and resubmit your details below.'
          : 'Your verification details are on file. You can update and resubmit them if needed.'}
      </p>
      <form onSubmit={handleSubmit(onSubmit)} className="card bg-base-200 p-4 space-y-4">
        <FormInput
          label="Business Registration Number"
          registration={register('businessRegistration')}
          error={errors.businessRegistration}
        />
        <FormInput
          label="Company Website"
          registration={register('website')}
          error={errors.website}
          placeholder="https://example.com"
        />
        <FormTextarea label="Office Address" registration={register('address')} error={errors.address} rows={3} />
        <FormInput label="Contact Person & Title" registration={register('contact')} error={errors.contact} />
        <button type="submit" className="btn btn-primary" disabled={submitting}>
          {submitting ? <LoadingSpinner /> : 'Resubmit Verification'}
        </button>
      </form>
    </div>
  );
}
