import { useEffect, useState } from 'react';
import { ApplicationKanban } from '../../components/jobseeker/ApplicationKanban';
import { LoadingFallback } from '../../components/ui/LoadingFallback';
import { useAuth } from '../../hooks/useAuth';
import type { Application } from '../../lib/application';
import { getMyApplications } from '../../lib/application';

export function ApplicationsPage() {
  const { user } = useAuth();
  const [applications, setApplications] = useState<Application[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!user) return;
    let cancelled = false;
    setLoading(true);
    getMyApplications()
      .then((data) => {
        if (!cancelled) setApplications(data);
      })
      .catch((err) => {
        if (!cancelled) setError(err instanceof Error ? err.message : 'Failed to load applications');
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });
    return () => {
      cancelled = true;
    };
  }, [user]);

  if (!user) {
    return (
      <div className="max-w-4xl mx-auto p-4">
        <p>Please log in to view your applications.</p>
      </div>
    );
  }

  if (loading) return <LoadingFallback text="Loading applications" />;

  if (error) {
    return (
      <div className="max-w-4xl mx-auto p-4">
        <div className="alert alert-error">{error}</div>
      </div>
    );
  }

  return (
    <div className="max-w-5xl mx-auto p-4 space-y-4">
      <h1 className="text-2xl font-bold">My Applications</h1>
      <ApplicationKanban applications={applications} />
    </div>
  );
}
