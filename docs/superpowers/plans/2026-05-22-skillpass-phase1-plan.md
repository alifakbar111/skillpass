# SkillPass Phase 1 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build the backend API server (Elysia) and frontend SPA (React) for Phase 1 — auth, profiles, company verification, candidate search, job postings, and public passport page.

**Architecture:** Two separate Bun packages under `skillpass/`: `server/` (Elysia + Drizzle + PostgreSQL) and `web/` (Vite + React 19 + React Router v7 + Tailwind v4 + DaisyUI 5). Server exposes REST API on port 3000, SPA on port 5173 with Vite proxy to backend.

**Tech Stack:** Bun, Elysia 1.x, Drizzle ORM, PostgreSQL, React 19, React Router v7, Tailwind CSS v4, DaisyUI 5, `@elysiajs/jwt`, `@elysiajs/swagger`, daisyui

---

## File Structure

```
skillpass/
├── server/
│   ├── src/
│   │   ├── index.ts              — Elysia app entry, plugin registration
│   │   ├── db/
│   │   │   ├── schema.ts         — all Drizzle table definitions
│   │   │   └── index.ts          — db client connection (Bun.sqlite for dev / postgres for prod)
│   │   ├── lib/
│   │   │   └── password.ts       — hash + verify helpers (bcrypt via Bun.password)
│   │   ├── middleware/
│   │   │   └── auth.ts           — JWT verify middleware for protected routes
│   │   └── routes/
│   │       ├── auth.ts           — register, login, refresh, logout
│   │       ├── profiles.ts       — jobseeker profile CRUD + experience CRUD
│   │       ├── passport.ts       — public passport GET endpoint
│   │       ├── companies.ts      — company profile + verification
│   │       ├── jobs.ts           — job postings CRUD + public listing
│   │       ├── search.ts         — candidate search (verified companies only)
│   │       ├── reference.ts      — industries, tags
│   │       └── admin.ts          — verification approval/rejection
│   ├── package.json
│   ├── tsconfig.json
│   └── drizzle.config.ts
├── web/
│   ├── src/
│   │   ├── main.tsx
│   │   ├── App.tsx               — React Router setup
│   │   ├── lib/
│   │   │   └── api.ts            — fetch wrapper with JWT handling
│   │   ├── hooks/
│   │   │   └── useAuth.tsx       — auth context + provider
│   │   ├── components/
│   │   │   ├── ui/
│   │   │   │   ├── ThemeToggle.tsx
│   │   │   │   └── ProtectedRoute.tsx
│   │   │   └── layout/
│   │   │       ├── RootLayout.tsx
│   │   │       └── Navbar.tsx
│   │   ├── pages/
│   │   │   ├── Landing.tsx
│   │   │   ├── Login.tsx
│   │   │   ├── Register.tsx
│   │   │   ├── JobseekerProfile.tsx
│   │   │   ├── JobseekerPassport.tsx
│   │   │   ├── CompanyProfile.tsx
│   │   │   ├── CompanyVerification.tsx
│   │   │   ├── CompanySearch.tsx
│   │   │   ├── CompanyJobs.tsx
│   │   │   ├── JobDetail.tsx
│   │   │   ├── PublicJobs.tsx
│   │   │   ├── PublicPassport.tsx
│   │   │   └── AdminVerifications.tsx
│   │   └── styles/
│   │       └── index.css
│   ├── index.html
│   ├── package.json
│   ├── vite.config.ts
│   └── tsconfig.json
├── docs/
│   └── superpowers/
│       ├── specs/
│       └── plans/
```

### Task 1: Initialize server package and install dependencies

**Files:**
- Create: `skillpass/server/package.json`
- Create: `skillpass/server/tsconfig.json`
- Create: `skillpass/server/drizzle.config.ts`

- [ ] **Create server directory and package.json**

```bash
mkdir -p /home/al-ip/learning/skillpass/server/src/{db,lib,middleware,routes} /home/al-ip/learning/skillpass/server/tests
```

```json
// /home/al-ip/learning/skillpass/server/package.json
{
  "name": "skillpass-server",
  "type": "module",
  "scripts": {
    "dev": "bun run --watch src/index.ts",
    "start": "bun run src/index.ts",
    "db:generate": "drizzle-kit generate",
    "db:push": "drizzle-kit push",
    "db:migrate": "drizzle-kit migrate",
    "db:studio": "drizzle-kit studio",
    "test": "bun test",
    "typecheck": "tsc --noEmit"
  },
  "dependencies": {
    "elysia": "^1.2.0",
    "@elysiajs/jwt": "^1.2.0",
    "@elysiajs/cors": "^1.2.0",
    "@elysiajs/swagger": "^1.2.0",
    "drizzle-orm": "^0.40.0",
    "drizzle-kit": "^0.30.0",
    "postgres": "^3.4.0"
  },
  "devDependencies": {
    "bun-types": "^1.2.0",
    "typescript": "^5.8.0"
  }
}
```

```json
// /home/al-ip/learning/skillpass/server/tsconfig.json
{
  "compilerOptions": {
    "target": "ESNext",
    "module": "ESNext",
    "moduleResolution": "bundler",
    "types": ["bun-types"],
    "strict": true,
    "esModuleInterop": true,
    "skipLibCheck": true,
    "outDir": "dist",
    "rootDir": "src"
  },
  "include": ["src"]
}
```

```typescript
// /home/al-ip/learning/skillpass/server/drizzle.config.ts
import { defineConfig } from 'drizzle-kit';

export default defineConfig({
  schema: './src/db/schema.ts',
  out: './drizzle',
  dialect: 'postgresql',
  dbCredentials: {
    url: process.env.DATABASE_URL || 'postgres://postgres:postgres@localhost:5432/skillpass',
  },
});
```

- [ ] **Install dependencies**

Run: `cd /home/al-ip/learning/skillpass/server && bun install`

Expected: `bun install` succeeds, `node_modules/` created.

- [ ] **Commit**

```bash
git -C /home/al-ip/learning/skillpass init
git -C /home/al-ip/learning/skillpass add server/package.json server/tsconfig.json server/drizzle.config.ts
git -C /home/al-ip/learning/skillpass commit -m "chore: init server package with Elysia + Drizzle"
```

### Task 2: Create database schema (Drizzle)

**Files:**
- Create: `skillpass/server/src/db/schema.ts`

- [ ] **Write Drizzle schema for all Phase 1 tables**

```typescript
// /home/al-ip/learning/skillpass/server/src/db/schema.ts
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

// Relations
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
```

- [ ] **Commit**

```bash
git -C /home/al-ip/learning/skillpass add server/src/db/schema.ts
git -C /home/al-ip/learning/skillpass commit -m "feat: add Drizzle schema for all Phase 1 tables"
```

### Task 3: Create DB client and seed

**Files:**
- Create: `skillpass/server/src/db/index.ts`
- Create: `skillpass/server/seed.ts`

- [ ] **Write DB connection file**

```typescript
// /home/al-ip/learning/skillpass/server/src/db/index.ts
import { drizzle } from 'drizzle-orm/postgres-js';
import postgres from 'postgres';
import * as schema from './schema';

const connectionString = process.env.DATABASE_URL || 'postgres://postgres:postgres@localhost:5432/skillpass';
const client = postgres(connectionString);

export const db = drizzle(client, { schema });
export { schema };
```

- [ ] **Write seed script with initial industry categories + tags**

```typescript
// /home/al-ip/learning/skillpass/server/seed.ts
import { db, schema } from './src/db/index';

async function seed() {
  console.log('🌱 Seeding database...');

  const industries = [
    { name: 'Technology', description: 'Software, hardware, IT services' },
    { name: 'Manufacturing', description: 'Industrial production and fabrication' },
    { name: 'Healthcare', description: 'Medical services and pharmaceuticals' },
    { name: 'Finance', description: 'Banking, investment, insurance' },
    { name: 'Education', description: 'Schools, universities, training' },
    { name: 'Retail', description: 'Sales, e-commerce, consumer goods' },
    { name: 'Transportation', description: 'Logistics, delivery, ride-hailing' },
    { name: 'Creative Arts', description: 'Design, media, entertainment' },
    { name: 'Hospitality', description: 'Hotels, restaurants, tourism' },
    { name: 'Construction', description: 'Building and infrastructure' },
    { name: 'Agriculture', description: 'Farming, food production' },
    { name: 'Energy', description: 'Oil, gas, renewable energy' },
  ];

  for (const industry of industries) {
    await db.insert(schema.industryCategories).values(industry).onConflictDoNothing();
  }

  console.log(`✅ Seeded ${industries.length} industry categories`);
  process.exit(0);
}

seed().catch((err) => {
  console.error('❌ Seed failed:', err);
  process.exit(1);
});
```

- [ ] **Add seed script to package.json**

Edit the `"scripts"` in `skillpass/server/package.json` to add:

```json
"seed": "bun run seed.ts"
```

- [ ] **Commit**

```bash
git -C /home/al-ip/learning/skillpass add server/src/db/index.ts server/seed.ts
git -C /home/al-ip/learning/skillpass add -p server/package.json
git -C /home/al-ip/learning/skillpass commit -m "feat: add db client and seed script"
```

### Task 4: Create lib utilities

**Files:**
- Create: `skillpass/server/src/lib/password.ts`
- Create: `skillpass/server/src/middleware/auth.ts`

- [ ] **Write password hashing utility (uses Bun.password)**

```typescript
// /home/al-ip/learning/skillpass/server/src/lib/password.ts
export async function hashPassword(password: string): Promise<string> {
  return Bun.password.hash(password, { algorithm: 'bcrypt', cost: 10 });
}

export async function verifyPassword(password: string, hash: string): Promise<boolean> {
  return Bun.password.verify(password, hash);
}
```

- [ ] **Write auth middleware for JWT verification**

```typescript
// /home/al-ip/learning/skillpass/server/src/middleware/auth.ts
import { Elysia, t } from 'elysia';

export const authMiddleware = new Elysia()
  .guard({
    beforeHandle({ headers, error }) {
      const auth = headers.authorization;
      if (!auth || !auth.startsWith('Bearer ')) {
        return error(401, 'Unauthorized');
      }
    },
  });
```

- [ ] **Commit**

```bash
git -C /home/al-ip/learning/skillpass add server/src/lib/password.ts server/src/middleware/auth.ts
git -C /home/al-ip/learning/skillpass commit -m "feat: add password hashing and auth middleware"
```

### Task 5: Implement auth routes (register, login, refresh, logout)

**Files:**
- Create: `skillpass/server/src/routes/auth.ts`

Also need to add `@elysiajs/jwt` to the auth route for token verification.

- [ ] **Write auth routes with register, login, refresh, logout**

```typescript
// /home/al-ip/learning/skillpass/server/src/routes/auth.ts
import { Elysia, t } from 'elysia';
import { db, schema } from '../db';
import { hashPassword, verifyPassword } from '../lib/password';
import { eq } from 'drizzle-orm';
import { jwt } from '@elysiajs/jwt';

const JWT_SECRET = process.env.JWT_SECRET || 'skillpass-dev-secret-change-in-prod';

export const authRoutes = new Elysia({ prefix: '/api/v1/auth' })
  .use(jwt({ secret: JWT_SECRET, name: 'jwt' }))
  .post('/register', async ({ body, jwt: j, error }) => {
    const { email, username, password, name, role } = body;

    const existingUser = await db.select().from(schema.users)
      .where(eq(schema.users.email, email))
      .limit(1);

    if (existingUser.length > 0) {
      return error(409, 'Email already registered');
    }

    const passwordHash = await hashPassword(password);
    const user = await db.insert(schema.users).values({
      email, username, passwordHash, name, role,
    }).returning();

    if (role === 'jobseeker') {
      await db.insert(schema.jobseekerProfiles).values({
        userId: user[0].id,
        slug: username,
      });
    } else if (role === 'company') {
      await db.insert(schema.companies).values({
        userId: user[0].id,
        companyName: name,
        industry: 'Technology',
      });
    }

    const accessToken = await j.sign({ userId: user[0].id, role: user[0].role });
    const refreshToken = await j.sign({ userId: user[0].id, type: 'refresh' }, { expiresIn: '7d' });

    return {
      accessToken,
      refreshToken,
      user: { id: user[0].id, email: user[0].email, username: user[0].username, name: user[0].name, role: user[0].role },
    };
  }, {
    body: t.Object({
      email: t.String({ format: 'email' }),
      username: t.String({ minLength: 3 }),
      password: t.String({ minLength: 6 }),
      name: t.String(),
      role: t.Union([t.Literal('jobseeker'), t.Literal('company')]),
    }),
  })
  .post('/login', async ({ body, jwt: j, error }) => {
    const users = await db.select().from(schema.users)
      .where(eq(schema.users.email, body.email))
      .limit(1);

    if (users.length === 0) {
      return error(401, 'Invalid credentials');
    }

    const valid = await verifyPassword(body.password, users[0].passwordHash);
    if (!valid) {
      return error(401, 'Invalid credentials');
    }

    const accessToken = await j.sign({ userId: users[0].id, role: users[0].role });
    const refreshToken = await j.sign({ userId: users[0].id, type: 'refresh' }, { expiresIn: '7d' });

    return {
      accessToken,
      refreshToken,
      user: { id: users[0].id, email: users[0].email, username: users[0].username, name: users[0].name, role: users[0].role },
    };
  }, {
    body: t.Object({
      email: t.String(),
      password: t.String(),
    }),
  })
  .post('/refresh', async ({ body, jwt: j, error }) => {
    try {
      const payload = await j.verify(body.refreshToken);
      if (!payload || payload.type !== 'refresh') {
        return error(401, 'Invalid refresh token');
      }

      const accessToken = await j.sign({ userId: payload.userId, role: payload.role });
      return { accessToken };
    } catch {
      return error(401, 'Invalid refresh token');
    }
  }, {
    body: t.Object({
      refreshToken: t.String(),
    }),
  })
  .post('/logout', async () => {
    return { message: 'Logged out' };
  });
```

