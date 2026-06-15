import { useQuery } from '@tanstack/react-query';
import { LoadingFallback } from '../../../components/ui/LoadingFallback';
import { useAuth } from '../../../hooks/useAuth';
import type { AISuggestion, Feedback } from '../../../lib/feedback';
import { getMyFeedback, getMySuggestions } from '../../../lib/feedback';

export function FeedbackPage() {
  const { user } = useAuth();
  const { data: feedback = [], isLoading: loadingFeedback } = useQuery({
    queryKey: ['feedback', 'me'],
    enabled: !!user,
    queryFn: getMyFeedback,
  });
  const { data: suggestions = [], isLoading: loadingSuggestions } = useQuery({
    queryKey: ['feedback', 'suggestions'],
    enabled: !!user,
    queryFn: getMySuggestions,
  });
  if (!user)
    return (
      <div className="max-w-4xl mx-auto p-4">
        <p>Please log in to view feedback.</p>
      </div>
    );
  if (loadingFeedback || loadingSuggestions) return <LoadingFallback text="Loading feedback" />;
  return (
    <div className="max-w-5xl mx-auto p-4 space-y-6">
      <h1 className="text-2xl font-bold">Feedback Received</h1>
      {feedback.length === 0 ? (
        <div className="alert alert-info">
          <span>No feedback received yet.</span>
        </div>
      ) : (
        <div className="space-y-4">
          {feedback.map((item: Feedback) => (
            <div key={item.id} className="card bg-base-200 shadow-md">
              <div className="card-body">
                <p className="text-sm text-base-content/60">From Company</p>
                <p className="text-xs text-base-content/40">{new Date(item.createdAt).toLocaleDateString()}</p>
                <p className="mt-2">{item.content}</p>
                {item.ratingAreas.length > 0 && (
                  <div className="mt-4">
                    <h3 className="font-semibold text-sm mb-2">Skill Ratings</h3>
                    <div className="grid grid-cols-2 gap-2">
                      {item.ratingAreas.map((area, i) => (
                        <div key={i} className="flex items-center gap-2">
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
      <h2 className="text-xl font-bold mt-8">AI Learning Suggestions</h2>
      {suggestions.length === 0 ? (
        <div className="alert alert-info">
          <span>No suggestions yet.</span>
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {suggestions.map((s: AISuggestion, i: number) => (
            <div key={i} className="card bg-base-100 border border-base-300">
              <div className="card-body">
                <h3 className="card-title text-sm">{s.area}</h3>
                <p className="text-sm">{s.tip}</p>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
