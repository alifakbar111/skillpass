import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Brain, RefreshCw } from 'lucide-react';
import { EvaluationScoreBadge } from '@/components/jobseeker/EvaluationScoreBadge';
import { SkillScoresChart } from '@/components/jobseeker/SkillScoresChart';
import { LoadingSpinner } from '@/components/ui/LoadingFallback';
import { getLatestEvaluation, triggerEvaluation } from '@/lib/evaluation';

export function AIEvaluationSection() {
  const queryClient = useQueryClient();

  const { data: evaluation, isLoading } = useQuery({
    queryKey: ['evaluation', 'latest'],
    queryFn: getLatestEvaluation,
  });

  const triggerMutation = useMutation({
    mutationFn: triggerEvaluation,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['evaluation', 'latest'] });
    },
  });

  if (isLoading) {
    return (
      <div className="card bg-base-200 p-6">
        <LoadingSpinner size="sm" />
      </div>
    );
  }

  return (
    <div className="card bg-base-200 p-6">
      <div className="flex items-center justify-between mb-4">
        <h2 className="font-semibold flex items-center gap-2">
          <Brain size={20} /> AI Evaluation
        </h2>
      </div>

      {!evaluation ? (
        <div className="text-center py-6 space-y-4">
          <p className="text-sm opacity-60">Get your skills evaluated by AI to unlock personalized job matches.</p>
          <button
            type="button"
            className="btn btn-primary"
            onClick={() => triggerMutation.mutate()}
            disabled={triggerMutation.isPending}
          >
            {triggerMutation.isPending ? <LoadingSpinner size="sm" /> : 'Run AI Evaluation'}
          </button>
        </div>
      ) : (
        <div className="space-y-4">
          <div className="flex justify-center">
            <EvaluationScoreBadge overallScore={evaluation.overallScore ?? 0} />
          </div>

          {evaluation.strengths && evaluation.strengths.length > 0 && (
            <div>
              <h3 className="font-medium text-sm text-success mb-2">Strengths</h3>
              <ul className="space-y-1">
                {evaluation.strengths.map((s) => (
                  <li key={s.skill} className="text-sm flex items-start gap-2">
                    <span className="text-success mt-0.5" aria-hidden="true">
                      ✓
                    </span>
                    <span>
                      <strong>{s.skill}</strong> — {s.note}
                    </span>
                  </li>
                ))}
              </ul>
            </div>
          )}

          {evaluation.weaknesses && evaluation.weaknesses.length > 0 && (
            <div>
              <h3 className="font-medium text-sm text-warning mb-2">Areas to Improve</h3>
              <ul className="space-y-1">
                {evaluation.weaknesses.map((w) => (
                  <li key={w.skill} className="text-sm flex items-start gap-2">
                    <span className="text-warning mt-0.5" aria-hidden="true">
                      !
                    </span>
                    <span>
                      <strong>{w.skill}</strong> — {w.note}
                    </span>
                  </li>
                ))}
              </ul>
            </div>
          )}

          {evaluation.suggestions && evaluation.suggestions.length > 0 && (
            <div>
              <h3 className="font-medium text-sm text-info mb-2">Suggestions</h3>
              <ul className="space-y-1">
                {evaluation.suggestions.map((s) => (
                  <li key={s.area} className="text-sm">
                    <strong>{s.area}:</strong> {s.tip}
                  </li>
                ))}
              </ul>
            </div>
          )}

          {evaluation.skillScores && evaluation.skillScores.length > 0 && (
            <SkillScoresChart skillScores={evaluation.skillScores} />
          )}

          <div className="text-center pt-2">
            <button
              type="button"
              className="btn btn-ghost btn-xs gap-1"
              onClick={() => triggerMutation.mutate()}
              disabled={triggerMutation.isPending}
            >
              {triggerMutation.isPending ? <LoadingSpinner size="xs" /> : <RefreshCw size={14} aria-hidden="true" />}
              Refresh Evaluation
            </button>
          </div>
        </div>
      )}
    </div>
  );
}
