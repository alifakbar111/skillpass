import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { beforeEach, describe, expect, it } from 'vitest';
import { Navbar } from '@/components/layout/Navbar';
import { AuthProvider } from '@/hooks/useAuth';

function wrapper({ children }: { children: React.ReactNode }) {
  return (
    <MemoryRouter>
      <AuthProvider>{children}</AuthProvider>
    </MemoryRouter>
  );
}

describe('Navbar', () => {
  beforeEach(() => {
    localStorage.clear();
  });

  it('renders app name', () => {
    render(<Navbar />, { wrapper });
    expect(screen.getByText(/skillpass/i)).toBeInTheDocument();
  });

  it('shows login link when not authenticated', () => {
    render(<Navbar />, { wrapper });
    expect(screen.getByRole('link', { name: /login/i })).toBeInTheDocument();
  });

  it('shows register link when not authenticated', () => {
    render(<Navbar />, { wrapper });
    expect(screen.getByRole('link', { name: /register/i })).toBeInTheDocument();
  });
});
