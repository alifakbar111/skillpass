import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Check, X } from 'lucide-react';
import { useState } from 'react';
import { ApiError, api } from '../../lib/api';
import type { Company } from './type';

export function AdminVerifications() {
  const queryClient = useQueryClient();
  const [error, setError] = useState<string | null>(null);
  const [loadingId, setLoadingId] = useState<string | null>(null);

  const { data: pending = [], error: loadError } = useQuery({
    queryKey: ['admin', 'verifications', 'pending'],
    queryFn: () => api<Company[]>('/admin/verifications/pending'),
  });

  const loadErrorMessage = loadError
    ? loadError instanceof ApiError
      ? (loadError.serverMessage ?? loadError.message)
      : 'Failed to load pending verifications'
    : null;

  const actionMutation = useMutation({
    mutationFn: ({ id, action }: { id: string; action: 'approve' | 'reject' }) =>
      api(`/admin/verifications/${encodeURIComponent(id)}`, {
        method: 'POST',
        body: JSON.stringify({ action }),
      }),
    onMutate: ({ id }) => {
      setError(null);
      setLoadingId(id);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin', 'verifications', 'pending'] });
    },
    onError: (err) => {
      setError(err instanceof ApiError ? (err.serverMessage ?? err.message) : 'Action failed');
    },
    onSettled: () => setLoadingId(null),
  });

  const handleAction = (id: string, action: 'approve' | 'reject') => actionMutation.mutate({ id, action });

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

      {loadErrorMessage && (
        <div className="alert alert-error">
          <span>{loadErrorMessage}</span>
        </div>
      )}

      {pending.length === 0
        ? !loadErrorMessage && (
            <div className="card bg-base-200 p-8 text-center">
              <p className="opacity-50">No pending verifications</p>
            </div>
          )
        : pending.map((company) => (
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
                    disabled={loadingId === company.id}
                    onClick={() => handleAction(company.id, 'approve')}
                  >
                    <Check size={16} aria-hidden="true" /> Approve
                  </button>
                  <button
                    type="button"
                    className="btn btn-error btn-sm"
                    disabled={loadingId === company.id}
                    onClick={() => handleAction(company.id, 'reject')}
                  >
                    <X size={16} aria-hidden="true" /> Reject
                  </button>
                </div>
              </div>
            </div>
          ))}
    </div>
  );
}
