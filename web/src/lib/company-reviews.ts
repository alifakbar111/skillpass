import { api } from './api';

export interface CompanyReview {
  id: string;
  companyId: string;
  candidateId: string;
  rating: number;
  review: string | null;
  interactionType: 'applied' | 'interviewed';
  createdAt: string;
}

export interface Reputation {
  averageRating: number;
  reviewCount: number;
}

export interface CreateReviewRequest {
  rating: number;
  review?: string;
  interactionType: 'applied' | 'interviewed';
}

export async function createCompanyReview(companyId: string, data: CreateReviewRequest): Promise<CompanyReview> {
  return api<CompanyReview>(`/companies/${companyId}/reviews`, {
    method: 'POST',
    body: JSON.stringify(data),
  });
}

export async function getCompanyReputation(companyId: string): Promise<Reputation> {
  return api<Reputation>(`/companies/${companyId}/reputation`);
}
