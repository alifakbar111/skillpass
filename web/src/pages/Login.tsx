import { useState, type FormEvent } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { useAuth } from '../hooks/useAuth';

export function Login() {
  const { login } = useAuth();
  const navigate = useNavigate();
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setError('');
    setLoading(true);
    try {
      await login(email, password);
      navigate('/');
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Login failed');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="hero min-h-[60vh]">
      <div className="hero-content w-full max-w-sm">
        <div className="card bg-base-200 w-full p-6">
          <h2 className="text-2xl font-bold mb-6 text-center">Sign In</h2>
          <form onSubmit={handleSubmit} className="space-y-4">
            <label className="form-control w-full">
              <span className="label-text">Email</span>
              <input type="email" className="input input-bordered w-full" value={email}
                onChange={(e) => setEmail(e.target.value)} required />
            </label>
            <label className="form-control w-full">
              <span className="label-text">Password</span>
              <input type="password" className="input input-bordered w-full" value={password}
                onChange={(e) => setPassword(e.target.value)} required />
            </label>
            {error && <p className="text-error text-sm">{error}</p>}
            <button type="submit" className="btn btn-primary w-full" disabled={loading}>
              {loading ? <span className="loading loading-spinner" /> : 'Sign In'}
            </button>
          </form>
          <p className="text-sm text-center mt-4">
            Don't have an account? <Link to="/auth/register" className="link link-primary">Register</Link>
          </p>
        </div>
      </div>
    </div>
  );
}
