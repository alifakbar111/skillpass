export interface Experience {
  id: string;
  type: string;
  title: string;
  organization: string;
  startDate: string;
  endDate?: string;
  isCurrent: boolean;
  description?: string;
  industry?: string;
  skillsUsed?: string[];
  url?: string;
}

export interface Profile {
  id: string;
  headline?: string;
  about?: string;
  yearsOfExperience?: number;
  slug: string;
  name?: string;
  avatarUrl?: string | null;
  experiences: Experience[];
}
