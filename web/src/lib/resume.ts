import { api } from '@/lib/api';

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
  /** Raw Markdown extracted from the PDF by MarkItDown (available on upload). */
  rawMarkdown?: string;
}

export async function parseResume(text: string): Promise<ParsedResume> {
  return api<ParsedResume>('/profiles/me/resume-parse', {
    method: 'POST',
    body: { text },
  });
}

export async function uploadResume(file: File): Promise<ParsedResume> {
  const form = new FormData();
  form.append('file', file);
  return api<ParsedResume>('/profiles/me/resume-upload', { method: 'POST', body: form });
}
