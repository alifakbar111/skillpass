import { zodResolver } from '@hookform/resolvers/zod';
import { useState } from 'react';
import { useForm } from 'react-hook-form';
import { Link, useNavigate } from 'react-router-dom';
import { FormInput } from '../../components/ui/FormField';
import { LoadingSpinner } from '../../components/ui/LoadingFallback';
import { useAuth } from '../../hooks/useAuth';
import { ApiError } from '../../lib/api';
import { type LoginForm, loginSchema } from '../../lib/schemas';

export function Login() {
  const { login } = useAuth();
  const navigate = useNavigate();
  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<LoginForm>({ resolver: zodResolver(loginSchema) });
  const [serverError, setServerError] = useState('');
  const [loading, setLoading] = useState(false);

  const onSubmit = async (data: LoginForm) => {
    setServerError('');
    setLoading(true);
    try {
      const user = await login(data.email, data.password);
      if (user.role === 'admin') {
        navigate('/admin/verifications');
      } else {
        navigate('/');
      }
    } catch (err) {
      if (err instanceof ApiError) {
        setServerError(err.serverMessage ?? err.message);
      } else {
        setServerError(err instanceof Error ? err.message : 'Login failed');
      }
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="hero min-h-[60vh]">
      <div className="hero-content w-full max-w-sm">
        <div className="card bg-base-200 w-full p-6">
          <h1 className="text-2xl font-bold mb-6 text-center">Sign In</h1>
          <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
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
              autoComplete="current-password"
            />
            {serverError && (
              <p className="text-error text-sm" id="login-error" role="alert">
                {serverError}
              </p>
            )}
            <button
              type="submit"
              className="btn btn-primary w-full"
              disabled={loading}
              aria-describedby={serverError ? 'login-error' : undefined}
            >
              {loading ? <LoadingSpinner /> : 'Sign In'}
            </button>
          </form>
          <p className="text-sm text-center mt-4">
            Don't have an account?{' '}
            <Link to="/auth/register" className="link link-primary">
              Register
            </Link>
          </p>
        </div>
      </div>
    </div>
  );
}
