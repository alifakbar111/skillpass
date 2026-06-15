import { api } from '@/lib/api';

export interface AttendanceRow {
  employeeName: string;
  employeeCode: string;
  date: string;
  clockIn: string;
  clockOut: string;
  workHours: string;
  status: string;
  shiftName: string;
}

export interface DeptBreakdown {
  department: string;
  headcount: number;
  newHires: number;
  exits: number;
}

export interface AnalyticsSnapshot {
  id: string;
  snapshotMonth: string;
  totalHeadcount: number;
  newHires: number;
  terminations: number;
  turnoverRate: number;
  avgTenureMonths: number;
  departmentBreakdown: DeptBreakdown[];
  createdAt: string;
}

export interface HeadcountStats {
  totalActive: number;
  byDepartment: { department: string; count: number }[];
  byBranch: { branch: string; count: number }[];
  byStatus: { status: string; count: number }[];
  avgTenureMonths: number;
  genderBreakdown: { gender: string; count: number }[];
}

export function exportAttendance(from: string, to: string) {
  return api<AttendanceRow[]>(`/hris/reports/attendance-export?from=${from}&to=${to}`);
}

export function getHeadcountStats() {
  return api<HeadcountStats>('/hris/reports/headcount');
}

export function generateSnapshot(month: string) {
  return api<AnalyticsSnapshot>('/hris/reports/snapshots', {
    method: 'POST',
    body: JSON.stringify({ month }),
  });
}

export function listSnapshots() {
  return api<AnalyticsSnapshot[]>('/hris/reports/snapshots');
}
