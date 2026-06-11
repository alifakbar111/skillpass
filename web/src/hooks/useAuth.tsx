import { createContext, type ReactNode, useCallback, useContext, useEffect, useMemo, useState } from 'react';
import {
  type AuthUser,
  api,
  login as apiLogin,
  logout as apiLogout,
  register as apiRegister,
  clearTokens,
  isAuthError,
} from '../lib/api';
import { queryClient } from '../lib/queryClient';

type User = AuthUser;

interface AuthContextType {
  user: User | null;
  loading: boolean;
  login: (email: string, password: string) => Promise<User>;
  register: (data: {
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
  }) => Promise<void>;
  logout: () => Promise<void>;
  refreshUser: () => Promise<void>;
}

const AuthContext = createContext<AuthContextType | null>(null);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const onStorage = (e: StorageEvent) => {
      if (e.key === 'accessToken' && !e.newValue) {
        queryClient.clear();
        setUser(null);
      }
    };
    const onAuthLogout = () => {
      queryClient.clear();
      setUser(null);
    };
    window.addEventListener('storage', onStorage);
    window.addEventListener('auth:logout', onAuthLogout);
    return () => {
      window.removeEventListener('storage', onStorage);
      window.removeEventListener('auth:logout', onAuthLogout);
    };
  }, []);

  useEffect(() => {
    const accessToken = localStorage.getItem('accessToken');
    if (accessToken) {
      // /auth/me works for every role (unlike /profiles/me, which is
      // jobseeker-only and used to log company users out on refresh).
      api<User>('/auth/me')
        .then((u) => setUser(u))
        .catch((err) => {
          if (isAuthError(err)) {
            clearTokens();
            setUser(null);
          }
        })
        .finally(() => setLoading(false));
    } else {
      setLoading(false);
    }
  }, []);

  const login = useCallback(async (email: string, password: string) => {
    const data = await apiLogin(email, password);
    if (!data.user) throw new Error('Login response missing user');
    setUser(data.user);
    return data.user;
  }, []);

  const register = useCallback(
    async (data: {
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
    }) => {
      const res = await apiRegister(data);
      if (res.user) setUser(res.user);
    },
    [],
  );

  const logout = useCallback(async () => {
    await apiLogout();
    queryClient.clear();
    setUser(null);
  }, []);

  // refreshUser re-reads /auth/me, e.g. after email verification.
  const refreshUser = useCallback(async () => {
    try {
      const u = await api<User>('/auth/me');
      setUser(u);
    } catch {
      // keep current state; this is a soft refresh
    }
  }, []);

  const value = useMemo<AuthContextType>(
    () => ({ user, loading, login, register, logout, refreshUser }),
    [user, loading, login, register, logout, refreshUser],
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth() {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error('useAuth must be used within AuthProvider');
  return ctx;
}
