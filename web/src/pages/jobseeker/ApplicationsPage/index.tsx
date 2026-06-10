import { useQuery } from '@tanstack/react-query';
import { ApplicationKanban } from '../../../components/jobseeker/ApplicationKanban';
import { LoadingFallback } from '../../../components/ui/LoadingFallback';
import { useAuth } from '../../../hooks/useAuth';
import { getMyApplications } from '../../../lib/application';

export function ApplicationsPage() {
  const { user } = useAuth();
  const {
    data: applications = [],
    error,
    isLoading,
  } = useQuery({
    queryKey: ['applications', 'me'],
    enabled: !!user,
    queryFn: getMyApplications,
  });

  if (!user) {
    return (
      <div className="max-w-4xl mx-auto p-4">
        <p>Please log in to view your applications.</p>
      </div>
    );
  }

  if (isLoading) return <LoadingFallback text="Loading applications" />;

  if (error) {
    return (
      <div className="max-w-4xl mx-auto p-4">
        <div className="alert alert-error">
          {error instanceof Error ? error.message : 'Failed to load applications'}
        </div>
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
