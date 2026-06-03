const BASE_URL = '/api/v1';

function getTokens() {
  const accessToken = localStorage.getItem('accessToken');
  const refreshToken = localStorage.getItem('refreshToken');
  return { accessToken, refreshToken };
}

function setTokens(accessToken: string, refreshToken?: string) {
  localStorage.setItem('accessToken', accessToken);
  if (refreshToken) localStorage.setItem('refreshToken', refreshToken);
}

export function clearTokens() {
  localStorage.removeItem('accessToken');
  localStorage.removeItem('refreshToken');
}

export function isAuthError(err: unknown): boolean {
  if (!(err instanceof Error)) return false;
  return err.message.includes('401') || err.message.includes('Unauthorized') || err.message.includes('Session expired');
}

async function fetchJson(url: string, options?: RequestInit): Promise<{ ok: boolean; status: number; body: string }> {
  const res = await fetch(url, options);
  const body = await res.text().catch(() => '');
  return { ok: res.ok, status: res.status, body };
}

function parseJson<T>(body: string): T {
  return JSON.parse(body) as T;
}

async function refreshAccessToken(): Promise<string | null> {
  const refreshToken = localStorage.getItem('refreshToken');
  if (!refreshToken) return null;

  try {
    const res = await fetchJson(`${BASE_URL}/auth/refresh`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ refreshToken }),
    });

    if (!res.ok) {
      clearTokens();
      return null;
    }

    let data: { accessToken?: string };
    try {
      data = parseJson<{ accessToken?: string }>(res.body);
    } catch {
      clearTokens();
      return null;
    }

    if (!data.accessToken || typeof data.accessToken !== 'string') {
      clearTokens();
      return null;
    }

    setTokens(data.accessToken);
    return data.accessToken;
  } catch {
    clearTokens();
    return null;
  }
}

export async function api<T = unknown>(path: string, options: RequestInit = {}): Promise<T> {
  const { accessToken } = getTokens();
  const headers = new Headers(options.headers);
  headers.set('Content-Type', 'application/json');
  if (accessToken) headers.set('Authorization', `Bearer ${accessToken}`);

  try {
    let res = await fetchJson(`${BASE_URL}${path}`, { ...options, headers });

    if (res.status === 401) {
      const newToken = await refreshAccessToken();
      if (newToken) {
        headers.set('Authorization', `Bearer ${newToken}`);
        res = await fetchJson(`${BASE_URL}${path}`, { ...options, headers });
      }
    }

    if (!res.ok) {
      throw new Error(res.body || `HTTP ${res.status}`);
    }

    return parseJson<T>(res.body);
  } catch (err) {
    if (err instanceof TypeError && err.message === 'Failed to fetch') {
      throw new Error('Network error — please check your connection');
    }
    throw err;
  }
}

export interface LoginResponse {
  accessToken: string;
  refreshToken: string;
  user: { id: string; email: string; username: string; name: string; role: 'jobseeker' | 'company' };
}

export async function login(email: string, password: string): Promise<LoginResponse> {
  const data = await api<LoginResponse>('/auth/login', {
    method: 'POST',
    body: JSON.stringify({ email, password }),
  });
  setTokens(data.accessToken, data.refreshToken);
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
  setTokens(data.accessToken, data.refreshToken);
  return data;
}

export async function logout(): Promise<void> {
  try {
    await api('/auth/logout', { method: 'POST' });
  } catch {}
  clearTokens();
}
