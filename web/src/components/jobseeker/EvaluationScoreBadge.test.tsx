import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { EvaluationScoreBadge } from '@/components/jobseeker/EvaluationScoreBadge';

describe('EvaluationScoreBadge', () => {
  it('renders score and label', () => {
    render(<EvaluationScoreBadge overallScore={85} />);
    expect(screen.getByText('85')).toBeInTheDocument();
  });

  it('shows Expert for high score', () => {
    render(<EvaluationScoreBadge overallScore={250} />);
    expect(screen.getByText('Expert')).toBeInTheDocument();
  });

  it('shows Advanced for mid-high score', () => {
    render(<EvaluationScoreBadge overallScore={150} />);
    expect(screen.getByText('Advanced')).toBeInTheDocument();
  });

  it('shows Intermediate for medium score', () => {
    render(<EvaluationScoreBadge overallScore={75} />);
    expect(screen.getByText('Intermediate')).toBeInTheDocument();
  });

  it('shows Beginner for low score', () => {
    render(<EvaluationScoreBadge overallScore={30} />);
    expect(screen.getByText('Beginner')).toBeInTheDocument();
  });
});
