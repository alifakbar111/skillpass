import type { FetchOptions } from 'ofetch';
import { FetchError, ofetch } from 'ofetch';
import type { z } from 'zod';
import type { LoginResponse, User } from './api-types';

export type { LoginResponse, User };
export type AuthUser = User;

const BASE_URL = import.meta.env.VITE_API_BASE_PATH ?? '/api/v1';
const ACCESS_TOKEN_KEY = 'accessToken';

// ── Token helpers ───────────────────────────────────────────────

export function getAccessToken(): string | null {
  return localStorage.getItem(ACCESS_TOKEN_KEY);
}

export function clearTokens() {
  localStorage.removeItem(ACCESS_TOKEN_KEY);
}

function setAccessToken(token: string) {
  localStorage.setItem(ACCESS_TOKEN_KEY, token);
}

// ── Error classes (backwards-compatible) ────────────────────────

export class ApiError extends Error {
  constructor(
    public status: number,
    public body: string,
    public serverMessage: string | null,
  ) {
    super(serverMessage ?? `HTTP ${status}`);
    this.name = 'ApiError';
  }
}

export class AuthError extends ApiError {
  constructor(status: number, body: string, serverMessage: string | null) {
    super(status, body, serverMessage);
    this.name = 'AuthError';
  }
}

export function isAuthError(err: unknown): err is AuthError {
  return err instanceof AuthError;
}

// ── Refresh dedup ───────────────────────────────────────────────

let refreshPromise: Promise<string | null> | null = null;

async function doRefresh(): Promise<string | null> {
  if (refreshPromise) return refreshPromise;
  refreshPromise = (async () => {
    try {
      const res = await fetch(`${BASE_URL}/auth/refresh`, {
        method: 'POST',
        credentials: 'include',
        headers: { 'Content-Type': 'application/json' },
      });
      if (!res.ok) {
        clearTokens();
        return null;
      }
      const data = (await res.json()) as { accessToken?: string };
      if (!data.accessToken || typeof data.accessToken !== 'string') {
        clearTokens();
        return null;
      }
      setAccessToken(data.accessToken);
      return data.accessToken;
    } catch {
      return null;
    } finally {
      refreshPromise = null;
    }
  })();
  return refreshPromise;
}

// ── Tuned ofetch instance ───────────────────────────────────────

const _ofetch = ofetch.create({
  baseURL: BASE_URL,
  credentials: 'include',
  onRequest({ options }) {
    const token = getAccessToken();
    if (token) {
      options.headers = new Headers(options.headers);
      options.headers.set('Authorization', `Bearer ${token}`);
    }
  },
});

// ── Exported api() — wraps ofetch with 401 auto-refresh ────────

export async function api<T = unknown>(path: string, options: FetchOptions<'json'> = {}): Promise<T> {
  try {
    return await _ofetch<T>(path, options);
  } catch (err) {
    // On 401, try refreshing the token and retry once
    if (err instanceof FetchError && err.status === 401) {
      const newToken = await doRefresh();
      if (newToken) {
        const headers = new Headers(options.headers as HeadersInit | undefined);
        headers.set('Authorization', `Bearer ${newToken}`);
        return _ofetch<T>(path, { ...options, headers });
      }
      // Refresh failed — throw AuthError
      const body = typeof err.data === 'string' ? err.data : JSON.stringify(err.data ?? '');
      throw new AuthError(401, body, 'Session expired. Please log in again.');
    }

    // All other FetchErrors → ApiError (backwards compat)
    if (err instanceof FetchError) {
      const body = typeof err.data === 'string' ? err.data : JSON.stringify(err.data ?? '');
      const message = (err.data as { error?: string } | null)?.error ?? err.message;
      throw new ApiError(err.status ?? 500, body, message);
    }

    // Network errors, etc. — rethrow as-is
    throw err;
  }
}

// ── Zod-validated fetch ────────────────────────────────────────

export async function apiWithSchema<T extends z.ZodTypeAny>(
  schema: T,
  path: string,
  options?: FetchOptions<'json'>,
): Promise<z.infer<T>> {
  return api(path, options).then((data) => schema.parse(data));
}

// ── File uploads (backwards-compatible thin wrapper) ────────────
// ofetch handles FormData natively — no special Content-Type needed.

export async function apiUpload<T = unknown>(path: string, form: FormData): Promise<T> {
  return api<T>(path, { method: 'POST', body: form });
}

// ── Convenience wrappers ───────────────────────────────────────

export async function login(email: string, password: string): Promise<LoginResponse> {
  const data = await api<LoginResponse>('/auth/login', {
    method: 'POST',
    body: { email, password },
  });
  if (data.accessToken) setAccessToken(data.accessToken);
  return data;
}

export async function register(body: {
  email: string;
  username: string;
  password: string;
  name?: string;
  role: 'jobseeker' | 'company';
  companyName?: string;
  businessRegistration?: string;
  website?: string;
  address?: string;
  contact?: string;
}): Promise<LoginResponse> {
  const data = await api<LoginResponse>('/auth/register', {
    method: 'POST',
    body,
  });
  if (data.accessToken) setAccessToken(data.accessToken);
  return data;
}

export async function logout(): Promise<void> {
  try {
    await api('/auth/logout', { method: 'POST' });
  } catch {
    // Swallow — we clear tokens either way
  }
  clearTokens();
}
