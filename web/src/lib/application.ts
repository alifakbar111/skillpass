import { api } from './api';

export type ApplicationStatus = 'applied' | 'reviewed' | 'interviewed' | 'offered' | 'rejected';

export interface Application {
  id: string;
  jobseekerId: string;
  jobPostingId: string;
  status: ApplicationStatus;
  createdAt: string;
  updatedAt: string;
  jobTitle?: string;
  companyName?: string;
}

export async function applyToJob(jobId: string): Promise<Application> {
  return api<Application>(`/jobs/${jobId}/apply`, { method: 'POST' });
}

export async function getMyApplications(): Promise<Application[]> {
  return api<Application[]>('/applications/me');
}
