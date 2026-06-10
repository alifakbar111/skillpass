import { useState } from 'react';
import { Link, useNavigate, useSearchParams } from 'react-router-dom';
import { LoadingSpinner } from '../../components/ui/LoadingFallback';
import { ApiError, api } from '../../lib/api';

export function ResetPassword() {
  const [params] = useSearchParams();
  const navigate = useNavigate();
  const token = params.get('token') ?? '';

  const [password, setPassword] = useState('');
  const [confirm, setConfirm] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const onSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (loading) return;
    if (password.length < 8) {
      setError('Password must be at least 8 characters');
      return;
    }
    if (password !== confirm) {
      setError('Passwords do not match');
      return;
    }
    setLoading(true);
    setError('');
    try {
      await api('/auth/reset-password', {
        method: 'POST',
        body: JSON.stringify({ token, newPassword: password }),
      });
      navigate('/auth/login', { replace: true });
    } catch (err) {
      setError(err instanceof ApiError ? (err.serverMessage ?? err.message) : 'Reset failed');
    } finally {
      setLoading(false);
    }
  };

  if (!token) {
    return (
      <div className="max-w-sm mx-auto p-8 text-center">
        <p className="text-error mb-4">This reset link is missing its token.</p>
        <Link to="/auth/forgot-password" className="link link-primary text-sm">
          Request a new link
        </Link>
      </div>
    );
  }

  return (
    <div className="hero min-h-[60vh]">
      <div className="hero-content w-full max-w-sm">
        <div className="card bg-base-200 w-full p-6">
          <h1 className="text-2xl font-bold mb-4 text-center">Set New Password</h1>
          <form onSubmit={onSubmit} className="space-y-4">
            <label className="form-control w-full">
              <span className="label-text mb-1">New password</span>
              <input
                type="password"
                className="input input-bordered w-full"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                autoComplete="new-password"
                required
                minLength={8}
              />
            </label>
            <label className="form-control w-full">
              <span className="label-text mb-1">Confirm password</span>
              <input
                type="password"
                className="input input-bordered w-full"
                value={confirm}
                onChange={(e) => setConfirm(e.target.value)}
                autoComplete="new-password"
                required
              />
            </label>
            {error && (
              <p className="text-error text-sm" role="alert">
                {error}
              </p>
            )}
            <button type="submit" className="btn btn-primary w-full" disabled={loading}>
              {loading ? <LoadingSpinner /> : 'Update password'}
            </button>
          </form>
        </div>
      </div>
    </div>
  );
}
