import { api, apiUpload } from './api';

export interface ParsedExperience {
  type: string;
  title: string;
  organization: string;
  startDate: string;
  endDate: string;
  isCurrent: boolean;
  description: string;
  skillsUsed: string[];
}

export interface ParsedResume {
  headline: string;
  about: string;
  yearsOfExperience: number;
  experiences: ParsedExperience[];
}

export async function parseResume(text: string): Promise<ParsedResume> {
  return api<ParsedResume>('/profiles/me/resume-parse', {
    method: 'POST',
    body: JSON.stringify({ text }),
  });
}

export async function uploadResume(file: File): Promise<ParsedResume> {
  const form = new FormData();
  form.append('file', file);
  return apiUpload<ParsedResume>('/profiles/me/resume-upload', form);
}
