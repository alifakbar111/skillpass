import { jwt } from '@elysiajs/jwt';
import { and, eq, sql } from 'drizzle-orm';
import { Elysia, status, t } from 'elysia';
import { db, schema } from '../db';

const JWT_SECRET = process.env.JWT_SECRET || 'skillpass-dev-secret-change-in-prod';

export const jobRoutes = new Elysia({ prefix: '/api/v1/jobs' }).use(jwt({ secret: JWT_SECRET, name: 'jwt' }));

// Public listing — no auth required
jobRoutes.get('/', async ({ query }) => {
  let conditions = and(eq(schema.jobPostings.status, 'open'));

  if (query.industry) {
    conditions = and(conditions, eq(schema.jobPostings.industry, query.industry as string));
  }
  if (query.experience_level) {
    conditions = and(conditions, eq(schema.jobPostings.experienceLevel, sql`${query.experience_level}::experience_level`));
  }

  const jobs = await db.select().from(schema.jobPostings).where(conditions).orderBy(schema.jobPostings.createdAt);
  return jobs;
});

// Public single job
jobRoutes.get('/:id', async ({ params: { id }, set }) => {
  const [job] = await db.select().from(schema.jobPostings).where(eq(schema.jobPostings.id, id)).limit(1);
  if (!job) { set.status = 404; return { error: 'Job not found' }; }
  return job;
});

// Company routes — require auth + company role
const _companyJobs = jobRoutes.group('/me', (app) =>
  app
    .derive(async ({ headers, jwt: j }) => {
      const auth = headers.authorization;
      if (!auth?.startsWith('Bearer ')) return status(401, 'Unauthorized');
      const payload = await j.verify(auth.slice(7));
      if (!payload) return status(401, 'Unauthorized');
      if (payload.role !== 'company') return status(403, 'Forbidden');
      return { userId: payload.userId as string };
    })
    .get('', async ({ userId }) => {
      const [company] = await db.select().from(schema.companies).where(eq(schema.companies.userId, userId)).limit(1);

      if (!company) return [];
      return db
        .select()
        .from(schema.jobPostings)
        .where(eq(schema.jobPostings.companyId, company.id))
        .orderBy(schema.jobPostings.createdAt);
    }),
);

jobRoutes.post(
  '/',
  async ({ headers, body, jwt: j, set }) => {
    const auth = headers.authorization;
    if (!auth?.startsWith('Bearer ')) {
      set.status = 401;
      return 'Unauthorized';
    }
    const payload = await j.verify(auth.slice(7));
    if (!payload || payload.role !== 'company') {
      set.status = 401;
      return 'Unauthorized';
    }

    const [company] = await db
      .select()
      .from(schema.companies)
      .where(eq(schema.companies.userId, payload.userId as string))
      .limit(1);

    if (!company) {
      set.status = 404;
      return 'Company not found';
    }
    if (company.verificationStatus !== 'verified') {
      set.status = 403;
      return 'Company not verified';
    }

    const [job] = await db
      .insert(schema.jobPostings)
      .values({
        companyId: company.id,
        ...body,
      })
      .returning();

    return job;
  },
  {
    body: t.Object({
      title: t.String(),
      description: t.String(),
      industry: t.String(),
      tags: t.Optional(t.Array(t.String())),
      requiredSkills: t.Optional(t.Array(t.String())),
      experienceLevel: t.Optional(
        t.Union([t.Literal('entry'), t.Literal('mid'), t.Literal('senior'), t.Literal('lead')]),
      ),
      location: t.Optional(t.String()),
      salaryRange: t.Optional(t.String()),
    }),
  },
);

jobRoutes.put(
  '/:id',
  async ({ headers, params, body, jwt: j, set }) => {
    const auth = headers.authorization;
    if (!auth?.startsWith('Bearer ')) {
      set.status = 401;
      return 'Unauthorized';
    }
    const payload = await j.verify(auth.slice(7));
    if (!payload || payload.role !== 'company') {
      set.status = 401;
      return 'Unauthorized';
    }

    const [company] = await db
      .select()
      .from(schema.companies)
      .where(eq(schema.companies.userId, payload.userId as string))
      .limit(1);
    if (!company) {
      set.status = 404;
      return 'Company not found';
    }

    const [job] = await db
      .update(schema.jobPostings)
      .set(body)
      .where(and(eq(schema.jobPostings.id, params.id), eq(schema.jobPostings.companyId, company.id)))
      .returning();

    if (!job) {
      set.status = 404;
      return 'Job not found';
    }
    return job;
  },
  {
    body: t.Partial(
      t.Object({
        title: t.String(),
        description: t.String(),
        industry: t.String(),
        tags: t.Optional(t.Array(t.String())),
        requiredSkills: t.Optional(t.Array(t.String())),
        experienceLevel: t.Optional(
          t.Union([t.Literal('entry'), t.Literal('mid'), t.Literal('senior'), t.Literal('lead')]),
        ),
        location: t.Optional(t.String()),
        salaryRange: t.Optional(t.String()),
        status: t.Optional(t.Union([t.Literal('open'), t.Literal('closed')])),
      }),
    ),
  },
);

jobRoutes.delete('/:id', async ({ headers, params, jwt: j, set }) => {
  const auth = headers.authorization;
  if (!auth?.startsWith('Bearer ')) {
    set.status = 401;
    return 'Unauthorized';
  }
  const payload = await j.verify(auth.slice(7));
  if (!payload || payload.role !== 'company') {
    set.status = 401;
    return 'Unauthorized';
  }

  const [company] = await db
    .select()
    .from(schema.companies)
    .where(eq(schema.companies.userId, payload.userId as string))
    .limit(1);
  if (!company) {
    set.status = 404;
    return 'Company not found';
  }

  const [deleted] = await db
    .delete(schema.jobPostings)
    .where(and(eq(schema.jobPostings.id, params.id), eq(schema.jobPostings.companyId, company.id)))
    .returning();

  if (!deleted) {
    set.status = 404;
    return 'Job not found';
  }
  return { message: 'Deleted' };
});
