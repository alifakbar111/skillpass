import { useState } from 'react';
import { useForm } from 'react-hook-form';
import { Link, useNavigate } from 'react-router-dom';
import { LoadingSpinner } from '../components/ui/LoadingFallback';
import { useAuth } from '../hooks/useAuth';

type RegisterForm = {
  email: string;
  username: string;
  password: string;
  name: string;
  role: 'jobseeker' | 'company';
};

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
    defaultValues: { email: '', username: '', password: '', name: '', role: 'jobseeker' },
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
            <fieldset className="border border-base-300 rounded-lg p-4" aria-label="Account type">
              <legend className="font-semibold px-1">I am a&hellip;</legend>
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
            <fieldset className="border border-base-300 rounded-lg p-4">
              <legend className="font-semibold px-1">Account Details</legend>
              <div className="space-y-4">
                <label className="form-control w-full">
                  <div className="label-text mb-1">Full Name</div>
                  <input
                    className="input input-bordered w-full"
                    autoComplete="name"
                    {...register('name', { required: 'Name is required' })}
                    aria-describedby={errors.name ? 'register-name-error' : undefined}
                  />
                  {errors.name && (
                    <span className="text-error text-xs mt-1" id="register-name-error" role="alert">
                      {errors.name.message}
                    </span>
                  )}
                </label>
                <label className="form-control w-full">
                  <div className="label-text mb-1">Username</div>
                  <input
                    className="input input-bordered w-full"
                    autoComplete="username"
                    {...register('username', {
                      required: 'Username is required',
                      minLength: { value: 3, message: 'At least 3 characters' },
                    })}
                    aria-describedby={errors.username ? 'register-username-error' : undefined}
                  />
                  {errors.username && (
                    <span className="text-error text-xs mt-1" id="register-username-error" role="alert">
                      {errors.username.message}
                    </span>
                  )}
                </label>
                <label className="form-control w-full">
                  <div className="label-text mb-1">Email</div>
                  <input
                    type="email"
                    className="input input-bordered w-full"
                    autoComplete="email"
                    {...register('email', {
                      required: 'Email is required',
                      pattern: { value: /^\S+@\S+$/, message: 'Invalid email' },
                    })}
                    aria-describedby={errors.email ? 'register-email-error' : undefined}
                  />
                  {errors.email && (
                    <span className="text-error text-xs mt-1" id="register-email-error" role="alert">
                      {errors.email.message}
                    </span>
                  )}
                </label>
                <label className="form-control w-full">
                  <div className="label-text mb-1">Password</div>
                  <input
                    type="password"
                    className="input input-bordered w-full"
                    autoComplete="new-password"
                    {...register('password', {
                      required: 'Password is required',
                      minLength: { value: 6, message: 'At least 6 characters' },
                    })}
                    aria-describedby={errors.password ? 'register-password-error' : undefined}
                  />
                  {errors.password && (
                    <span className="text-error text-xs mt-1" id="register-password-error" role="alert">
                      {errors.password.message}
                    </span>
                  )}
                </label>
              </div>
            </fieldset>
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
