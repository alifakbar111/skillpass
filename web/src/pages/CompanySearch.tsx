import { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { api } from '../lib/api';

interface Candidate {
  id: string;
  name: string;
  avatarUrl?: string;
  headline?: string;
  about?: string;
  yearsOfExperience?: number;
  slug: string;
  skills: string[];
}

export function CompanySearch() {
  const [candidates, setCandidates] = useState<Candidate[]>([]);
  const [query, setQuery] = useState('');
  const [industry, setIndustry] = useState('');
  const [skills, setSkills] = useState('');
  const [industries, setIndustries] = useState<Array<{ id: string; name: string }>>([]);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    api<Array<{ id: string; name: string }>>('/industries').then(setIndustries);
  }, []);

  const search = async () => {
    setLoading(true);
    const params = new URLSearchParams();
    if (query) params.set('q', query);
    if (industry) params.set('industry', industry);
    if (skills) params.set('skills', skills);
    const data = await api<Candidate[]>(`/search/candidates?${params}`);
    setCandidates(data);
    setLoading(false);
  };

  // biome-ignore lint/correctness/useExhaustiveDependencies: search intentionally not in deps to avoid auto-rerun on input change
  useEffect(() => {
    search();
  }, []);

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
          <select className="select select-bordered" value={industry} onChange={(e) => setIndustry(e.target.value)} aria-label="Filter by industry">
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
            {loading ? <span className="loading loading-spinner" aria-hidden="true" /> : 'Search'}
          </button>
        </div>
      </search>

      <div className="space-y-2">
        {candidates.map((c) => (
          <Link
            key={c.id}
            to={`/profiles/${c.slug}`}
            className="card bg-base-200 p-4 block hover:bg-base-300 transition-colors"
          >
            <div className="flex items-center gap-3">
              <div className="avatar placeholder">
                <div className="bg-neutral text-neutral-content rounded-full w-12">
                  <span>{c.name?.charAt(0)}</span>
                </div>
              </div>
              <div className="flex-1">
                <h3 className="font-semibold">{c.name}</h3>
                {c.headline && <p className="text-sm opacity-70">{c.headline}</p>}
                {c.skills.length > 0 && (
                  <div className="flex flex-wrap gap-1 mt-1">
                    {c.skills.slice(0, 5).map((s) => (
                      <span key={s} className="badge badge-sm">
                        {s}
                      </span>
                    ))}
                    {c.skills.length > 5 && <span className="text-xs opacity-50">+{c.skills.length - 5}</span>}
                  </div>
                )}
              </div>
            </div>
          </Link>
        ))}
        {candidates.length === 0 && !loading && <p className="text-center opacity-50 py-8">No candidates found</p>}
      </div>
    </div>
  );
}