import { api } from '@/lib/api';

export type AtsStageType = 'screening' | 'phone_screen' | 'technical' | 'hr_interview' | 'final' | 'offer' | 'hired';

export interface AtsStage {
  id: string;
  name: string;
  stageType: AtsStageType;
  sortOrder: number;
}

export interface AtsPipeline {
  id: string;
  name: string;
  isDefault: boolean;
  createdAt: string;
  stages?: AtsStage[];
}

export type AtsCandidateStatus = 'active' | 'hired' | 'rejected' | 'withdrawn';

export interface AtsCandidate {
  id: string;
  pipelineId: string;
  currentStageId?: string;
  currentStageName?: string;
  applicationId?: string;
  jobPostingId?: string;
  jobTitle?: string;
  candidateName: string;
  candidateEmail: string;
  status: AtsCandidateStatus;
  createdAt: string;
  updatedAt: string;
}

export interface AtsScorecard {
  id: string;
  candidateId: string;
  stageId?: string;
  evaluatorName: string;
  scores?: Record<string, number>;
  overallRating?: number;
  recommendation?: string;
  notes?: string;
  createdAt: string;
}

export interface AtsInterview {
  id: string;
  candidateId: string;
  stageId?: string;
  scheduledAt: string;
  mode: string;
  location?: string;
  meetingLink?: string;
  interviewer?: string;
  notes?: string;
  status: string;
  createdAt: string;
}

export interface AtsOffer {
  id: string;
  candidateId: string;
  templateId?: string;
  positionTitle: string;
  salary?: string;
  startDate?: string;
  body: string;
  status: string;
  acceptToken?: string;
  signatureName?: string;
  signedAt?: string;
  sentAt?: string;
  createdAt: string;
}

export interface AtsOfferTemplate {
  id: string;
  name: string;
  body: string;
  createdAt: string;
}

// ── Pipelines ──
export const listPipelines = () => api<AtsPipeline[]>('/hris/ats/pipelines');
export const createPipeline = (body: { name: string; stages: { name: string; stageType: AtsStageType }[] }) =>
  api<AtsPipeline>('/hris/ats/pipelines', { method: 'POST', body });
export const updatePipeline = (
  id: string,
  body: { name: string; stages: { name: string; stageType: AtsStageType }[] },
) => api<AtsPipeline>(`/hris/ats/pipelines/${id}`, { method: 'PUT', body });
export const deletePipeline = (id: string) => api<void>(`/hris/ats/pipelines/${id}`, { method: 'DELETE' });

// ── Candidates ──
export const listCandidates = (pipelineId?: string) =>
  api<AtsCandidate[]>(`/hris/ats/candidates${pipelineId ? `?pipelineId=${pipelineId}` : ''}`);
export const addCandidate = (body: {
  candidateName: string;
  candidateEmail: string;
  pipelineId?: string;
  applicationId?: string;
}) => api<AtsCandidate>('/hris/ats/candidates', { method: 'POST', body });
export const getCandidate = (id: string) => api<AtsCandidate>(`/hris/ats/candidates/${id}`);
export const moveCandidate = (id: string, stageId: string) =>
  api<AtsCandidate>(`/hris/ats/candidates/${id}/move`, { method: 'PUT', body: { stageId } });
export const setCandidateStatus = (id: string, status: AtsCandidateStatus) =>
  api<AtsCandidate>(`/hris/ats/candidates/${id}/status`, { method: 'PUT', body: { status } });

// ── Scorecards ──
export const listScorecards = (candidateId: string) =>
  api<AtsScorecard[]>(`/hris/ats/candidates/${candidateId}/scorecards`);
export const addScorecard = (
  candidateId: string,
  body: {
    stageId?: string;
    scores?: Record<string, number>;
    overallRating?: number;
    recommendation?: string;
    notes?: string;
  },
) => api<AtsScorecard>(`/hris/ats/candidates/${candidateId}/scorecards`, { method: 'POST', body });

// ── Interviews ──
export const listInterviews = (candidateId: string) =>
  api<AtsInterview[]>(`/hris/ats/candidates/${candidateId}/interviews`);
export const scheduleInterview = (
  candidateId: string,
  body: {
    stageId?: string;
    scheduledAt: string;
    mode: string;
    location?: string;
    meetingLink?: string;
    interviewer?: string;
    notes?: string;
  },
) => api<AtsInterview>(`/hris/ats/candidates/${candidateId}/interviews`, { method: 'POST', body });
export const updateInterviewStatus = (interviewId: string, status: string) =>
  api<AtsInterview>(`/hris/ats/interviews/${interviewId}/status`, { method: 'PUT', body: { status } });

// ── Offers ──
export const listOffers = (candidateId: string) => api<AtsOffer[]>(`/hris/ats/candidates/${candidateId}/offers`);
export const createOffer = (
  candidateId: string,
  body: { templateId?: string; positionTitle: string; salary?: string; startDate?: string; body?: string },
) => api<AtsOffer>(`/hris/ats/candidates/${candidateId}/offers`, { method: 'POST', body });
export const sendOffer = (offerId: string) => api<AtsOffer>(`/hris/ats/offers/${offerId}/send`, { method: 'POST' });

// ── Offer templates ──
export const listOfferTemplates = () => api<AtsOfferTemplate[]>('/hris/ats/offer-templates');
export const createOfferTemplate = (body: { name: string; body: string }) =>
  api<AtsOfferTemplate>('/hris/ats/offer-templates', { method: 'POST', body });
export const updateOfferTemplate = (id: string, body: { name: string; body: string }) =>
  api<AtsOfferTemplate>(`/hris/ats/offer-templates/${id}`, { method: 'PUT', body });
export const deleteOfferTemplate = (id: string) => api<void>(`/hris/ats/offer-templates/${id}`, { method: 'DELETE' });
