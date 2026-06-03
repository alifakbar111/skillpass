import { createContext, type ReactNode, useCallback, useContext, useEffect, useState } from 'react';
import {
  api,
  login as apiLogin,
  logout as apiLogout,
  register as apiRegister,
  clearTokens,
  isAuthError,
} from '../lib/api';

interface User {
  id: string;
  email: string;
  username: string;
  name: string;
  role: 'jobseeker' | 'company' | 'admin';
}

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
}

const AuthContext = createContext<AuthContextType | null>(null);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const accessToken = localStorage.getItem('accessToken');
    if (accessToken) {
      api<User>('/profiles/me')
        .then((u) => setUser(u))
        .catch((err) => {
          if (isAuthError(err)) clearTokens();
        })
        .finally(() => setLoading(false));
    } else {
      setLoading(false);
    }
  }, []);

  const login = useCallback(async (email: string, password: string) => {
    const data = await apiLogin(email, password);
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
