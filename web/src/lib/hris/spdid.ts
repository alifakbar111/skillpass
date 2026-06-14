import { api } from '@/lib/api';

export interface DIDRecord {
  id: string;
  companyId: string;
  employeeId: string;
  didString: string;
  status: string;
  createdAt: string;
}

export function createDID(employeeId: string): Promise<DIDRecord> {
  return api<DIDRecord>(`/hris/employees/${employeeId}/did`, { method: 'POST' });
}

export function getDID(employeeId: string): Promise<DIDRecord> {
  return api<DIDRecord>(`/hris/employees/${employeeId}/did`);
}
