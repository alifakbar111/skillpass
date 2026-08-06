import { api } from '@/lib/api';

export interface DidIdentity {
  id: string;
  did: string;
  publicKey: string;
  algorithm: string;
  issuerDid: string;
  createdAt: string;
}

export interface SignedCredential {
  id: string;
  credentialType: string;
  issuerDid: string;
  subjectDataHash: string;
  signature: string;
  algorithm: string;
  issuedAt: string;
  verified: boolean;
}

export function getEmployeeDid(employeeId: string): Promise<DidIdentity> {
  return api<DidIdentity>(`/hris/identity/employees/${employeeId}/did`);
}

export function issueEmployeeDid(employeeId: string): Promise<DidIdentity> {
  return api<DidIdentity>(`/hris/identity/employees/${employeeId}/did`, { method: 'POST' });
}

export function listEmployeeCredentials(employeeId: string): Promise<SignedCredential[]> {
  return api<SignedCredential[]>(`/hris/identity/employees/${employeeId}/credentials`);
}

// ── Sprint 6: signed skill attestations ──

export interface SkillAttestation {
  id: string;
  skillName: string;
  score: number;
  hash: string;
  signature: string;
  algorithm: string;
  issuedAt: string;
  revoked: boolean;
  verified: boolean;
}

export function attestEmployeeSkills(employeeId: string, evaluationId?: string): Promise<{ attested: number }> {
  return api(`/hris/identity/employees/${employeeId}/attest`, {
    method: 'POST',
    body: evaluationId ? { evaluationId } : {},
  });
}

export function listEmployeeAttestations(employeeId: string): Promise<SkillAttestation[]> {
  return api<SkillAttestation[]>(`/hris/identity/employees/${employeeId}/attestations`);
}

// ── Sprint 6: external identity verification (Dukcapil / PDDikti / manual) ──

export type IdentityProvider = 'dukcapil' | 'pddikti' | 'manual';

export interface IdentityVerification {
  id: string;
  provider: string;
  status: string;
  detail?: string;
  verifiedAt?: string;
  createdAt: string;
}

export function runVerification(employeeId: string, provider: IdentityProvider): Promise<IdentityVerification> {
  return api<IdentityVerification>(`/hris/identity/employees/${employeeId}/verify`, {
    method: 'POST',
    body: { provider },
  });
}

export function listVerifications(employeeId: string): Promise<IdentityVerification[]> {
  return api<IdentityVerification[]>(`/hris/identity/employees/${employeeId}/verifications`);
}

// ── Sprint 6: public Skill Passport settings ──

export interface PassportSettings {
  employeeId: string;
  slug: string;
  isPublic: boolean;
  publicPath: string;
  updatedAt: string;
}

export function getEmployeePassport(employeeId: string): Promise<PassportSettings> {
  return api<PassportSettings>(`/hris/identity/employees/${employeeId}/passport`);
}

export function setEmployeePassportVisibility(employeeId: string, isPublic: boolean): Promise<PassportSettings> {
  return api<PassportSettings>(`/hris/identity/employees/${employeeId}/passport`, {
    method: 'PUT',
    body: { isPublic },
  });
}
