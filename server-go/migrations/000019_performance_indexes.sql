-- Performance indexes for high-traffic queries
-- Addresses N+1 query optimization and common lookup patterns

-- 1. Applications: support ORDER BY created_at DESC per jobseeker
--    Used by ListForJobseeker CTE query
CREATE INDEX IF NOT EXISTS applications_jobseeker_created_idx 
    ON applications(jobseeker_id, created_at DESC);

-- 2. Applications: support ORDER BY created_at DESC per company  
--    Used by ListForCompany CTE query
CREATE INDEX IF NOT EXISTS applications_posting_created_idx 
    ON applications(job_posting_id, created_at DESC);

-- 3. Job postings: support company's job list + status filter
--    Used by ListMyJobs handler
CREATE INDEX IF NOT EXISTS job_postings_company_status_idx 
    ON job_postings(company_id, status);

-- 4. AI evaluations: support ORDER BY created_at DESC per profile
--    Used by getLatestEvaluation in matching service
CREATE INDEX IF NOT EXISTS ai_evaluations_profile_created_idx 
    ON ai_evaluations(profile_id, created_at DESC);

-- 5. Refresh tokens: support token lookup during refresh
--    Used by Refresh handler
CREATE INDEX IF NOT EXISTS refresh_tokens_token_hash_idx 
    ON refresh_tokens(token_hash);
