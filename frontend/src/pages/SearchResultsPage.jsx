import { useEffect, useMemo, useState, useCallback } from 'react';
import { useSearchParams } from 'react-router-dom';
import { listingsApi } from '../api/listings';
import { favoritesApi } from '../api/favorites';
import { useAuth } from '../contexts/AuthContext';
import ListingCard from '../components/listings/ListingCard';
import FiltersPanel, { defaultFilters, filtersToParams } from '../components/listings/FiltersPanel';
import Pagination from '../components/common/Pagination';
import Spinner from '../components/common/Spinner';
import EmptyState from '../components/common/EmptyState';
import { formatNumber } from '../utils/format';

const LIMIT = 20;

function filtersFromParams(sp) {
  const f = defaultFilters();
  f.platform = sp.getAll('platform');
  f.condition = sp.get('condition') || '';
  f.size_rus = sp.getAll('size_rus');
  f.size_eu = sp.getAll('size_eu');
  f.size_us = sp.getAll('size_us');
  f.price_min = sp.get('price_min') || '';
  f.price_max = sp.get('price_max') || '';
  f.cityInput = sp.get('city') || '';
  f.brandInput = sp.get('brand') || '';
  const ins = sp.get('include_no_size');
  if (ins !== null) f.include_no_size = ins !== 'false';
  const inp = sp.get('include_no_price');
  if (inp !== null) f.include_no_price = inp !== 'false';
  const inc = sp.get('include_no_city');
  if (inc !== null) f.include_no_city = inc !== 'false';
  return f;
}

export default function SearchResultsPage() {
  const [sp, setSp] = useSearchParams();
  const { user } = useAuth();

  const [items, setItems] = useState([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(true);
  const [filtersOpen, setFiltersOpen] = useState(false);
  const [favIds, setFavIds] = useState(new Set());

  const q = sp.get('q') || '';
  const sort = sp.get('sort') || 'date_desc';
  const offset = Math.max(0, parseInt(sp.get('offset') || '0', 10));

  const filters = useMemo(() => filtersFromParams(sp), [sp]);

  const load = useCallback(async () => {
    setLoading(true);
    try {
      const params = {
        q: q || undefined,
        sort,
        limit: LIMIT,
        offset,
        ...filtersToParams(filters),
      };
      const res = await listingsApi.search(params);
      setItems(res.items || []);
      setTotal(res.total || 0);
    } catch {
      setItems([]);
      setTotal(0);
    } finally {
      setLoading(false);
    }
  }, [q, sort, offset, filters]);

  useEffect(() => {
    load();
  }, [load]);

  useEffect(() => {
    if (!user) {
      setFavIds(new Set());
      return;
    }
    let cancelled = false;
    (async () => {
      try {
        const res = await favoritesApi.list({ limit: 100 });
        if (!cancelled) {
          setFavIds(new Set((res.items || []).map((f) => f.listing.id)));
        }
      } catch {
        // silent
      }
    })();
    return () => {
      cancelled = true;
    };
  }, [user]);

  const applyFilters = (f) => {
    const p = new URLSearchParams();
    if (q) p.set('q', q);
    if (sort && sort !== 'date_desc') p.set('sort', sort);
    const params = filtersToParams(f);
    for (const [k, v] of Object.entries(params)) {
      if (Array.isArray(v)) v.forEach((x) => p.append(k, x));
      else if (v !== '' && v !== undefined && v !== null) p.set(k, String(v));
    }
    setSp(p);
  };

  const setSort = (s) => {
    const p = new URLSearchParams(sp);
    p.set('sort', s);
    p.delete('offset');
    setSp(p);
  };

  const setOffset = (o) => {
    const p = new URLSearchParams(sp);
    if (o) p.set('offset', String(o));
    else p.delete('offset');
    setSp(p);
  };

  const onFavToggle = (id, nowFav) => {
    setFavIds((s) => {
      const next = new Set(s);
      if (nowFav) next.add(id);
      else next.delete(id);
      return next;
    });
  };

  return (
    <div className="results-layout">
      <FiltersPanel
        value={filters}
        onApply={applyFilters}
        isOpen={filtersOpen}
        onToggleOpen={() => setFiltersOpen((v) => !v)}
      />
      <main className="results-main">
        <div className="results-header">
          <div>
            <div className="results-query">
              {q ? (
                <>
                  Результаты: <span>«{q}»</span>
                </>
              ) : (
                'Все объявления'
              )}
            </div>
            <div className="results-count" style={{ marginTop: 4 }}>
              {loading ? 'Загрузка…' : `Найдено ${formatNumber(total)} объявлений`}
            </div>
          </div>
          <select className="sort-select" value={sort} onChange={(e) => setSort(e.target.value)}>
            <option value="date_desc">По дате ↓</option>
            <option value="price_asc">По цене ↑</option>
            <option value="price_desc">По цене ↓</option>
          </select>
        </div>

        {loading ? (
          <Spinner />
        ) : items.length === 0 ? (
          <EmptyState
            icon="🔍"
            title="Ничего не найдено"
            subtitle="Попробуйте изменить фильтры или поисковый запрос"
          />
        ) : (
          <>
            <div className="cards-grid">
              {items.map((it) => (
                <ListingCard
                  key={it.id}
                  listing={it}
                  isFavorite={favIds.has(it.id)}
                  onToggleFavorite={onFavToggle}
                />
              ))}
            </div>
            <Pagination total={total} limit={LIMIT} offset={offset} onChange={setOffset} />
          </>
        )}
      </main>
    </div>
  );
}