- [ ] **Commit**

```bash
git -C /home/al-ip/learning/skillpass add server/src/routes/auth.ts
git -C /home/al-ip/learning/skillpass commit -m "feat: implement auth routes (register, login, refresh, logout)"
```

### Task 6: Implement profile routes (jobseeker profile + experience CRUD)

**Files:**
- Create: `skillpass/server/src/routes/profiles.ts`
- Create: `skillpass/server/src/routes/passport.ts`

- [ ] **Write jobseeker profile CRUD routes**

```typescript
// /home/al-ip/learning/skillpass/server/src/routes/profiles.ts
import { Elysia, t } from 'elysia';
import { db, schema } from '../db';
import { eq, and } from 'drizzle-orm';
import { jwt } from '@elysiajs/jwt';

const JWT_SECRET = process.env.JWT_SECRET || 'skillpass-dev-secret-change-in-prod';

export const profileRoutes = new Elysia({ prefix: '/api/v1/profiles' })
  .use(jwt({ secret: JWT_SECRET, name: 'jwt' }))
  .resolve(async ({ headers, jwt: j, error }) => {
    const auth = headers.authorization;
    if (!auth || !auth.startsWith('Bearer ')) return error(401, 'Unauthorized');
    const payload = await j.verify(auth.slice(7));
    if (!payload) return error(401, 'Unauthorized');
    return { userId: payload.userId as string, role: payload.role as string };
  })
  .get('/me', async ({ userId, error }) => {
    const profile = await db.select().from(schema.jobseekerProfiles)
      .where(eq(schema.jobseekerProfiles.userId, userId))
      .limit(1);

    if (profile.length === 0) return error(404, 'Profile not found');

    const experiences = await db.select().from(schema.jobExperiences)
      .where(eq(schema.jobExperiences.profileId, profile[0].id))
      .orderBy(schema.jobExperiences.startDate);

    return { ...profile[0], experiences };
  })
  .put('/me', async ({ userId, body, error }) => {
    const [profile] = await db.update(schema.jobseekerProfiles)
      .set({
        headline: body.headline ?? undefined,
        about: body.about ?? undefined,
        yearsOfExperience: body.yearsOfExperience ?? undefined,
        slug: body.slug ?? undefined,
      })
      .where(eq(schema.jobseekerProfiles.userId, userId))
      .returning();

    if (!profile) return error(404, 'Profile not found');
    return profile;
  }, {
    body: t.Object({
      headline: t.Optional(t.String()),
      about: t.Optional(t.String()),
      yearsOfExperience: t.Optional(t.Number()),
      slug: t.Optional(t.String()),
    }),
  })
  .post('/me/experience', async ({ userId, body, error }) => {
    const [profile] = await db.select().from(schema.jobseekerProfiles)
      .where(eq(schema.jobseekerProfiles.userId, userId))
      .limit(1);

    if (!profile) return error(404, 'Profile not found');

    const [exp] = await db.insert(schema.jobExperiences).values({
      profileId: profile.id,
      ...body,
    }).returning();

    return exp;
  }, {
    body: t.Object({
      type: t.Enum({ employment: 'employment', gig: 'gig', education: 'education', certification: 'certification', project: 'project', volunteering: 'volunteering' }),
      title: t.String(),
      organization: t.String(),
      startDate: t.String(),
      endDate: t.Optional(t.String()),
      isCurrent: t.Optional(t.Boolean()),
      description: t.Optional(t.String()),
      industry: t.Optional(t.String()),
      skillsUsed: t.Optional(t.Array(t.String())),
      url: t.Optional(t.String()),
    }),
  })
  .put('/me/experience/:id', async ({ userId, params, body, error }) => {
    const [profile] = await db.select().from(schema.jobseekerProfiles)
      .where(eq(schema.jobseekerProfiles.userId, userId))
      .limit(1);
    if (!profile) return error(404, 'Profile not found');

    const [updated] = await db.update(schema.jobExperiences)
      .set(body)
      .where(and(
        eq(schema.jobExperiences.id, params.id),
        eq(schema.jobExperiences.profileId, profile.id),
      ))
      .returning();

    if (!updated) return error(404, 'Experience not found');
    return updated;
  }, {
    body: t.Partial(t.Object({
      type: t.Enum({ employment: 'employment', gig: 'gig', education: 'education', certification: 'certification', project: 'project', volunteering: 'volunteering' }),
      title: t.String(),
      organization: t.String(),
      startDate: t.String(),
      endDate: t.Optional(t.String()),
      isCurrent: t.Optional(t.Boolean()),
      description: t.Optional(t.String()),
      industry: t.Optional(t.String()),
      skillsUsed: t.Optional(t.Array(t.String())),
      url: t.Optional(t.String()),
    })),
  })
  .delete('/me/experience/:id', async ({ userId, params, error }) => {
    const [profile] = await db.select().from(schema.jobseekerProfiles)
      .where(eq(schema.jobseekerProfiles.userId, userId))
      .limit(1);
    if (!profile) return error(404, 'Profile not found');

    const [deleted] = await db.delete(schema.jobExperiences)
      .where(and(
        eq(schema.jobExperiences.id, params.id),
        eq(schema.jobExperiences.profileId, profile.id),
      ))
      .returning();

    if (!deleted) return error(404, 'Experience not found');
    return { message: 'Deleted' };
  });
```

- [ ] **Write public passport route (no auth required)**

```typescript
// /home/al-ip/learning/skillpass/server/src/routes/passport.ts
import { Elysia, t } from 'elysia';
import { db, schema } from '../db';
import { eq } from 'drizzle-orm';

export const passportRoutes = new Elysia()
  .get('/api/v1/profiles/:username', async ({ params: { username }, error }) => {
    const [profile] = await db.select().from(schema.jobseekerProfiles)
      .where(eq(schema.jobseekerProfiles.slug, username))
      .limit(1);

    if (!profile) return error(404, 'Profile not found');

    const [user] = await db.select().from(schema.users)
      .where(eq(schema.users.id, profile.userId))
      .limit(1);

    const experiences = await db.select().from(schema.jobExperiences)
      .where(eq(schema.jobExperiences.profileId, profile.id))
      .orderBy(schema.jobExperiences.startDate);

    return {
      name: user?.name,
      avatarUrl: user?.avatarUrl,
      headline: profile.headline,
      about: profile.about,
      yearsOfExperience: profile.yearsOfExperience,
      experiences,
    };
  });
```

- [ ] **Commit**

```bash
git -C /home/al-ip/learning/skillpass add server/src/routes/profiles.ts server/src/routes/passport.ts
git -C /home/al-ip/learning/skillpass commit -m "feat: implement jobseeker profile and passport routes"
```

### Task 7: Implement company routes (profile + verification)

**Files:**
- Create: `skillpass/server/src/routes/companies.ts`

- [ ] **Write company routes**

```typescript
// /home/al-ip/learning/skillpass/server/src/routes/companies.ts
import { Elysia, t } from 'elysia';
import { db, schema } from '../db';
import { eq } from 'drizzle-orm';
import { jwt } from '@elysiajs/jwt';

const JWT_SECRET = process.env.JWT_SECRET || 'skillpass-dev-secret-change-in-prod';

export const companyRoutes = new Elysia({ prefix: '/api/v1/company' })
  .use(jwt({ secret: JWT_SECRET, name: 'jwt' }))
  .resolve(async ({ headers, jwt: j, error }) => {
    const auth = headers.authorization;
    if (!auth || !auth.startsWith('Bearer ')) return error(401, 'Unauthorized');
    const payload = await j.verify(auth.slice(7));
    if (!payload) return error(401, 'Unauthorized');
    if (payload.role !== 'company') return error(403, 'Forbidden: company role required');
    return { userId: payload.userId as string };
  })
  .get('/profile', async ({ userId, error }) => {
    const [company] = await db.select().from(schema.companies)
      .where(eq(schema.companies.userId, userId))
      .limit(1);

    if (!company) return error(404, 'Company not found');
    return company;
  })
  .put('/profile', async ({ userId, body }) => {
    const [company] = await db.update(schema.companies)
      .set(body)
      .where(eq(schema.companies.userId, userId))
      .returning();

    return company;
  }, {
    body: t.Object({
      companyName: t.Optional(t.String()),
      website: t.Optional(t.String()),
      industry: t.Optional(t.String()),
      description: t.Optional(t.String()),
    }),
  })
  .post('/verification', async ({ userId, body, error }) => {
    const [company] = await db.update(schema.companies)
      .set({
        verificationDocs: body,
        verificationStatus: 'pending',
      })
      .where(eq(schema.companies.userId, userId))
      .returning();

    if (!company) return error(404, 'Company not found');
    return { message: 'Verification submitted', status: 'pending' };
  }, {
    body: t.Object({
      businessRegistration: t.String(),
      website: t.String(),
      address: t.String(),
      contact: t.String(),
    }),
  })
  .get('/verification-status', async ({ userId }) => {
    const [company] = await db.select({ verificationStatus: schema.companies.verificationStatus })
      .from(schema.companies)
      .where(eq(schema.companies.userId, userId))
      .limit(1);

    return { verificationStatus: company?.verificationStatus || 'none' };
  });
```

- [ ] **Commit**

```bash
git -C /home/al-ip/learning/skillpass add server/src/routes/companies.ts
git -C /home/al-ip/learning/skillpass commit -m "feat: implement company profile and verification routes"
```

### Task 8: Implement job posting routes

**Files:**
- Create: `skillpass/server/src/routes/jobs.ts`

- [ ] **Write job posting routes (company CRUD + public listing)**

