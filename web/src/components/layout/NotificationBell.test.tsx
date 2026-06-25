import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { describe, expect, it } from 'vitest';
import { NotificationBell } from '@/components/layout/NotificationBell';

function wrapper({ children }: { children: React.ReactNode }) {
  return <MemoryRouter>{children}</MemoryRouter>;
}

describe('NotificationBell', () => {
  it('renders bell icon button', () => {
    render(<NotificationBell />, { wrapper });
    expect(screen.getByRole('button', { name: /notifications/i })).toBeInTheDocument();
  });
});
