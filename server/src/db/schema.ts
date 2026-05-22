import { pgTable, text, uuid, timestamp, boolean, integer, jsonb, pgEnum } from 'drizzle-orm/pg-core';
import { relations } from 'drizzle-orm';

export const roleEnum = pgEnum('role', ['jobseeker', 'company']);
export const experienceTypeEnum = pgEnum('experience_type', ['employment', 'gig', 'education', 'certification', 'project', 'volunteering']);
export const verificationStatusEnum = pgEnum('verification_status', ['pending', 'verified', 'rejected']);
export const experienceLevelEnum = pgEnum('experience_level', ['entry', 'mid', 'senior', 'lead']);
export const jobStatusEnum = pgEnum('job_status', ['open', 'closed']);

export const users = pgTable('users', {
  id: uuid('id').defaultRandom().primaryKey(),
  email: text('email').unique().notNull(),
  username: text('username').unique().notNull(),
  passwordHash: text('password_hash').notNull(),
  role: roleEnum('role').notNull(),
  name: text('name').notNull(),
  avatarUrl: text('avatar_url'),
  isVerified: boolean('is_verified').default(false).notNull(),
  createdAt: timestamp('created_at').defaultNow().notNull(),
});

export const companies = pgTable('companies', {
  id: uuid('id').defaultRandom().primaryKey(),
  userId: uuid('user_id').references(() => users.id).notNull().unique(),
  companyName: text('company_name').notNull(),
  website: text('website'),
  industry: text('industry').notNull(),
  description: text('description'),
  verificationStatus: verificationStatusEnum('verification_status').default('pending').notNull(),
  verificationDocs: jsonb('verification_docs'),
  verifiedAt: timestamp('verified_at'),
  createdAt: timestamp('created_at').defaultNow().notNull(),
});

export const jobseekerProfiles = pgTable('jobseeker_profiles', {
  id: uuid('id').defaultRandom().primaryKey(),
  userId: uuid('user_id').references(() => users.id).notNull().unique(),
  headline: text('headline'),
  about: text('about'),
  yearsOfExperience: integer('years_of_experience'),
  slug: text('slug').unique().notNull(),
});

export const jobExperiences = pgTable('job_experiences', {
  id: uuid('id').defaultRandom().primaryKey(),
  profileId: uuid('profile_id').references(() => jobseekerProfiles.id).notNull(),
  type: experienceTypeEnum('type').notNull(),
  title: text('title').notNull(),
  organization: text('organization').notNull(),
  startDate: text('start_date').notNull(),
  endDate: text('end_date'),
  isCurrent: boolean('is_current').default(false).notNull(),
  description: text('description'),
  industry: text('industry'),
  skillsUsed: text('skills_used').array(),
  url: text('url'),
});

export const industryCategories = pgTable('industry_categories', {
  id: uuid('id').defaultRandom().primaryKey(),
  name: text('name').unique().notNull(),
  description: text('description'),
});

export const tags = pgTable('tags', {
  id: uuid('id').defaultRandom().primaryKey(),
  name: text('name').notNull(),
  industryCategoryId: uuid('industry_category_id').references(() => industryCategories.id),
});

export const jobPostings = pgTable('job_postings', {
  id: uuid('id').defaultRandom().primaryKey(),
  companyId: uuid('company_id').references(() => companies.id).notNull(),
  title: text('title').notNull(),
  description: text('description').notNull(),
  industry: text('industry').notNull(),
  tags: text('tags').array(),
  requiredSkills: text('required_skills').array(),
  experienceLevel: experienceLevelEnum('experience_level'),
  location: text('location'),
  salaryRange: text('salary_range'),
  status: jobStatusEnum('status').default('open').notNull(),
  createdAt: timestamp('created_at').defaultNow().notNull(),
});

export const usersRelations = relations(users, ({ one }) => ({
  company: one(companies, { fields: [users.id], references: [companies.userId] }),
  profile: one(jobseekerProfiles, { fields: [users.id], references: [jobseekerProfiles.userId] }),
}));

export const companiesRelations = relations(companies, ({ one, many }) => ({
  user: one(users, { fields: [companies.userId], references: [users.id] }),
  jobPostings: many(jobPostings),
}));

export const jobseekerProfilesRelations = relations(jobseekerProfiles, ({ one, many }) => ({
  user: one(users, { fields: [jobseekerProfiles.userId], references: [users.id] }),
  experiences: many(jobExperiences),
}));

export const jobExperiencesRelations = relations(jobExperiences, ({ one }) => ({
  profile: one(jobseekerProfiles, { fields: [jobExperiences.profileId], references: [jobseekerProfiles.id] }),
}));

export const jobPostingsRelations = relations(jobPostings, ({ one }) => ({
  company: one(companies, { fields: [jobPostings.companyId], references: [companies.id] }),
}));