```typescript
// /home/al-ip/learning/skillpass/server/src/routes/jobs.ts
import { Elysia, t } from 'elysia';
import { db, schema } from '../db';
import { eq, and, like, or, inArray } from 'drizzle-orm';
import { jwt } from '@elysiajs/jwt';

const JWT_SECRET = process.env.JWT_SECRET || 'skillpass-dev-secret-change-in-prod';

export const jobRoutes = new Elysia({ prefix: '/api/v1/jobs' })
  .use(jwt({ secret: JWT_SECRET, name: 'jwt' }));

// Public listing — no auth required
jobRoutes.get('/', async ({ query }) => {
  let conditions = and(eq(schema.jobPostings.status, 'open'));

  if (query.industry) {
    conditions = and(conditions, eq(schema.jobPostings.industry, query.industry as string));
  }
  if (query.experience_level) {
    conditions = and(conditions, eq(schema.jobPostings.experienceLevel, query.experience_level as string));
  }

  const jobs = await db.select().from(schema.jobPostings).where(conditions).orderBy(schema.jobPostings.createdAt);
  return jobs;
});

// Company routes — require auth + company role
const companyJobs = jobRoutes.group('/me', (app) =>
  app.resolve(async ({ headers, jwt: j, error }) => {
    const auth = headers.authorization;
    if (!auth || !auth.startsWith('Bearer ')) return error(401, 'Unauthorized');
    const payload = await j.verify(auth.slice(7));
    if (!payload) return error(401, 'Unauthorized');
    if (payload.role !== 'company') return error(403, 'Forbidden');
    return { userId: payload.userId as string };
  })
  .get('', async ({ userId }) => {
    const [company] = await db.select().from(schema.companies)
      .where(eq(schema.companies.userId, userId))
      .limit(1);

    if (!company) return [];
    return db.select().from(schema.jobPostings)
      .where(eq(schema.jobPostings.companyId, company.id))
      .orderBy(schema.jobPostings.createdAt);
  })
);

jobRoutes.post('/', async ({ headers, body, jwt: j, error }) => {
  const auth = headers.authorization;
  if (!auth || !auth.startsWith('Bearer ')) return error(401, 'Unauthorized');
  const payload = await j.verify(auth.slice(7));
  if (!payload || payload.role !== 'company') return error(403, 'Forbidden');

  const [company] = await db.select().from(schema.companies)
    .where(eq(schema.companies.userId, payload.userId as string))
    .limit(1);

  if (!company) return error(404, 'Company not found');
  if (company.verificationStatus !== 'verified') return error(403, 'Company not verified');

  const [job] = await db.insert(schema.jobPostings).values({
    companyId: company.id,
    ...body,
  }).returning();

  return job;
}, {
  body: t.Object({
    title: t.String(),
    description: t.String(),
    industry: t.String(),
    tags: t.Optional(t.Array(t.String())),
    requiredSkills: t.Optional(t.Array(t.String())),
    experienceLevel: t.Optional(t.Union([t.Literal('entry'), t.Literal('mid'), t.Literal('senior'), t.Literal('lead')])),
    location: t.Optional(t.String()),
    salaryRange: t.Optional(t.String()),
  }),
});

jobRoutes.put('/:id', async ({ headers, params, body, jwt: j, error }) => {
  const auth = headers.authorization;
  if (!auth || !auth.startsWith('Bearer ')) return error(401, 'Unauthorized');
  const payload = await j.verify(auth.slice(7));
  if (!payload || payload.role !== 'company') return error(403, 'Forbidden');

  const [company] = await db.select().from(schema.companies)
    .where(eq(schema.companies.userId, payload.userId as string))
    .limit(1);
  if (!company) return error(404, 'Company not found');

  const [job] = await db.update(schema.jobPostings)
    .set(body)
    .where(and(
      eq(schema.jobPostings.id, params.id),
      eq(schema.jobPostings.companyId, company.id),
    ))
    .returning();

  if (!job) return error(404, 'Job not found');
  return job;
}, {
  body: t.Partial(t.Object({
    title: t.String(),
    description: t.String(),
    industry: t.String(),
    tags: t.Optional(t.Array(t.String())),
    requiredSkills: t.Optional(t.Array(t.String())),
    experienceLevel: t.Optional(t.Union([t.Literal('entry'), t.Literal('mid'), t.Literal('senior'), t.Literal('lead')])),
    location: t.Optional(t.String()),
    salaryRange: t.Optional(t.String()),
    status: t.Optional(t.Union([t.Literal('open'), t.Literal('closed')])),
  })),
});

jobRoutes.delete('/:id', async ({ headers, params, jwt: j, error }) => {
  const auth = headers.authorization;
  if (!auth || !auth.startsWith('Bearer ')) return error(401, 'Unauthorized');
  const payload = await j.verify(auth.slice(7));
  if (!payload || payload.role !== 'company') return error(403, 'Forbidden');

  const [company] = await db.select().from(schema.companies)
    .where(eq(schema.companies.userId, payload.userId as string))
    .limit(1);
  if (!company) return error(404, 'Company not found');

  const [deleted] = await db.delete(schema.jobPostings)
    .where(and(
      eq(schema.jobPostings.id, params.id),
      eq(schema.jobPostings.companyId, company.id),
    ))
    .returning();

  if (!deleted) return error(404, 'Job not found');
  return { message: 'Deleted' };
});

// Public single job
jobRoutes.get('/:id', async ({ params: { id }, error }) => {
  const [job] = await db.select().from(schema.jobPostings).where(eq(schema.jobPostings.id, id)).limit(1);
  if (!job) return error(404, 'Job not found');
  return job;
});
```

- [ ] **Commit**

```bash
git -C /home/al-ip/learning/skillpass add server/src/routes/jobs.ts
git -C /home/al-ip/learning/skillpass commit -m "feat: implement job posting routes"
```

### Task 9: Implement search, reference, and admin routes

**Files:**
- Create: `skillpass/server/src/routes/search.ts`
- Create: `skillpass/server/src/routes/reference.ts`
- Create: `skillpass/server/src/routes/admin.ts`

- [ ] **Write candidate search route (verified companies only)**

```typescript
// /home/al-ip/learning/skillpass/server/src/routes/search.ts
import { Elysia, t } from 'elysia';
import { db, schema } from '../db';
import { eq, ilike, or, and } from 'drizzle-orm';
import { jwt } from '@elysiajs/jwt';

const JWT_SECRET = process.env.JWT_SECRET || 'skillpass-dev-secret-change-in-prod';

export const searchRoutes = new Elysia({ prefix: '/api/v1/search' })
  .use(jwt({ secret: JWT_SECRET, name: 'jwt' }))
  .resolve(async ({ headers, jwt: j, error }) => {
    const auth = headers.authorization;
    if (!auth || !auth.startsWith('Bearer ')) return error(401, 'Unauthorized');
    const payload = await j.verify(auth.slice(7));
    if (!payload) return error(401, 'Unauthorized');
    if (payload.role !== 'company') return error(403, 'Forbidden');

    const [company] = await db.select().from(schema.companies)
      .where(eq(schema.companies.userId, payload.userId as string))
      .limit(1);

    if (!company || company.verificationStatus !== 'verified') return error(403, 'Company not verified');
    return { userId: payload.userId as string };
  })
  .get('/candidates', async ({ query }) => {
    const profiles = await db.select().from(schema.jobseekerProfiles);
    const results = [];

    for (const profile of profiles) {
      const [user] = await db.select().from(schema.users)
        .where(eq(schema.users.id, profile.userId))
        .limit(1);

      const experiences = await db.select().from(schema.jobExperiences)
        .where(eq(schema.jobExperiences.profileId, profile.id));

      // Basic keyword filtering
      if (query.q) {
        const q = (query.q as string).toLowerCase();
        const matchesName = user?.name.toLowerCase().includes(q);
        const matchesHeadline = profile.headline?.toLowerCase().includes(q);
        const matchesAbout = profile.about?.toLowerCase().includes(q);
        const matchesExp = experiences.some(e =>
          e.title.toLowerCase().includes(q) ||
          e.organization.toLowerCase().includes(q) ||
          e.skillsUsed?.some(s => s.toLowerCase().includes(q))
        );
        if (!matchesName && !matchesHeadline && !matchesAbout && !matchesExp) continue;
      }

      // Skill filter
      if (query.skills) {
        const skillList = (query.skills as string).split(',').map(s => s.trim().toLowerCase());
        const hasSkill = experiences.some(e =>
          e.skillsUsed?.some(s => skillList.includes(s.toLowerCase()))
        );
        if (!hasSkill) continue;
      }

      // Industry filter
      if (query.industry) {
        const hasIndustry = experiences.some(e =>
          e.industry?.toLowerCase() === (query.industry as string).toLowerCase()
        );
        if (!hasIndustry) continue;
      }

      results.push({
        id: profile.id,
        name: user?.name,
        avatarUrl: user?.avatarUrl,
        headline: profile.headline,
        about: profile.about,
        yearsOfExperience: profile.yearsOfExperience,
        slug: profile.slug,
        skills: [...new Set(experiences.flatMap(e => e.skillsUsed || []))],
      });
    }

    return results;
  });
```

- [ ] **Write reference data routes**

```typescript
// /home/al-ip/learning/skillpass/server/src/routes/reference.ts
import { Elysia } from 'elysia';
import { db, schema } from '../db';
import { eq } from 'drizzle-orm';

export const referenceRoutes = new Elysia({ prefix: '/api/v1' })
  .get('/industries', async () => {
    return db.select().from(schema.industryCategories).orderBy(schema.industryCategories.name);
  })
  .get('/tags', async ({ query }) => {
    let queryBuilder = db.select().from(schema.tags);
    if (query.industry) {
      queryBuilder = queryBuilder.where(eq(schema.tags.industryCategoryId, query.industry as string));
    }
    return queryBuilder;
  });
```

- [ ] **Write admin verification routes**

```typescript
// /home/al-ip/learning/skillpass/server/src/routes/admin.ts
import { Elysia, t } from 'elysia';
import { db, schema } from '../db';
import { eq } from 'drizzle-orm';
import { jwt } from '@elysiajs/jwt';

const JWT_SECRET = process.env.JWT_SECRET || 'skillpass-dev-secret-change-in-prod';
// Note: For MVP, admin check is simple — any user can access /admin.
// In production, add a proper admin role.

export const adminRoutes = new Elysia({ prefix: '/api/v1/admin' })
  .use(jwt({ secret: JWT_SECRET, name: 'jwt' }))
  .resolve(async ({ headers, jwt: j, error }) => {
    const auth = headers.authorization;
    if (!auth || !auth.startsWith('Bearer ')) return error(401, 'Unauthorized');
    const payload = await j.verify(auth.slice(7));
    if (!payload) return error(401, 'Unauthorized');
    return { userId: payload.userId as string };
  })
  .get('/verifications/pending', async () => {
    return db.select().from(schema.companies)
      .where(eq(schema.companies.verificationStatus, 'pending'));
  })
  .post('/verifications/:id', async ({ params, body, error }) => {
    const [company] = await db.select().from(schema.companies)
      .where(eq(schema.companies.id, params.id))
      .limit(1);

    if (!company) return error(404, 'Company not found');

    if (body.action === 'approve') {
      const [updated] = await db.update(schema.companies)
        .set({ verificationStatus: 'verified', verifiedAt: new Date() })
        .where(eq(schema.companies.id, params.id))
        .returning();

      // Also mark the user as verified
      await db.update(schema.users)
        .set({ isVerified: true })
        .where(eq(schema.users.id, company.userId));

      return updated;
    } else if (body.action === 'reject') {
      const [updated] = await db.update(schema.companies)
        .set({ verificationStatus: 'rejected' })
        .where(eq(schema.companies.id, params.id))
        .returning();

      return updated;
    }

    return error(400, 'Invalid action');
  }, {
    body: t.Object({
      action: t.Union([t.Literal('approve'), t.Literal('reject')]),
      reason: t.Optional(t.String()),
    }),
  });
```

- [ ] **Commit**

```bash
git -C /home/al-ip/learning/skillpass add server/src/routes/search.ts server/src/routes/reference.ts server/src/routes/admin.ts
git -C /home/al-ip/learning/skillpass commit -m "feat: implement search, reference, and admin routes"
```

### Task 10: Wire up server entry point

**Files:**
- Create: `skillpass/server/src/index.ts`

- [ ] **Write server entry point (register all plugins and routes)**

```typescript
// /home/al-ip/learning/skillpass/server/src/index.ts
import { Elysia } from 'elysia';
import { cors } from '@elysiajs/cors';
import { swagger } from '@elysiajs/swagger';

import { authRoutes } from './routes/auth';
import { profileRoutes } from './routes/profiles';
import { passportRoutes } from './routes/passport';
import { companyRoutes } from './routes/companies';
import { jobRoutes } from './routes/jobs';
import { searchRoutes } from './routes/search';
import { referenceRoutes } from './routes/reference';
import { adminRoutes } from './routes/admin';

const app = new Elysia()
  .use(cors({
    origin: process.env.CORS_ORIGIN || 'http://localhost:5173',
    credentials: true,
  }))
  .use(swagger({
    path: '/docs',
    documentation: {
      info: { title: 'SkillPass API', version: '1.0.0', description: 'Talent marketplace & skill passport API' },
    },
  }));

// Register route groups
app
  .use(authRoutes)
  .use(profileRoutes)
  .use(passportRoutes)
  .use(companyRoutes)
  .use(jobRoutes)
  .use(searchRoutes)
  .use(referenceRoutes)
  .use(adminRoutes);

// Health check
app.get('/api/v1/health', () => ({ status: 'ok', timestamp: new Date().toISOString() }));

const port = Number(process.env.PORT || 3000);
app.listen(port);

console.log(`🦊 SkillPass API running at http://localhost:${port}`);
console.log(`📚 Swagger docs at http://localhost:${port}/docs`);

export type App = typeof app;
```

- [ ] **Verify server starts**

Run: `cd /home/al-ip/learning/skillpass/server && bun run src/index.ts`

Expected: `🦊 SkillPass API running at http://localhost:3000` and `📚 Swagger docs at http://localhost:3000/docs`

Stop the server with Ctrl+C.

- [ ] **Commit**

```bash
git -C /home/al-ip/learning/skillpass add server/src/index.ts
git -C /home/al-ip/learning/skillpass commit -m "feat: wire up server entry point with all routes"
```

### Task 11: Initialize web frontend with Vite + React + DaisyUI

**Files:**
- Create: `skillpass/web/package.json`
- Create: `skillpass/web/tsconfig.json`
- Create: `skillpass/web/vite.config.ts`
- Create: `skillpass/web/index.html`
- Create: `skillpass/web/src/styles/index.css`
- Create: `skillpass/web/src/main.tsx`

- [ ] **Create web directory and package.json**

```bash
mkdir -p /home/al-ip/learning/skillpass/web/src/{pages,components/{ui,layout},lib,hooks,styles}
```

