import { api } from './api';

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
