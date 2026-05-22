import { useState, useEffect, type FormEvent } from 'react';
import { api } from '../lib/api';
import { useAuth } from '../hooks/useAuth';

export function CompanyVerification() {
  const { user } = useAuth();
  const [status, setStatus] = useState<string | null>(null);
  const [form, setForm] = useState({ businessRegistration: '', website: '', address: '', contact: '' });
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    api<{ verificationStatus: string }>('/company/verification-status')
      .then(data => setStatus(data.verificationStatus));
  }, []);

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setSubmitting(true);
    await api('/company/verification', { method: 'POST', body: JSON.stringify(form) });
    setStatus('pending');
    setSubmitting(false);
  };

  if (!user || user.role !== 'company') return <div className="text-center p-8 text-error">Access denied</div>;

  if (status === 'verified') return (
    <div className="max-w-lg mx-auto p-4 text-center">
      <div className="card bg-base-200 p-6">
        <span className="text-4xl mb-2">✅</span>
        <h2 className="text-xl font-bold">Verified!</h2>
        <p className="opacity-70">Your company is verified. You can search candidates and post jobs.</p>
      </div>
    </div>
  );

  if (status === 'pending') return (
    <div className="max-w-lg mx-auto p-4 text-center">
      <div className="card bg-base-200 p-6">
        <span className="loading loading-spinner loading-lg mb-2" />
        <h2 className="text-xl font-bold">Verification Pending</h2>
        <p className="opacity-70">We're reviewing your documents. Check back soon.</p>
      </div>
    </div>
  );

  return (
    <div className="max-w-lg mx-auto p-4">
      <h1 className="text-2xl font-bold mb-6">Verify Your Company</h1>
      <p className="opacity-70 mb-4">Submit your business details to get verified. Verified companies can search candidates and post jobs.</p>
      <form onSubmit={handleSubmit} className="card bg-base-200 p-4 space-y-4">
        <label className="form-control">
          <span className="label-text">Business Registration Number</span>
          <input className="input input-bordered" value={form.businessRegistration}
            onChange={e => setForm({ ...form, businessRegistration: e.target.value })} required />
        </label>
        <label className="form-control">
          <span className="label-text">Company Website</span>
          <input className="input input-bordered" value={form.website}
            onChange={e => setForm({ ...form, website: e.target.value })} required />
        </label>
        <label className="form-control">
          <span className="label-text">Office Address</span>
          <textarea className="textarea textarea-bordered" value={form.address}
            onChange={e => setForm({ ...form, address: e.target.value })} required />
        </label>
        <label className="form-control">
          <span className="label-text">Contact Person & Title</span>
          <input className="input input-bordered" value={form.contact}
            onChange={e => setForm({ ...form, contact: e.target.value })} required />
        </label>
        <button type="submit" className="btn btn-primary" disabled={submitting}>
          {submitting ? <span className="loading loading-spinner" /> : 'Submit Verification'}
        </button>
      </form>
    </div>
  );
}
