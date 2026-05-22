import { useState, type FormEvent } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { useAuth } from '../hooks/useAuth';

export function Register() {
  const { register } = useAuth();
  const navigate = useNavigate();
  const [form, setForm] = useState({ email: '', username: '', password: '', name: '', role: 'jobseeker' as const });
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setError('');
    setLoading(true);
    try {
      await register(form);
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
          <form onSubmit={handleSubmit} className="space-y-4">
            <div className="flex gap-2">
              <button type="button" className={`btn flex-1 ${form.role === 'jobseeker' ? 'btn-primary' : 'btn-outline'}`}
                onClick={() => setForm({ ...form, role: 'jobseeker' })}>Jobseeker</button>
              <button type="button" className={`btn flex-1 ${form.role === 'company' ? 'btn-primary' : 'btn-outline'}`}
                onClick={() => setForm({ ...form, role: 'company' })}>Company</button>
            </div>
            <label className="form-control">
              <span className="label-text">Full Name</span>
              <input className="input input-bordered" value={form.name}
                onChange={(e) => setForm({ ...form, name: e.target.value })} required />
            </label>
            <label className="form-control">
              <span className="label-text">Username</span>
              <input className="input input-bordered" value={form.username}
                onChange={(e) => setForm({ ...form, username: e.target.value })} required minLength={3} />
            </label>
            <label className="form-control">
              <span className="label-text">Email</span>
              <input type="email" className="input input-bordered" value={form.email}
                onChange={(e) => setForm({ ...form, email: e.target.value })} required />
            </label>
            <label className="form-control">
              <span className="label-text">Password</span>
              <input type="password" className="input input-bordered" value={form.password}
                onChange={(e) => setForm({ ...form, password: e.target.value })} required minLength={6} />
            </label>
            {error && <p className="text-error text-sm">{error}</p>}
            <button type="submit" className="btn btn-primary w-full" disabled={loading}>
              {loading ? <span className="loading loading-spinner" /> : 'Create Account'}
            </button>
          </form>
          <p className="text-sm text-center mt-4">
            Already have an account? <Link to="/auth/login" className="link link-primary">Sign In</Link>
          </p>
        </div>
      </div>
    </div>
  );
}
