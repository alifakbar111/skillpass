import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import type { Application } from '../../lib/application';
import { ApplicationKanban } from './ApplicationKanban';

const baseApp: Application = {
  id: 'app-1',
  jobseekerId: 'js-1',
  jobPostingId: 'job-1',
  status: 'applied',
  createdAt: '2026-06-01T00:00:00Z',
  updatedAt: '2026-06-01T00:00:00Z',
  jobTitle: 'Backend Engineer',
  companyName: 'Acme Corp',
};

describe('ApplicationKanban', () => {
  it('renders applications in their status column', () => {
    render(<ApplicationKanban applications={[baseApp]} />);

    expect(screen.getByText('Applied')).toBeInTheDocument();
    expect(screen.getByText('Backend Engineer')).toBeInTheDocument();
    expect(screen.getByText('Acme Corp')).toBeInTheDocument();
  });

  it('shows the latest company note when present', () => {
    render(<ApplicationKanban applications={[{ ...baseApp, latestNote: "We'd love to chat next week" }]} />);
    expect(screen.getByText(/We'd love to chat next week/)).toBeInTheDocument();
  });

  it('shows empty state when a column has no applications', () => {
    render(<ApplicationKanban applications={[]} />);
    expect(screen.getAllByText('No applications').length).toBeGreaterThan(0);
  });
});
