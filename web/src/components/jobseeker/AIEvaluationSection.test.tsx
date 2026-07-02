import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { AIEvaluationSection } from '@/components/jobseeker/AIEvaluationSection';

const mocks = vi.hoisted(() => ({
  getLatestEvaluation: vi.fn(),
  triggerEvaluation: vi.fn(),
}));

vi.mock('@/lib/evaluation', () => ({
  getLatestEvaluation: mocks.getLatestEvaluation,
  triggerEvaluation: mocks.triggerEvaluation,
}));

function createWrapper() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  return function Wrapper({ children }: { children: React.ReactNode }) {
    return <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>;
  };
}

describe('AIEvaluationSection', () => {
  it('shows loading spinner while fetching', () => {
    mocks.getLatestEvaluation.mockReturnValue(new Promise(() => {}));
    render(<AIEvaluationSection />, { wrapper: createWrapper() });
    expect(document.querySelector('.loading')).toBeInTheDocument();
  });

  it('shows empty state when no evaluation exists', async () => {
    mocks.getLatestEvaluation.mockResolvedValue(null);
    render(<AIEvaluationSection />, { wrapper: createWrapper() });

    expect(await screen.findByText(/Run AI Evaluation/)).toBeInTheDocument();
    expect(screen.getByText(/get your skills evaluated/i)).toBeInTheDocument();
  });

  it('shows evaluation results when data is available', async () => {
    mocks.getLatestEvaluation.mockResolvedValue({
      overallScore: 85,
      strengths: [{ skill: 'React', score: 90, note: 'Strong component architecture' }],
      weaknesses: [{ skill: 'Go', score: 40, note: 'Needs more backend experience' }],
      suggestions: [{ area: 'Testing', tip: 'Add unit tests to your workflow' }],
      skillScores: [
        { skill: 'React', category: 'frontend', score: 90 },
        { skill: 'Go', category: 'backend', score: 40 },
      ],
    });
    render(<AIEvaluationSection />, { wrapper: createWrapper() });

    expect(await screen.findByText('85')).toBeInTheDocument();
    expect(screen.getByText(/Strong component architecture/)).toBeInTheDocument();
    expect(screen.getByText(/Needs more backend experience/)).toBeInTheDocument();
    expect(screen.getByText(/Add unit tests to your workflow/)).toBeInTheDocument();
    // Skill names appear in both strengths/weaknesses and the chart
    expect(screen.getAllByText('React').length).toBe(2);
    expect(screen.getAllByText('Go').length).toBe(2);
  });

  it('shows refresh button when evaluation exists', async () => {
    mocks.getLatestEvaluation.mockResolvedValue({
      overallScore: 60,
      strengths: [],
      weaknesses: [],
      suggestions: [],
      skillScores: [],
    });
    render(<AIEvaluationSection />, { wrapper: createWrapper() });

    expect(await screen.findByText(/Refresh Evaluation/)).toBeInTheDocument();
  });
});
