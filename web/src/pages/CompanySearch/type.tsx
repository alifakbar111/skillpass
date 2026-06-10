export interface Candidate {
  id: string;
  name: string;
  avatarUrl?: string;
  headline?: string;
  about?: string;
  yearsOfExperience?: number;
  slug: string;
  skills: string[];
}
