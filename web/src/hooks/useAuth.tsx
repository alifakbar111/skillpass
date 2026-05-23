import { createContext, type ReactNode, useCallback, useContext, useEffect, useState } from 'react';
import { api, login as apiLogin, logout as apiLogout, register as apiRegister, clearTokens } from '../lib/api';

interface User {
  id: string;
  email: string;
  username: string;
  name: string;
  role: 'jobseeker' | 'company';
}

interface AuthContextType {
  user: User | null;
  loading: boolean;
  login: (email: string, password: string) => Promise<void>;
  register: (data: {
    email: string;
    username: string;
    password: string;
    name: string;
    role: 'jobseeker' | 'company';
  }) => Promise<void>;
  logout: () => Promise<void>;
}

const AuthContext = createContext<AuthContextType | null>(null);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const accessToken = localStorage.getItem('accessToken');
    if (accessToken) {
      // Try fetching current user — if fails, tokens are stale
      api<{ id: string; email: string; username: string; name: string; role: 'jobseeker' | 'company' }>('/profiles/me')
        .then((u) => setUser(u as unknown as User))
        .catch(() => clearTokens())
        .finally(() => setLoading(false));
    } else {
      setLoading(false);
    }
  }, []);

  const login = useCallback(async (email: string, password: string) => {
    const data = await apiLogin(email, password);
    setUser(data.user);
  }, []);

  const register = useCallback(
    async (data: {
      email: string;
      username: string;
      password: string;
      name: string;
      role: 'jobseeker' | 'company';
    }) => {
      const res = await apiRegister(data);
      setUser(res.user);
    },
    [],
  );

  const logout = useCallback(async () => {
    await apiLogout();
    setUser(null);
  }, []);

  return <AuthContext.Provider value={{ user, loading, login, register, logout }}>{children}</AuthContext.Provider>;
}

export function useAuth() {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error('useAuth must be used within AuthProvider');
  return ctx;
}
