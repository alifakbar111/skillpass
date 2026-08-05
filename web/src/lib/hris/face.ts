import { api } from '@/lib/api';

export interface FaceStatus {
  enrolled: boolean;
  enrolledAt?: string;
  livenessScore?: number;
}

export interface FaceEnrollResult {
  enrolled: boolean;
  livenessScore: number;
  enrolledAt: string;
}

export function getFaceStatus(): Promise<FaceStatus> {
  return api<FaceStatus>('/hris/face/status');
}

export function getEmployeeFaceStatus(employeeId: string): Promise<FaceStatus> {
  return api<FaceStatus>(`/hris/face/employees/${employeeId}`);
}

// enrollFace sends a base64 (data-URL or raw) image to enrol the current user.
export function enrollFace(image: string): Promise<FaceEnrollResult> {
  return api<FaceEnrollResult>('/hris/face/enroll', { method: 'POST', body: { image } });
}
