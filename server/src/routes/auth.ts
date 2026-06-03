import { jwt } from '@elysiajs/jwt';
import { eq } from 'drizzle-orm';
import { Elysia, t } from 'elysia';
import { config } from '../config';
import { db, schema } from '../db';
import { hashPassword, verifyPassword } from '../lib/password';

export const authRoutes = new Elysia({ prefix: '/api/v1/auth' })
  .use(jwt({ secret: config.jwtSecret, name: 'jwt' }))
  .post(
    '/register',
    async ({ body, jwt: j, set }) => {
      const { email, username, password, name, role } = body;

      const existingUser = await db.select().from(schema.users).where(eq(schema.users.email, email)).limit(1);

      if (existingUser.length > 0) {
        set.status = 409;
        return { error: 'Email already registered' };
      }

      const coFields = body as {
        companyName?: string;
        businessRegistration?: string;
        website?: string;
        address?: string;
        contact?: string;
      };
      const displayName = role === 'company' ? coFields.companyName || name : name;

      const passwordHash = await hashPassword(password);
      const user = await db
        .insert(schema.users)
        .values({
          email,
          username,
          passwordHash,
          name: displayName,
          role,
        })
        .returning();

      if (role === 'jobseeker') {
        await db.insert(schema.jobseekerProfiles).values({
          userId: user[0].id,
          slug: username,
        });
      } else if (role === 'company') {
        const coBody = body as {
          companyName?: string;
          businessRegistration?: string;
          website?: string;
          address?: string;
          contact?: string;
        };
        await db.insert(schema.companies).values({
          userId: user[0].id,
          companyName: displayName,
          industry: 'Technology',
          verificationDocs: {
            businessRegistration: coBody.businessRegistration ?? '',
            website: coBody.website ?? '',
            address: coBody.address ?? '',
            contact: coBody.contact ?? '',
          },
          verificationStatus: 'pending',
        });
      }

      const accessToken = await j.sign({ userId: user[0].id, role: user[0].role });
      const refreshToken = await j.sign({
        userId: user[0].id,
        type: 'refresh',
        exp: Math.floor(Date.now() / 1000) + 7 * 24 * 60 * 60,
      });

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
        companyName: t.Optional(t.String()),
        businessRegistration: t.Optional(t.String()),
        website: t.Optional(t.String()),
        address: t.Optional(t.String()),
        contact: t.Optional(t.String()),
      }),
    },
  )
  .post(
    '/login',
    async ({ body, jwt: j, set }) => {
      const users = await db.select().from(schema.users).where(eq(schema.users.email, body.email)).limit(1);

      if (users.length === 0) {
        set.status = 401;
        return { error: 'Invalid credentials' };
      }

      const valid = await verifyPassword(body.password, users[0].passwordHash);
      if (!valid) {
        set.status = 401;
        return { error: 'Invalid credentials' };
      }

      const accessToken = await j.sign({ userId: users[0].id, role: users[0].role });
      const refreshToken = await j.sign({
        userId: users[0].id,
        type: 'refresh',
        exp: Math.floor(Date.now() / 1000) + 7 * 24 * 60 * 60,
      });

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
    async ({ body, jwt: j, set }) => {
      try {
        const payload = await j.verify(body.refreshToken);
        if (!payload || payload.type !== 'refresh') {
          set.status = 401;
          return { error: 'Invalid refresh token' };
        }

        const accessToken = await j.sign({ userId: payload.userId, role: payload.role });
        return { accessToken };
      } catch {
        set.status = 401;
        return { error: 'Invalid refresh token' };
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
