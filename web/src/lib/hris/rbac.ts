import { api } from '@/lib/api';

export interface HrisRole {
  id: string;
  name: string;
  description?: string;
  isSystem: boolean;
}

export interface MyPermissions {
  permissions: string[];
  roles: HrisRole[];
}

export function getMyPermissions(): Promise<MyPermissions> {
  return api<MyPermissions>('/hris/me/permissions');
}

export function listRoles(): Promise<HrisRole[]> {
  return api<HrisRole[]>('/hris/roles');
}

export function assignRole(employeeId: string, roleId: string): Promise<void> {
  return api(`/hris/employees/${employeeId}/roles`, {
    method: 'POST',
    body: JSON.stringify({ roleId }),
  });
}

export function removeRole(employeeId: string, roleId: string): Promise<void> {
  return api(`/hris/employees/${employeeId}/roles/${roleId}`, {
    method: 'DELETE',
  });
}
