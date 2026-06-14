import { api } from '@/lib/api';

export interface ShiftTemplate {
  id: string;
  companyId: string;
  name: string;
  startTime: string;
  endTime: string;
  breakDurationMinutes: number;
  lateToleranceMinutes: number;
  overtimeMultiplier: number;
  applicableDays: number[];
  isDefault: boolean;
  createdAt: string;
}

export interface EmployeeShift {
  id: string;
  employeeId: string;
  shiftId: string;
  effectiveDate: string;
  endDate?: string;
  shiftName: string;
  createdAt: string;
}

export interface AttendanceLog {
  id: string;
  companyId: string;
  employeeId: string;
  employeeName?: string;
  date: string;
  clockIn?: string;
  clockOut?: string;
  clockInLat?: number;
  clockInLng?: number;
  clockOutLat?: number;
  clockOutLng?: number;
  branchId?: string;
  isInGeofence?: boolean;
  isLate: boolean;
  lateMinutes: number;
  isEarlyOut: boolean;
  overtimeMinutes: number;
  attendanceCode: string;
  createdAt: string;
}

export interface DashboardStats {
  totalEmployees: number;
  present: number;
  late: number;
  absent: number;
  onLeave: number;
}

export interface DashboardResponse {
  stats: DashboardStats;
  logs: AttendanceLog[];
}

export interface AttendanceException {
  id: string;
  companyId: string;
  employeeId: string;
  employeeName?: string;
  date: string;
  exceptionType: string;
  reason: string;
  attachmentUrl?: string;
  status: string;
  reviewerId?: string;
  reviewerComment?: string;
  reviewedAt?: string;
  createdAt: string;
}

// Shift Templates
export function listShiftTemplates(): Promise<ShiftTemplate[]> {
  return api<ShiftTemplate[]>('/hris/shifts');
}

export function createShiftTemplate(data: Partial<ShiftTemplate>): Promise<ShiftTemplate> {
  return api<ShiftTemplate>('/hris/shifts', { method: 'POST', body: JSON.stringify(data) });
}

export function updateShiftTemplate(id: string, data: Partial<ShiftTemplate>): Promise<void> {
  return api(`/hris/shifts/${id}`, { method: 'PUT', body: JSON.stringify(data) });
}

export function deleteShiftTemplate(id: string): Promise<void> {
  return api(`/hris/shifts/${id}`, { method: 'DELETE' });
}

// Employee Shifts
export function assignShift(
  employeeId: string,
  data: { shiftId: string; effectiveDate: string; endDate?: string },
): Promise<EmployeeShift> {
  return api<EmployeeShift>(`/hris/employees/${employeeId}/shifts`, { method: 'POST', body: JSON.stringify(data) });
}

export function listEmployeeShifts(employeeId: string): Promise<EmployeeShift[]> {
  return api<EmployeeShift[]>(`/hris/employees/${employeeId}/shifts`);
}

// Attendance
export function clockIn(data: { lat: number; lng: number; branchId?: string }): Promise<AttendanceLog> {
  return api<AttendanceLog>('/hris/attendance/clock-in', { method: 'POST', body: JSON.stringify(data) });
}

export function clockOut(data: { lat: number; lng: number }): Promise<AttendanceLog> {
  return api<AttendanceLog>('/hris/attendance/clock-out', { method: 'POST', body: JSON.stringify(data) });
}

export function getAttendanceDashboard(date?: string): Promise<DashboardResponse> {
  const params = date ? `?date=${date}` : '';
  return api<DashboardResponse>(`/hris/attendance/dashboard${params}`);
}

export function getMyAttendance(month?: string): Promise<AttendanceLog[]> {
  const params = month ? `?month=${month}` : '';
  return api<AttendanceLog[]>(`/hris/attendance/my${params}`);
}

// Exceptions
export function createException(data: {
  date: string;
  exceptionType: string;
  reason: string;
  attachmentUrl?: string;
}): Promise<AttendanceException> {
  return api<AttendanceException>('/hris/attendance-exceptions', { method: 'POST', body: JSON.stringify(data) });
}

export function listExceptions(status?: string): Promise<AttendanceException[]> {
  const params = status ? `?status=${status}` : '';
  return api<AttendanceException[]>(`/hris/attendance-exceptions${params}`);
}

export function reviewException(id: string, data: { status: string; comment: string }): Promise<void> {
  return api(`/hris/attendance-exceptions/${id}/review`, { method: 'PUT', body: JSON.stringify(data) });
}
