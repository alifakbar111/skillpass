import { z } from 'zod';

/** Matches the OpenAPI JobResponse shape — validates at the fetch boundary. */
export const JobSchema = z.object({
  id: z.string().optional(),
  title: z.string().optional(),
  description: z.string().optional(),
  industry: z.string().optional(),
  location: z.string().optional(),
  salaryRange: z.string().optional(),
  experienceLevel: z.string().optional(),
  status: z.string().optional(),
  createdAt: z.string().optional(),
  requiredSkills: z.array(z.string()).optional(),
  tags: z.array(z.string()).optional(),
  companyId: z.string().optional(),
});

export type Job = z.infer<typeof JobSchema>;