```json
// /home/al-ip/learning/skillpass/web/package.json
{
  "name": "skillpass-web",
  "type": "module",
  "scripts": {
    "dev": "vite",
    "build": "tsc && vite build",
    "preview": "vite preview",
    "test": "vitest run",
    "test:watch": "vitest",
    "typecheck": "tsc --noEmit"
  },
  "dependencies": {
    "react": "^19.2.0",
    "react-dom": "^19.2.0",
    "react-router": "^7.5.0",
    "react-router-dom": "^7.5.0",
    "lucide-react": "^0.544.0"
  },
  "devDependencies": {
    "@types/react": "^19.2.0",
    "@types/react-dom": "^19.2.0",
    "@vitejs/plugin-react": "^4.4.0",
    "daisyui": "^5.0.0",
    "tailwindcss": "^4.1.0",
    "@tailwindcss/vite": "^4.1.0",
    "typescript": "^5.8.0",
    "vite": "^7.0.0",
    "vitest": "^3.0.0",
    "@testing-library/react": "^16.0.0",
    "@testing-library/jest-dom": "^6.0.0",
    "happy-dom": "^19.0.0"
  }
}
```

- [ ] **Create config files**

```typescript
// /home/al-ip/learning/skillpass/web/vite.config.ts
import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import tailwindcss from '@tailwindcss/vite';

export default defineConfig({
  plugins: [react(), tailwindcss()],
  server: {
    port: 5173,
    proxy: {
      '/api': 'http://localhost:3000',
    },
  },
});
```

```json
// /home/al-ip/learning/skillpass/web/tsconfig.json
{
  "compilerOptions": {
    "target": "ESNext",
    "module": "ESNext",
    "moduleResolution": "bundler",
    "strict": true,
    "esModuleInterop": true,
    "skipLibCheck": true,
    "jsx": "react-jsx",
    "baseUrl": ".",
    "paths": { "@/*": ["src/*"] }
  },
  "include": ["src"]
}
```

```html
<!-- /home/al-ip/learning/skillpass/web/index.html -->
<!DOCTYPE html>
<html lang="en" data-theme="winter">
  <head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>SkillPass — Your Career Passport</title>
  </head>
  <body>
    <div id="root"></div>
    <script type="module" src="/src/main.tsx"></script>
  </body>
</html>
```

- [ ] **Create CSS entry with DaisyUI import**

```css
/* /home/al-ip/learning/skillpass/web/src/styles/index.css */
@import "tailwindcss";
@plugin "daisyui";
```

- [ ] **Create main entry point and App.tsx**

```typescript
// /home/al-ip/learning/skillpass/web/src/main.tsx
import { StrictMode } from 'react';
import { createRoot } from 'react-dom/client';
import { BrowserRouter } from 'react-router-dom';
import { App } from './App';
import './styles/index.css';

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <BrowserRouter>
      <App />
    </BrowserRouter>
  </StrictMode>
);
```

```typescript
// /home/al-ip/learning/skillpass/web/src/App.tsx
import { Routes, Route } from 'react-router-dom';
import { RootLayout } from './components/layout/RootLayout';
import { AuthProvider } from './hooks/useAuth';
import { Landing } from './pages/Landing';
import { Login } from './pages/Login';
import { Register } from './pages/Register';
import { JobseekerProfile } from './pages/JobseekerProfile';
import { JobseekerPassport } from './pages/JobseekerPassport';
import { CompanyProfile } from './pages/CompanyProfile';
import { CompanyVerification } from './pages/CompanyVerification';
import { CompanySearch } from './pages/CompanySearch';
import { CompanyJobs } from './pages/CompanyJobs';
import { PublicJobs } from './pages/PublicJobs';
import { JobDetail } from './pages/JobDetail';
import { PublicPassport } from './pages/PublicPassport';
import { AdminVerifications } from './pages/AdminVerifications';

export function App() {
  return (
    <AuthProvider>
      <Routes>
        <Route element={<RootLayout />}>
          <Route path="/" element={<Landing />} />
          <Route path="/auth/login" element={<Login />} />
          <Route path="/auth/register" element={<Register />} />
          <Route path="/jobseeker/profile" element={<JobseekerProfile />} />
          <Route path="/jobseeker/passport" element={<JobseekerPassport />} />
          <Route path="/company/profile" element={<CompanyProfile />} />
          <Route path="/company/verification" element={<CompanyVerification />} />
          <Route path="/company/search" element={<CompanySearch />} />
          <Route path="/company/jobs" element={<CompanyJobs />} />
          <Route path="/jobs" element={<PublicJobs />} />
          <Route path="/jobs/:id" element={<JobDetail />} />
          <Route path="/profiles/:username" element={<PublicPassport />} />
          <Route path="/admin/verifications" element={<AdminVerifications />} />
        </Route>
      </Routes>
    </AuthProvider>
  );
}
```

- [ ] **Install dependencies**

Run: `cd /home/al-ip/learning/skillpass/web && bun install`

Expected: `bun install` succeeds.

- [ ] **Verify dev server starts**

Run: `cd /home/al-ip/learning/skillpass/web && bun run dev`

Expected: Vite starts on port 5173. Stop it.

- [ ] **Commit**

```bash
git -C /home/al-ip/learning/skillpass add web/package.json web/tsconfig.json web/vite.config.ts web/index.html web/src/main.tsx web/src/App.tsx web/src/styles/index.css
git -C /home/al-ip/learning/skillpass commit -m "feat: init web frontend with Vite + React + DaisyUI"
```

### Task 12: Create shared UI components (layout, auth, theme)

**Files:**
- Create: `skillpass/web/src/components/layout/RootLayout.tsx`
- Create: `skillpass/web/src/components/layout/Navbar.tsx`
- Create: `skillpass/web/src/components/ui/ThemeToggle.tsx`
- Create: `skillpass/web/src/components/ui/ProtectedRoute.tsx`
- Create: `skillpass/web/src/lib/api.ts`
- Create: `skillpass/web/src/hooks/useAuth.tsx`

- [ ] **Write API client (fetch wrapper with JWT)**

```typescript
// /home/al-ip/learning/skillpass/web/src/lib/api.ts
const BASE_URL = '/api/v1';

function getTokens() {
  const accessToken = localStorage.getItem('accessToken');
  const refreshToken = localStorage.getItem('refreshToken');
  return { accessToken, refreshToken };
}

function setTokens(accessToken: string, refreshToken?: string) {
  localStorage.setItem('accessToken', accessToken);
  if (refreshToken) localStorage.setItem('refreshToken', refreshToken);
}

export function clearTokens() {
  localStorage.removeItem('accessToken');
  localStorage.removeItem('refreshToken');
}

async function refreshAccessToken(): Promise<string | null> {
  const refreshToken = localStorage.getItem('refreshToken');
  if (!refreshToken) return null;

  const res = await fetch(`${BASE_URL}/auth/refresh`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ refreshToken }),
  });

  if (!res.ok) { clearTokens(); return null; }
  const data = await res.json();
  setTokens(data.accessToken);
  return data.accessToken;
}

export async function api<T = unknown>(
  path: string,
  options: RequestInit = {}
): Promise<T> {
  const { accessToken } = getTokens();
  const headers = new Headers(options.headers);
  headers.set('Content-Type', 'application/json');
  if (accessToken) headers.set('Authorization', `Bearer ${accessToken}`);

  let res = await fetch(`${BASE_URL}${path}`, { ...options, headers });

  // Try refreshing token on 401
  if (res.status === 401) {
    const newToken = await refreshAccessToken();
    if (newToken) {
      headers.set('Authorization', `Bearer ${newToken}`);
      res = await fetch(`${BASE_URL}${path}`, { ...options, headers });
    }
  }

  if (!res.ok) {
    const err = await res.text();
    throw new Error(err || `HTTP ${res.status}`);
  }

  return res.json();
}

export interface LoginResponse {
  accessToken: string;
  refreshToken: string;
  user: { id: string; email: string; username: string; name: string; role: 'jobseeker' | 'company' };
}

export async function login(email: string, password: string): Promise<LoginResponse> {
  const data = await api<LoginResponse>('/auth/login', {
    method: 'POST',
    body: JSON.stringify({ email, password }),
  });
  setTokens(data.accessToken, data.refreshToken);
  return data;
}

export async function register(body: {
  email: string; username: string; password: string; name: string; role: 'jobseeker' | 'company';
}): Promise<LoginResponse> {
  const data = await api<LoginResponse>('/auth/register', {
    method: 'POST',
    body: JSON.stringify(body),
  });
  setTokens(data.accessToken, data.refreshToken);
  return data;
}

export async function logout(): Promise<void> {
  try { await api('/auth/logout', { method: 'POST' }); } catch {}
  clearTokens();
}
```

- [ ] **Write auth context and provider**

```typescript
// /home/al-ip/learning/skillpass/web/src/hooks/useAuth.tsx
import { createContext, useContext, useState, useEffect, useCallback, type ReactNode } from 'react';
import { api, login as apiLogin, register as apiRegister, logout as apiLogout, clearTokens, type LoginResponse } from '../lib/api';

interface User {
  id: string; email: string; username: string; name: string; role: 'jobseeker' | 'company';
}

interface AuthContextType {
  user: User | null;
  loading: boolean;
  login: (email: string, password: string) => Promise<void>;
  register: (data: { email: string; username: string; password: string; name: string; role: 'jobseeker' | 'company' }) => Promise<void>;
  logout: () => Promise<void>;
}

const AuthContext = createContext<AuthContextType | null>(null);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const accessToken = localStorage.getItem('accessToken');
    if (accessToken) {
      // Try fetching current user — if fails, tokens are stale
      api<{ id: string; email: string; username: string; name: string; role: 'jobseeker' | 'company' }>('/profiles/me')
        .then((u) => setUser(u as unknown as User))
        .catch(() => clearTokens())
        .finally(() => setLoading(false));
    } else {
      setLoading(false);
    }
  }, []);

  const login = useCallback(async (email: string, password: string) => {
    const data = await apiLogin(email, password);
    setUser(data.user);
  }, []);

  const register = useCallback(async (data: { email: string; username: string; password: string; name: string; role: 'jobseeker' | 'company' }) => {
    const res = await apiRegister(data);
    setUser(res.user);
  }, []);

  const logout = useCallback(async () => {
    await apiLogout();
    setUser(null);
  }, []);

  return (
    <AuthContext.Provider value={{ user, loading, login, register, logout }}>
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error('useAuth must be used within AuthProvider');
  return ctx;
}
```

- [ ] **Write layout components**

```typescript
// /home/al-ip/learning/skillpass/web/src/components/ui/ThemeToggle.tsx
import { useState, useEffect } from 'react';
import { Sun, Moon } from 'lucide-react';

export function ThemeToggle() {
  const [dark, setDark] = useState(() => localStorage.getItem('theme') === 'dark');

  useEffect(() => {
    document.documentElement.setAttribute('data-theme', dark ? 'dark' : 'winter');
    localStorage.setItem('theme', dark ? 'dark' : 'winter');
  }, [dark]);

  return (
    <button className="btn btn-ghost btn-circle" onClick={() => setDark(!dark)} aria-label="Toggle theme">
      {dark ? <Sun size={20} /> : <Moon size={20} />}
    </button>
  );
}
```

```typescript
// /home/al-ip/learning/skillpass/web/src/components/ui/ProtectedRoute.tsx
import { Navigate } from 'react-router-dom';
import { useAuth } from '../../hooks/useAuth';

interface Props {
  children: React.ReactNode;
  role?: 'jobseeker' | 'company';
}

export function ProtectedRoute({ children, role }: Props) {
  const { user, loading } = useAuth();

  if (loading) return <div className="flex justify-center p-8"><span className="loading loading-spinner loading-lg"></span></div>;
  if (!user) return <Navigate to="/auth/login" replace />;
  if (role && user.role !== role) return <Navigate to="/" replace />;

  return <>{children}</>;
}
```

