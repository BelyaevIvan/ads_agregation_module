import { createContext, useCallback, useContext, useEffect, useRef, useState } from 'react';

const ToastContext = createContext(null);

export function ToastProvider({ children }) {
  const [toasts, setToasts] = useState([]);
  const idRef = useRef(0);

  const remove = useCallback((id) => {
    setToasts((list) => list.filter((t) => t.id !== id));
  }, []);

  const push = useCallback(
    (toast) => {
      const id = ++idRef.current;
      const item = { id, type: 'error', duration: 5000, ...toast };
      setToasts((list) => [...list, item]);
      if (item.duration) {
        setTimeout(() => remove(id), item.duration);
      }
      return id;
    },
    [remove]
  );

  const error = useCallback(
    (message, sub) =>
      push({
        type: 'error',
        title: message || 'Ошибка. Не удалось выполнить запрос',
        sub,
      }),
    [push]
  );

  const success = useCallback(
    (message, sub) => push({ type: 'success', title: message, sub }),
    [push]
  );

  return (
    <ToastContext.Provider value={{ toasts, push, error, success, remove }}>
      {children}
    </ToastContext.Provider>
  );
}

export function useToast() {
  const ctx = useContext(ToastContext);
  if (!ctx) throw new Error('useToast must be used inside ToastProvider');
  return ctx;
}
