import { jwt } from '@elysiajs/jwt';
import { and, eq, sql } from 'drizzle-orm';
import { Elysia, status, t } from 'elysia';
import { config } from '../config';
import { db, schema } from '../db';

export const jobRoutes = new Elysia({ prefix: '/api/v1/jobs' }).use(jwt({ secret: config.jwtSecret, name: 'jwt' }));

// ── Public routes (no auth) ──

jobRoutes.get('/', async ({ query }) => {
  let conditions = and(eq(schema.jobPostings.status, 'open'));

  if (query.industry) {
    conditions = and(conditions, eq(schema.jobPostings.industry, query.industry as string));
  }
  if (query.experience_level) {
    conditions = and(
      conditions,
      eq(schema.jobPostings.experienceLevel, sql`${query.experience_level}::experience_level`),
    );
  }

  const jobs = await db.select().from(schema.jobPostings).where(conditions).orderBy(schema.jobPostings.createdAt);
  return jobs;
});

jobRoutes.get('/:id', async ({ params: { id }, set }) => {
  const [job] = await db.select().from(schema.jobPostings).where(eq(schema.jobPostings.id, id)).limit(1);
  if (!job) {
    set.status = 404;
    return { error: 'Job not found' };
  }
  return job;
});

// ── Protected routes (company auth) — derive applies to everything chained below ──

const jobBody = t.Object({
  title: t.String(),
  description: t.String(),
  industry: t.String(),
  tags: t.Optional(t.Array(t.String())),
  requiredSkills: t.Optional(t.Array(t.String())),
  experienceLevel: t.Optional(t.Union([t.Literal('entry'), t.Literal('mid'), t.Literal('senior'), t.Literal('lead')])),
  location: t.Optional(t.String()),
  salaryRange: t.Optional(t.String()),
});

const jobUpdateBody = t.Partial(
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
);

jobRoutes
  .derive(async ({ headers, jwt: j }) => {
    const auth = headers.authorization;
    if (!auth?.startsWith('Bearer ')) return status(401, 'Unauthorized');
    const payload = await j.verify(auth.slice(7));
    if (!payload) return status(401, 'Unauthorized');
    if (payload.role !== 'company') return status(401, 'Unauthorized');

    const [company] = await db
      .select()
      .from(schema.companies)
      .where(eq(schema.companies.userId, payload.userId as string))
      .limit(1);

    if (!company) return status(404, 'Company not found');
    if (company.verificationStatus !== 'verified') return status(403, 'Company not verified');

    return { userId: payload.userId as string, companyId: company.id };
  })
  .get('/me', async ({ userId }) => {
    const [company] = await db.select().from(schema.companies).where(eq(schema.companies.userId, userId)).limit(1);
    if (!company) return [];
    return db
      .select()
      .from(schema.jobPostings)
      .where(eq(schema.jobPostings.companyId, company.id))
      .orderBy(schema.jobPostings.createdAt);
  })
  .post(
    '/',
    async ({ body, companyId }) => {
      const [job] = await db
        .insert(schema.jobPostings)
        .values({
          companyId,
          ...body,
        })
        .returning();
      return job;
    },
    { body: jobBody },
  )
  .put(
    '/:id',
    async ({ params, body, companyId, set }) => {
      const [job] = await db
        .update(schema.jobPostings)
        .set(body)
        .where(and(eq(schema.jobPostings.id, params.id), eq(schema.jobPostings.companyId, companyId)))
        .returning();
      if (!job) {
        set.status = 404;
        return 'Job not found';
      }
      return job;
    },
    { body: jobUpdateBody },
  )
  .delete('/:id', async ({ params, companyId, set }) => {
    const [deleted] = await db
      .delete(schema.jobPostings)
      .where(and(eq(schema.jobPostings.id, params.id), eq(schema.jobPostings.companyId, companyId)))
      .returning();
    if (!deleted) {
      set.status = 404;
      return 'Job not found';
    }
    return { message: 'Deleted' };
  });