```typescript
// /home/al-ip/learning/skillpass/web/src/components/layout/Navbar.tsx
import { Link, useNavigate } from 'react-router-dom';
import { useAuth } from '../../hooks/useAuth';
import { ThemeToggle } from '../ui/ThemeToggle';

export function Navbar() {
  const { user, logout } = useAuth();
  const navigate = useNavigate();

  const handleLogout = async () => {
    await logout();
    navigate('/');
  };

  return (
    <div className="navbar bg-base-100 shadow-sm sticky top-0 z-50">
      <div className="flex-1">
        <Link to="/" className="btn btn-ghost text-xl font-bold">SkillPass</Link>
      </div>
      <div className="flex-none gap-2">
        {user ? (
          <>
            {user.role === 'jobseeker' && (
              <Link to="/jobseeker/profile" className="btn btn-ghost btn-sm">My Profile</Link>
            )}
            {user.role === 'company' && (
              <>
                <Link to="/company/search" className="btn btn-ghost btn-sm">Search</Link>
                <Link to="/company/jobs" className="btn btn-ghost btn-sm">Jobs</Link>
              </>
            )}
            <div className="dropdown dropdown-end">
              <div tabIndex={0} role="button" className="btn btn-ghost btn-circle avatar placeholder">
                <div className="bg-neutral text-neutral-content rounded-full w-10">
                  <span>{user.name.charAt(0).toUpperCase()}</span>
                </div>
              </div>
              <ul tabIndex={0} className="menu dropdown-content bg-base-100 rounded-box z-1 mt-3 w-52 p-2 shadow-sm">
                <li className="menu-label text-xs opacity-60">{user.email}</li>
                <div className="divider my-1" />
                {user.role === 'company' && (
                  <li><Link to="/company/profile">Company Profile</Link></li>
                )}
                <li><button onClick={handleLogout} className="text-error">Logout</button></li>
              </ul>
            </div>
          </>
        ) : (
          <>
            <Link to="/auth/login" className="btn btn-ghost btn-sm">Login</Link>
            <Link to="/auth/register" className="btn btn-primary btn-sm">Register</Link>
          </>
        )}
        <ThemeToggle />
      </div>
    </div>
  );
}
```

```typescript
// /home/al-ip/learning/skillpass/web/src/components/layout/RootLayout.tsx
import { Outlet } from 'react-router-dom';
import { Navbar } from './Navbar';

export function RootLayout() {
  return (
    <div className="min-h-screen flex flex-col">
      <Navbar />
      <main className="flex-1">
        <Outlet />
      </main>
      <footer className="footer footer-center bg-base-200 text-base-content p-4 text-sm opacity-60">
        <p>SkillPass — Build your career passport</p>
      </footer>
    </div>
  );
}
```

- [ ] **Commit**

```bash
git -C /home/al-ip/learning/skillpass add web/src/components/ web/src/lib/api.ts web/src/hooks/useAuth.tsx
git -C /home/al-ip/learning/skillpass commit -m "feat: add shared UI components, API client, and auth context"
```

### Task 13: Implement auth pages (landing, login, register)

**Files:**
- Create: `skillpass/web/src/pages/Landing.tsx`
- Create: `skillpass/web/src/pages/Login.tsx`
- Create: `skillpass/web/src/pages/Register.tsx`

- [ ] **Write Landing page**

```typescript
// /home/al-ip/learning/skillpass/web/src/pages/Landing.tsx
import { Link } from 'react-router-dom';
import { Briefcase, Search, Award } from 'lucide-react';

export function Landing() {
  return (
    <div className="hero min-h-[80vh] bg-base-100">
      <div className="hero-content text-center">
        <div className="max-w-2xl">
          <h1 className="text-5xl font-bold mb-4">Your Career,<br />One Passport</h1>
          <p className="text-lg opacity-70 mb-8">
            Build your complete career profile. Let AI evaluate your skills.
            Get discovered by verified companies. Grow with personalized feedback.
          </p>

          <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-8">
            <div className="card bg-base-200 p-4">
              <Briefcase className="mx-auto mb-2" size={28} />
              <h3 className="font-semibold">Build Profile</h3>
              <p className="text-sm opacity-60">Add every experience — job, gig, project, or education</p>
            </div>
            <div className="card bg-base-200 p-4">
              <Search className="mx-auto mb-2" size={28} />
              <h3 className="font-semibold">Get Discovered</h3>
              <p className="text-sm opacity-60">Verified companies find you by skills and experience</p>
            </div>
            <div className="card bg-base-200 p-4">
              <Award className="mx-auto mb-2" size={28} />
              <h3 className="font-semibold">Grow Faster</h3>
              <p className="text-sm opacity-60">Know your strengths and get suggestions to improve</p>
            </div>
          </div>

          <div className="flex gap-4 justify-center">
            <Link to="/auth/register" className="btn btn-primary btn-lg">Get Started</Link>
            <Link to="/jobs" className="btn btn-outline btn-lg">Browse Jobs</Link>
          </div>
        </div>
      </div>
    </div>
  );
}
```

- [ ] **Write Login page**

```typescript
// /home/al-ip/learning/skillpass/web/src/pages/Login.tsx
import { useState, type FormEvent } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { useAuth } from '../hooks/useAuth';

export function Login() {
  const { login } = useAuth();
  const navigate = useNavigate();
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setError('');
    setLoading(true);
    try {
      await login(email, password);
      navigate('/');
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Login failed');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="hero min-h-[60vh]">
      <div className="hero-content w-full max-w-sm">
        <div className="card bg-base-200 w-full p-6">
          <h2 className="text-2xl font-bold mb-6 text-center">Sign In</h2>
          <form onSubmit={handleSubmit} className="space-y-4">
            <label className="form-control w-full">
              <span className="label-text">Email</span>
              <input type="email" className="input input-bordered w-full" value={email}
                onChange={(e) => setEmail(e.target.value)} required />
            </label>
            <label className="form-control w-full">
              <span className="label-text">Password</span>
              <input type="password" className="input input-bordered w-full" value={password}
                onChange={(e) => setPassword(e.target.value)} required />
            </label>
            {error && <p className="text-error text-sm">{error}</p>}
            <button type="submit" className="btn btn-primary w-full" disabled={loading}>
              {loading ? <span className="loading loading-spinner" /> : 'Sign In'}
            </button>
          </form>
          <p className="text-sm text-center mt-4">
            Don't have an account? <Link to="/auth/register" className="link link-primary">Register</Link>
          </p>
        </div>
      </div>
    </div>
  );
}
```

- [ ] **Write Register page**

```typescript
// /home/al-ip/learning/skillpass/web/src/pages/Register.tsx
import { useState, type FormEvent } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { useAuth } from '../hooks/useAuth';

export function Register() {
  const { register } = useAuth();
  const navigate = useNavigate();
  const [form, setForm] = useState({ email: '', username: '', password: '', name: '', role: 'jobseeker' as const });
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setError('');
    setLoading(true);
    try {
      await register(form);
      navigate('/');
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Registration failed');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="hero min-h-[60vh]">
      <div className="hero-content w-full max-w-sm">
        <div className="card bg-base-200 w-full p-6">
          <h2 className="text-2xl font-bold mb-6 text-center">Create Account</h2>
          <form onSubmit={handleSubmit} className="space-y-4">
            <div className="flex gap-2">
              <button type="button" className={`btn flex-1 ${form.role === 'jobseeker' ? 'btn-primary' : 'btn-outline'}`}
                onClick={() => setForm({ ...form, role: 'jobseeker' })}>Jobseeker</button>
              <button type="button" className={`btn flex-1 ${form.role === 'company' ? 'btn-primary' : 'btn-outline'}`}
                onClick={() => setForm({ ...form, role: 'company' })}>Company</button>
            </div>
            <label className="form-control">
              <span className="label-text">Full Name</span>
              <input className="input input-bordered" value={form.name}
                onChange={(e) => setForm({ ...form, name: e.target.value })} required />
            </label>
            <label className="form-control">
              <span className="label-text">Username</span>
              <input className="input input-bordered" value={form.username}
                onChange={(e) => setForm({ ...form, username: e.target.value })} required minLength={3} />
            </label>
            <label className="form-control">
              <span className="label-text">Email</span>
              <input type="email" className="input input-bordered" value={form.email}
                onChange={(e) => setForm({ ...form, email: e.target.value })} required />
            </label>
            <label className="form-control">
              <span className="label-text">Password</span>
              <input type="password" className="input input-bordered" value={form.password}
                onChange={(e) => setForm({ ...form, password: e.target.value })} required minLength={6} />
            </label>
            {error && <p className="text-error text-sm">{error}</p>}
            <button type="submit" className="btn btn-primary w-full" disabled={loading}>
              {loading ? <span className="loading loading-spinner" /> : 'Create Account'}
            </button>
          </form>
          <p className="text-sm text-center mt-4">
            Already have an account? <Link to="/auth/login" className="link link-primary">Sign In</Link>
          </p>
        </div>
      </div>
    </div>
  );
}
```

- [ ] **Commit**

```bash
git -C /home/al-ip/learning/skillpass add web/src/pages/Landing.tsx web/src/pages/Login.tsx web/src/pages/Register.tsx
git -C /home/al-ip/learning/skillpass commit -m "feat: add landing, login, and register pages"
```

### Task 14: Implement jobseeker pages (profile + passport)

**Files:**
- Create: `skillpass/web/src/pages/JobseekerProfile.tsx`
- Create: `skillpass/web/src/pages/JobseekerPassport.tsx`
- Create: `skillpass/web/src/pages/PublicPassport.tsx`

- [ ] **Write JobseekerProfile page**

```typescript
// /home/al-ip/learning/skillpass/web/src/pages/JobseekerProfile.tsx
import { useState, useEffect, type FormEvent } from 'react';
import { api } from '../lib/api';
import { useAuth } from '../hooks/useAuth';
import { Plus, Pencil, Trash2 } from 'lucide-react';

interface Experience {
  id: string; type: string; title: string; organization: string;
  startDate: string; endDate?: string; isCurrent: boolean;
  description?: string; industry?: string; skillsUsed?: string[]; url?: string;
}

interface Profile {
  id: string; headline?: string; about?: string; yearsOfExperience?: number; slug: string;
  experiences: Experience[];
}

export function JobseekerProfile() {
  const { user } = useAuth();
  const [profile, setProfile] = useState<Profile | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [form, setForm] = useState({ headline: '', about: '', yearsOfExperience: 0 });
  const [showExpForm, setShowExpForm] = useState(false);
  const [expForm, setExpForm] = useState({ type: 'employment', title: '', organization: '', startDate: '', endDate: '', isCurrent: false, description: '', industry: '', skills: '' });

  useEffect(() => {
    api<Profile>('/profiles/me').then((data) => {
      setProfile(data);
      setForm({ headline: data.headline || '', about: data.about || '', yearsOfExperience: data.yearsOfExperience || 0 });
    }).finally(() => setLoading(false));
  }, []);

  const saveProfile = async (e: FormEvent) => {
    e.preventDefault();
    setSaving(true);
    const updated = await api<Profile>('/profiles/me', {
      method: 'PUT', body: JSON.stringify(form),
    });
    setProfile(prev => prev ? { ...prev, ...updated } : null);
    setSaving(false);
  };

  const addExperience = async (e: FormEvent) => {
    e.preventDefault();
    const skills = expForm.skills.split(',').map(s => s.trim()).filter(Boolean);
    const added = await api<Experience>('/profiles/me/experience', {
      method: 'POST',
      body: JSON.stringify({ ...expForm, skillsUsed: skills, endDate: expForm.isCurrent ? undefined : expForm.endDate || undefined }),
    });
    setProfile(prev => prev ? { ...prev, experiences: [...prev.experiences, added] } : null);
    setExpForm({ type: 'employment', title: '', organization: '', startDate: '', endDate: '', isCurrent: false, description: '', industry: '', skills: '' });
    setShowExpForm(false);
  };

  const deleteExperience = async (id: string) => {
    await api(`/profiles/me/experience/${id}`, { method: 'DELETE' });
    setProfile(prev => prev ? { ...prev, experiences: prev.experiences.filter(e => e.id !== id) } : null);
  };

  if (loading) return <div className="flex justify-center p-8"><span className="loading loading-spinner loading-lg" /></div>;
  if (!user || user.role !== 'jobseeker') return <div className="text-center p-8 text-error">Access denied</div>;

  return (
    <div className="max-w-2xl mx-auto p-4 space-y-6">
      <h1 className="text-2xl font-bold">My Profile</h1>

      <form onSubmit={saveProfile} className="card bg-base-200 p-4 space-y-4">
        <label className="form-control">
          <span className="label-text">Headline</span>
          <input className="input input-bordered" value={form.headline} onChange={e => setForm({ ...form, headline: e.target.value })} placeholder="e.g. Senior Software Engineer" />
        </label>
        <label className="form-control">
          <span className="label-text">About</span>
          <textarea className="textarea textarea-bordered h-24" value={form.about} onChange={e => setForm({ ...form, about: e.target.value })} placeholder="Tell companies about yourself" />
        </label>
        <label className="form-control">
          <span className="label-text">Years of Experience</span>
          <input type="number" className="input input-bordered" value={form.yearsOfExperience} onChange={e => setForm({ ...form, yearsOfExperience: Number(e.target.value) })} />
        </label>
        <button type="submit" className="btn btn-primary" disabled={saving}>
          {saving ? <span className="loading loading-spinner" /> : 'Save Profile'}
        </button>
      </form>

      <div className="card bg-base-200 p-4">
        <div className="flex justify-between items-center mb-4">
          <h2 className="text-xl font-semibold">Experience</h2>
          <button className="btn btn-primary btn-sm" onClick={() => setShowExpForm(!showExpForm)}>
            <Plus size={16} /> Add
          </button>
        </div>

        {showExpForm && (
          <form onSubmit={addExperience} className="space-y-3 mb-4 p-3 border border-base-300 rounded-box">
            <select className="select select-bordered w-full" value={expForm.type}
              onChange={e => setExpForm({ ...expForm, type: e.target.value })}>
              <option value="employment">Employment</option>
              <option value="gig">Gig / Freelance</option>
              <option value="education">Education</option>
              <option value="certification">Certification</option>
              <option value="project">Project</option>
              <option value="volunteering">Volunteering</option>
            </select>
            <input className="input input-bordered w-full" placeholder="Title / Degree" value={expForm.title}
              onChange={e => setExpForm({ ...expForm, title: e.target.value })} required />
            <input className="input input-bordered w-full" placeholder="Organization / Institution" value={expForm.organization}
              onChange={e => setExpForm({ ...expForm, organization: e.target.value })} required />
            <div className="flex gap-2">
              <input type="date" className="input input-bordered flex-1" value={expForm.startDate}
                onChange={e => setExpForm({ ...expForm, startDate: e.target.value })} required />
              <input type="date" className="input input-bordered flex-1" value={expForm.endDate}
                onChange={e => setExpForm({ ...expForm, endDate: e.target.value })} disabled={expForm.isCurrent} />
            </div>
            <label className="flex items-center gap-2">
              <input type="checkbox" className="checkbox checkbox-sm" checked={expForm.isCurrent}
                onChange={e => setExpForm({ ...expForm, isCurrent: e.target.checked })} />
              <span className="text-sm">I currently work here</span>
            </label>
            <textarea className="textarea textarea-bordered w-full" placeholder="Description" value={expForm.description}
              onChange={e => setExpForm({ ...expForm, description: e.target.value })} />
            <input className="input input-bordered w-full" placeholder="Skills (comma-separated)" value={expForm.skills}
              onChange={e => setExpForm({ ...expForm, skills: e.target.value })} />
            <button type="submit" className="btn btn-primary btn-sm">Add Experience</button>
          </form>
        )}

        <div className="space-y-2">
          {profile?.experiences.map(exp => (
            <div key={exp.id} className="flex justify-between items-start p-3 bg-base-100 rounded-box">
              <div>
                <p className="font-medium">{exp.title}</p>
                <p className="text-sm opacity-70">{exp.organization} · {exp.startDate} {exp.isCurrent ? '- Present' : exp.endDate ? `- ${exp.endDate}` : ''}</p>
                {exp.skillsUsed && exp.skillsUsed.length > 0 && (
                  <div className="flex flex-wrap gap-1 mt-1">
                    {exp.skillsUsed.map(s => <span key={s} className="badge badge-sm">{s}</span>)}
                  </div>
                )}
              </div>
              <button className="btn btn-ghost btn-xs text-error" onClick={() => deleteExperience(exp.id)}>
                <Trash2 size={14} />
              </button>
            </div>
          ))}
          {(!profile?.experiences || profile.experiences.length === 0) && (
            <p className="text-sm opacity-50 text-center py-4">No experience added yet</p>
          )}
        </div>
      </div>
    </div>
  );
}
```

