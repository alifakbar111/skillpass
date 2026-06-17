import { api } from '@/lib/api';

export interface SkillRequirement {
  skill: string;
  level: number;
  weight: number;
}

export interface ProgressionStep {
  fromRole: string;
  toRole: string;
}

export interface CareerPath {
  id: string;
  title: string;
  description: string;
  skillRequirements: SkillRequirement[];
  typicalProgression: ProgressionStep[];
  industry: string;
}

export interface SkillGapItem {
  skill: string;
  currentLevel: number;
  requiredLevel: number;
  gap: number;
}

export interface SkillGapResult {
  industry: string;
  skills: SkillGapItem[];
}

export interface CareerPrediction {
  predictedRole: string;
  timeline: string;
  similarProfiles: number;
  strengths: string[];
  weaknesses: string[];
  recommendations: string[];
}

export async function getCareerPaths(industry?: string): Promise<CareerPath[]> {
  const params = industry ? `?industry=${encodeURIComponent(industry)}` : '';
  return api<CareerPath[]>(`/career/paths${params}`);
}

export async function getSkillGap(): Promise<SkillGapResult> {
  return api<SkillGapResult>('/career/skill-gap/me');
}

export async function getCareerPrediction(): Promise<CareerPrediction> {
  return api<CareerPrediction>('/career/path/me');
}
