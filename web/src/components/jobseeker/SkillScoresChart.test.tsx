import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { SkillScoresChart } from '@/components/jobseeker/SkillScoresChart';

const mockSkills = [
  { skill: 'Go', score: 90, category: 'backend' },
  { skill: 'React', score: 75, category: 'frontend' },
  { skill: 'PostgreSQL', score: 85, category: 'backend' },
];

describe('SkillScoresChart', () => {
  it('renders skill names', () => {
    render(<SkillScoresChart skillScores={mockSkills} />);
    expect(screen.getByText('Go')).toBeInTheDocument();
    expect(screen.getByText('React')).toBeInTheDocument();
    expect(screen.getByText('PostgreSQL')).toBeInTheDocument();
  });

  it('renders scores', () => {
    render(<SkillScoresChart skillScores={mockSkills} />);
    expect(screen.getByText('90')).toBeInTheDocument();
    expect(screen.getByText('75')).toBeInTheDocument();
    expect(screen.getByText('85')).toBeInTheDocument();
  });

  it('shows empty state', () => {
    render(<SkillScoresChart skillScores={[]} />);
    expect(screen.getByText(/no skill scores/i)).toBeInTheDocument();
  });

  it('groups by category', () => {
    render(<SkillScoresChart skillScores={mockSkills} />);
    expect(screen.getByText('backend')).toBeInTheDocument();
    expect(screen.getByText('frontend')).toBeInTheDocument();
  });
});
