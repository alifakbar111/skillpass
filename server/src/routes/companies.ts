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
