export interface Job {
  id: string;
  title: string;
  description: string;
  industry: string;
  tags?: string[];
  requiredSkills?: string[];
  location?: string;
  experienceLevel?: string;
  salaryRange?: string;
  status: string;
  createdAt: string;
}
