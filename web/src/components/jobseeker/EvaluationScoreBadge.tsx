interface Props {
  overallScore: number;
}

function getScoreLevel(score: number): { label: string; color: string } {
  if (score >= 200) return { label: 'Expert', color: 'badge-success' };
  if (score >= 100) return { label: 'Advanced', color: 'badge-info' };
  if (score >= 50) return { label: 'Intermediate', color: 'badge-warning' };
  return { label: 'Beginner', color: 'badge-ghost' };
}

export function EvaluationScoreBadge({ overallScore }: Props) {
  const { label, color } = getScoreLevel(overallScore);
  return (
    <div className="flex items-center gap-2">
      <div className={`badge badge-lg ${color} gap-1`}>
        <span className="font-bold">{overallScore}</span>
        <span className="text-xs">{label}</span>
      </div>
    </div>
  );
}
