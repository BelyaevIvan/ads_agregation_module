import { createContext, useContext, useEffect, useState } from 'react';
import { filtersApi } from '../api/filters';

const FiltersContext = createContext(null);

export function FiltersProvider({ children }) {
  const [sizes, setSizes] = useState(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let cancelled = false;
    (async () => {
      try {
        const data = await filtersApi.getSizes();
        if (!cancelled) setSizes(data);
      } catch {
        // тост уже показан клиентом; оставляем sizes = null,
        // секция фильтра размеров просто не отрисуется
      } finally {
        if (!cancelled) setLoading(false);
      }
    })();
    return () => {
      cancelled = true;
    };
  }, []);

  return (
    <FiltersContext.Provider value={{ sizes, loading }}>
      {children}
    </FiltersContext.Provider>
  );
}

// Хук для компонентов, которым нужны размеры.
// Возвращает { sizes, loading }. sizes === null, пока идёт загрузка либо
// если запрос упал. Потребитель должен просто не рендерить секцию, если
// sizes null, либо какой-то из массивов пустой.
export function useSizeFilters() {
  const ctx = useContext(FiltersContext);
  if (!ctx) throw new Error('useSizeFilters must be used inside FiltersProvider');
  return ctx;
}
