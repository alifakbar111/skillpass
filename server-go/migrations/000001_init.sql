-- server-go/migrations/000001_init.sql
-- Initial schema extracted from server/src/db/schema.ts

-- Enums
DO $$ BEGIN
  CREATE TYPE role AS ENUM ('jobseeker', 'company', 'admin');
EXCEPTION WHEN duplicate_object THEN null;
END $$;

DO $$ BEGIN
  CREATE TYPE experience_type AS ENUM ('employment', 'gig', 'education', 'certification', 'project', 'volunteering');
EXCEPTION WHEN duplicate_object THEN null;
END $$;

DO $$ BEGIN
  CREATE TYPE verification_status AS ENUM ('pending', 'verified', 'rejected');
EXCEPTION WHEN duplicate_object THEN null;
END $$;

DO $$ BEGIN
  CREATE TYPE experience_level AS ENUM ('entry', 'mid', 'senior', 'lead');
EXCEPTION WHEN duplicate_object THEN null;
END $$;

DO $$ BEGIN
  CREATE TYPE job_status AS ENUM ('open', 'closed');
EXCEPTION WHEN duplicate_object THEN null;
END $$;

-- Users
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email TEXT NOT NULL UNIQUE,
    username TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    role role NOT NULL,
    name TEXT NOT NULL,
    avatar_url TEXT,
    is_verified BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Companies
CREATE TABLE IF NOT EXISTS companies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    company_name TEXT NOT NULL,
    website TEXT,
    industry TEXT NOT NULL,
    description TEXT,
    verification_status verification_status NOT NULL DEFAULT 'pending',
    verification_docs JSONB,
    verified_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Jobseeker profiles
CREATE TABLE IF NOT EXISTS jobseeker_profiles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    headline TEXT,
    about TEXT,
    years_of_experience INTEGER,
    slug TEXT NOT NULL UNIQUE
);

-- Job experiences
CREATE TABLE IF NOT EXISTS job_experiences (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    profile_id UUID NOT NULL REFERENCES jobseeker_profiles(id) ON DELETE CASCADE,
    type experience_type NOT NULL,
    title TEXT NOT NULL,
    organization TEXT NOT NULL,
    start_date TEXT NOT NULL,
    end_date TEXT,
    is_current BOOLEAN NOT NULL DEFAULT FALSE,
    description TEXT,
    industry TEXT,
    skills_used TEXT[],
    url TEXT
);

CREATE INDEX IF NOT EXISTS experience_profile_idx ON job_experiences(profile_id);

-- Industry categories
CREATE TABLE IF NOT EXISTS industry_categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL UNIQUE,
    description TEXT
);

-- Tags
CREATE TABLE IF NOT EXISTS tags (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    industry_category_id UUID REFERENCES industry_categories(id)
);

CREATE INDEX IF NOT EXISTS tags_industry_idx ON tags(industry_category_id);

-- Job postings
CREATE TABLE IF NOT EXISTS job_postings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    description TEXT NOT NULL,
    industry TEXT NOT NULL,
    tags TEXT[],
    required_skills TEXT[],
    experience_level experience_level,
    location TEXT,
    salary_range TEXT,
    status job_status NOT NULL DEFAULT 'open',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS job_status_created_idx ON job_postings(status, created_at);
CREATE INDEX IF NOT EXISTS job_company_idx ON job_postings(company_id);
