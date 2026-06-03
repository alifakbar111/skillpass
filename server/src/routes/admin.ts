import { jwt } from '@elysiajs/jwt';
import { eq } from 'drizzle-orm';
import { Elysia, status, t } from 'elysia';
import { config } from '../config';
import { db, schema } from '../db';

export const adminRoutes = new Elysia({ prefix: '/api/v1/admin' })
  .use(jwt({ secret: config.jwtSecret, name: 'jwt' }))
  .derive(async ({ headers, jwt: j }) => {
    const auth = headers.authorization;
    if (!auth?.startsWith('Bearer ')) return status(401, 'Unauthorized');
    const payload = await j.verify(auth.slice(7));
    if (!payload) return status(401, 'Unauthorized');
    if (payload.role !== 'admin') return status(403, 'Forbidden: admin role required');
    return { userId: payload.userId as string };
  })
  .get('/verifications/pending', async () => {
    return db.select().from(schema.companies).where(eq(schema.companies.verificationStatus, 'pending'));
  })
  .post(
    '/verifications/:id',
    async ({ params, body, set }) => {
      const [company] = await db.select().from(schema.companies).where(eq(schema.companies.id, params.id)).limit(1);

      if (!company) {
        set.status = 404;
        return { error: 'Company not found' };
      }

      if (body.action === 'approve') {
        const [updated] = await db
          .update(schema.companies)
          .set({ verificationStatus: 'verified', verifiedAt: new Date() })
          .where(eq(schema.companies.id, params.id))
          .returning();

        await db.update(schema.users).set({ isVerified: true }).where(eq(schema.users.id, company.userId));

        return updated;
      } else if (body.action === 'reject') {
        const [updated] = await db
          .update(schema.companies)
          .set({ verificationStatus: 'rejected' })
          .where(eq(schema.companies.id, params.id))
          .returning();

        return updated;
      }

      set.status = 400;
      return { error: 'Invalid action' };
    },
    {
      body: t.Object({
        action: t.Union([t.Literal('approve'), t.Literal('reject')]),
        reason: t.Optional(t.String()),
      }),
    },
  );
