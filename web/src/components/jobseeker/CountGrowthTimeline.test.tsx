import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { CountGrowthTimeline } from '@/components/jobseeker/CountGrowthTimeline';

function getByTextContent(text: string) {
  return screen.getByText((content) => content.includes(text));
}

describe('CountGrowthTimeline', () => {
  const sampleHistory = [
    { id: '1', overallScore: 1000, createdAt: '2026-01-15T00:00:00Z' },
    { id: '2', overallScore: 1500, createdAt: '2026-04-15T00:00:00Z' },
    { id: '3', overallScore: 2022, createdAt: '2026-07-08T00:00:00Z' },
  ];

  it('renders timeline entries', () => {
    render(<CountGrowthTimeline history={sampleHistory} />);
    expect(getByTextContent('1,000')).toBeInTheDocument();
    expect(getByTextContent('1,500')).toBeInTheDocument();
    expect(getByTextContent('2,022')).toBeInTheDocument();
  });

  it('shows growth indicators', () => {
    render(<CountGrowthTimeline history={sampleHistory} />);
    expect(screen.getByText('+500')).toBeInTheDocument();
    expect(screen.getByText('+522')).toBeInTheDocument();
  });

  it('handles empty history', () => {
    render(<CountGrowthTimeline history={[]} />);
    expect(screen.getByText(/No evaluation history/)).toBeInTheDocument();
  });
});
