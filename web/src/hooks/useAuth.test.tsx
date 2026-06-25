import { act, renderHook, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { beforeEach, describe, expect, it } from 'vitest';
import { AuthProvider, useAuth } from '@/hooks/useAuth';

function wrapper({ children }: { children: React.ReactNode }) {
  return (
    <MemoryRouter>
      <AuthProvider>{children}</AuthProvider>
    </MemoryRouter>
  );
}

describe('useAuth', () => {
  beforeEach(() => {
    localStorage.clear();
  });

  it('returns unauthenticated state by default', async () => {
    const { result } = renderHook(() => useAuth(), { wrapper });
    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(result.current.user).toBeNull();
  });

  it('sets user on successful login', async () => {
    const { result } = renderHook(() => useAuth(), { wrapper });
    await waitFor(() => expect(result.current.loading).toBe(false));

    await act(async () => {
      await result.current.login('test@example.com', 'password123');
    });

    expect(result.current.user?.email).toBe('test@example.com');
  });

  it('clears user on logout', async () => {
    const { result } = renderHook(() => useAuth(), { wrapper });
    await waitFor(() => expect(result.current.loading).toBe(false));

    await act(async () => {
      await result.current.login('test@example.com', 'password123');
    });

    expect(result.current.user).not.toBeNull();

    await act(async () => {
      await result.current.logout();
    });

    expect(result.current.user).toBeNull();
  });

  it('reads existing tokens from localStorage and loads user', async () => {
    localStorage.setItem('accessToken', 'existing-token');
    const { result } = renderHook(() => useAuth(), { wrapper });
    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(result.current.user).not.toBeNull();
    expect(result.current.user?.email).toBe('test@example.com');
  });
});
