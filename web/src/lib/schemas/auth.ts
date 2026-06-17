import { z } from 'zod';

/** Matches the OpenAPI UserResponse shape. */
export const UserSchema = z.object({
  id: z.string(),
  email: z.string().email().optional(),
  username: z.string().optional(),
  name: z.string().optional().nullable(),
  role: z.enum(['jobseeker', 'company', 'admin']).optional(),
  isVerified: z.boolean().optional(),
  avatarUrl: z.string().optional().nullable(),
  createdAt: z.string().optional(),
});

export type AuthUser = z.infer<typeof UserSchema>;

/** Matches the OpenAPI LoginResponse shape. */
export const LoginResponseSchema = z.object({
  accessToken: z.string().optional(),
  user: UserSchema.optional(),
});
