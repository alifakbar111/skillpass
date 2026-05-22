import { Elysia, t } from 'elysia';
import { db, schema } from '../db';
import { eq, ilike, or, and } from 'drizzle-orm';
import { jwt } from '@elysiajs/jwt';

const JWT_SECRET = process.env.JWT_SECRET || 'skillpass-dev-secret-change-in-prod';

export const searchRoutes = new Elysia({ prefix: '/api/v1/search' })
  .use(jwt({ secret: JWT_SECRET, name: 'jwt' }))
  .resolve(async ({ headers, jwt: j, error }) => {
    const auth = headers.authorization;
    if (!auth || !auth.startsWith('Bearer ')) return error(401, 'Unauthorized');
    const payload = await j.verify(auth.slice(7));
    if (!payload) return error(401, 'Unauthorized');
    if (payload.role !== 'company') return error(403, 'Forbidden');

    const [company] = await db.select().from(schema.companies)
      .where(eq(schema.companies.userId, payload.userId as string))
      .limit(1);

    if (!company || company.verificationStatus !== 'verified') return error(403, 'Company not verified');
    return { userId: payload.userId as string };
  })
  .get('/candidates', async ({ query }) => {
    const profiles = await db.select().from(schema.jobseekerProfiles);
    const results = [];

    for (const profile of profiles) {
      const [user] = await db.select().from(schema.users)
        .where(eq(schema.users.id, profile.userId))
        .limit(1);

      const experiences = await db.select().from(schema.jobExperiences)
        .where(eq(schema.jobExperiences.profileId, profile.id));

      if (query.q) {
        const q = (query.q as string).toLowerCase();
        const matchesName = user?.name.toLowerCase().includes(q);
        const matchesHeadline = profile.headline?.toLowerCase().includes(q);
        const matchesAbout = profile.about?.toLowerCase().includes(q);
        const matchesExp = experiences.some(e =>
          e.title.toLowerCase().includes(q) ||
          e.organization.toLowerCase().includes(q) ||
          e.skillsUsed?.some(s => s.toLowerCase().includes(q))
        );
        if (!matchesName && !matchesHeadline && !matchesAbout && !matchesExp) continue;
      }

      if (query.skills) {
        const skillList = (query.skills as string).split(',').map(s => s.trim().toLowerCase());
        const hasSkill = experiences.some(e =>
          e.skillsUsed?.some(s => skillList.includes(s.toLowerCase()))
        );
        if (!hasSkill) continue;
      }

      if (query.industry) {
        const hasIndustry = experiences.some(e =>
          e.industry?.toLowerCase() === (query.industry as string).toLowerCase()
        );
        if (!hasIndustry) continue;
      }

      results.push({
        id: profile.id,
        name: user?.name,
        avatarUrl: user?.avatarUrl,
        headline: profile.headline,
        about: profile.about,
        yearsOfExperience: profile.yearsOfExperience,
        slug: profile.slug,
        skills: [...new Set(experiences.flatMap(e => e.skillsUsed || []))],
      });
    }

    return results;
  });