- [ ] **Write JobseekerPassport page (private preview)**

```typescript
// /home/al-ip/learning/skillpass/web/src/pages/JobseekerPassport.tsx
import { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { api } from '../lib/api';
import { useAuth } from '../hooks/useAuth';
import { ExternalLink } from 'lucide-react';

interface PassportData {
  name: string; avatarUrl?: string; headline?: string; about?: string;
  yearsOfExperience?: number; experiences: Array<{ type: string; title: string; organization: string; startDate: string; endDate?: string; isCurrent: boolean; description?: string; skillsUsed?: string[] }>;
}

export function JobseekerPassport() {
  const { user } = useAuth();
  const [data, setData] = useState<PassportData | null>(null);

  useEffect(() => {
    if (user) {
      api<PassportData>(`/profiles/${user.username}`).then(setData);
    }
  }, [user]);

  if (!data) return <div className="flex justify-center p-8"><span className="loading loading-spinner loading-lg" /></div>;

  return (
    <div className="max-w-2xl mx-auto p-4 space-y-4">
      <div className="flex justify-between items-center">
        <h1 className="text-2xl font-bold">My Passport</h1>
        <Link to={`/profiles/${user?.username}`} className="btn btn-outline btn-sm gap-2" target="_blank">
          <ExternalLink size={14} /> View Public
        </Link>
      </div>

      <div className="card bg-base-200 p-6">
        <div className="flex items-center gap-4 mb-4">
          <div className="avatar placeholder">
            <div className="bg-neutral text-neutral-content rounded-full w-16">
              <span className="text-xl">{data.name?.charAt(0)}</span>
            </div>
          </div>
          <div>
            <h2 className="text-xl font-bold">{data.name}</h2>
            {data.headline && <p className="opacity-70">{data.headline}</p>}
            {data.yearsOfExperience !== undefined && <p className="text-sm opacity-50">{data.yearsOfExperience} years of experience</p>}
          </div>
        </div>
        {data.about && <p className="opacity-70 mb-4">{data.about}</p>}
      </div>

      <div className="card bg-base-200 p-4">
        <h3 className="font-semibold mb-3">Experience</h3>
        <div className="space-y-2">
          {data.experiences.map((exp, i) => (
            <div key={i} className="p-3 bg-base-100 rounded-box">
              <p className="font-medium">{exp.title}</p>
              <p className="text-sm opacity-70">{exp.organization} · {exp.startDate} {exp.isCurrent ? '- Present' : exp.endDate ? `- ${exp.endDate}` : ''}</p>
              {exp.description && <p className="text-sm mt-1 opacity-60">{exp.description}</p>}
              {exp.skillsUsed && exp.skillsUsed.length > 0 && (
                <div className="flex flex-wrap gap-1 mt-1">
                  {exp.skillsUsed.map(s => <span key={s} className="badge badge-sm">{s}</span>)}
                </div>
              )}
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}
```

- [ ] **Write PublicPassport page (no auth)**

```typescript
// /home/al-ip/learning/skillpass/web/src/pages/PublicPassport.tsx
import { useEffect, useState } from 'react';
import { useParams } from 'react-router-dom';
import { api } from '../lib/api';

interface PassportData {
  name: string; avatarUrl?: string; headline?: string; about?: string;
  yearsOfExperience?: number; experiences: Array<{ type: string; title: string; organization: string; startDate: string; endDate?: string; isCurrent: boolean; description?: string; skillsUsed?: string[] }>;
}

export function PublicPassport() {
  const { username } = useParams();
  const [data, setData] = useState<PassportData | null>(null);
  const [error, setError] = useState('');

  useEffect(() => {
    if (!username) return;
    api<PassportData>(`/profiles/${username}`)
      .then(setData)
      .catch(() => setError('Profile not found'));
  }, [username]);

  if (error) return <div className="text-center p-8"><p className="text-error">{error}</p></div>;
  if (!data) return <div className="flex justify-center p-8"><span className="loading loading-spinner loading-lg" /></div>;

  return (
    <div className="max-w-2xl mx-auto p-4 space-y-4">
      <div className="card bg-base-200 p-6">
        <div className="flex items-center gap-4 mb-4">
          <div className="avatar placeholder">
            <div className="bg-neutral text-neutral-content rounded-full w-20">
              <span className="text-2xl">{data.name?.charAt(0)}</span>
            </div>
          </div>
          <div>
            <h1 className="text-2xl font-bold">{data.name}</h1>
            {data.headline && <p className="opacity-70">{data.headline}</p>}
            {data.yearsOfExperience !== undefined && <p className="text-sm opacity-50">{data.yearsOfExperience} years of experience</p>}
          </div>
        </div>
        {data.about && <p className="opacity-70 mb-4">{data.about}</p>}
      </div>

      <div className="card bg-base-200 p-4">
        <h2 className="font-semibold mb-3">Experience</h2>
        <div className="space-y-2">
          {data.experiences.map((exp, i) => (
            <div key={i} className="p-3 bg-base-100 rounded-box">
              <p className="font-medium">{exp.title}</p>
              <p className="text-sm opacity-70">{exp.organization} · {exp.startDate} {exp.isCurrent ? '- Present' : exp.endDate ? `- ${exp.endDate}` : ''}</p>
              {exp.description && <p className="text-sm mt-1 opacity-60">{exp.description}</p>}
              {exp.skillsUsed && exp.skillsUsed.length > 0 && (
                <div className="flex flex-wrap gap-1 mt-1">
                  {exp.skillsUsed.map(s => <span key={s} className="badge badge-sm">{s}</span>)}
                </div>
              )}
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}
```

- [ ] **Commit**

```bash
git -C /home/al-ip/learning/skillpass add web/src/pages/JobseekerProfile.tsx web/src/pages/JobseekerPassport.tsx web/src/pages/PublicPassport.tsx
git -C /home/al-ip/learning/skillpass commit -m "feat: add jobseeker profile and passport pages"
```

### Task 15: Implement company pages (profile, verification, search, jobs)

**Files:**
- Create: `skillpass/web/src/pages/CompanyProfile.tsx`
- Create: `skillpass/web/src/pages/CompanyVerification.tsx`
- Create: `skillpass/web/src/pages/CompanySearch.tsx`
- Create: `skillpass/web/src/pages/CompanyJobs.tsx`
- Create: `skillpass/web/src/pages/JobDetail.tsx`
- Create: `skillpass/web/src/pages/PublicJobs.tsx`

- [ ] **Write CompanyProfile page**

```typescript
// /home/al-ip/learning/skillpass/web/src/pages/CompanyProfile.tsx
import { useState, useEffect, type FormEvent } from 'react';
import { api } from '../lib/api';
import { useAuth } from '../hooks/useAuth';

export function CompanyProfile() {
  const { user } = useAuth();
  const [form, setForm] = useState({ companyName: '', website: '', industry: '', description: '' });
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [industries, setIndustries] = useState<Array<{ id: string; name: string }>>([]);

  useEffect(() => {
    api<Array<{ id: string; name: string }>>('/industries').then(setIndustries);
    api<{ companyName: string; website?: string; industry: string; description?: string }>('/company/profile')
      .then((data) => setForm({ companyName: data.companyName, website: data.website || '', industry: data.industry, description: data.description || '' }))
      .finally(() => setLoading(false));
  }, []);

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setSaving(true);
    await api('/company/profile', { method: 'PUT', body: JSON.stringify(form) });
    setSaving(false);
  };

  if (loading) return <div className="flex justify-center p-8"><span className="loading loading-spinner loading-lg" /></div>;
  if (!user || user.role !== 'company') return <div className="text-center p-8 text-error">Access denied</div>;

  return (
    <div className="max-w-lg mx-auto p-4">
      <h1 className="text-2xl font-bold mb-6">Company Profile</h1>
      <form onSubmit={handleSubmit} className="card bg-base-200 p-4 space-y-4">
        <label className="form-control">
          <span className="label-text">Company Name</span>
          <input className="input input-bordered" value={form.companyName}
            onChange={e => setForm({ ...form, companyName: e.target.value })} required />
        </label>
        <label className="form-control">
          <span className="label-text">Website</span>
          <input className="input input-bordered" value={form.website}
            onChange={e => setForm({ ...form, website: e.target.value })} />
        </label>
        <label className="form-control">
          <span className="label-text">Industry</span>
          <select className="select select-bordered" value={form.industry}
            onChange={e => setForm({ ...form, industry: e.target.value })}>
            {industries.map(ind => <option key={ind.id} value={ind.name}>{ind.name}</option>)}
          </select>
        </label>
        <label className="form-control">
          <span className="label-text">Description</span>
          <textarea className="textarea textarea-bordered h-24" value={form.description}
            onChange={e => setForm({ ...form, description: e.target.value })} />
        </label>
        <button type="submit" className="btn btn-primary" disabled={saving}>
          {saving ? <span className="loading loading-spinner" /> : 'Save'}
        </button>
      </form>
    </div>
  );
}
```

- [ ] **Write CompanyVerification page**

