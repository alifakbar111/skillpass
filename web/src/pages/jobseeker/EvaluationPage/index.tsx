import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Sparkles, X } from 'lucide-react';
import { useState } from 'react';
import { EvaluationScoreBadge } from '../../../components/jobseeker/EvaluationScoreBadge';
import { SkillScoresChart } from '../../../components/jobseeker/SkillScoresChart';
import { LoadingFallback } from '../../../components/ui/LoadingFallback';
import { useAuth } from '../../../hooks/useAuth';
import { ApiError } from '../../../lib/api';
import { getLatestEvaluation, triggerEvaluation } from '../../../lib/evaluation';
import { CareerPathSection } from './CareerPathSection';

export function EvaluationPage() {
  const { user } = useAuth();
  const queryClient = useQueryClient();
  const [error, setError] = useState<string | null>(null);

  const { data: evaluation, isLoading } = useQuery({
    queryKey: ['evaluation', 'latest'],
    enabled: !!user,
    queryFn: getLatestEvaluation,
    retry: (count, err) => count < 1 && !(err instanceof ApiError && err.status === 404),
  });

  const evaluateMutation = useMutation({
    mutationFn: triggerEvaluation,
    onMutate: () => {
      setError(null);
    },
    onSuccess: (result) => {
      queryClient.setQueryData(['evaluation', 'latest'], result);
    },
    onError: (err) => {
      setError(err instanceof ApiError ? (err.serverMessage ?? err.message) : 'Evaluation failed. Please try again.');
    },
  });

  if (!user) {
    return (
      <div className="max-w-2xl mx-auto p-4">
        <p>Please log in to view your evaluation.</p>
      </div>
    );
  }

  if (isLoading) return <LoadingFallback text="Loading evaluation" />;

  return (
    <div className="max-w-2xl mx-auto p-4 space-y-6">
      <div className="flex justify-between items-center">
        <h1 className="text-2xl font-bold">AI Profile Evaluation</h1>
        <button
          type="button"
          className="btn btn-primary gap-2"
          onClick={() => evaluateMutation.mutate()}
          disabled={evaluateMutation.isPending}
        >
          <Sparkles size={16} aria-hidden="true" />
          {evaluateMutation.isPending ? 'Evaluating...' : evaluation ? 'Re-evaluate' : 'Evaluate My Profile'}
        </button>
      </div>

      {error && (
        <div className="alert alert-error">
          <span>{error}</span>
          <button type="button" title="close" className="btn btn-ghost btn-xs" onClick={() => setError(null)}>
            <X size={14} />
          </button>
        </div>
      )}

      {evaluateMutation.isSuccess && (
        <div className="alert alert-success">
          <span>Evaluation complete!</span>
          <button type="button" title="close" className="btn btn-ghost btn-xs" onClick={() => evaluateMutation.reset()}>
            <X size={14} />
          </button>
        </div>
      )}

      {evaluation ? (
        <>
          <div className="card bg-base-200 p-6">
            <div className="flex items-center justify-between">
              <div>
                <h2 className="text-lg font-semibold">Overall Score</h2>
                <p className="text-sm opacity-60">Unlimited cumulative scoring — every skill adds points</p>
              </div>
              <EvaluationScoreBadge overallScore={evaluation.overallScore ?? 0} />
            </div>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div className="card bg-base-200 p-4">
              <h3 className="font-semibold mb-3 text-success">Strengths</h3>
              {(evaluation.strengths ?? []).length === 0 ? (
                <p className="text-sm opacity-60">No strengths identified yet.</p>
              ) : (
                <ul className="space-y-2">
                  {(evaluation.strengths ?? []).map((s) => (
                    <li key={s.skill} className="p-2 bg-base-100 rounded-box">
                      <div className="flex justify-between items-center">
                        <span className="font-medium">{s.skill}</span>
                        <span className="badge badge-success badge-sm">{s.score}</span>
                      </div>
                      {s.note && <p className="text-xs opacity-60 mt-1">{s.note}</p>}
                    </li>
                  ))}
                </ul>
              )}
            </div>

            <div className="card bg-base-200 p-4">
              <h3 className="font-semibold mb-3 text-warning">Areas to Improve</h3>
              {(evaluation.weaknesses ?? []).length === 0 ? (
                <p className="text-sm opacity-60">No weaknesses identified.</p>
              ) : (
                <ul className="space-y-2">
                  {(evaluation.weaknesses ?? []).map((w) => (
                    <li key={w.skill} className="p-2 bg-base-100 rounded-box">
                      <div className="flex justify-between items-center">
                        <span className="font-medium">{w.skill}</span>
                        <span className="badge badge-warning badge-sm">{w.score}</span>
                      </div>
                      {w.note && <p className="text-xs opacity-60 mt-1">{w.note}</p>}
                    </li>
                  ))}
                </ul>
              )}
            </div>
          </div>

          <div className="card bg-base-200 p-4">
            <h3 className="font-semibold mb-3 text-info">Suggestions</h3>
            {(evaluation.suggestions ?? []).length === 0 ? (
              <p className="text-sm opacity-60">No suggestions yet.</p>
            ) : (
              <ul className="space-y-2">
                {(evaluation.suggestions ?? []).map((s) => (
                  <li key={`suggest-${s.area}-${s.tip}`} className="p-3 bg-base-100 rounded-box">
                    <p className="font-medium capitalize">{s.area}</p>
                    <p className="text-sm opacity-70">{s.tip}</p>
                  </li>
                ))}
              </ul>
            )}
          </div>

          <div className="card bg-base-200 p-4">
            <h3 className="font-semibold mb-3">Skill Scores</h3>
            <SkillScoresChart skillScores={evaluation.skillScores ?? []} />
          </div>

          <CareerPathSection />

          <p className="text-xs opacity-50 text-center">
            Last evaluated: {evaluation.createdAt ? new Date(evaluation.createdAt).toLocaleString() : 'unknown'}
          </p>
        </>
      ) : (
        <div className="card bg-base-200 p-8 text-center">
          <p className="text-lg mb-4">You haven't evaluated your profile yet.</p>
          <p className="text-sm opacity-60 mb-6">
            Click "Evaluate My Profile" to get AI-powered insights on your skills, strengths, and areas for growth.
          </p>
        </div>
      )}
    </div>
  );
}
