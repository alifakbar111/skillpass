import type {
  CareerPathResult,
  DevelopmentStep,
  EvaluationResponse,
  SkillNote,
  SkillScoreItem,
  SuggestedRole,
  Suggestion,
} from './api-types';
import { ApiError, api } from './api';

// EvaluationResult keeps the historical name used across the UI; it is the
// generated server response type.
export type EvaluationResult = EvaluationResponse;
export type { CareerPathResult, DevelopmentStep, SkillNote, SkillScoreItem, SuggestedRole, Suggestion };

export async function triggerEvaluation(): Promise<EvaluationResult> {
  return api<EvaluationResult>('/evaluate/me', { method: 'POST' });
}

/** Returns `null` when the jobseeker has never been evaluated (server responds 404). */
export async function getLatestEvaluation(): Promise<EvaluationResult | null> {
  try {
    return await api<EvaluationResult>('/evaluate/me/results');
  } catch (err) {
    if (err instanceof ApiError && err.status === 404) {
      return null;
    }
    throw err;
  }
}

export async function getCareerPath(): Promise<CareerPathResult> {
  return api<CareerPathResult>('/evaluate/me/career-path', { method: 'POST' });
}
