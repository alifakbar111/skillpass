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
  latestNote?: string;
}

export interface ApplicationMessage {
  id: string;
  senderName: string;
  body: string;
  createdAt: string;
}

export async function applyToJob(jobId: string): Promise<Application> {
  return api<Application>(`/jobs/${jobId}/apply`, { method: 'POST' });
}

export async function getMyApplications(): Promise<Application[]> {
  return api<Application[]>('/applications/me');
}

export async function getApplicationMessages(applicationId: string): Promise<ApplicationMessage[]> {
  return api<ApplicationMessage[]>(`/applications/${applicationId}/messages`);
}

export async function addApplicationMessage(applicationId: string, body: string): Promise<ApplicationMessage> {
  return api<ApplicationMessage>(`/applications/${applicationId}/messages`, {
    method: 'POST',
    body: JSON.stringify({ body }),
  });
}
