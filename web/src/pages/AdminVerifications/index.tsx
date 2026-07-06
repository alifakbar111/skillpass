import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Building2, Check, FileText, Globe, MapPin, User, X } from 'lucide-react';
import { useState } from 'react';
import { ApiError, api } from '@/lib/api';
import type { PendingCompany as Company } from '@/lib/api-types';

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
        body: { action },
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
        <div className="flex flex-row justify-between alert alert-error">
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
                <div className="space-y-4">
                  <div className="flex flex-row justify-between">
                    <div>
                      <h3 className="font-semibold">{company.companyName}</h3>
                      <p className="text-sm opacity-70">
                        {company.industry} {company.website ? `· ${company.website}` : ''}
                      </p>
                    </div>
                    <div>
                      <div className="flex gap-2">
                        <button
                          type="button"
                          className="btn btn-success btn-sm"
                          disabled={loadingId === company.id}
                          onClick={() => company.id && handleAction(company.id, 'approve')}
                        >
                          <Check size={16} aria-hidden="true" /> Approve
                        </button>
                        <button
                          type="button"
                          className="btn btn-error btn-sm"
                          disabled={loadingId === company.id}
                          onClick={() => company.id && handleAction(company.id, 'reject')}
                        >
                          <X size={16} aria-hidden="true" /> Reject
                        </button>
                      </div>
                    </div>
                  </div>
                  {company.description && <p className="text-sm mt-1">{company.description}</p>}
                  {company.verificationDocs &&
                    (() => {
                      const docs = company.verificationDocs as unknown as Record<string, string>;
                      return (
                        <div className="mt-3 space-y-2 p-3 bg-base-100 rounded-box">
                          <p className="text-xs font-medium opacity-50 flex items-center gap-1 mb-2">
                            <FileText size={12} aria-hidden="true" />
                            Verification Documents
                          </p>
                          <dl className="space-y-2">
                            <div className="flex items-start gap-2">
                              <dt className="flex items-center gap-1.5 text-xs font-medium opacity-70 w-36 shrink-0">
                                <Building2 size={14} aria-hidden="true" />
                                Business Registration
                              </dt>
                              <dd className="text-sm">{docs.businessRegistration ?? '—'}</dd>
                            </div>
                            <div className="flex items-start gap-2">
                              <dt className="flex items-center gap-1.5 text-xs font-medium opacity-70 w-36 shrink-0">
                                <Globe size={14} aria-hidden="true" />
                                Website
                              </dt>
                              <dd className="text-sm">
                                {docs.website ? (
                                  <a
                                    href={docs.website}
                                    className="link link-primary"
                                    target="_blank"
                                    rel="noopener noreferrer"
                                  >
                                    {docs.website}
                                  </a>
                                ) : (
                                  '—'
                                )}
                              </dd>
                            </div>
                            <div className="flex items-start gap-2">
                              <dt className="flex items-center gap-1.5 text-xs font-medium opacity-70 w-36 shrink-0">
                                <MapPin size={14} aria-hidden="true" />
                                Address
                              </dt>
                              <dd className="text-sm whitespace-pre-line">{docs.address ?? '—'}</dd>
                            </div>
                            <div className="flex items-start gap-2">
                              <dt className="flex items-center gap-1.5 text-xs font-medium opacity-70 w-36 shrink-0">
                                <User size={14} aria-hidden="true" />
                                Contact Person
                              </dt>
                              <dd className="text-sm">{docs.contact ?? '—'}</dd>
                            </div>
                          </dl>
                        </div>
                      );
                    })()}
                </div>
              </div>
            </div>
          ))}
    </div>
  );
}
