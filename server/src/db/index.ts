import { drizzle } from 'drizzle-orm/postgres-js';
import postgres from 'postgres';
import * as schema from './schema';

const connectionString = process.env.DATABASE_URL || 'postgres://postgres:postgres@localhost:5432/skillpass';
const client = postgres(connectionString, {
  onnotice: () => {},
  max: 1,
});

// Test connectivity at startup — fail fast with a helpful message
try {
  await client`SELECT 1`;
} catch (err: unknown) {
  const e = err as Error;
  console.error('');
  console.error('❌ Cannot connect to PostgreSQL database.');
  console.error('');
  if (e.message.includes('ECONNREFUSED') || e.message.includes('connect')) {
    console.error('   The database server is not running on this machine.');
    console.error('   Start it with: docker compose up db -d');
    console.error('   Then run:      bun run db:push && bun run db:seed');
  } else if (e.message.includes('password')) {
    console.error('   Authentication failed — check your DATABASE_URL credentials.');
  } else {
    console.error('   Error:', e.message);
  }
  console.error('');
  console.error('   DATABASE_URL:', connectionString.replace(/\/\/[^@]+@/, '//***:***@'));
  console.error('');
  process.exit(1);
}

export const db = drizzle(client, { schema });
export { schema };
