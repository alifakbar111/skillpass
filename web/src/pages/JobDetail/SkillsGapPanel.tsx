import { Target } from 'lucide-react';
import { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { getSkillsGap, type SkillsGap } from '../../lib/matching';

export function SkillsGapPanel({ jobId }: { jobId: string }) {
  const [gap, setGap] = useState<SkillsGap | null>(null);

  useEffect(() => {
    let cancelled = false;
    getSkillsGap(jobId)
      .then((data) => {
        if (!cancelled) setGap(data);
      })
      .catch(() => {
        // Panel is an enhancement — hide on any error.
      });
    return () => {
      cancelled = true;
    };
  }, [jobId]);

  if (!gap) return null;
  if (gap.matchedSkills.length === 0 && gap.missingSkills.length === 0) return null;

  return (
    <div className="mt-4 p-4 bg-base-100 rounded-box">
      <div className="flex items-center justify-between mb-2">
        <h3 className="font-semibold flex items-center gap-2 text-sm">
          <Target size={16} className="text-primary" aria-hidden="true" /> Your Fit
        </h3>
        {gap.hasEvaluation && (
          <span className="badge badge-sm badge-primary">{Math.round(gap.matchPercent)}% match</span>
        )}
      </div>

      {!gap.hasEvaluation ? (
        <p className="text-xs opacity-70">
          Run an{' '}
          <Link to="/jobseeker/evaluation" className="link link-primary">
            AI evaluation
          </Link>{' '}
          to see how your skills match this job.
        </p>
      ) : (
        <div className="space-y-2">
          {gap.matchedSkills.length > 0 && (
            <div>
              <p className="text-xs opacity-60 mb-1">You have:</p>
              <div className="flex flex-wrap gap-1">
                {gap.matchedSkills.map((s) => (
                  <span key={s} className="badge badge-success badge-sm">
                    {s}
                  </span>
                ))}
              </div>
            </div>
          )}
          {gap.missingSkills.length > 0 && (
            <div>
              <p className="text-xs opacity-60 mb-1">To develop:</p>
              <div className="flex flex-wrap gap-1">
                {gap.missingSkills.map((s) => (
                  <span key={s} className="badge badge-outline badge-sm">
                    {s}
                  </span>
                ))}
              </div>
            </div>
          )}
        </div>
      )}
    </div>
  );
}
