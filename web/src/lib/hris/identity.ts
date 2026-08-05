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
