import { CheckCircle, XCircle } from 'lucide-react';
import { useEffect, useState } from 'react';
import { Link, useSearchParams } from 'react-router-dom';
import { LoadingFallback } from '../../components/ui/LoadingFallback';
import { useAuth } from '../../hooks/useAuth';
import { ApiError, api } from '../../lib/api';

export function VerifyEmail() {
  const [params] = useSearchParams();
  const { user, refreshUser } = useAuth();
  const token = params.get('token') ?? '';
  const [state, setState] = useState<'verifying' | 'done' | 'error'>('verifying');
  const [message, setMessage] = useState('');

  useEffect(() => {
    if (!token) {
      setState('error');
      setMessage('This verification link is missing its token.');
      return;
    }
    let cancelled = false;
    api(`/auth/verify-email?token=${encodeURIComponent(token)}`)
      .then(() => {
        if (cancelled) return;
        setState('done');
        // Refresh auth state so the "verify your email" banner disappears.
        refreshUser();
      })
      .catch((err) => {
        if (cancelled) return;
        setState('error');
        setMessage(
          err instanceof ApiError ? (err.serverMessage ?? err.message) : 'Verification failed. Please try again.',
        );
      });
    return () => {
      cancelled = true;
    };
  }, [token, refreshUser]);

  if (state === 'verifying') return <LoadingFallback text="Verifying your email" />;

  return (
    <div className="max-w-sm mx-auto p-8">
      <div className="card bg-base-200 p-6 text-center space-y-3">
        {state === 'done' ? (
          <>
            <CheckCircle size={40} className="text-success mx-auto" aria-hidden="true" />
            <h1 className="text-xl font-bold">Email verified!</h1>
            <p className="text-sm text-muted-strong">Your account is confirmed. Welcome aboard.</p>
            <Link to={user ? '/' : '/auth/login'} className="btn btn-primary btn-sm">
              {user ? 'Continue' : 'Sign in'}
            </Link>
          </>
        ) : (
          <>
            <XCircle size={40} className="text-error mx-auto" aria-hidden="true" />
            <h1 className="text-xl font-bold">Verification failed</h1>
            <p className="text-sm text-muted-strong">{message}</p>
            {user && !user.isVerified && (
              <p className="text-xs text-muted">You can request a new link from the banner after signing in.</p>
            )}
          </>
        )}
      </div>
    </div>
  );
}
