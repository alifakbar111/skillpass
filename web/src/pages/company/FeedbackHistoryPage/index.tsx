import { useQuery } from '@tanstack/react-query';
import { LoadingFallback } from '@/components/ui/LoadingFallback';
import { useAuth } from '@/hooks/useAuth';
import { api } from '@/lib/api';
import type { Feedback } from '../../../lib/feedback';

async function getCompanyFeedback(): Promise<Feedback[]> {
  return api<Feedback[]>('/feedback/company');
}

export function FeedbackHistoryPage() {
  const { user } = useAuth();
  const {
    data: feedback = [],
    isLoading,
    isError,
    error,
  } = useQuery({
    queryKey: ['feedback', 'company'],
    enabled: !!user,
    queryFn: getCompanyFeedback,
  });
  if (!user)
    return (
      <div className="max-w-4xl mx-auto p-4">
        <p>Please log in to view feedback history.</p>
      </div>
    );
  if (isLoading) return <LoadingFallback text="Loading feedback history" />;
  if (isError)
    return (
      <div className="max-w-4xl mx-auto p-4">
        <div className="alert alert-error">
          <span>{error.message}</span>
        </div>
      </div>
    );
  return (
    <div className="max-w-5xl mx-auto p-4 space-y-6">
      <h1 className="text-2xl font-bold">Feedback History</h1>
      {feedback.length === 0 ? (
        <div className="alert alert-info">
          <span>You haven't given any feedback yet.</span>
        </div>
      ) : (
        <div className="space-y-4">
          {feedback.map((item: Feedback) => (
            <div key={item.id} className="card bg-base-200 shadow-md">
              <div className="card-body">
                <p className="text-sm text-base-content/60">Candidate Profile</p>
                <p className="text-xs text-base-content/40">{new Date(item.createdAt).toLocaleDateString()}</p>
                <p className="mt-2">{item.content}</p>
                {item.ratingAreas.length > 0 && (
                  <div className="mt-4">
                    <h3 className="font-semibold text-sm mb-2">Skill Ratings Given</h3>
                    <div className="grid grid-cols-2 gap-2">
                      {item.ratingAreas.map((area) => (
                        <div key={area.skill} className="flex items-center gap-2">
                          <span className="text-sm">{area.skill}</span>
                          <div className="badge badge-primary badge-sm">{area.rating}/5</div>
                        </div>
                      ))}
                    </div>
                  </div>
                )}
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
