import { api } from '@/lib/api';

// Public verification + offer-acceptance APIs (no auth required — the token or
// slug is the credential). Phase 2 · Sprints 5 & 6.

export interface PublicOffer {
  companyName: string;
  candidateName: string;
  positionTitle: string;
  salary?: string;
  startDate?: string;
  body: string;
  status: string;
}

export interface AcceptOfferResult {
  status: string;
  employeeLinked: boolean;
  message: string;
}

export const getPublicOffer = (token: string) => api<PublicOffer>(`/ats/offers/${token}`);
export const acceptOffer = (token: string, signatureName: string) =>
  api<AcceptOfferResult>(`/ats/offers/${token}/accept`, { method: 'POST', body: { signatureName } });
export const declineOffer = (token: string) => api<void>(`/ats/offers/${token}/decline`, { method: 'POST' });

export interface VerifiedCredential {
  id: string;
  type: string;
  skillName: string;
  score: number;
  issuerDid: string;
  issuedAt: string;
  revoked: boolean;
  verified: boolean;
}

export const verifyCredential = (id: string) =>
  api<VerifiedCredential>(`/verify/credential?id=${encodeURIComponent(id)}`);

export interface PublicSkillBadge {
  attestationId: string;
  skillName: string;
  score: number;
  verified: boolean;
  verifyPath: string;
}

export interface PublicPassport {
  name: string;
  companyName?: string;
  issuerDid: string;
  identityVerified: boolean;
  skills?: PublicSkillBadge[];
}

export const getPublicPassport = (slug: string) => api<PublicPassport>(`/verify/passport/${slug}`);