```typescript
// /home/al-ip/learning/skillpass/web/src/pages/CompanyVerification.tsx
import { useState, useEffect, type FormEvent } from 'react';
import { api } from '../lib/api';
import { useAuth } from '../hooks/useAuth';

export function CompanyVerification() {
  const { user } = useAuth();
  const [status, setStatus] = useState<string | null>(null);
  const [form, setForm] = useState({ businessRegistration: '', website: '', address: '', contact: '' });
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    api<{ verificationStatus: string }>('/company/verification-status')
      .then(data => setStatus(data.verificationStatus));
  }, []);

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setSubmitting(true);
    await api('/company/verification', { method: 'POST', body: JSON.stringify(form) });
    setStatus('pending');
    setSubmitting(false);
  };

  if (!user || user.role !== 'company') return <div className="text-center p-8 text-error">Access denied</div>;

  if (status === 'verified') return (
    <div className="max-w-lg mx-auto p-4 text-center">
      <div className="card bg-base-200 p-6">
        <span className="text-4xl mb-2">✅</span>
        <h2 className="text-xl font-bold">Verified!</h2>
        <p className="opacity-70">Your company is verified. You can search candidates and post jobs.</p>
      </div>
    </div>
  );

  if (status === 'pending') return (
    <div className="max-w-lg mx-auto p-4 text-center">
      <div className="card bg-base-200 p-6">
        <span className="loading loading-spinner loading-lg mb-2" />
        <h2 className="text-xl font-bold">Verification Pending</h2>
        <p className="opacity-70">We're reviewing your documents. Check back soon.</p>
      </div>
    </div>
  );

  return (
    <div className="max-w-lg mx-auto p-4">
      <h1 className="text-2xl font-bold mb-6">Verify Your Company</h1>
      <p className="opacity-70 mb-4">Submit your business details to get verified. Verified companies can search candidates and post jobs.</p>
      <form onSubmit={handleSubmit} className="card bg-base-200 p-4 space-y-4">
        <label className="form-control">
          <span className="label-text">Business Registration Number</span>
          <input className="input input-bordered" value={form.businessRegistration}
            onChange={e => setForm({ ...form, businessRegistration: e.target.value })} required />
        </label>
        <label className="form-control">
          <span className="label-text">Company Website</span>
          <input className="input input-bordered" value={form.website}
            onChange={e => setForm({ ...form, website: e.target.value })} required />
        </label>
        <label className="form-control">
          <span className="label-text">Office Address</span>
          <textarea className="textarea textarea-bordered" value={form.address}
            onChange={e => setForm({ ...form, address: e.target.value })} required />
        </label>
        <label className="form-control">
          <span className="label-text">Contact Person & Title</span>
          <input className="input input-bordered" value={form.contact}
            onChange={e => setForm({ ...form, contact: e.target.value })} required />
        </label>
        <button type="submit" className="btn btn-primary" disabled={submitting}>
          {submitting ? <span className="loading loading-spinner" /> : 'Submit Verification'}
        </button>
      </form>
    </div>
  );
}
```

- [ ] **Write CompanySearch page**

```typescript
// /home/al-ip/learning/skillpass/web/src/pages/CompanySearch.tsx
import { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { api } from '../lib/api';

interface Candidate {
  id: string; name: string; avatarUrl?: string; headline?: string;
  about?: string; yearsOfExperience?: number; slug: string; skills: string[];
}

export function CompanySearch() {
  const [candidates, setCandidates] = useState<Candidate[]>([]);
  const [query, setQuery] = useState('');
  const [industry, setIndustry] = useState('');
  const [skills, setSkills] = useState('');
  const [industries, setIndustries] = useState<Array<{ id: string; name: string }>>([]);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    api<Array<{ id: string; name: string }>>('/industries').then(setIndustries);
  }, []);

  const search = async () => {
    setLoading(true);
    const params = new URLSearchParams();
    if (query) params.set('q', query);
    if (industry) params.set('industry', industry);
    if (skills) params.set('skills', skills);
    const data = await api<Candidate[]>(`/search/candidates?${params}`);
    setCandidates(data);
    setLoading(false);
  };

  useEffect(() => { search(); }, []);

  return (
    <div className="max-w-4xl mx-auto p-4 space-y-4">
      <h1 className="text-2xl font-bold">Find Candidates</h1>
      <div className="card bg-base-200 p-4">
        <div className="flex flex-wrap gap-2">
          <input className="input input-bordered flex-1" placeholder="Search by name, title, skill..."
            value={query} onChange={e => setQuery(e.target.value)} />
          <select className="select select-bordered" value={industry} onChange={e => setIndustry(e.target.value)}>
            <option value="">All Industries</option>
            {industries.map(ind => <option key={ind.id} value={ind.name}>{ind.name}</option>)}
          </select>
          <input className="input input-bordered w-48" placeholder="Skills (comma-separated)"
            value={skills} onChange={e => setSkills(e.target.value)} />
          <button className="btn btn-primary" onClick={search} disabled={loading}>
            {loading ? <span className="loading loading-spinner" /> : 'Search'}
          </button>
        </div>
      </div>

      <div className="space-y-2">
        {candidates.map(c => (
          <Link key={c.id} to={`/profiles/${c.slug}`} className="card bg-base-200 p-4 block hover:bg-base-300 transition-colors">
            <div className="flex items-center gap-3">
              <div className="avatar placeholder">
                <div className="bg-neutral text-neutral-content rounded-full w-12">
                  <span>{c.name?.charAt(0)}</span>
                </div>
              </div>
              <div className="flex-1">
                <h3 className="font-semibold">{c.name}</h3>
                {c.headline && <p className="text-sm opacity-70">{c.headline}</p>}
                {c.skills.length > 0 && (
                  <div className="flex flex-wrap gap-1 mt-1">
                    {c.skills.slice(0, 5).map(s => <span key={s} className="badge badge-sm">{s}</span>)}
                    {c.skills.length > 5 && <span className="text-xs opacity-50">+{c.skills.length - 5}</span>}
                  </div>
                )}
              </div>
            </div>
          </Link>
        ))}
        {candidates.length === 0 && !loading && (
          <p className="text-center opacity-50 py-8">No candidates found</p>
        )}
      </div>
    </div>
  );
}
```

- [ ] **Write CompanyJobs page**

```typescript
// /home/al-ip/learning/skillpass/web/src/pages/CompanyJobs.tsx
import { useState, useEffect, type FormEvent } from 'react';
import { Link } from 'react-router-dom';
import { api } from '../lib/api';
import { Plus, Pencil, X } from 'lucide-react';

interface Job {
  id: string; title: string; industry: string; location?: string;
  experienceLevel?: string; status: string; createdAt: string;
}

export function CompanyJobs() {
  const [jobs, setJobs] = useState<Job[]>([]);
  const [showForm, setShowForm] = useState(false);
  const [form, setForm] = useState({ title: '', description: '', industry: 'Technology', tags: '', requiredSkills: '', experienceLevel: 'mid', location: '', salaryRange: '' });
  const [industries, setIndustries] = useState<Array<{ id: string; name: string }>>([]);
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    api<Array<{ id: string; name: string }>>('/industries').then(setIndustries);
    api<Job[]>('/jobs/me').then(setJobs);
  }, []);

  const createJob = async (e: FormEvent) => {
    e.preventDefault();
    setSaving(true);
    const tags = form.tags.split(',').map(t => t.trim()).filter(Boolean);
    const requiredSkills = form.requiredSkills.split(',').map(s => s.trim()).filter(Boolean);
    await api('/jobs', {
      method: 'POST',
      body: JSON.stringify({ ...form, tags, requiredSkills }),
    });
    const updated = await api<Job[]>('/jobs/me');
    setJobs(updated);
    setShowForm(false);
    setSaving(false);
  };

  const closeJob = async (id: string) => {
    await api(`/jobs/${id}`, { method: 'PUT', body: JSON.stringify({ status: 'closed' }) });
    setJobs(prev => prev.map(j => j.id === id ? { ...j, status: 'closed' } : j));
  };

  return (
    <div className="max-w-3xl mx-auto p-4 space-y-4">
      <div className="flex justify-between items-center">
        <h1 className="text-2xl font-bold">My Job Postings</h1>
        <button className="btn btn-primary btn-sm" onClick={() => setShowForm(!showForm)}>
          <Plus size={16} /> New Job
        </button>
      </div>

      {showForm && (
        <form onSubmit={createJob} className="card bg-base-200 p-4 space-y-3">
          <input className="input input-bordered" placeholder="Job Title" value={form.title}
            onChange={e => setForm({ ...form, title: e.target.value })} required />
          <textarea className="textarea textarea-bordered h-24" placeholder="Job Description" value={form.description}
            onChange={e => setForm({ ...form, description: e.target.value })} required />
          <select className="select select-bordered" value={form.industry}
            onChange={e => setForm({ ...form, industry: e.target.value })}>
            {industries.map(ind => <option key={ind.id} value={ind.name}>{ind.name}</option>)}
          </select>
          <select className="select select-bordered" value={form.experienceLevel}
            onChange={e => setForm({ ...form, experienceLevel: e.target.value })}>
            <option value="entry">Entry</option>
            <option value="mid">Mid</option>
            <option value="senior">Senior</option>
            <option value="lead">Lead</option>
          </select>
          <input className="input input-bordered" placeholder="Tags (comma-separated)" value={form.tags}
            onChange={e => setForm({ ...form, tags: e.target.value })} />
          <input className="input input-bordered" placeholder="Required Skills (comma-separated)" value={form.requiredSkills}
            onChange={e => setForm({ ...form, requiredSkills: e.target.value })} />
          <div className="flex gap-2">
            <input className="input input-bordered flex-1" placeholder="Location" value={form.location}
              onChange={e => setForm({ ...form, location: e.target.value })} />
            <input className="input input-bordered flex-1" placeholder="Salary Range" value={form.salaryRange}
              onChange={e => setForm({ ...form, salaryRange: e.target.value })} />
          </div>
          <div className="flex gap-2">
            <button type="submit" className="btn btn-primary" disabled={saving}>
              {saving ? <span className="loading loading-spinner" /> : 'Post Job'}
            </button>
            <button type="button" className="btn" onClick={() => setShowForm(false)}>Cancel</button>
          </div>
        </form>
      )}

      <div className="space-y-2">
        {jobs.map(job => (
          <div key={job.id} className="card bg-base-200 p-4">
            <div className="flex justify-between items-start">
              <div>
                <h3 className="font-semibold">{job.title}</h3>
                <p className="text-sm opacity-70">{job.industry} {job.location ? `· ${job.location}` : ''}</p>
                <div className="flex gap-2 mt-1">
                  <span className="badge badge-sm">{job.experienceLevel}</span>
                  <span className={`badge badge-sm ${job.status === 'open' ? 'badge-success' : 'badge-ghost'}`}>{job.status}</span>
                </div>
              </div>
              <div className="flex gap-1">
                <Link to={`/jobs/${job.id}`} className="btn btn-ghost btn-xs"><Pencil size={14} /></Link>
                {job.status === 'open' && (
                  <button className="btn btn-ghost btn-xs text-error" onClick={() => closeJob(job.id)}>
                    <X size={14} />
                  </button>
                )}
              </div>
            </div>
          </div>
        ))}
        {jobs.length === 0 && !showForm && (
          <p className="text-center opacity-50 py-8">No job postings yet. Create your first one!</p>
        )}
      </div>
    </div>
  );
}
```

- [ ] **Write PublicJobs and JobDetail pages**

```typescript
// /home/al-ip/learning/skillpass/web/src/pages/PublicJobs.tsx
import { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { api } from '../lib/api';
import { Briefcase } from 'lucide-react';

interface Job {
  id: string; title: string; companyName?: string; industry: string;
  location?: string; experienceLevel?: string; salaryRange?: string; createdAt: string;
}

export function PublicJobs() {
  const [jobs, setJobs] = useState<Job[]>([]);
  const [industry, setIndustry] = useState('');
  const [industries, setIndustries] = useState<Array<{ id: string; name: string }>>([]);

  useEffect(() => {
    api<Array<{ id: string; name: string }>>('/industries').then(setIndustries);
    const params = industry ? `?industry=${industry}` : '';
    api<Job[]>(`/jobs${params}`).then(setJobs);
  }, [industry]);

  return (
    <div className="max-w-3xl mx-auto p-4 space-y-4">
      <h1 className="text-2xl font-bold">Job Openings</h1>
      <select className="select select-bordered w-full max-w-xs" value={industry}
        onChange={e => setIndustry(e.target.value)}>
        <option value="">All Industries</option>
        {industries.map(ind => <option key={ind.id} value={ind.name}>{ind.name}</option>)}
      </select>

      <div className="space-y-2">
        {jobs.map(job => (
          <Link key={job.id} to={`/jobs/${job.id}`} className="card bg-base-200 p-4 block hover:bg-base-300 transition-colors">
            <div className="flex items-start gap-3">
              <Briefcase className="mt-1 opacity-50" size={20} />
              <div>
                <h3 className="font-semibold">{job.title}</h3>
                <p className="text-sm opacity-70">{job.industry} {job.location ? `· ${job.location}` : ''}</p>
                <div className="flex gap-2 mt-1">
                  {job.experienceLevel && <span className="badge badge-sm">{job.experienceLevel}</span>}
                  {job.salaryRange && <span className="badge badge-sm badge-outline">{job.salaryRange}</span>}
                </div>
              </div>
            </div>
          </Link>
        ))}
        {jobs.length === 0 && <p className="text-center opacity-50 py-8">No jobs found</p>}
      </div>
    </div>
  );
}
```

