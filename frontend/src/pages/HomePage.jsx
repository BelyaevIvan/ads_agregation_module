import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { listingsApi } from '../api/listings';
import { formatNumber } from '../utils/format';

const POPULAR_TAGS = ['Nike', 'Stone Island', 'New Balance', 'Balenciaga', 'iPhone', 'Jordan', 'Adidas'];

export default function HomePage() {
  const navigate = useNavigate();
  const [query, setQuery] = useState('');
  const [totalListings, setTotalListings] = useState(null);

  useEffect(() => {
    let cancelled = false;
    (async () => {
      try {
        const res = await listingsApi.search({ limit: 1 });
        if (!cancelled) setTotalListings(res.total);
      } catch {
        // silent
      }
    })();
    return () => {
      cancelled = true;
    };
  }, []);

  const search = (q) => {
    const trimmed = (q ?? query).trim();
    const params = new URLSearchParams();
    if (trimmed) params.set('q', trimmed);
    navigate('/search' + (params.toString() ? `?${params}` : ''));
  };

  return (
    <div className="hero">
      <div className="hero-glow" />
      <div className="hero-glow2" />
      {totalListings !== null && (
        <div className="hero-badge">
          <span />
          {formatNumber(totalListings)} объявлений уже в базе
        </div>
      )}
      <h1 className="hero-title">
        Бренды с рук —<br />
        <span className="grad">в одном месте</span>
      </h1>
      <p className="hero-sub">
        Агрегируем объявления из Telegram и ВКонтакте. Nike, Stone Island, Balenciaga — всё найдётся
        по выгодной цене.
      </p>
      <form
        className="search-wrap"
        onSubmit={(e) => {
          e.preventDefault();
          search();
        }}
      >
        <input
          type="text"
          placeholder="Найти: Nike Air Max 90, Stone Island, iPhone 15…"
          value={query}
          onChange={(e) => setQuery(e.target.value)}
        />
        <button type="submit" className="search-btn">
          Найти →
        </button>
      </form>
      <div className="hero-tags">
        {POPULAR_TAGS.map((t) => (
          <span key={t} className="tag" onClick={() => search(t)}>
            {t}
          </span>
        ))}
      </div>
      <div className="hero-stats">
        <div className="stat-item">
          <div className="stat-num">{totalListings !== null ? formatNumber(totalListings) : '—'}</div>
          <div className="stat-label">объявлений</div>
        </div>
        <div className="stat-item">
          <div className="stat-num">24/7</div>
          <div className="stat-label">мониторинг</div>
        </div>
        <div className="stat-item">
          <div className="stat-num">TG + VK</div>
          <div className="stat-label">источники</div>
        </div>
      </div>
    </div>
  );
}
