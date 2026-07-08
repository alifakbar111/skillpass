import { render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { VolunteeringSection } from '@/components/jobseeker/VolunteeringSection';

describe('VolunteeringSection', () => {
  it('renders empty state and add button', () => {
    render(<VolunteeringSection experiences={[]} onAdd={vi.fn()} onEdit={vi.fn()} onDelete={vi.fn()} />);
    expect(screen.getByRole('heading', { name: /Volunteering/ })).toBeInTheDocument();
    expect(screen.getByText(/Add Volunteering/)).toBeInTheDocument();
  });
});
