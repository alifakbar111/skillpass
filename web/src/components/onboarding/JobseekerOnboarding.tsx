import { useEffect, useState } from 'react';
import { getLatestEvaluation } from '../../lib/evaluation';
import { ChecklistCard, type ChecklistStep } from './ChecklistCard';

interface Props {
  hasHeadline: boolean;
  experienceCount: number;
  /** Opens the experience entry UI on the profile page (resume import / form). */
  onAddExperience: () => void;
}

// JobseekerOnboarding guides a fresh account to its first "aha":
// profile basics -> first experience -> AI evaluation (which unlocks matches).
export function JobseekerOnboarding({ hasHeadline, experienceCount, onAddExperience }: Props) {
  const [hasEvaluation, setHasEvaluation] = useState<boolean | null>(null);

  useEffect(() => {
    let cancelled = false;
    getLatestEvaluation()
      .then(() => {
        if (!cancelled) setHasEvaluation(true);
      })
      .catch(() => {
        if (!cancelled) setHasEvaluation(false);
      });
    return () => {
      cancelled = true;
    };
  }, []);

  // Wait for the evaluation check so the card doesn't flash a wrong state.
  if (hasEvaluation === null) return null;

  const steps: ChecklistStep[] = [
    {
      id: 'headline',
      label: 'Add a headline to your profile',
      done: hasHeadline,
    },
    {
      id: 'experience',
      label: 'Add your first experience (paste or upload your resume)',
      done: experienceCount > 0,
      onAction: onAddExperience,
      actionLabel: 'Add now',
    },
    {
      id: 'evaluation',
      label: 'Run your AI skill evaluation to unlock job matches',
      done: hasEvaluation,
      to: '/jobseeker/evaluation',
      actionLabel: 'Evaluate',
    },
  ];

  return <ChecklistCard title="Set up your skill passport" steps={steps} />;
}
