import { useQuery } from '@tanstack/react-query';
import { useState } from 'react';
import { Link } from 'react-router-dom';
import type { CandidateResult as Candidate } from '@/lib/api-types';
import { LoadingSpinner } from '../../components/ui/LoadingFallback';
import { useIndustries } from '../../hooks/useIndustries';
import { ApiError, api } from '../../lib/api';

interface SearchParams {
  query: string;
  industry: string;
  skills: string;
}

export function CompanySearch() {
  const [query, setQuery] = useState('');
  const [industry, setIndustry] = useState('');
  const [skills, setSkills] = useState('');
  // Committed params: only updated when the user clicks Search (not on keystroke).
  const [committed, setCommitted] = useState<SearchParams>({ query: '', industry: '', skills: '' });

  const { data: industries = [] } = useIndustries();

  const {
    data: candidates = [],
    error,
    isFetching: loading,
  } = useQuery({
    queryKey: ['candidates', committed],
    queryFn: () => {
      const params = new URLSearchParams();
      if (committed.query) params.set('q', committed.query);
      if (committed.industry) params.set('industry', committed.industry);
      if (committed.skills) params.set('skills', committed.skills);
      return api<Candidate[]>(`/search/candidates?${params}`);
    },
  });

  const errorMessage = error
    ? error instanceof ApiError
      ? (error.serverMessage ?? error.message)
      : 'Search failed'
    : null;

  const search = () => setCommitted({ query, industry, skills });

  return (
    <div className="max-w-4xl mx-auto p-4 space-y-4">
      <h1 className="text-2xl font-bold">Find Candidates</h1>
      <search className="card bg-base-200 p-4" aria-label="Search candidates">
        <div className="flex flex-wrap gap-2">
          <input
            className="input input-bordered flex-1"
            placeholder="Search by name, title, skill..."
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            aria-label="Search query"
          />
          <select
            className="select select-bordered"
            value={industry}
            onChange={(e) => setIndustry(e.target.value)}
            aria-label="Filter by industry"
          >
            <option value="">All Industries</option>
            {industries.map((ind) => (
              <option key={ind.id} value={ind.name}>
                {ind.name}
              </option>
            ))}
          </select>
          <input
            className="input input-bordered w-48"
            placeholder="Skills (comma-separated)"
            value={skills}
            onChange={(e) => setSkills(e.target.value)}
            aria-label="Filter by skills"
          />
          <button type="button" className="btn btn-primary" onClick={search} disabled={loading}>
            {loading ? <LoadingSpinner /> : 'Search'}
          </button>
        </div>
      </search>

      {errorMessage && <div className="alert alert-error">{errorMessage}</div>}

      <div className="space-y-2">
        {candidates.map((c) => {
          const content = (
            // biome-ignore lint/correctness/useJsxKeyInIterable: keys are on the returned Link/div below, not this extracted fragment
            <div className="flex items-center gap-3">
              <div className="avatar avatar-placeholder">
                <div className="bg-neutral text-neutral-content rounded-full w-12">
                  <span>{c.name?.charAt(0)}</span>
                </div>
              </div>
              <div className="flex-1">
                <h3 className="font-semibold">{c.name}</h3>
                {c.headline && <p className="text-sm opacity-70">{c.headline}</p>}
                {(c.skills ?? []).length > 0 && (
                  <div className="flex flex-wrap gap-1 mt-1">
                    {(c.skills ?? []).slice(0, 5).map((s) => (
                      <span key={s} className="badge badge-sm">
                        {s}
                      </span>
                    ))}
                    {(c.skills ?? []).length > 5 && (
                      <span className="text-xs opacity-50">+{(c.skills ?? []).length - 5}</span>
                    )}
                  </div>
                )}
              </div>
            </div>
          );

          // Blind mode masks the slug; render a non-clickable card with a hint.
          return c.slug ? (
            <Link
              key={c.id}
              to={`/profiles/${c.slug}`}
              className="card bg-base-200 p-4 block hover:bg-base-300 transition-colors"
            >
              {content}
            </Link>
          ) : (
            <div key={c.id} className="card bg-base-200 p-4">
              {content}
              <p className="text-xs opacity-50 mt-2">
                Identity hidden by blind screening — move this candidate forward to reveal.
              </p>
            </div>
          );
        })}
        {candidates.length === 0 && !loading && <p className="text-center opacity-50 py-8">No candidates found</p>}
      </div>
    </div>
  );
}
