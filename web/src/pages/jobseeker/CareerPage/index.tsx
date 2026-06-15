import { useQuery } from '@tanstack/react-query';
import { LoadingFallback } from '@/components/ui/LoadingFallback';
import { useAuth } from '@/hooks/useAuth';
import { getCareerPrediction, getSkillGap } from '@/lib/career';
import type { SkillGapItem } from '../../../lib/career';

export function CareerPage() {
  const { user } = useAuth();
  const {
    data: skillGap,
    isLoading: loadingGap,
    isError: isErrorGap,
    error: errorGap,
  } = useQuery({
    queryKey: ['career', 'skill-gap'],
    enabled: !!user,
    queryFn: getSkillGap,
  });
  const {
    data: prediction,
    isLoading: loadingPrediction,
    isError: isErrorPrediction,
    error: errorPrediction,
  } = useQuery({
    queryKey: ['career', 'prediction'],
    enabled: !!user,
    queryFn: getCareerPrediction,
  });
  if (!user)
    return (
      <div className="max-w-4xl mx-auto p-4">
        <p>Please log in to view career insights.</p>
      </div>
    );
  if (loadingGap || loadingPrediction) return <LoadingFallback text="Loading career insights" />;
  if (isErrorGap || isErrorPrediction)
    return (
      <div className="max-w-4xl mx-auto p-4">
        <div className="alert alert-error">
          <span>{errorGap?.message || errorPrediction?.message}</span>
        </div>
      </div>
    );
  return (
    <div className="max-w-5xl mx-auto p-4 space-y-6">
      <h1 className="text-2xl font-bold">Career Growth</h1>
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <div className="card bg-base-200 shadow-md">
          <div className="card-body">
            <h2 className="card-title">Skill Gap Radar</h2>
            {skillGap && skillGap.skills.length > 0 ? (
              <div className="space-y-3">
                <p className="text-sm text-base-content/60">Industry: {skillGap.industry || 'Not specified'}</p>
                {skillGap.skills.map((skill: SkillGapItem) => (
                  <div key={skill.skill} className="flex items-center gap-4">
                    <span className="w-24 text-sm font-medium">{skill.skill}</span>
                    <div className="flex-1">
                      <div className="h-4 bg-base-300 rounded-full overflow-hidden">
                        <div className="h-full bg-primary" style={{ width: `${Math.min(skill.currentLevel, 100)}%` }} />
                      </div>
                    </div>
                    <span className="text-xs text-base-content/60 w-16 text-right">
                      {skill.currentLevel}/{skill.requiredLevel}
                    </span>
                    {skill.gap > 0 && <span className="badge badge-warning badge-sm">Gap: {skill.gap}</span>}
                  </div>
                ))}
              </div>
            ) : (
              <p className="text-sm text-base-content/60">Complete an evaluation to see skill gap analysis.</p>
            )}
          </div>
        </div>
        <div className="card bg-base-200 shadow-md">
          <div className="card-body">
            <h2 className="card-title">Career Path Prediction</h2>
            {prediction ? (
              <div className="space-y-4">
                <div className="text-center p-4 bg-primary/10 rounded-lg">
                  <p className="text-lg font-bold text-primary">{prediction.predictedRole}</p>
                  <p className="text-sm text-base-content/60">Timeline: {prediction.timeline}</p>
                  <p className="text-xs text-base-content/40">Based on {prediction.similarProfiles} similar profiles</p>
                </div>
                <div>
                  <h3 className="font-semibold text-sm mb-2">Strengths</h3>
                  <div className="flex flex-wrap gap-2">
                    {prediction.strengths.map((s) => (
                      <span key={s} className="badge badge-success badge-outline">
                        {s}
                      </span>
                    ))}
                  </div>
                </div>
                <div>
                  <h3 className="font-semibold text-sm mb-2">Areas to Develop</h3>
                  <div className="flex flex-wrap gap-2">
                    {prediction.weaknesses.map((w) => (
                      <span key={w} className="badge badge-warning badge-outline">
                        {w}
                      </span>
                    ))}
                  </div>
                </div>
                <div>
                  <h3 className="font-semibold text-sm mb-2">Recommendations</h3>
                  <ul className="list-disc list-inside text-sm space-y-1">
                    {prediction.recommendations.map((r) => (
                      <li key={r}>{r}</li>
                    ))}
                  </ul>
                </div>
              </div>
            ) : (
              <p className="text-sm text-base-content/60">Complete an evaluation to get career path prediction.</p>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
