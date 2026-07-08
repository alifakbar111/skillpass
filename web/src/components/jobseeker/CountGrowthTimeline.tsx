import { TrendingUp } from 'lucide-react';

interface HistoryEntry {
  id?: string;
  overallScore?: number;
  createdAt?: string;
}

interface Props {
  history: HistoryEntry[];
}

export function CountGrowthTimeline({ history }: Props) {
  const entries = history.filter(
    (e): e is Required<Pick<HistoryEntry, 'id' | 'overallScore' | 'createdAt'>> & HistoryEntry =>
      e.id != null && e.overallScore != null && e.createdAt != null,
  );

  if (entries.length === 0) {
    return (
      <div className="card bg-base-200 p-4">
        <p className="text-sm opacity-60 text-center">No evaluation history available.</p>
      </div>
    );
  }

  return (
    <div className="card bg-base-200 p-4">
      <h3 className="font-semibold flex items-center gap-2 mb-4">
        <TrendingUp size={18} aria-hidden="true" /> Count Growth
      </h3>
      <div className="space-y-3">
        {entries.map((entry, index) => {
          const prev = index > 0 ? entries[index - 1].overallScore : null;
          const growth = prev !== null ? entry.overallScore - prev : null;
          const date = new Date(entry.createdAt).toLocaleDateString('en-US', {
            month: 'short',
            year: 'numeric',
          });

          return (
            <div key={entry.id} className="flex items-center gap-3">
              <div className="flex flex-col items-center">
                <div className="w-3 h-3 rounded-full bg-primary" aria-hidden="true" />
                {index < entries.length - 1 && <div className="w-0.5 h-8 bg-base-300" aria-hidden="true" />}
              </div>
              <div className="flex-1 flex justify-between items-center">
                <div>
                  <p className="font-medium">{entry.overallScore.toLocaleString()} Count</p>
                  <p className="text-xs opacity-60">{date}</p>
                </div>
                {growth !== null && growth !== 0 && (
                  <span className={`text-sm font-semibold ${growth > 0 ? 'text-success' : 'text-error'}`}>
                    {growth > 0 ? '+' : ''}
                    {growth.toLocaleString()}
                  </span>
                )}
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
}
