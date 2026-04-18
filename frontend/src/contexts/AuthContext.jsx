import { createContext, useCallback, useContext, useEffect, useState } from 'react';
import { authApi } from '../api/auth';
import {
  getToken,
  setToken,
  setErrorHandler,
  setUnauthorizedHandler,
} from '../api/client';
import { useToast } from './ToastContext';

const AuthContext = createContext(null);

export function AuthProvider({ children }) {
  const toast = useToast();
  const [user, setUser] = useState(null);
  const [loading, setLoading] = useState(true);

  const logout = useCallback(async () => {
    try {
      if (getToken()) await authApi.logout();
    } catch {
      // token might already be invalid — ignore
    }
    setToken(null);
    setUser(null);
  }, []);

  useEffect(() => {
    setErrorHandler(({ network, status, data }) => {
      if (network) {
        toast.error('Ошибка сети. Проверьте подключение');
        return;
      }
      const serverMsg = data && data.message;
      toast.error('Ошибка. Не удалось выполнить запрос', serverMsg);
    });
    setUnauthorizedHandler(() => {
      setToken(null);
      setUser(null);
    });
  }, [toast]);

  useEffect(() => {
    let cancelled = false;
    (async () => {
      const token = getToken();
      if (!token) {
        setLoading(false);
        return;
      }
      try {
        const me = await authApi.me();
        if (!cancelled) setUser(me);
      } catch {
        if (!cancelled) {
          setToken(null);
          setUser(null);
        }
      } finally {
        if (!cancelled) setLoading(false);
      }
    })();
    return () => {
      cancelled = true;
    };
  }, []);

  const login = useCallback(async (email, password) => {
    const tokenResp = await authApi.login(email, password);
    setToken(tokenResp.access_token);
    const me = await authApi.me();
    setUser(me);
    return me;
  }, []);

  const register = useCallback(async (email, password) => {
    await authApi.register(email, password);
    const tokenResp = await authApi.login(email, password);
    setToken(tokenResp.access_token);
    const me = await authApi.me();
    setUser(me);
    return me;
  }, []);

  const updateUser = useCallback((data) => {
    setUser((u) => (u ? { ...u, ...data } : u));
  }, []);

  return (
    <AuthContext.Provider
      value={{ user, loading, login, register, logout, updateUser, isAdmin: user?.role === 'admin' }}
    >
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error('useAuth must be used inside AuthProvider');
  return ctx;
}
