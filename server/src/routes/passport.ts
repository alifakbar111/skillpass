import { eq } from 'drizzle-orm';
import { Elysia } from 'elysia';
import { db, schema } from '../db';

export const passportRoutes = new Elysia().get(
  '/api/v1/profiles/:username',
  async ({ params: { username }, error }) => {
    const [profile] = await db
      .select()
      .from(schema.jobseekerProfiles)
      .where(eq(schema.jobseekerProfiles.slug, username))
      .limit(1);

    if (!profile) return error(404, 'Profile not found');

    const [user] = await db.select().from(schema.users).where(eq(schema.users.id, profile.userId)).limit(1);

    const experiences = await db
      .select()
      .from(schema.jobExperiences)
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
  },
);
