import { api } from './api';

export interface JobMatch {
  jobPostingId: string;
  title: string;
  companyName: string;
  industry: string;
  location: string | null;
  salaryRange: string | null;
  experienceLevel: string | null;
  matchScore: number;
  matchReason: string;
}

export interface CandidateMatch {
  profileId: string;
  name: string;
  headline: string | null;
  overallScore: number;
  topSkills: string[];
  matchScore: number;
  matchReason: string;
}

export async function getJobMatches(): Promise<JobMatch[]> {
  return api<JobMatch[]>('/jobs/matches');
}

export async function getCandidateMatches(jobId: string): Promise<CandidateMatch[]> {
  return api<CandidateMatch[]>(`/candidates/matches?jobId=${encodeURIComponent(jobId)}`);
}
