import { api, apiUpload, getAccessToken } from '@/lib/api';

export type DocumentCategory = 'identity' | 'contract' | 'certificate' | 'payslip' | 'tax' | 'other';
export type ScanStatus = 'pending' | 'clean' | 'infected' | 'error';

export interface HrisDocument {
  id: string;
  employeeId?: string;
  employeeName?: string;
  category: DocumentCategory;
  filename: string;
  mimeType: string;
  fileSize: number;
  scanStatus: ScanStatus;
  uploadedByName?: string;
  createdAt: string;
}

export interface DocumentAuditEntry {
  id: string;
  documentId: string;
  filename: string;
  accessedByName?: string;
  action: string;
  ipAddress?: string;
  createdAt: string;
}

export function listDocuments(params: { category?: string; employeeId?: string } = {}): Promise<HrisDocument[]> {
  const q = new URLSearchParams();
  if (params.category) q.set('category', params.category);
  if (params.employeeId) q.set('employeeId', params.employeeId);
  const qs = q.toString();
  return api<HrisDocument[]>(`/hris/documents${qs ? `?${qs}` : ''}`);
}

export function uploadDocument(file: File, category: DocumentCategory, employeeId?: string): Promise<HrisDocument> {
  const form = new FormData();
  form.append('file', file);
  form.append('category', category);
  if (employeeId) form.append('employeeId', employeeId);
  return apiUpload<HrisDocument>('/hris/documents', form);
}

export function deleteDocument(id: string): Promise<void> {
  return api(`/hris/documents/${id}`, { method: 'DELETE' });
}

export function getDocumentAuditLog(): Promise<DocumentAuditEntry[]> {
  return api<DocumentAuditEntry[]>('/hris/documents-audit-log');
}

// Download streams through an authenticated endpoint (documents are private —
// there's no public URL), so we fetch with the bearer token and save the blob.
export async function downloadDocument(id: string, filename: string): Promise<void> {
  const res = await fetch(`/api/v1/hris/documents/${id}/download`, {
    headers: { Authorization: `Bearer ${getAccessToken() ?? ''}` },
  });
  if (!res.ok) throw new Error('Download failed');
  const blob = await res.blob();
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = filename;
  a.click();
  URL.revokeObjectURL(url);
}
