import { api } from '@/lib/api';

export interface LeaveType {
  id: string;
  companyId: string;
  name: string;
  code: string;
  defaultDaysPerYear: number;
  isPaid: boolean;
  requiresAttachment: boolean;
  isActive: boolean;
  createdAt: string;
}

export interface LeaveBalance {
  id: string;
  employeeId: string;
  leaveTypeId: string;
  leaveTypeName: string;
  year: number;
  totalDays: number;
  usedDays: number;
  carryOverDays: number;
  remaining: number;
  createdAt: string;
}

export interface LeaveRequest {
  id: string;
  companyId: string;
  employeeId: string;
  employeeName?: string;
  leaveTypeId: string;
  leaveTypeName?: string;
  startDate: string;
  endDate: string;
  totalDays: number;
  reason: string;
  attachmentUrl?: string;
  status: string;
  reviewerId?: string;
  reviewerComment?: string;
  reviewedAt?: string;
  createdAt: string;
}

export interface Holiday {
  id: string;
  companyId: string;
  name: string;
  date: string;
  isRecurring: boolean;
  createdAt: string;
}

// Leave Types
export function listLeaveTypes(): Promise<LeaveType[]> {
  return api<LeaveType[]>('/hris/leave-types');
}

export function createLeaveType(data: Partial<LeaveType>): Promise<LeaveType> {
  return api<LeaveType>('/hris/leave-types', { method: 'POST', body: JSON.stringify(data) });
}

export function updateLeaveType(id: string, data: Partial<LeaveType>): Promise<void> {
  return api(`/hris/leave-types/${id}`, { method: 'PUT', body: JSON.stringify(data) });
}

export function deleteLeaveType(id: string): Promise<void> {
  return api(`/hris/leave-types/${id}`, { method: 'DELETE' });
}

// Leave Balances
export function getLeaveBalances(employeeId: string, year?: number): Promise<LeaveBalance[]> {
  const params = year ? `?year=${year}` : '';
  return api<LeaveBalance[]>(`/hris/employees/${employeeId}/leave-balances${params}`);
}

export function initLeaveBalances(employeeId: string): Promise<void> {
  return api(`/hris/employees/${employeeId}/leave-balances/init`, { method: 'POST' });
}

// Leave Requests
export function createLeaveRequest(data: {
  leaveTypeId: string;
  startDate: string;
  endDate: string;
  totalDays: number;
  reason: string;
}): Promise<LeaveRequest> {
  return api<LeaveRequest>('/hris/leave-requests', { method: 'POST', body: JSON.stringify(data) });
}

export function listLeaveRequests(status?: string): Promise<LeaveRequest[]> {
  const params = status ? `?status=${status}` : '';
  return api<LeaveRequest[]>(`/hris/leave-requests${params}`);
}

export function myLeaveRequests(): Promise<LeaveRequest[]> {
  return api<LeaveRequest[]>('/hris/leave-requests/my');
}

export function reviewLeaveRequest(id: string, data: { status: string; comment: string }): Promise<void> {
  return api(`/hris/leave-requests/${id}/review`, { method: 'PUT', body: JSON.stringify(data) });
}

export function cancelLeaveRequest(id: string): Promise<void> {
  return api(`/hris/leave-requests/${id}/cancel`, { method: 'PUT' });
}

// Holidays
export function listHolidays(year?: number): Promise<Holiday[]> {
  const params = year ? `?year=${year}` : '';
  return api<Holiday[]>(`/hris/holidays${params}`);
}

export function createHoliday(data: { name: string; date: string; isRecurring: boolean }): Promise<Holiday> {
  return api<Holiday>('/hris/holidays', { method: 'POST', body: JSON.stringify(data) });
}

export function updateHoliday(id: string, data: Partial<Holiday>): Promise<void> {
  return api(`/hris/holidays/${id}`, { method: 'PUT', body: JSON.stringify(data) });
}

export function deleteHoliday(id: string): Promise<void> {
  return api(`/hris/holidays/${id}`, { method: 'DELETE' });
}
