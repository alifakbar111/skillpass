import { cors } from '@elysiajs/cors';
import { swagger } from '@elysiajs/swagger';
import { Elysia } from 'elysia';
import { adminRoutes } from './routes/admin';
import { authRoutes } from './routes/auth';
import { companyRoutes } from './routes/companies';
import { jobRoutes } from './routes/jobs';
import { passportRoutes } from './routes/passport';
import { profileRoutes } from './routes/profiles';
import { referenceRoutes } from './routes/reference';
import { searchRoutes } from './routes/search';

const app = new Elysia()
  .use(
    cors({
      origin: process.env.CORS_ORIGIN || 'http://localhost:4200',
      credentials: true,
    }),
  )
  .use(
    swagger({
      path: '/docs',
      documentation: {
        info: { title: 'SkillPass API', version: '1.0.0', description: 'Talent marketplace & skill passport API' },
      },
    }),
  );

// Register route groups
app
  .use(authRoutes)
  .use(profileRoutes)
  .use(passportRoutes)
  .use(companyRoutes)
  .use(jobRoutes)
  .use(searchRoutes)
  .use(referenceRoutes)
  .use(adminRoutes);

// Health check
app.get('/api/v1/health', () => ({ status: 'ok', timestamp: new Date().toISOString() }));

const port = Number(process.env.PORT || 8800);
app.listen(port);

console.log(`🦊 SkillPass API running at http://localhost:${port}`);
console.log(`📚 Swagger docs at http://localhost:${port}/docs`);

export type App = typeof app;
