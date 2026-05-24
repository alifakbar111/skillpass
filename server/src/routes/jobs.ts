import { jwt } from '@elysiajs/jwt';
import { and, eq } from 'drizzle-orm';
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
    conditions = and(conditions, eq(schema.jobPostings.experienceLevel, query.experience_level as string));
  }

  const jobs = await db.select().from(schema.jobPostings).where(conditions).orderBy(schema.jobPostings.createdAt);
  return jobs;
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
  async ({ headers, body, jwt: j, error }) => {
    const auth = headers.authorization;
    if (!auth?.startsWith('Bearer ')) return error(401, 'Unauthorized');
    const payload = await j.verify(auth.slice(7));
    if (!payload || payload.role !== 'company') return error(403, 'Forbidden');

    const [company] = await db
      .select()
      .from(schema.companies)
      .where(eq(schema.companies.userId, payload.userId as string))
      .limit(1);

    if (!company) return error(404, 'Company not found');
    if (company.verificationStatus !== 'verified') return error(403, 'Company not verified');

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
  async ({ headers, params, body, jwt: j, error }) => {
    const auth = headers.authorization;
    if (!auth?.startsWith('Bearer ')) return error(401, 'Unauthorized');
    const payload = await j.verify(auth.slice(7));
    if (!payload || payload.role !== 'company') return error(403, 'Forbidden');

    const [company] = await db
      .select()
      .from(schema.companies)
      .where(eq(schema.companies.userId, payload.userId as string))
      .limit(1);
    if (!company) return error(404, 'Company not found');

    const [job] = await db
      .update(schema.jobPostings)
      .set(body)
      .where(and(eq(schema.jobPostings.id, params.id), eq(schema.jobPostings.companyId, company.id)))
      .returning();

    if (!job) return error(404, 'Job not found');
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

jobRoutes.delete('/:id', async ({ headers, params, jwt: j, error }) => {
  const auth = headers.authorization;
  if (!auth?.startsWith('Bearer ')) return error(401, 'Unauthorized');
  const payload = await j.verify(auth.slice(7));
  if (!payload || payload.role !== 'company') return error(403, 'Forbidden');

  const [company] = await db
    .select()
    .from(schema.companies)
    .where(eq(schema.companies.userId, payload.userId as string))
    .limit(1);
  if (!company) return error(404, 'Company not found');

  const [deleted] = await db
    .delete(schema.jobPostings)
    .where(and(eq(schema.jobPostings.id, params.id), eq(schema.jobPostings.companyId, company.id)))
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
