import { jwt } from '@elysiajs/jwt';
import { eq } from 'drizzle-orm';
import { Elysia, t } from 'elysia';
import { db, schema } from '../db';
import { hashPassword, verifyPassword } from '../lib/password';

const JWT_SECRET = process.env.JWT_SECRET || 'skillpass-dev-secret-change-in-prod';

export const authRoutes = new Elysia({ prefix: '/api/v1/auth' })
  .use(jwt({ secret: JWT_SECRET, name: 'jwt' }))
  .post(
    '/register',
    async ({ body, jwt: j, error }) => {
      const { email, username, password, name, role } = body;

      const existingUser = await db.select().from(schema.users).where(eq(schema.users.email, email)).limit(1);

      if (existingUser.length > 0) {
        return error(409, 'Email already registered');
      }

      const passwordHash = await hashPassword(password);
      const user = await db
        .insert(schema.users)
        .values({
          email,
          username,
          passwordHash,
          name,
          role,
        })
        .returning();

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
        user: {
          id: user[0].id,
          email: user[0].email,
          username: user[0].username,
          name: user[0].name,
          role: user[0].role,
        },
      };
    },
    {
      body: t.Object({
        email: t.String({ format: 'email' }),
        username: t.String({ minLength: 3 }),
        password: t.String({ minLength: 6 }),
        name: t.String(),
        role: t.Union([t.Literal('jobseeker'), t.Literal('company')]),
      }),
    },
  )
  .post(
    '/login',
    async ({ body, jwt: j, error }) => {
      const users = await db.select().from(schema.users).where(eq(schema.users.email, body.email)).limit(1);

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
        user: {
          id: users[0].id,
          email: users[0].email,
          username: users[0].username,
          name: users[0].name,
          role: users[0].role,
        },
      };
    },
    {
      body: t.Object({
        email: t.String(),
        password: t.String(),
      }),
    },
  )
  .post(
    '/refresh',
    async ({ body, jwt: j, error }) => {
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
    },
    {
      body: t.Object({
        refreshToken: t.String(),
      }),
    },
  )
  .post('/logout', async () => {
    return { message: 'Logged out' };
  });
