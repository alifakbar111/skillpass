import { Check, X } from 'lucide-react';
import { useEffect, useState } from 'react';
import { api } from '../lib/api';

interface Company {
  id: string;
  companyName: string;
  website?: string;
  industry: string;
  description?: string;
  verificationDocs?: Record<string, string>;
  createdAt: string;
}

export function AdminVerifications() {
  const [pending, setPending] = useState<Company[]>([]);
  const [actionLoading, setActionLoading] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);

  const loadPending = () => {
    api<Company[]>('/admin/verifications/pending')
      .then(setPending)
      .catch((err) => setError(err instanceof Error ? err.message : 'Failed to load verifications'));
  };

  useEffect(loadPending, []);

  const handleAction = async (id: string, action: 'approve' | 'reject') => {
    setActionLoading(id);
    setError(null);
    try {
      await api(`/admin/verifications/${id}`, { method: 'POST', body: JSON.stringify({ action }) });
      setPending((prev) => prev.filter((c) => c.id !== id));
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Action failed');
    } finally {
      setActionLoading(null);
    }
  };

  return (
    <div className="max-w-3xl mx-auto p-4 space-y-4">
      <h1 className="text-2xl font-bold">Company Verifications</h1>

      {error && (
        <div className="alert alert-error">
          <span>{error}</span>
          <button type="button" title="close" className="btn btn-ghost btn-xs" onClick={() => setError(null)}>
            <X size={14} />
          </button>
        </div>
      )}

      {pending.length === 0 ? (
        <div className="card bg-base-200 p-8 text-center">
          <p className="opacity-50">No pending verifications</p>
        </div>
      ) : (
        pending.map((company) => (
          <div key={company.id} className="card bg-base-200 p-4">
            <div className="flex justify-between items-start">
              <div>
                <h3 className="font-semibold">{company.companyName}</h3>
                <p className="text-sm opacity-70">
                  {company.industry} {company.website ? `· ${company.website}` : ''}
                </p>
                {company.description && <p className="text-sm mt-1">{company.description}</p>}
                {company.verificationDocs && (
                  <div className="mt-2 p-2 bg-base-100 rounded-box text-sm">
                    {Object.entries(company.verificationDocs as Record<string, string>).map(([key, val]) => (
                      <p key={key}>
                        <span className="font-medium">{key}:</span> {val}
                      </p>
                    ))}
                  </div>
                )}
              </div>
              <div className="flex gap-2">
                <button
                  type="button"
                  className="btn btn-success btn-sm"
                  disabled={actionLoading === company.id}
                  onClick={() => handleAction(company.id, 'approve')}
                >
                  <Check size={16} aria-hidden="true" /> Approve
                </button>
                <button
                  type="button"
                  className="btn btn-error btn-sm"
                  disabled={actionLoading === company.id}
                  onClick={() => handleAction(company.id, 'reject')}
                >
                  <X size={16} aria-hidden="true" /> Reject
                </button>
              </div>
            </div>
          </div>
        ))
      )}
    </div>
  );
}
