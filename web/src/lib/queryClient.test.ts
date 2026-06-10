import { afterEach, describe, expect, it, vi } from 'vitest';
import { AuthError } from './api';
import { handleQueryError } from './queryClient';

describe('handleQueryError', () => {
  afterEach(() => {
    localStorage.clear();
    vi.restoreAllMocks();
  });

  it('clears tokens and dispatches auth:logout on AuthError', () => {
    localStorage.setItem('accessToken', 'tok');
    const spy = vi.fn();
    window.addEventListener('auth:logout', spy);

    handleQueryError(new AuthError(401, '', 'unauthorized'));

    expect(localStorage.getItem('accessToken')).toBeNull();
    expect(spy).toHaveBeenCalledTimes(1);
    window.removeEventListener('auth:logout', spy);
  });

  it('ignores non-auth errors', () => {
    localStorage.setItem('accessToken', 'tok');
    const spy = vi.fn();
    window.addEventListener('auth:logout', spy);

    handleQueryError(new Error('network'));

    expect(localStorage.getItem('accessToken')).toBe('tok');
    expect(spy).not.toHaveBeenCalled();
    window.removeEventListener('auth:logout', spy);
  });
});