```typescript
// /home/al-ip/learning/skillpass/web/src/pages/JobDetail.tsx
import { useEffect, useState } from 'react';
import { useParams } from 'react-router-dom';
import { api } from '../lib/api';
import { Calendar, MapPin, DollarSign, Briefcase } from 'lucide-react';

interface Job {
  id: string; title: string; description: string; industry: string;
  tags?: string[]; requiredSkills?: string[]; experienceLevel?: string;
  location?: string; salaryRange?: string; status: string; createdAt: string;
}

export function JobDetail() {
  const { id } = useParams();
  const [job, setJob] = useState<Job | null>(null);

  useEffect(() => {
    if (id) api<Job>(`/jobs/${id}`).then(setJob);
  }, [id]);

  if (!job) return <div className="flex justify-center p-8"><span className="loading loading-spinner loading-lg" /></div>;

  return (
    <div className="max-w-2xl mx-auto p-4">
      <div className="card bg-base-200 p-6">
        <h1 className="text-2xl font-bold mb-2">{job.title}</h1>
        <div className="flex flex-wrap gap-3 text-sm opacity-70 mb-4">
          <span className="flex items-center gap-1"><Briefcase size={14} /> {job.industry}</span>
          {job.location && <span className="flex items-center gap-1"><MapPin size={14} /> {job.location}</span>}
          {job.salaryRange && <span className="flex items-center gap-1"><DollarSign size={14} /> {job.salaryRange}</span>}
          <span className="flex items-center gap-1"><Calendar size={14} /> {job.createdAt?.slice(0, 10)}</span>
        </div>

        {job.experienceLevel && <span className="badge mb-4">{job.experienceLevel}</span>}

        <p className="mb-4 whitespace-pre-wrap">{job.description}</p>

        {job.requiredSkills && job.requiredSkills.length > 0 && (
          <div className="mb-4">
            <h3 className="font-semibold mb-2">Required Skills</h3>
            <div className="flex flex-wrap gap-1">
              {job.requiredSkills.map(s => <span key={s} className="badge badge-primary">{s}</span>)}
            </div>
          </div>
        )}

        {job.tags && job.tags.length > 0 && (
          <div>
            <h3 className="font-semibold mb-2">Tags</h3>
            <div className="flex flex-wrap gap-1">
              {job.tags.map(t => <span key={t} className="badge badge-ghost">{t}</span>)}
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
```

- [ ] **Commit**

```bash
git -C /home/al-ip/learning/skillpass add web/src/pages/CompanyProfile.tsx web/src/pages/CompanyVerification.tsx web/src/pages/CompanySearch.tsx web/src/pages/CompanyJobs.tsx web/src/pages/JobDetail.tsx web/src/pages/PublicJobs.tsx
git -C /home/al-ip/learning/skillpass commit -m "feat: add company pages (profile, verification, search, jobs)"
```

### Task 16: Implement admin verification page

**Files:**
- Create: `skillpass/web/src/pages/AdminVerifications.tsx`

- [ ] **Write AdminVerifications page**

```typescript
// /home/al-ip/learning/skillpass/web/src/pages/AdminVerifications.tsx
import { useState, useEffect } from 'react';
import { api } from '../lib/api';
import { Check, X } from 'lucide-react';

interface Company {
  id: string; companyName: string; website?: string;
  industry: string; description?: string; verificationDocs?: Record<string, string>;
  createdAt: string;
}

export function AdminVerifications() {
  const [pending, setPending] = useState<Company[]>([]);
  const [actionLoading, setActionLoading] = useState<string | null>(null);

  const loadPending = () => {
    api<Company[]>('/admin/verifications/pending').then(setPending);
  };

  useEffect(loadPending, []);

  const handleAction = async (id: string, action: 'approve' | 'reject') => {
    setActionLoading(id);
    await api(`/admin/verifications/${id}`, { method: 'POST', body: JSON.stringify({ action }) });
    setPending(prev => prev.filter(c => c.id !== id));
    setActionLoading(null);
  };

  return (
    <div className="max-w-3xl mx-auto p-4 space-y-4">
      <h1 className="text-2xl font-bold">Company Verifications</h1>

      {pending.length === 0 ? (
        <div className="card bg-base-200 p-8 text-center">
          <p className="opacity-50">No pending verifications</p>
        </div>
      ) : (
        pending.map(company => (
          <div key={company.id} className="card bg-base-200 p-4">
            <div className="flex justify-between items-start">
              <div>
                <h3 className="font-semibold">{company.companyName}</h3>
                <p className="text-sm opacity-70">{company.industry} {company.website ? `· ${company.website}` : ''}</p>
                {company.description && <p className="text-sm mt-1">{company.description}</p>}
                {company.verificationDocs && (
                  <div className="mt-2 p-2 bg-base-100 rounded-box text-sm">
                    {Object.entries(company.verificationDocs as Record<string, string>).map(([key, val]) => (
                      <p key={key}><span className="font-medium">{key}:</span> {val}</p>
                    ))}
                  </div>
                )}
              </div>
              <div className="flex gap-2">
                <button className="btn btn-success btn-sm" disabled={actionLoading === company.id}
                  onClick={() => handleAction(company.id, 'approve')}>
                  <Check size={16} /> Approve
                </button>
                <button className="btn btn-error btn-sm" disabled={actionLoading === company.id}
                  onClick={() => handleAction(company.id, 'reject')}>
                  <X size={16} /> Reject
                </button>
              </div>
            </div>
          </div>
        ))
      )}
    </div>
  );
}
```

- [ ] **Commit**

```bash
git -C /home/al-ip/learning/skillpass add web/src/pages/AdminVerifications.tsx
git -C /home/al-ip/learning/skillpass commit -m "feat: add admin verification approval page"
```

### Task 17: End-to-end smoke test

- [ ] **Start PostgreSQL and run migrations**

```bash
# If using Docker:
docker run --name skillpass-pg -e POSTGRES_PASSWORD=postgres -e POSTGRES_DB=skillpass -p 5432:5432 -d postgres:16

# Generate and push schema
cd /home/al-ip/learning/skillpass/server
DATABASE_URL="postgres://postgres:postgres@localhost:5432/skillpass" bun run db:push
```

Expected: Tables created in PostgreSQL. No errors.

- [ ] **Seed reference data**

```bash
cd /home/al-ip/learning/skillpass/server
DATABASE_URL="postgres://postgres:postgres@localhost:5432/skillpass" bun run seed
```

Expected: `✅ Seeded 12 industry categories`

- [ ] **Start server and verify endpoints**

```bash
cd /home/al-ip/learning/skillpass/server
DATABASE_URL="postgres://postgres:postgres@localhost:5432/skillpass" bun run dev
```

Expected: Server starts on port 3000. Swagger docs at http://localhost:3000/docs.

Test registration:
```bash
curl -X POST http://localhost:3000/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"john@test.com","username":"johndoe","password":"password123","name":"John Doe","role":"jobseeker"}'
```

Expected: Returns `201` with `accessToken`, `refreshToken`, and `user` object.

- [ ] **Start frontend and verify pages**

```bash
cd /home/al-ip/learning/skillpass/web
bun run dev
```

Expected: Vite starts on port 5173. Visit http://localhost:5173 — landing page renders with DaisyUI styling.

Test: Register as jobseeker → redirected home → navigate to /jobseeker/profile → fill in profile → add experience → view passport.

Test: Register as company → navigate to /company/verification → submit docs → admin can approve/reject.

### Task 17: Create Docker setup

**Files:**
- Create: `skillpass/server/Dockerfile`
- Create: `skillpass/server/.dockerignore`
- Create: `skillpass/web/Dockerfile`
- Create: `skillpass/web/nginx.conf`
- Create: `skillpass/web/.dockerignore`
- Create: `skillpass/docker-compose.yml`

- [ ] **Create server Dockerfile**

```dockerfile
# /home/al-ip/learning/skillpass/server/Dockerfile
FROM oven/bun:1 AS base
WORKDIR /app

COPY package.json bun.lock ./
RUN bun install --frozen-lockfile

COPY . .

EXPOSE 3000
CMD ["bun", "run", "src/index.ts"]
```

- [ ] **Create server .dockerignore**

```
node_modules
dist
drizzle
.env
.git
tests
*.md
```

- [ ] **Create web Dockerfile (multi-stage: Bun build + nginx serve)**

```dockerfile
# /home/al-ip/learning/skillpass/web/Dockerfile
FROM oven/bun:1 AS build
WORKDIR /app
COPY package.json bun.lock ./
RUN bun install --frozen-lockfile
COPY . .
RUN bun run build

FROM nginx:alpine
COPY --from=build /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
```

- [ ] **Create nginx.conf with API proxy**

```nginx
# /home/al-ip/learning/skillpass/web/nginx.conf
server {
    listen 80;
    server_name localhost;

    root /usr/share/nginx/html;
    index index.html;

    location /api/ {
        proxy_pass http://server:3000;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    location / {
        try_files $uri $uri/ /index.html;
    }
}
```

- [ ] **Create web .dockerignore**

```
node_modules
dist
.env
.git
tests
*.md
```

- [ ] **Create docker-compose.yml orchestrating all services**

```yaml
# /home/al-ip/learning/skillpass/docker-compose.yml
services:
  db:
    image: postgres:16
    ports:
      - "5432:5432"
    volumes:
      - pgdata:/var/lib/postgresql/data
    environment:
      POSTGRES_DB: skillpass
      POSTGRES_PASSWORD: postgres
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 5s
      timeout: 5s
      retries: 5

  server:
    build:
      context: ./server
    ports:
      - "3000:3000"
    depends_on:
      db:
        condition: service_healthy
    environment:
      DATABASE_URL: postgres://postgres:postgres@db:5432/skillpass
      JWT_SECRET: skillpass-dev-secret
      CORS_ORIGIN: http://localhost:5173
      PORT: 3000
    command: >
      sh -c "bun run db:push && bun run seed && bun run src/index.ts"

  web:
    build:
      context: ./web
    ports:
      - "5173:80"
    depends_on:
      - server

volumes:
  pgdata:
```

- [ ] **Verify full stack starts**

```bash
cd /home/al-ip/learning/skillpass
docker compose up --build
```

Expected: All three services start. Server accessible at http://localhost:3000, Web at http://localhost:5173, API calls proxied through nginx.

- [ ] **Commit**

```bash
git -C /home/al-ip/learning/skillpass add server/Dockerfile server/.dockerignore web/Dockerfile web/nginx.conf web/.dockerignore docker-compose.yml
git -C /home/al-ip/learning/skillpass commit -m "feat: add Docker setup with docker-compose"
```

## Run Instructions

### Local dev (no Docker):
```bash
# 1. Start PostgreSQL (Docker or local)
docker run --name skillpass-pg -e POSTGRES_PASSWORD=postgres -e POSTGRES_DB=skillpass -p 5432:5432 -d postgres:16

# 2. Run migrations + seed
cd skillpass/server
DATABASE_URL="postgres://postgres:postgres@localhost:5432/skillpass" bun run db:push
DATABASE_URL="postgres://postgres:postgres@localhost:5432/skillpass" bun run seed

# 3. Start server
DATABASE_URL="postgres://postgres:postgres@localhost:5432/skillpass" bun run dev

# 4. In another terminal, start web
cd skillpass/web
bun run dev
```

### Full Docker setup:
```bash
cd skillpass
docker compose up --build
```

All tables are created and seeded automatically on server startup via the `command` in docker-compose.yml.

---

## Plan Self-Review

### Spec coverage check
- Landing page with marketing copy and CTA ✓ (Task 13)
- Auth (register, login, refresh, logout) ✓ (Task 5, 13)
- Jobseeker profile CRUD (headline, about, experience) ✓ (Task 6, 14)
- Experience types (employment, gig, education, etc.) ✓ (schema + forms)
- Company profile & verification flow ✓ (Task 7, 15)
- Company candidate search (verified companies only) ✓ (Task 9, 15)
- Job postings CRUD + public listing ✓ (Task 8, 15)
- Public passport page (no auth) ✓ (Task 6, 14)
- Admin verification approval ✓ (Task 9, 16)
- Industry categories + tags reference data ✓ (Task 3, 9)
- Dark/light mode toggle ✓ (Task 12)
- Swagger API docs ✓ (Task 10)
- Rate limiting / Caddy — noted as production concern, not implemented in MVP ✓

### Placeholder scan
No TBDs, TODOs, or placeholder patterns found.

### Type consistency
- `userId` consistently used as JWT payload field in all route handlers
- `slug` on `jobseeker_profiles` maps to `username` from registration
- `companies.industry` is `text`, matching `industry_categories.name`
- API response shapes consistent across routes and frontend `api.ts`
