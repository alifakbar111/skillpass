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
  process.exit(0);
}

seed().catch((err) => {
  console.error('❌ Seed failed:', err);
  process.exit(1);
});
