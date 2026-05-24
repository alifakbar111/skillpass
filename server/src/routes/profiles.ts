import { jwt } from '@elysiajs/jwt';
import { and, eq } from 'drizzle-orm';
import { Elysia, status, t } from 'elysia';
import { db, schema } from '../db';

const JWT_SECRET = process.env.JWT_SECRET || 'skillpass-dev-secret-change-in-prod';

export const profileRoutes = new Elysia({ prefix: '/api/v1/profiles' })
  .use(jwt({ secret: JWT_SECRET, name: 'jwt' }))
  .derive(async ({ headers, jwt: j }) => {
    const auth = headers.authorization;
    if (!auth?.startsWith('Bearer ')) return status(401, 'Unauthorized');
    const payload = await j.verify(auth.slice(7));
    if (!payload) return status(401, 'Unauthorized');
    return { userId: payload.userId as string, role: payload.role as string };
  })
  .get('/me', async ({ userId, error }) => {
    const profile = await db
      .select()
      .from(schema.jobseekerProfiles)
      .where(eq(schema.jobseekerProfiles.userId, userId))
      .limit(1);

    if (profile.length === 0) return error(404, 'Profile not found');

    const [user] = await db
      .select({
        name: schema.users.name,
        email: schema.users.email,
        username: schema.users.username,
        role: schema.users.role,
        avatarUrl: schema.users.avatarUrl,
      })
      .from(schema.users)
      .where(eq(schema.users.id, profile[0].userId))
      .limit(1);

    const experiences = await db
      .select()
      .from(schema.jobExperiences)
      .where(eq(schema.jobExperiences.profileId, profile[0].id))
      .orderBy(schema.jobExperiences.startDate);

    return { ...profile[0], ...user, experiences };
  })
  .put(
    '/me',
    async ({ userId, body, error }) => {
      const [profile] = await db
        .update(schema.jobseekerProfiles)
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
    },
    {
      body: t.Object({
        headline: t.Optional(t.String()),
        about: t.Optional(t.String()),
        yearsOfExperience: t.Optional(t.Number()),
        slug: t.Optional(t.String()),
      }),
    },
  )
  .post(
    '/me/experience',
    async ({ userId, body, error }) => {
      const [profile] = await db
        .select()
        .from(schema.jobseekerProfiles)
        .where(eq(schema.jobseekerProfiles.userId, userId))
        .limit(1);

      if (!profile) return error(404, 'Profile not found');

      const [exp] = await db
        .insert(schema.jobExperiences)
        .values({
          profileId: profile.id,
          ...body,
        })
        .returning();

      return exp;
    },
    {
      body: t.Object({
        type: t.Enum({
          employment: 'employment',
          gig: 'gig',
          education: 'education',
          certification: 'certification',
          project: 'project',
          volunteering: 'volunteering',
        }),
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
    },
  )
  .put(
    '/me/experience/:id',
    async ({ userId, params, body, error }) => {
      const [profile] = await db
        .select()
        .from(schema.jobseekerProfiles)
        .where(eq(schema.jobseekerProfiles.userId, userId))
        .limit(1);
      if (!profile) return error(404, 'Profile not found');

      const [updated] = await db
        .update(schema.jobExperiences)
        .set(body)
        .where(and(eq(schema.jobExperiences.id, params.id), eq(schema.jobExperiences.profileId, profile.id)))
        .returning();

      if (!updated) return error(404, 'Experience not found');
      return updated;
    },
    {
      body: t.Partial(
        t.Object({
          type: t.Enum({
            employment: 'employment',
            gig: 'gig',
            education: 'education',
            certification: 'certification',
            project: 'project',
            volunteering: 'volunteering',
          }),
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
      ),
    },
  )
  .delete('/me/experience/:id', async ({ userId, params, error }) => {
    const [profile] = await db
      .select()
      .from(schema.jobseekerProfiles)
      .where(eq(schema.jobseekerProfiles.userId, userId))
      .limit(1);
    if (!profile) return error(404, 'Profile not found');

    const [deleted] = await db
      .delete(schema.jobExperiences)
      .where(and(eq(schema.jobExperiences.id, params.id), eq(schema.jobExperiences.profileId, profile.id)))
      .returning();

    if (!deleted) return error(404, 'Experience not found');
    return { message: 'Deleted' };
  });
