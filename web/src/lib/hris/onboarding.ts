import { api } from '@/lib/api';

export interface TemplateTask {
  id: string;
  templateId: string;
  title: string;
  description: string;
  sortOrder: number;
  dueDays: number;
  assigneeRole: string;
}

export interface Template {
  id: string;
  name: string;
  description: string;
  isActive: boolean;
  tasks: TemplateTask[];
  createdAt: string;
}

export interface ChecklistItem {
  id: string;
  checklistId: string;
  title: string;
  description: string;
  sortOrder: number;
  dueDate: string | null;
  isCompleted: boolean;
  completedAt: string | null;
  completedBy: string | null;
  notes: string;
}

export interface Checklist {
  id: string;
  employeeId: string;
  employeeName: string;
  employeeCode: string;
  templateName: string;
  status: string;
  startedAt: string;
  completedAt: string | null;
  items: ChecklistItem[];
  progress: number;
}

export function listTemplates() {
  return api<Template[]>('/hris/onboarding/templates');
}

export function getTemplate(id: string) {
  return api<Template>(`/hris/onboarding/templates/${id}`);
}

export function createTemplate(data: { name: string; description: string; tasks: Partial<TemplateTask>[] }) {
  return api<Template>('/hris/onboarding/templates', {
    method: 'POST',
    body: JSON.stringify(data),
  });
}

export function updateTemplate(id: string, data: { name: string; description: string; isActive: boolean }) {
  return api<void>(`/hris/onboarding/templates/${id}`, {
    method: 'PUT',
    body: JSON.stringify(data),
  });
}

export function deleteTemplate(id: string) {
  return api<void>(`/hris/onboarding/templates/${id}`, { method: 'DELETE' });
}

export function listChecklists() {
  return api<Checklist[]>('/hris/onboarding/checklists');
}

export function getChecklist(id: string) {
  return api<Checklist>(`/hris/onboarding/checklists/${id}`);
}

export function getMyChecklist() {
  return api<Checklist | null>('/hris/onboarding/my');
}

export function assignChecklist(employeeId: string, templateId: string) {
  return api<Checklist>(`/hris/employees/${employeeId}/onboarding`, {
    method: 'POST',
    body: JSON.stringify({ templateId }),
  });
}

export function completeItem(itemId: string, notes = '') {
  return api<void>(`/hris/onboarding/items/${itemId}/complete`, {
    method: 'POST',
    body: JSON.stringify({ notes }),
  });
}

export function uncompleteItem(itemId: string) {
  return api<void>(`/hris/onboarding/items/${itemId}/uncomplete`, { method: 'POST' });
}
