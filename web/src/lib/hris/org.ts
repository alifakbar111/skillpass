import { api } from '@/lib/api';

export interface Branch {
  id: string;
  companyId: string;
  name: string;
  branchType: string;
  parentBranchId?: string;
  address?: string;
  city?: string;
  province?: string;
  latitude?: number;
  longitude?: number;
  geofenceRadiusMeters: number;
  isActive: boolean;
  createdAt: string;
}

export interface Department {
  id: string;
  companyId: string;
  name: string;
  parentDepartmentId?: string;
  createdAt: string;
}

export interface Position {
  id: string;
  companyId: string;
  name: string;
  departmentId?: string;
  level: string;
  createdAt: string;
}

export interface OrgNode {
  id: string;
  name: string;
  type: string;
  parentId?: string;
  level?: string;
  employeeCount?: number;
  children?: OrgNode[];
}

export function listBranches(): Promise<Branch[]> {
  return api<Branch[]>('/hris/branches');
}

export function createBranch(data: Partial<Branch>): Promise<Branch> {
  return api<Branch>('/hris/branches', { method: 'POST', body: JSON.stringify(data) });
}

export function updateBranch(id: string, data: Partial<Branch>): Promise<Branch> {
  return api<Branch>(`/hris/branches/${id}`, { method: 'PUT', body: JSON.stringify(data) });
}

export function deleteBranch(id: string): Promise<void> {
  return api(`/hris/branches/${id}`, { method: 'DELETE' });
}

export function listDepartments(): Promise<Department[]> {
  return api<Department[]>('/hris/departments');
}

export function createDepartment(data: { name: string; parentDepartmentId?: string }): Promise<Department> {
  return api<Department>('/hris/departments', { method: 'POST', body: JSON.stringify(data) });
}

export function updateDepartment(id: string, data: Partial<Department>): Promise<Department> {
  return api<Department>(`/hris/departments/${id}`, { method: 'PUT', body: JSON.stringify(data) });
}

export function deleteDepartment(id: string): Promise<void> {
  return api(`/hris/departments/${id}`, { method: 'DELETE' });
}

export function listPositions(): Promise<Position[]> {
  return api<Position[]>('/hris/positions');
}

export function createPosition(data: { name: string; departmentId?: string; level: string }): Promise<Position> {
  return api<Position>('/hris/positions', { method: 'POST', body: JSON.stringify(data) });
}

export function updatePosition(id: string, data: Partial<Position>): Promise<Position> {
  return api<Position>(`/hris/positions/${id}`, { method: 'PUT', body: JSON.stringify(data) });
}

export function deletePosition(id: string): Promise<void> {
  return api(`/hris/positions/${id}`, { method: 'DELETE' });
}

export function getOrgTree(): Promise<OrgNode[]> {
  return api<OrgNode[]>('/hris/org/tree');
}

export interface OrgChartNode {
  id: string;
  name: string;
  positionName?: string;
  level?: string;
  departmentId?: string;
  photoUrl?: string;
  managerId?: string;
  children: OrgChartNode[];
}

export function getOrgChart(): Promise<OrgChartNode[]> {
  return api<OrgChartNode[]>('/hris/org/chart');
}
