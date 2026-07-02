ALTER TABLE job_postings
  ADD COLUMN is_fresh_grad_friendly BOOLEAN NOT NULL DEFAULT false;
