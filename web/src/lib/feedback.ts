import { api } from '@/lib/api';

export interface RatingArea {
  skill: string;
  rating: number;
  note: string;
}

export interface AISuggestion {
  area: string;
  tip: string;
  resource?: string;
}

export interface Feedback {
  id: string;
  profileId: string;
  companyId: string;
  content: string;
  ratingAreas: RatingArea[];
  aiSuggestions: AISuggestion[];
  createdAt: string;
}

export interface CreateFeedbackRequest {
  content: string;
  ratingAreas: RatingArea[];
}

export async function createFeedback(profileId: string, data: CreateFeedbackRequest): Promise<Feedback> {
  return api<Feedback>(`/feedback/${profileId}`, {
    method: 'POST',
    body: data,
  });
}

export async function getMyFeedback(): Promise<Feedback[]> {
  return api<Feedback[]>('/feedback/me');
}

export async function getMySuggestions(): Promise<AISuggestion[]> {
  return api<AISuggestion[]>('/feedback/suggestions/me');
}
