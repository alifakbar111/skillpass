import { ApiError, api } from './api';

export interface SkillNote {
  skill: string;
  score: number;
  note: string;
}

export interface Suggestion {
  area: string;
  tip: string;
}

export interface SkillScoreItem {
  skill: string;
  category: string;
  score: number;
}

export interface EvaluationResult {
  id: string;
  overallScore: number;
  strengths: SkillNote[];
  weaknesses: SkillNote[];
  suggestions: Suggestion[];
  skillScores: SkillScoreItem[];
  createdAt: string;
}

export interface SuggestedRole {
  title: string;
  reason: string;
  readiness: 'ready' | 'stretch' | 'long-term';
}

export interface DevelopmentStep {
  area: string;
  action: string;
}

export interface CareerPathResult {
  currentPosition: string;
  suggestedRoles: SuggestedRole[];
  steps: DevelopmentStep[];
}

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
