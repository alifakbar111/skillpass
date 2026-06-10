import { useState } from 'react';
import { Link } from 'react-router-dom';
import { LoadingSpinner } from '../../components/ui/LoadingFallback';
import { ApiError, api } from '../../lib/api';

export function ForgotPassword() {
  const [email, setEmail] = useState('');
  const [sent, setSent] = useState(false);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const onSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (loading || !email.trim()) return;
    setLoading(true);
    setError('');
    try {
      await api('/auth/forgot-password', {
        method: 'POST',
        body: JSON.stringify({ email: email.trim() }),
      });
      setSent(true);
    } catch (err) {
      setError(err instanceof ApiError ? (err.serverMessage ?? err.message) : 'Something went wrong');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="hero min-h-[60vh]">
      <div className="hero-content w-full max-w-sm">
        <div className="card bg-base-200 w-full p-6">
          <h1 className="text-2xl font-bold mb-2 text-center">Forgot Password</h1>
          {sent ? (
            <div className="text-center space-y-4">
              <p className="text-sm text-muted-strong">
                If that email belongs to an account, a reset link is on its way. Check your inbox (and spam folder).
              </p>
              <Link to="/auth/login" className="link link-primary text-sm">
                Back to sign in
              </Link>
            </div>
          ) : (
            <>
              <p className="text-sm text-muted mb-4 text-center">
                Enter your account email and we'll send you a reset link.
              </p>
              <form onSubmit={onSubmit} className="space-y-4">
                <label className="form-control w-full">
                  <span className="label-text mb-1">Email</span>
                  <input
                    type="email"
                    className="input input-bordered w-full"
                    value={email}
                    onChange={(e) => setEmail(e.target.value)}
                    autoComplete="email"
                    required
                  />
                </label>
                {error && (
                  <p className="text-error text-sm" role="alert">
                    {error}
                  </p>
                )}
                <button type="submit" className="btn btn-primary w-full" disabled={loading}>
                  {loading ? <LoadingSpinner /> : 'Send reset link'}
                </button>
              </form>
              <p className="text-sm text-center mt-4">
                <Link to="/auth/login" className="link link-primary">
                  Back to sign in
                </Link>
              </p>
            </>
          )}
        </div>
      </div>
    </div>
  );
}
