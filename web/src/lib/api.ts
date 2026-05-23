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

async function refreshAccessToken(): Promise<string | null> {
  const refreshToken = localStorage.getItem('refreshToken');
  if (!refreshToken) return null;

  const res = await fetch(`${BASE_URL}/auth/refresh`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ refreshToken }),
  });

  if (!res.ok) {
    clearTokens();
    return null;
  }
  const data = await res.json();
  setTokens(data.accessToken);
  return data.accessToken;
}

export async function api<T = unknown>(path: string, options: RequestInit = {}): Promise<T> {
  const { accessToken } = getTokens();
  const headers = new Headers(options.headers);
  headers.set('Content-Type', 'application/json');
  if (accessToken) headers.set('Authorization', `Bearer ${accessToken}`);

  let res = await fetch(`${BASE_URL}${path}`, { ...options, headers });

  if (res.status === 401) {
    const newToken = await refreshAccessToken();
    if (newToken) {
      headers.set('Authorization', `Bearer ${newToken}`);
      res = await fetch(`${BASE_URL}${path}`, { ...options, headers });
    }
  }

  if (!res.ok) {
    const err = await res.text();
    throw new Error(err || `HTTP ${res.status}`);
  }

  return res.json();
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
  name: string;
  role: 'jobseeker' | 'company';
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
