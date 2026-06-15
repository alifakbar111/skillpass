import { useQuery } from '@tanstack/react-query';
import { LoadingFallback } from '../../../components/ui/LoadingFallback';
import { useAuth } from '../../../hooks/useAuth';
import { api } from '../../../lib/api';
import { getCompanyReputation } from '../../../lib/company-reviews';

interface CompanyReview {
  id: string;
  rating: number;
  review: string | null;
  interactionType: string;
  createdAt: string;
}

async function getCompanyReviews(companyId: string): Promise<CompanyReview[]> {
  return api<CompanyReview[]>(`/companies/${companyId}/reviews`);
}

async function getCompanyProfile(): Promise<{ id: string }> {
  return api<{ id: string }>('/company/profile');
}

export function ReputationPage() {
  const { user } = useAuth();
  const { data: companyProfile } = useQuery({
    queryKey: ['company', 'profile'],
    enabled: !!user,
    queryFn: getCompanyProfile,
  });
  const companyId = companyProfile?.id ?? '';
  const { data: reputation, isLoading: loadingReputation } = useQuery({
    queryKey: ['company', 'reputation'],
    enabled: !!companyId,
    queryFn: () => getCompanyReputation(companyId),
  });
  const { data: reviews = [], isLoading: loadingReviews } = useQuery({
    queryKey: ['company', 'reviews'],
    enabled: !!companyId,
    queryFn: () => getCompanyReviews(companyId),
  });
  if (!user)
    return (
      <div className="max-w-4xl mx-auto p-4">
        <p>Please log in to view reputation.</p>
      </div>
    );
  if (loadingReputation || loadingReviews) return <LoadingFallback text="Loading reputation" />;
  return (
    <div className="max-w-5xl mx-auto p-4 space-y-6">
      <h1 className="text-2xl font-bold">Company Reputation</h1>
      {reputation && (
        <div className="grid grid-cols-2 gap-4">
          <div className="stat bg-base-200 rounded-box p-4">
            <div className="stat-title text-xs">Average Rating</div>
            <div className="stat-value text-xl text-primary">
              {reputation.averageRating > 0 ? reputation.averageRating.toFixed(1) : '—'}
            </div>
            <div className="stat-desc">out of 5 stars</div>
          </div>
          <div className="stat bg-base-200 rounded-box p-4">
            <div className="stat-title text-xs">Total Reviews</div>
            <div className="stat-value text-xl">{reputation.reviewCount}</div>
          </div>
        </div>
      )}
      <div className="card bg-base-200 shadow-md">
        <div className="card-body">
          <h2 className="card-title">Reviews</h2>
          {reviews.length === 0 ? (
            <p className="text-sm text-base-content/60">No reviews yet.</p>
          ) : (
            <div className="space-y-4">
              {reviews.map((review: CompanyReview) => (
                <div key={review.id} className="border-b border-base-300 pb-4 last:border-0">
                  <div className="flex items-center gap-2 mb-2">
                    <div className="rating rating-sm">
                      {[1, 2, 3, 4, 5].map((star) => (
                        <input
                          key={star}
                          type="radio"
                          name={`rating-${review.id}`}
                          className="mask mask-star-2 bg-warning"
                          checked={star === review.rating}
                          readOnly
                        />
                      ))}
                    </div>
                    <span className="badge badge-outline badge-sm">{review.interactionType}</span>
                    <span className="text-xs text-base-content/40">
                      {new Date(review.createdAt).toLocaleDateString()}
                    </span>
                  </div>
                  {review.review && <p className="text-sm">{review.review}</p>}
                </div>
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
