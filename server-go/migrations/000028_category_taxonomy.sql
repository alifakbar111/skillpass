-- Skill category taxonomy for industry-agnostic matching

-- 27 universal skill categories
CREATE TABLE IF NOT EXISTS skill_categories (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        TEXT NOT NULL UNIQUE,
    description TEXT NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Category weights per job posting (0-10 scale)
CREATE TABLE IF NOT EXISTS job_category_weights (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    job_posting_id  UUID NOT NULL REFERENCES job_postings(id) ON DELETE CASCADE,
    category_id     UUID NOT NULL REFERENCES skill_categories(id) ON DELETE CASCADE,
    weight          INTEGER NOT NULL DEFAULT 1 CHECK (weight >= 0 AND weight <= 10),
    UNIQUE(job_posting_id, category_id)
);

CREATE INDEX IF NOT EXISTS idx_job_category_weights_job ON job_category_weights(job_posting_id);

-- Seed the 27 universal categories
INSERT INTO skill_categories (name, description) VALUES
    ('Software Engineering', 'Coding, web/mobile dev, architecture, testing'),
    ('Data & Analytics', 'Data science, ML, BI, statistics, data engineering'),
    ('Infrastructure & IT', 'Cloud, DevOps, networking, sysadmin, IT support'),
    ('Cybersecurity', 'Security testing, compliance, incident response'),
    ('Engineering (Hardware)', 'Mechanical, electrical, civil, industrial'),
    ('Clinical & Medical', 'Healthcare, nursing, diagnosis, patient care'),
    ('Research & Science', 'Lab work, methodology, experiments, clinical trials'),
    ('Finance & Accounting', 'Accounting, audit, tax, financial analysis'),
    ('Sales & Marketing', 'Revenue, growth, digital marketing, branding'),
    ('Education & Training', 'Teaching, curriculum, mentoring, instructional design'),
    ('Management & Leadership', 'Team leadership, project management, strategy'),
    ('HR & People', 'Recruiting, payroll, employee relations'),
    ('Legal & Compliance', 'Law, regulation, contracts, governance'),
    ('Design & Creative', 'UI/UX, visual design, video, animation, motion'),
    ('Media & Content', 'Writing, journalism, content strategy, PR'),
    ('Customer Service & Support', 'Client support, help desk, account management'),
    ('Operations & Logistics', 'Supply chain, procurement, scheduling, inventory'),
    ('Construction & Trades', 'Building, electrical, plumbing, carpentry'),
    ('Manufacturing & Production', 'Assembly, quality control, machining, factory ops'),
    ('Hospitality & Tourism', 'Food service, hotel, travel, events, culinary'),
    ('Social Services & Nonprofit', 'Counseling, community outreach, fundraising'),
    ('Sports & Fitness', 'Training, coaching, physical therapy, athletics'),
    ('Agriculture & Environment', 'Farming, forestry, fishing, environmental science'),
    ('Real Estate & Property', 'Sales, appraisal, property management, leasing'),
    ('Government & Public Policy', 'Public admin, policy, urban planning, diplomacy'),
    ('Beauty & Wellness', 'Hair, skincare, spa, massage, aesthetics'),
    ('Religious & Spiritual', 'Clergy, chaplaincy, pastoral care, theology')
ON CONFLICT (name) DO NOTHING;
