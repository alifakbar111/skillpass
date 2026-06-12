import { zodResolver } from '@hookform/resolvers/zod';
import { useState } from 'react';
import { useForm, useWatch } from 'react-hook-form';
import { Link, useNavigate } from 'react-router-dom';
import { Form } from '../../components/ui/Form';
import { FormInput } from '../../components/ui/FormInput';
import { FormTextarea } from '../../components/ui/FormTextarea';
import { LoadingSpinner } from '../../components/ui/LoadingFallback';
import { ToggleButtonGroup } from '../../components/ui/ToggleButtonGroup';
import { useAuth } from '../../hooks/useAuth';
import { ApiError } from '../../lib/api';
import { type RegisterForm, registerSchema } from '../../lib/schemas';

export function Register() {
  const { register: authRegister } = useAuth();
  const navigate = useNavigate();
  const methods = useForm<RegisterForm>({
    resolver: zodResolver(registerSchema),
    shouldUnregister: true,
    defaultValues: {
      email: '',
      username: '',
      password: '',
      role: 'jobseeker',
    },
  });
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);
  const role = useWatch({ control: methods.control, name: 'role', defaultValue: 'jobseeker' });

  const onSubmit = async (data: RegisterForm) => {
    setError('');
    setLoading(true);
    try {
      await authRegister(data);
      // Land new users on their first onboarding step instead of the homepage.
      navigate(data.role === 'company' ? '/company/verification' : '/jobseeker/profile');
    } catch (err) {
      if (err instanceof ApiError) {
        setError(err.serverMessage ?? err.message);
      } else {
        setError(err instanceof Error ? err.message : 'Registration failed');
      }
    } finally {
      setLoading(false);
    }
  };
  return (
    <div className="hero min-h-[60vh]">
      <div className="hero-content w-full max-w-xl">
        <div className="card bg-base-200 w-full p-6">
          <h1 className="text-2xl font-bold mb-6 text-center">Create Account</h1>
          <Form methods={methods} onSubmit={onSubmit} className="space-y-4">
            <ToggleButtonGroup
              name="role"
              legend="I am a&hellip;"
              options={[
                { value: 'jobseeker', label: 'Jobseeker' },
                { value: 'company', label: 'Company' },
              ]}
              aria-label="Account type"
            />
            <fieldset className="fieldset">
              <legend className="fieldset-legend">Account Details</legend>
              <div className="space-y-4">
                {role === 'company' ? (
                  <FormInput label="Company Name" name="companyName" autoComplete="organization" />
                ) : (
                  <FormInput label="Full Name" name="name" autoComplete="name" />
                )}
                <FormInput label="Username" name="username" autoComplete="username" />
                <FormInput label="Email" name="email" type="email" autoComplete="email" />
                <FormInput label="Password" name="password" type="password" autoComplete="new-password" />
              </div>
            </fieldset>
            {role === 'company' && (
              <>
                <fieldset className="fieldset">
                  <legend className="fieldset-legend">Company Verification</legend>
                  <div className="space-y-4">
                    <FormInput label="Business Registration Number" name="businessRegistration" />
                    <FormInput
                      label="Company Website"
                      name="website"
                      type="url"
                      placeholder="https://example.com"
                      autoComplete="url"
                    />
                    <FormTextarea label="Office Address" name="address" rows={3} />
                    <FormInput label="Contact Person & Title" name="contact" />
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
          </Form>
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
