import { jwt } from '@elysiajs/jwt';
import { eq } from 'drizzle-orm';
import { Elysia, t } from 'elysia';
import { db, schema } from '../db';

const JWT_SECRET = process.env.JWT_SECRET || 'skillpass-dev-secret-change-in-prod';

export const adminRoutes = new Elysia({ prefix: '/api/v1/admin' })
  .use(jwt({ secret: JWT_SECRET, name: 'jwt' }))
  .resolve(async ({ headers, jwt: j, error }) => {
    const auth = headers.authorization;
    if (!auth?.startsWith('Bearer ')) return error(401, 'Unauthorized');
    const payload = await j.verify(auth.slice(7));
    if (!payload) return error(401, 'Unauthorized');
    return { userId: payload.userId as string };
  })
  .get('/verifications/pending', async () => {
    return db.select().from(schema.companies).where(eq(schema.companies.verificationStatus, 'pending'));
  })
  .post(
    '/verifications/:id',
    async ({ params, body, error }) => {
      const [company] = await db.select().from(schema.companies).where(eq(schema.companies.id, params.id)).limit(1);

      if (!company) return error(404, 'Company not found');

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

      return error(400, 'Invalid action');
    },
    {
      body: t.Object({
        action: t.Union([t.Literal('approve'), t.Literal('reject')]),
        reason: t.Optional(t.String()),
      }),
    },
  );
