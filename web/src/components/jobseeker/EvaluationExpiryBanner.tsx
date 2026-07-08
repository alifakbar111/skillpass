import { AlertTriangle } from 'lucide-react';

interface Props {
  createdAt: string;
}

export function EvaluationExpiryBanner({ createdAt }: Props) {
  const created = new Date(createdAt);
  const now = new Date();
  const threeMonthsMs = 3 * 30 * 24 * 60 * 60 * 1000;
  const isExpired = now.getTime() - created.getTime() > threeMonthsMs;

  if (!isExpired) return null;

  const expiredDate = new Date(created.getTime() + threeMonthsMs);

  return (
    <div className="alert alert-warning shadow-sm" role="alert">
      <AlertTriangle size={20} aria-hidden="true" />
      <div>
        <p className="font-semibold">Evaluation Expired</p>
        <p className="text-sm">
          Your AI evaluation expired on {expiredDate.toLocaleDateString()}. Re-evaluate to get updated skill scores and
          improve your job matches.
        </p>
      </div>
    </div>
  );
}
