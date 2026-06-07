import type { SkillScoreItem } from '../../lib/evaluation';

interface Props {
  skillScores: SkillScoreItem[];
}

export function SkillScoresChart({ skillScores }: Props) {
  if (skillScores.length === 0) {
    return <p className="text-sm opacity-60">No skill scores available.</p>;
  }

  const grouped: Record<string, SkillScoreItem[]> = {};
  for (const item of skillScores) {
    const cat = item.category || 'uncategorized';
    if (!grouped[cat]) grouped[cat] = [];
    grouped[cat].push(item);
  }

  return (
    <div className="space-y-4">
      {Object.entries(grouped).map(([category, items]) => (
        <div key={category}>
          <h4 className="text-sm font-semibold capitalize mb-2">{category}</h4>
          <div className="space-y-2">
            {items.map((item) => (
              <div key={item.skill} className="flex items-center gap-2">
                <span className="text-sm w-24 truncate">{item.skill}</span>
                <progress
                  className="progress progress-primary flex-1"
                  value={item.score}
                  max={100}
                  aria-label={`${item.skill}: ${item.score}%`}
                />
                <span className="text-xs w-8 text-right opacity-70">{item.score}</span>
              </div>
            ))}
          </div>
        </div>
      ))}
    </div>
  );
}
