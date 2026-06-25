import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { ThemeToggle } from '@/components/ui/ThemeToggle';

describe('ThemeToggle', () => {
  it('renders toggle button', () => {
    render(<ThemeToggle />);
    expect(screen.getByRole('button', { name: /toggle theme/i })).toBeInTheDocument();
  });

  it('toggles theme on click', () => {
    render(<ThemeToggle />);
    const button = screen.getByRole('button', { name: /toggle theme/i });
    fireEvent.click(button);
    // Theme state is in localStorage or data attribute
    expect(document.documentElement.getAttribute('data-theme')).toBeTruthy();
  });
});
