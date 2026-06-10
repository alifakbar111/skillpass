export interface Job {
  id: string;
  title: string;
  description: string;
  industry: string;
  tags?: string[];
  requiredSkills?: string[];
  experienceLevel?: string;
  location?: string;
  salaryRange?: string;
  status: string;
  createdAt: string;
}
