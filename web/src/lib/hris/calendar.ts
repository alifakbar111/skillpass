import { api } from '@/lib/api';

export interface WorkingCalendar {
  id: string;
  companyId: string;
  branchId?: string;
  year: number;
  defaultWorkDays: number[];
  createdAt: string;
}

export function listCalendars(year?: number): Promise<WorkingCalendar[]> {
  const params = year ? `?year=${year}` : '';
  return api<WorkingCalendar[]>(`/hris/working-calendars${params}`);
}

export function createCalendar(data: {
  branchId?: string;
  year: number;
  defaultWorkDays: number[];
}): Promise<WorkingCalendar> {
  return api<WorkingCalendar>('/hris/working-calendars', {
    method: 'POST',
    body: JSON.stringify(data),
  });
}

export function updateCalendar(id: string, data: { defaultWorkDays: number[] }): Promise<WorkingCalendar> {
  return api<WorkingCalendar>(`/hris/working-calendars/${id}`, {
    method: 'PUT',
    body: JSON.stringify(data),
  });
}

export function deleteCalendar(id: string): Promise<void> {
  return api(`/hris/working-calendars/${id}`, { method: 'DELETE' });
}
