const BASE_URL = import.meta.env.VITE_API_BASE_PATH ?? '/api/v1';

const ACCESS_TOKEN_KEY = 'accessToken';

export function getAccessToken(): string | null {
  return localStorage.getItem(ACCESS_TOKEN_KEY);
}

export function clearTokens() {
  localStorage.removeItem(ACCESS_TOKEN_KEY);
}

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

interface ErrorBody {
  error?: string;
}

function parseServerMessage(body: string): string | null {
  if (!body) return null;
  try {
    const obj = JSON.parse(body) as ErrorBody;
    return typeof obj.error === 'string' ? obj.error : null;
  } catch {
    return null;
  }
}

async function fetchJson(url: string, options?: RequestInit): Promise<{ ok: boolean; status: number; body: string }> {
  const res = await fetch(url, options);
  const body = await res.text().catch(() => '');
  return { ok: res.ok, status: res.status, body };
}

function throwApiError(status: number, body: string): never {
  const serverMessage = parseServerMessage(body);
  if (status === 401) {
    throw new AuthError(status, body, serverMessage);
  }
  throw new ApiError(status, body, serverMessage);
}

let refreshInFlight: Promise<string | null> | null = null;

async function refreshAccessToken(): Promise<string | null> {
  if (refreshInFlight) {
    return refreshInFlight;
  }
  const promise = (async (): Promise<string | null> => {
    try {
      const res = await fetchJson(`${BASE_URL}/auth/refresh`, {
        method: 'POST',
        credentials: 'include',
        headers: { 'Content-Type': 'application/json' },
      });
      if (!res.ok) {
        return null;
      }
      const data = JSON.parse(res.body) as { accessToken?: string; refreshToken?: string };
      if (!data.accessToken || typeof data.accessToken !== 'string') {
        return null;
      }
      setAccessToken(data.accessToken);
      return data.accessToken;
    } catch {
      return null;
    } finally {
      refreshInFlight = null;
    }
  })();
  refreshInFlight = promise;
  return promise;
}

function setAccessToken(token: string) {
  localStorage.setItem(ACCESS_TOKEN_KEY, token);
}

export async function api<T = unknown>(path: string, options: RequestInit = {}): Promise<T> {
  const headers = new Headers(options.headers);
  headers.set('Content-Type', 'application/json');

  const accessToken = getAccessToken();
  if (accessToken) headers.set('Authorization', `Bearer ${accessToken}`);

  let res = await fetchJson(`${BASE_URL}${path}`, {
    ...options,
    headers,
    credentials: 'include',
  });

  if (res.status === 401) {
    const newToken = await refreshAccessToken();
    if (newToken) {
      headers.set('Authorization', `Bearer ${newToken}`);
      res = await fetchJson(`${BASE_URL}${path}`, {
        ...options,
        headers,
        credentials: 'include',
      });
    }
  }

  if (!res.ok) {
    if (res.status === 401) {
      clearTokens();
    }
    throwApiError(res.status, res.body);
  }

  if (res.status === 204 || res.body === '') {
    return undefined as T;
  }

  try {
    return JSON.parse(res.body) as T;
  } catch {
    throwApiError(res.status, res.body);
  }
}

// apiUpload sends multipart form data (file uploads). The browser sets the
// Content-Type boundary itself — do not set it manually.
export async function apiUpload<T = unknown>(path: string, form: FormData): Promise<T> {
  const headers = new Headers();
  const accessToken = getAccessToken();
  if (accessToken) headers.set('Authorization', `Bearer ${accessToken}`);

  let res = await fetchJson(`${BASE_URL}${path}`, {
    method: 'POST',
    body: form,
    headers,
    credentials: 'include',
  });

  if (res.status === 401) {
    const newToken = await refreshAccessToken();
    if (newToken) {
      headers.set('Authorization', `Bearer ${newToken}`);
      res = await fetchJson(`${BASE_URL}${path}`, {
        method: 'POST',
        body: form,
        headers,
        credentials: 'include',
      });
    }
  }

  if (!res.ok) {
    if (res.status === 401) {
      clearTokens();
    }
    throwApiError(res.status, res.body);
  }

  return JSON.parse(res.body) as T;
}

export interface AuthUser {
  id: string;
  email: string;
  username: string;
  name: string;
  role: 'jobseeker' | 'company' | 'admin';
  isVerified: boolean;
}

export interface LoginResponse {
  accessToken: string;
  refreshToken?: string;
  user: AuthUser;
}

export async function login(email: string, password: string): Promise<LoginResponse> {
  const data = await api<LoginResponse>('/auth/login', {
    method: 'POST',
    body: JSON.stringify({ email, password }),
  });
  setAccessToken(data.accessToken);
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
    body: JSON.stringify(body),
  });
  setAccessToken(data.accessToken);
  return data;
}

export async function logout(): Promise<void> {
  try {
    await api('/auth/logout', { method: 'POST' });
  } catch {}
  clearTokens();
}
