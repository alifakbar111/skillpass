import { render, screen } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import { ProjectSection } from '@/components/jobseeker/ProjectSection';

describe('ProjectSection', () => {
  it('renders empty state and add button', () => {
    render(<ProjectSection experiences={[]} onAdd={vi.fn()} onEdit={vi.fn()} onDelete={vi.fn()} />);
    expect(screen.getByText(/Projects/)).toBeInTheDocument();
    expect(screen.getByText(/Add Project/)).toBeInTheDocument();
  });
});
