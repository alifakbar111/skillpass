import { Elysia } from 'elysia';
import { db, schema } from '../db';
import { eq } from 'drizzle-orm';

export const referenceRoutes = new Elysia({ prefix: '/api/v1' })
  .get('/industries', async () => {
    return db.select().from(schema.industryCategories).orderBy(schema.industryCategories.name);
  })
  .get('/tags', async ({ query }) => {
    let queryBuilder = db.select().from(schema.tags);
    if (query.industry) {
      queryBuilder = queryBuilder.where(eq(schema.tags.industryCategoryId, query.industry as string));
    }
    return queryBuilder;
  });
