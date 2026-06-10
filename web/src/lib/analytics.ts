import { api } from './api';

export interface StatusCount {
  status: string;
  count: number;
}

export interface JobFunnel {
  jobPostingId: string;
  title: string;
  status: string;
  total: number;
  byStatus: StatusCount[];
}

export interface CompanyAnalytics {
  totalJobs: number;
  openJobs: number;
  totalApplications: number;
  applicationsByStatus: StatusCount[];
  avgDaysToDecision: number | null;
  jobs: JobFunnel[];
}

export interface JobseekerAnalytics {
  totalApplications: number;
  applicationsByStatus: StatusCount[];
  passportViews: number;
  responseRate: number | null;
}

export async function getCompanyAnalytics(): Promise<CompanyAnalytics> {
  return api<CompanyAnalytics>('/company/analytics');
}

export async function getJobseekerAnalytics(): Promise<JobseekerAnalytics> {
  return api<JobseekerAnalytics>('/profiles/me/analytics');
}
