import { CheckCircle2, Circle } from 'lucide-react';
import { Link } from 'react-router-dom';

export interface ChecklistStep {
  id: string;
  label: string;
  done: boolean;
  /** Route to navigate to for this step. */
  to?: string;
  /** Inline action instead of navigation (e.g. open a form on this page). */
  onAction?: () => void;
  actionLabel?: string;
}

interface Props {
  title: string;
  steps: ChecklistStep[];
}

// ChecklistCard renders onboarding progress. It disappears once every step
// is done, so established users never see it.
export function ChecklistCard({ title, steps }: Props) {
  const doneCount = steps.filter((s) => s.done).length;
  if (steps.length === 0 || doneCount === steps.length) return null;

  return (
    <div className="card bg-base-200 border border-primary/30 p-4" data-testid="onboarding-checklist">
      <div className="flex justify-between items-center mb-1">
        <h2 className="font-semibold">{title}</h2>
        <span className="text-xs text-muted">
          {doneCount}/{steps.length} done
        </span>
      </div>
      <progress className="progress progress-primary w-full mb-3" value={doneCount} max={steps.length} />
      <ul className="space-y-2">
        {steps.map((step) => (
          <li key={step.id} className="flex items-center gap-2 text-sm">
            {step.done ? (
              <CheckCircle2 size={16} className="text-success shrink-0" aria-hidden="true" />
            ) : (
              <Circle size={16} className="text-base-content/30 shrink-0" aria-hidden="true" />
            )}
            <span className={step.done ? 'line-through opacity-50' : ''}>{step.label}</span>
            {!step.done && step.to && (
              <Link to={step.to} className="link link-primary text-xs ml-auto shrink-0">
                {step.actionLabel ?? 'Go'}
              </Link>
            )}
            {!step.done && step.onAction && (
              <button type="button" className="link link-primary text-xs ml-auto shrink-0" onClick={step.onAction}>
                {step.actionLabel ?? 'Start'}
              </button>
            )}
          </li>
        ))}
      </ul>
    </div>
  );
}
