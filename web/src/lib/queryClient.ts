import { MutationCache, QueryCache, QueryClient } from '@tanstack/react-query';
import { clearTokens, isAuthError } from './api';

export const AUTH_LOGOUT_EVENT = 'auth:logout';

/**
 * Global handler for query/mutation errors. On an authentication failure
 * (401 after refresh failed), clear the stored token and notify AuthProvider
 * to reset `user` in the current tab (the `storage` event only fires cross-tab).
 */
export function handleQueryError(error: unknown): void {
  if (isAuthError(error)) {
    clearTokens();
    window.dispatchEvent(new CustomEvent(AUTH_LOGOUT_EVENT));
  }
}

export const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 60_000,
      retry: 1,
    },
  },
  queryCache: new QueryCache({
    onError: handleQueryError,
  }),
  mutationCache: new MutationCache({
    onError: handleQueryError,
  }),
});
