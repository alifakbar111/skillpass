import { EXPERIENCE_VALUES } from '../constants';
import { z } from 'zod';

export const loginSchema = z.object({
  email: z.string().email('Invalid email address'),
  password: z.string().min(1, 'Password is required').max(128),
});
export type LoginForm = z.infer<typeof loginSchema>;

export const registerSchema = z
  .object({
    role: z.enum(['jobseeker', 'company']),
    email: z.string().email('Invalid email address'),
    username: z
      .string()
      .min(3, 'Username must be at least 3 characters')
      .max(32, 'Username must be at most 32 characters')
      .regex(/^[a-z0-9_-]+$/i, 'Username may only contain letters, numbers, _ and -'),
    password: z.string().min(8, 'Password must be at least 8 characters').max(128),
    name: z.string().optional(),
    companyName: z.string().optional(),
    businessRegistration: z.string().optional(),
    website: z.string().optional(),
    address: z.string().optional(),
    contact: z.string().optional(),
  })
  .superRefine((data, ctx) => {
    if (data.role === 'jobseeker') {
      if (!data.name || data.name.length < 1) {
        ctx.addIssue({ code: 'custom', message: 'Name is required', path: ['name'] });
      }
    } else {
      if (!data.companyName || data.companyName.length < 1) {
        ctx.addIssue({
          code: 'custom',
          message: 'Company name is required',
          path: ['companyName'],
        });
      }
      if (!data.businessRegistration || data.businessRegistration.length < 1) {
        ctx.addIssue({
          code: 'custom',
          message: 'Business registration is required',
          path: ['businessRegistration'],
        });
      }
      if (data.website && data.website.length > 0) {
        try {
          z.string().url().parse(data.website);
        } catch {
          ctx.addIssue({ code: 'custom', message: 'Invalid URL', path: ['website'] });
        }
      }
      if (!data.address || data.address.length < 1) {
        ctx.addIssue({
          code: 'custom',
          message: 'Office address is required',
          path: ['address'],
        });
      }
      if (!data.contact || data.contact.length < 1) {
        ctx.addIssue({
          code: 'custom',
          message: 'Contact person is required',
          path: ['contact'],
        });
      }
    }
  });
export type RegisterForm = z.infer<typeof registerSchema>;

export const profileSchema = z.object({
  headline: z.string().optional(),
  about: z.string().optional(),
  yearsOfExperience: z.number().min(0).optional(),
});
export type ProfileForm = z.infer<typeof profileSchema>;

export const experienceSchema = z.object({
  type: z.enum(['employment', 'gig', 'education', 'certification', 'project', 'volunteering']),
  title: z.string().min(1, 'Title is required'),
  organization: z.string().min(1, 'Organization is required'),
  startDate: z.string().min(1, 'Start date is required'),
  endDate: z.string().optional(),
  isCurrent: z.boolean().optional(),
  description: z.string().optional(),
  industry: z.string().optional(),
  skills: z.string().optional(),
  url: z.string().url('Invalid URL').or(z.literal('')).optional(),
});
export type ExperienceForm = z.infer<typeof experienceSchema>;

export const companyProfileSchema = z.object({
  companyName: z.string().min(1, 'Company name is required'),
  website: z.string().url('Invalid URL').or(z.literal('')),
  industry: z.string().min(1, 'Industry is required'),
  description: z.string().optional(),
});
export type CompanyProfileForm = z.infer<typeof companyProfileSchema>;

export const jobSchema = z.object({
  title: z.string().min(1, 'Title is required'),
  description: z.string().min(1, 'Description is required'),
  industry: z.string().min(1, 'Industry is required'),
  tags: z.string().optional(),
  requiredSkills: z.string().optional(),
  experienceLevel: z.enum(EXPERIENCE_VALUES),
  location: z.string().optional(),
  salaryRange: z.string().optional(),
});
export type JobForm = z.infer<typeof jobSchema>;

export const verificationSchema = z.object({
  businessRegistration: z.string().min(1, 'Business registration is required'),
  website: z.string().url('Invalid URL'),
  address: z.string().min(1, 'Address is required'),
  contact: z.string().min(1, 'Contact is required'),
});
export type VerificationForm = z.infer<typeof verificationSchema>;
