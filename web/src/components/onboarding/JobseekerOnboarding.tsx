import { useQuery } from '@tanstack/react-query';
import { ChecklistCard, type ChecklistStep } from '@/components/onboarding/ChecklistCard';
import { getLatestEvaluation } from '@/lib/evaluation';

interface Props {
  hasHeadline: boolean;
  experienceCount: number;
  /** Opens the experience entry UI on the profile page (resume import / form). */
  onAddExperience: () => void;
}

// JobseekerOnboarding guides a fresh account to its first "aha":
// profile basics -> first experience -> AI evaluation (which unlocks matches).
export function JobseekerOnboarding({ hasHeadline, experienceCount, onAddExperience }: Props) {
  // Share the same query key as AIEvaluationSection so TanStack Query deduplicates.
  const { data: evaluation, isLoading } = useQuery({
    queryKey: ['evaluation', 'latest'],
    queryFn: getLatestEvaluation,
  });

  const hasEvaluation = evaluation !== undefined && evaluation !== null;

  // Wait for the evaluation check so the card doesn't flash a wrong state.
  if (isLoading) return null;

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
