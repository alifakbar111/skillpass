import { zodResolver } from '@hookform/resolvers/zod';
import { useState } from 'react';
import { useForm } from 'react-hook-form';
import { Link, useNavigate } from 'react-router-dom';
import { FormInput, FormTextarea } from '../components/ui/FormField';
import { LoadingSpinner } from '../components/ui/LoadingFallback';
import { useAuth } from '../hooks/useAuth';
import { type RegisterForm, registerSchema } from '../lib/schemas';

export function Register() {
  const { register: authRegister } = useAuth();
  const navigate = useNavigate();
  const {
    register,
    handleSubmit,
    watch,
    setValue,
    formState: { errors },
  } = useForm<RegisterForm>({
    resolver: zodResolver(registerSchema),
    defaultValues: {
      email: '',
      username: '',
      password: '',
      name: '',
      role: 'jobseeker',
      companyName: '',
      businessRegistration: '',
      website: '',
      address: '',
      contact: '',
    },
  });
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);
  const role = watch('role');

  const onSubmit = async (data: RegisterForm) => {
    setError('');
    setLoading(true);
    try {
      await authRegister(data);
      navigate('/');
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Registration failed');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="hero min-h-[60vh]">
      <div className="hero-content w-full max-w-sm">
        <div className="card bg-base-200 w-full p-6">
          <h1 className="text-2xl font-bold mb-6 text-center">Create Account</h1>
          <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
            <fieldset className="fieldset" aria-label="Account type">
              <legend className="fieldset-legend">I am a&hellip;</legend>
              <div className="flex gap-2">
                <button
                  type="button"
                  className={`btn flex-1 ${role === 'jobseeker' ? 'btn-primary' : 'btn-outline'}`}
                  onClick={() => setValue('role', 'jobseeker')}
                  aria-label="Jobseeker"
                >
                  Jobseeker
                </button>
                <button
                  type="button"
                  className={`btn flex-1 ${role === 'company' ? 'btn-primary' : 'btn-outline'}`}
                  onClick={() => setValue('role', 'company')}
                  aria-label="Company"
                >
                  Company
                </button>
              </div>
            </fieldset>
            <fieldset className="fieldset">
              <legend className="fieldset-legend">Account Details</legend>
              <div className="space-y-4">
                <FormInput
                  label={role === 'company' ? 'Company Name' : 'Full Name'}
                  registration={role === 'company' ? register('companyName') : register('name')}
                  error={errors.companyName || errors.name}
                  autoComplete={role === 'company' ? 'organization' : 'name'}
                />
                <FormInput
                  label="Username"
                  registration={register('username')}
                  error={errors.username}
                  autoComplete="username"
                />
                <FormInput
                  label="Email"
                  registration={register('email')}
                  error={errors.email}
                  type="email"
                  autoComplete="email"
                />
                <FormInput
                  label="Password"
                  registration={register('password')}
                  error={errors.password}
                  type="password"
                  autoComplete="new-password"
                />
              </div>
            </fieldset>
            {role === 'company' && (
              <>
                <fieldset className="fieldset">
                  <legend className="fieldset-legend">Company Verification</legend>
                  <div className="space-y-4">
                    <FormInput
                      label="Business Registration Number"
                      registration={register('businessRegistration')}
                      error={errors.businessRegistration}
                    />
                    <FormInput
                      label="Company Website"
                      registration={register('website')}
                      error={errors.website}
                      type="url"
                      placeholder="https://example.com"
                      autoComplete="url"
                    />
                    <FormTextarea
                      label="Office Address"
                      registration={register('address')}
                      error={errors.address}
                      rows={3}
                    />
                    <FormInput
                      label="Contact Person & Title"
                      registration={register('contact')}
                      error={errors.contact}
                    />
                  </div>
                </fieldset>
                <div className="alert alert-soft text-sm" role="note">
                  <span className="icon-[tabler--info-circle] size-5" aria-hidden="true" />
                  <span>Verification is required before you can post jobs. We review submissions within 48 hours.</span>
                </div>
              </>
            )}
            {error && (
              <p className="text-error text-sm" role="alert">
                {error}
              </p>
            )}
            <button type="submit" className="btn btn-primary w-full" disabled={loading}>
              {loading ? <LoadingSpinner /> : 'Create Account'}
            </button>
          </form>
          <p className="text-sm text-center mt-4">
            Already have an account?{' '}
            <Link to="/auth/login" className="link link-primary">
              Sign In
            </Link>
          </p>
        </div>
      </div>
    </div>
  );
}
