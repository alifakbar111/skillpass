import { ChevronDown, ExternalLink, MessageSquare, User } from 'lucide-react';
import { Fragment, useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { LoadingFallback } from '../../components/ui/LoadingFallback';
import type { CompanyApplicationResult as CompanyApplication } from '@/lib/api-types';
import { ApiError, api } from '../../lib/api';
import { ApplicationNotes } from './ApplicationNotes';

const STATUS_OPTIONS = ['reviewed', 'interviewed', 'offered', 'rejected'] as const;

const STATUS_BADGE: Record<string, string> = {
  applied: 'badge-ghost',
  reviewed: 'badge-info',
  interviewed: 'badge-warning',
  offered: 'badge-success',
  rejected: 'badge-error',
};

export function CompanyApplications() {
  const [applications, setApplications] = useState<CompanyApplication[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [updating, setUpdating] = useState<string | null>(null);
  const [notesAppId, setNotesAppId] = useState<string | null>(null);

  useEffect(() => {
    api<CompanyApplication[]>('/company/applications')
      .then(setApplications)
      .catch((err) => {
        setError(err instanceof ApiError ? (err.serverMessage ?? err.message) : 'Failed to load applications');
      })
      .finally(() => setLoading(false));
  }, []);

  async function handleStatusChange(applicationId: string, newStatus: string) {
    setUpdating(applicationId);
    try {
      const updated = await api<CompanyApplication>(`/applications/${applicationId}/status`, {
        method: 'PUT',
        body: JSON.stringify({ status: newStatus }),
      });
      setApplications((prev) =>
        prev.map((a) => (a.id === applicationId ? { ...a, status: updated.status, updatedAt: updated.updatedAt } : a)),
      );
    } catch (err) {
      const msg = err instanceof ApiError ? (err.serverMessage ?? err.message) : 'Failed to update status';
      setError(msg);
    } finally {
      setUpdating(null);
    }
  }

  if (loading) return <LoadingFallback text="Loading applications" />;

  const grouped = applications.reduce<Record<string, CompanyApplication[]>>((acc, app) => {
    const key = app.jobTitle ?? 'Untitled';
    if (!acc[key]) acc[key] = [];
    acc[key].push(app);
    return acc;
  }, {});

  return (
    <div className="max-w-4xl mx-auto p-4">
      <h1 className="text-2xl font-bold mb-6">Applications</h1>

      {error && (
        <div className="alert alert-error mb-4">
          <span>{error}</span>
          <button type="button" className="btn btn-sm btn-ghost" onClick={() => setError(null)}>
            Dismiss
          </button>
        </div>
      )}

      {applications.length === 0 && (
        <div className="text-center py-12 opacity-60">
          <p className="text-lg">No applications yet.</p>
          <p className="text-sm mt-1">Applications will appear here when jobseekers apply to your jobs.</p>
        </div>
      )}

      {Object.entries(grouped).map(([jobTitle, apps]) => (
        <div key={jobTitle} className="mb-8">
          <h2 className="text-lg font-semibold mb-3 flex items-center gap-2">
            {jobTitle}
            <span className="badge badge-sm badge-neutral">{apps.length}</span>
          </h2>

          <div className="overflow-x-auto">
            <table className="table table-zebra w-full">
              <thead>
                <tr>
                  <th>Candidate</th>
                  <th>Status</th>
                  <th>Applied</th>
                  <th>Actions</th>
                </tr>
              </thead>
              <tbody>
                {apps.map((app) => (
                  <Fragment key={app.id}>
                    <tr>
                      <td>
                        <div className="flex items-center gap-3">
                          <div className="avatar placeholder">
                            <div className="bg-neutral text-neutral-content w-8 rounded-full">
                              <User size={16} />
                            </div>
                          </div>
                          <div>
                            <div className="font-medium">{app.candidateName}</div>
                            {app.candidateHeadline && <div className="text-xs opacity-60">{app.candidateHeadline}</div>}
                          </div>
                        </div>
                      </td>
                      <td>
                        <span className={`badge ${STATUS_BADGE[app.status ?? ''] ?? 'badge-ghost'}`}>{app.status}</span>
                      </td>
                      <td className="text-sm opacity-70">{app.createdAt?.slice(0, 10)}</td>
                      <td>
                        <div className="flex items-center gap-2">
                          <Link
                            to={`/profiles/${app.candidateSlug}`}
                            className="btn btn-xs btn-ghost"
                            title="View passport"
                          >
                            <ExternalLink size={14} />
                          </Link>

                          <button
                            type="button"
                            className="btn btn-xs btn-ghost"
                            title="Notes"
                            onClick={() => setNotesAppId((prev) => (prev === app.id ? null : (app.id ?? null)))}
                          >
                            <MessageSquare size={14} />
                          </button>

                          {app.status !== 'rejected' && app.status !== 'offered' && (
                            <div className="dropdown dropdown-end">
                              <div tabIndex={0} role="button" className="btn btn-xs btn-outline">
                                {updating === app.id ? (
                                  <span className="loading loading-spinner loading-xs" />
                                ) : (
                                  <>
                                    Move <ChevronDown size={12} />
                                  </>
                                )}
                              </div>
                              <ul
                                tabIndex={0}
                                className="dropdown-content menu p-2 shadow bg-base-200 rounded-box w-40 z-10"
                              >
                                {STATUS_OPTIONS.filter((s) => s !== app.status).map((s) => (
                                  <li key={s}>
                                    <button type="button" onClick={() => app.id && handleStatusChange(app.id, s)}>
                                      {s.charAt(0).toUpperCase() + s.slice(1)}
                                    </button>
                                  </li>
                                ))}
                              </ul>
                            </div>
                          )}
                        </div>
                      </td>
                    </tr>
                    {notesAppId === app.id && (
                      <tr>
                        <td colSpan={4} className="bg-base-100">
                          <ApplicationNotes applicationId={app.id} />
                        </td>
                      </tr>
                    )}
                  </Fragment>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      ))}
    </div>
  );
}
