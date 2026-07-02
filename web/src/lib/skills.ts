import { api } from '@/lib/api';

export interface Skill {
  id: string;
  name: string;
}

export async function searchSkills(query: string): Promise<Skill[]> {
  if (!query.trim()) return [];
  return api<Skill[]>(`/skills?q=${encodeURIComponent(query)}`);
}

export async function getPopularSkills(): Promise<Skill[]> {
  return api<Skill[]>('/skills');
}
