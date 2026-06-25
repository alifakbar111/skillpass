import { render, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { beforeEach, describe, expect, it } from 'vitest';
import { ProtectedRoute } from '@/components/ui/ProtectedRoute';
import { AuthProvider } from '@/hooks/useAuth';

function wrapper({ children }: { children: React.ReactNode }) {
  return (
    <MemoryRouter>
      <AuthProvider>{children}</AuthProvider>
    </MemoryRouter>
  );
}

describe('ProtectedRoute', () => {
  beforeEach(() => {
    localStorage.clear();
  });

  it('renders children when authenticated', async () => {
    localStorage.setItem('accessToken', 'valid-token');
    const { getByText } = render(
      <ProtectedRoute>
        <div>Protected Content</div>
      </ProtectedRoute>,
      { wrapper },
    );
    await waitFor(() => {
      expect(getByText('Protected Content')).toBeInTheDocument();
    });
  });

  it('redirects to login when not authenticated', async () => {
    const { queryByText } = render(
      <ProtectedRoute>
        <div>Protected Content</div>
      </ProtectedRoute>,
      { wrapper },
    );
    await waitFor(() => {
      expect(queryByText('Protected Content')).not.toBeInTheDocument();
    });
  });

  it('redirects when role does not match', async () => {
    localStorage.setItem('accessToken', 'valid-token');
    const { queryByText } = render(
      <ProtectedRoute requiredRole="company">
        <div>Company Only</div>
      </ProtectedRoute>,
      { wrapper },
    );
    await waitFor(() => {
      expect(queryByText('Company Only')).not.toBeInTheDocument();
    });
  });
});
