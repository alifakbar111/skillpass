import { Elysia } from 'elysia';
import { cors } from '@elysiajs/cors';
import { swagger } from '@elysiajs/swagger';

import { authRoutes } from './routes/auth';
import { profileRoutes } from './routes/profiles';
import { passportRoutes } from './routes/passport';
import { companyRoutes } from './routes/companies';
import { jobRoutes } from './routes/jobs';
import { searchRoutes } from './routes/search';
import { referenceRoutes } from './routes/reference';
import { adminRoutes } from './routes/admin';

const app = new Elysia()
  .use(cors({
    origin: process.env.CORS_ORIGIN || 'http://localhost:5173',
    credentials: true,
  }))
  .use(swagger({
    path: '/docs',
    documentation: {
      info: { title: 'SkillPass API', version: '1.0.0', description: 'Talent marketplace & skill passport API' },
    },
  }));

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

const port = Number(process.env.PORT || 3000);
app.listen(port);

console.log(`🦊 SkillPass API running at http://localhost:${port}`);
console.log(`📚 Swagger docs at http://localhost:${port}/docs`);

export type App = typeof app;
