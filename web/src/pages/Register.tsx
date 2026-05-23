import { useState } from 'react';
import { useForm } from 'react-hook-form';
import { Link, useNavigate } from 'react-router-dom';
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
          <h2 className="text-2xl font-bold mb-6 text-center">Create Account</h2>
          <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
            <div className="flex gap-2">
              <button
                type="button"
                className={`btn flex-1 ${role === 'jobseeker' ? 'btn-primary' : 'btn-outline'}`}
                onClick={() => setValue('role', 'jobseeker')}
              >
                Jobseeker
              </button>
              <button
                type="button"
                className={`btn flex-1 ${role === 'company' ? 'btn-primary' : 'btn-outline'}`}
                onClick={() => setValue('role', 'company')}
              >
                Company
              </button>
            </div>
            <div>
              <label className="form-control">
                <span className="label-text">Full Name</span>
                <input className="input input-bordered" {...register('name', { required: 'Name is required' })} />
                {errors.name && <span className="text-error text-xs mt-1">{errors.name.message}</span>}
              </label>
              <label className="form-control">
                <span className="label-text">Username</span>
                <input
                  className="input input-bordered"
                  {...register('username', {
                    required: 'Username is required',
                    minLength: { value: 3, message: 'At least 3 characters' },
                  })}
                />
                {errors.username && <span className="text-error text-xs mt-1">{errors.username.message}</span>}
              </label>
              <label className="form-control">
                <span className="label-text">Email</span>
                <input
                  type="email"
                  className="input input-bordered"
                  {...register('email', {
                    required: 'Email is required',
                    pattern: { value: /^\S+@\S+$/, message: 'Invalid email' },
                  })}
                />
                {errors.email && <span className="text-error text-xs mt-1">{errors.email.message}</span>}
              </label>
              <label className="form-control">
                <span className="label-text">Password</span>
                <input
                  type="password"
                  className="input input-bordered"
                  {...register('password', {
                    required: 'Password is required',
                    minLength: { value: 6, message: 'At least 6 characters' },
                  })}
                />
                {errors.password && <span className="text-error text-xs mt-1">{errors.password.message}</span>}
              </label>
            </div>
            {error && <p className="text-error text-sm">{error}</p>}
            <button type="submit" className="btn btn-primary w-full" disabled={loading}>
              {loading ? <span className="loading loading-spinner" /> : 'Create Account'}
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
