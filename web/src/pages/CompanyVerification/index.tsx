import { zodResolver } from '@hookform/resolvers/zod';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { X } from 'lucide-react';
import { useState } from 'react';
import { useForm } from 'react-hook-form';
import { Link } from 'react-router-dom';
import { CompanyOnboarding } from '@/components/onboarding/CompanyOnboarding';
import { Form } from '@/components/ui/Form';
import { FormInput } from '@/components/ui/FormInput';
import { FormTextarea } from '@/components/ui/FormTextarea';
import { LoadingSpinner } from '@/components/ui/LoadingFallback';
import { ApiError, api } from '@/lib/api';
import { type VerificationForm, verificationSchema } from '@/lib/schemas';

export function CompanyVerification() {
  const queryClient = useQueryClient();
  const [queryErrorDismissed, setQueryErrorDismissed] = useState(false);

  const methods = useForm<VerificationForm>({
    resolver: zodResolver(verificationSchema),
  });

  const {
    data,
    error: queryError,
    isLoading,
  } = useQuery({
    queryKey: ['company', 'verification-status'],
    queryFn: () => api<{ verificationStatus: string }>('/company/verification-status'),
  });

  const status = data?.verificationStatus ?? null;

  const submitMutation = useMutation({
    mutationFn: (formData: VerificationForm) => api('/company/verification', { method: 'POST', body: formData }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['company', 'verification-status'] });
    },
  });

  const onSubmit = (formData: VerificationForm) => submitMutation.mutate(formData);

  const error =
    queryError && !queryErrorDismissed
      ? queryError instanceof ApiError
        ? (queryError.serverMessage ?? queryError.message)
        : 'Failed to load verification status'
      : submitMutation.error
        ? submitMutation.error instanceof ApiError
          ? (submitMutation.error.serverMessage ?? submitMutation.error.message)
          : 'Submission failed'
        : null;

  if (isLoading) {
    return (
      <div className="max-w-lg mx-auto p-4 flex justify-center">
        <LoadingSpinner />
      </div>
    );
  }

  if (error) {
    return (
      <div className="max-w-lg mx-auto p-4">
        <div className="alert alert-error">
          <span>{error}</span>
          <button
            type="button"
            title="close"
            className="btn btn-ghost btn-xs"
            onClick={() => {
              setQueryErrorDismissed(true);
              submitMutation.reset();
            }}
          >
            <X size={14} />
          </button>
        </div>
      </div>
    );
  }

  if (status === 'verified')
    return (
      <div className="max-w-lg mx-auto p-4 space-y-4">
        <div className="card bg-base-200 p-6 text-center">
          <span className="text-4xl mb-2" aria-hidden="true">
            &#10004;&#65039;
          </span>
          <h2 className="text-xl font-bold">Verified!</h2>
          <p className="text-muted-strong mb-4">Your company is verified. You can search candidates and post jobs.</p>
          <div className="flex gap-2 justify-center">
            <Link to="/company/jobs" className="btn btn-primary btn-sm">
              Post your first job
            </Link>
            <Link to="/company/search" className="btn btn-outline btn-sm">
              Browse candidates
            </Link>
          </div>
        </div>
        <CompanyOnboarding />
      </div>
    );

  if (status === 'pending')
    return (
      <div className="max-w-lg mx-auto p-4 space-y-4">
        <div className="card bg-base-200 p-6 text-center">
          <LoadingSpinner className="mb-2" />
          <h2 className="text-xl font-bold">Verification Pending</h2>
          <p className="text-muted-strong">
            Your verification details were submitted at registration. Our team reviews submissions within 48 hours and
            you'll get an email the moment you're approved.
          </p>
        </div>
        <div className="card bg-base-200 p-4">
          <h3 className="font-semibold mb-2 text-sm">While you wait</h3>
          <ul className="text-sm space-y-1 list-disc list-inside text-muted-strong">
            <li>
              <Link to="/company/profile" className="link link-primary">
                Complete your company profile
              </Link>{' '}
              — a description and website help candidates trust you
            </li>
            <li>
              <Link to="/jobs" className="link link-primary">
                Browse the job board
              </Link>{' '}
              to see how other companies present roles
            </li>
            <li>Draft your first job post — you can publish it the moment you're verified</li>
          </ul>
        </div>
        <CompanyOnboarding />
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
      <Form methods={methods} onSubmit={onSubmit} className="card bg-base-200 p-4 space-y-4">
        <FormInput label="Business Registration Number" name="businessRegistration" />
        <FormInput label="Company Website" name="website" placeholder="https://example.com" />
        <FormTextarea label="Office Address" name="address" rows={3} />
        <FormInput label="Contact Person & Title" name="contact" />
        <button type="submit" className="btn btn-primary" disabled={submitMutation.isPending}>
          {submitMutation.isPending ? <LoadingSpinner /> : 'Resubmit Verification'}
        </button>
      </Form>
    </div>
  );
}
