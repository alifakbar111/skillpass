import { render, screen } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import { WorkHistorySection } from '@/components/jobseeker/WorkHistorySection';

describe('WorkHistorySection', () => {
  it('renders with add button and empty state', () => {
    render(<WorkHistorySection experiences={[]} onAdd={vi.fn()} onEdit={vi.fn()} onDelete={vi.fn()} />);
    expect(screen.getByText(/Work History/)).toBeInTheDocument();
    expect(screen.getByText(/Add Work/)).toBeInTheDocument();
    expect(screen.getByText(/No work history added/)).toBeInTheDocument();
  });

  it('renders employment entries', () => {
    render(
      <WorkHistorySection
        experiences={[
          {
            id: '1',
            title: 'Senior Dev',
            organization: 'Co',
            startDate: '2020-01',
            endDate: null,
            isCurrent: true,
            type: 'employment',
            description: null,
            industry: 'Tech',
            skillsUsed: ['Go'],
            url: null,
          },
        ]}
        onAdd={vi.fn()}
        onEdit={vi.fn()}
        onDelete={vi.fn()}
      />,
    );
    expect(screen.getByText('Senior Dev')).toBeInTheDocument();
    expect(screen.getByText(/Co/)).toBeInTheDocument();
  });
});
