import { eq } from 'drizzle-orm';
import { db, schema } from './src/db/index';

async function seed() {
  console.log('🌱 Seeding database...');

  const industries = [
    { name: 'Technology', description: 'Software, hardware, IT services' },
    { name: 'Manufacturing', description: 'Industrial production and fabrication' },
    { name: 'Healthcare', description: 'Medical services and pharmaceuticals' },
    { name: 'Finance', description: 'Banking, investment, insurance' },
    { name: 'Education', description: 'Schools, universities, training' },
    { name: 'Retail', description: 'Sales, e-commerce, consumer goods' },
    { name: 'Transportation', description: 'Logistics, delivery, ride-hailing' },
    { name: 'Creative Arts', description: 'Design, media, entertainment' },
    { name: 'Hospitality', description: 'Hotels, restaurants, tourism' },
    { name: 'Construction', description: 'Building and infrastructure' },
    { name: 'Agriculture', description: 'Farming, food production' },
    { name: 'Energy', description: 'Oil, gas, renewable energy' },
  ];

  for (const industry of industries) {
    await db.insert(schema.industryCategories).values(industry).onConflictDoNothing();
  }

  console.log(`✅ Seeded ${industries.length} industry categories`);

  // Seed admin user
  const adminEmail = 'admin-skillpass@yopmail.com';
  const existingAdmin = await db.select().from(schema.users).where(eq(schema.users.email, adminEmail)).limit(1);

  if (existingAdmin.length === 0) {
    const passwordHash = await Bun.password.hash('admin123!!');
    await db.insert(schema.users).values({
      email: adminEmail,
      username: 'admin',
      passwordHash,
      name: 'Admin',
      role: 'admin',
    });
    console.log('✅ Seeded admin user (admin-skillpass@yopmail.com / admin123!!)');
  } else {
    console.log('ℹ️  Admin user already exists, skipping');
  }

  process.exit(0);
}

seed().catch((err) => {
  console.error('❌ Seed failed:', err);
  process.exit(1);
});
