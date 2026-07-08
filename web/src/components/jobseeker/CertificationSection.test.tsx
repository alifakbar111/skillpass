import { render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { CertificationSection } from '@/components/jobseeker/CertificationSection';

describe('CertificationSection', () => {
  it('renders empty state and add button', () => {
    render(<CertificationSection experiences={[]} onAdd={vi.fn()} onEdit={vi.fn()} onDelete={vi.fn()} />);
    expect(screen.getByText(/Certifications/)).toBeInTheDocument();
    expect(screen.getByText(/Add Certification/)).toBeInTheDocument();
  });
});
