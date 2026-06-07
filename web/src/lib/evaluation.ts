import { api } from './api';

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

export async function triggerEvaluation(): Promise<EvaluationResult> {
  return api<EvaluationResult>('/evaluate/me', { method: 'POST' });
}

export async function getLatestEvaluation(): Promise<EvaluationResult> {
  return api<EvaluationResult>('/evaluate/me/results');
}
