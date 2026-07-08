import { render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { EducationSection } from '@/components/jobseeker/EducationSection';

describe('EducationSection', () => {
  it('renders empty state and add button', () => {
    render(<EducationSection experiences={[]} onAdd={vi.fn()} onEdit={vi.fn()} onDelete={vi.fn()} />);
    expect(screen.getByRole('heading', { name: /Education/ })).toBeInTheDocument();
    expect(screen.getByText(/Add Education/)).toBeInTheDocument();
  });
});
