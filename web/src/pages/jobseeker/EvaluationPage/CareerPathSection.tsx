import { Compass } from 'lucide-react';
import { useState } from 'react';
import { ApiError } from '../../../lib/api';
import { type CareerPathResult, getCareerPath } from '../../../lib/evaluation';

const READINESS_BADGE: Record<string, string> = {
  ready: 'badge-success',
  stretch: 'badge-warning',
  'long-term': 'badge-info',
};

const READINESS_LABEL: Record<string, string> = {
  ready: 'Ready now',
  stretch: '6–12 months',
  'long-term': 'Long-term',
};

export function CareerPathSection() {
  const [result, setResult] = useState<CareerPathResult | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function handleGenerate() {
    if (loading) return;
    setLoading(true);
    setError(null);
    try {
      const data = await getCareerPath();
      setResult(data);
    } catch (err) {
      setError(err instanceof ApiError ? (err.serverMessage ?? err.message) : 'Failed to generate career path');
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="card bg-base-200 p-4">
      <div className="flex justify-between items-center mb-2">
        <h3 className="font-semibold flex items-center gap-2">
          <Compass size={16} className="text-primary" aria-hidden="true" /> Career Path
        </h3>
        <button type="button" className="btn btn-outline btn-sm" onClick={handleGenerate} disabled={loading}>
          {loading ? (
            <>
              <span className="loading loading-spinner loading-xs" /> Thinking…
            </>
          ) : result ? (
            'Regenerate'
          ) : (
            'Where could I go next?'
          )}
        </button>
      </div>

      {error && <p className="text-error text-sm mb-2">{error}</p>}

      {!result && !error && (
        <p className="text-sm opacity-60">
          AI-powered role recommendations based on your evaluation — see which roles you're ready for today and what to
          build toward.
        </p>
      )}

      {result && (
        <div className="space-y-3">
          {result.currentPosition && <p className="text-sm opacity-70 italic">{result.currentPosition}</p>}

          <div className="space-y-2">
            {(result.suggestedRoles ?? []).map((role) => (
              <div key={role.title} className="p-3 bg-base-100 rounded-box">
                <div className="flex justify-between items-center gap-2">
                  <span className="font-medium">{role.title}</span>
                  <span className={`badge badge-sm ${READINESS_BADGE[role.readiness ?? ''] ?? 'badge-ghost'}`}>
                    {READINESS_LABEL[role.readiness ?? ''] ?? role.readiness}
                  </span>
                </div>
                <p className="text-xs opacity-60 mt-1">{role.reason}</p>
              </div>
            ))}
          </div>

          {(result.steps ?? []).length > 0 && (
            <div>
              <p className="text-sm font-semibold mb-1">Next steps</p>
              <ul className="space-y-1">
                {(result.steps ?? []).map((step) => (
                  <li key={`${step.area}-${step.action}`} className="text-sm flex gap-2">
                    <span className="badge badge-xs badge-ghost mt-1 shrink-0 capitalize">{step.area}</span>
                    <span className="opacity-80">{step.action}</span>
                  </li>
                ))}
              </ul>
            </div>
          )}
        </div>
      )}
    </div>
  );
}
