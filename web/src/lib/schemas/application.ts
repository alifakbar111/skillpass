import { z } from 'zod';

/** Matches the OpenAPI ApplicationResult shape. */
export const ApplicationSchema = z.object({
  id: z.string(),
  jobId: z.string().optional(),
  userId: z.string().optional(),
  status: z.string().optional(),
  resumeUrl: z.string().optional().nullable(),
  coverLetter: z.string().optional().nullable(),
  createdAt: z.string().optional(),
  updatedAt: z.string().optional(),
  job: z
    .object({
      id: z.string().optional(),
      title: z.string().optional(),
      companyId: z.string().optional(),
    })
    .optional(),
});

export type Application = z.infer<typeof ApplicationSchema>;
