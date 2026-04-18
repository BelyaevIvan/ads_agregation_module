import { useEffect, useState } from 'react';
import { useOutletContext } from 'react-router-dom';
import { adminApi } from '../../api/admin';
import Spinner from '../../components/common/Spinner';
import { formatNumber, formatDateShort } from '../../utils/format';

export default function DashboardPage() {
  const { user, avatar } = useOutletContext();
  const [stats, setStats] = useState(null);
  const [period, setPeriod] = useState(30);
  const [byDay, setByDay] = useState([]);
  const [brands, setBrands] = useState([]);
  const [cities, setCities] = useState([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let cancelled = false;
    setLoading(true);
    (async () => {
      try {
        const [s, bd, tb, tc] = await Promise.all([
          adminApi.stats(),
          adminApi.listingsByDay(period),
          adminApi.topBrands(5),
          adminApi.topCities(5),
        ]);
        if (cancelled) return;
        setStats(s);
        setByDay((bd.items || []).slice().reverse());
        setBrands(tb.items || []);
        setCities(tc.items || []);
      } catch {
        // toast fired
      } finally {
        if (!cancelled) setLoading(false);
      }
    })();
    return () => {
      cancelled = true;
    };
  }, [period]);

  const maxByDay = Math.max(1, ...byDay.map((d) => d.count));
  const maxBrand = Math.max(1, ...brands.map((b) => b.count));
  const maxCity = Math.max(1, ...cities.map((c) => c.count));

  return (
    <>
      <div className="admin-topbar">
        <div className="admin-page-title">Дашборд</div>
        <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
          <div style={{ textAlign: 'right' }}>
            <div style={{ fontSize: 13, fontWeight: 600 }}>Администратор</div>
            <div style={{ fontSize: 11, color: 'var(--c-muted)' }}>{user?.email}</div>
          </div>
          <div className="admin-avatar">{avatar}</div>
        </div>
      </div>

      {loading && !stats ? (
        <Spinner />
      ) : stats ? (
        <>
          <div className="stats-grid">
            <StatCard
              icon="📊"
              value={stats.total_listings}
              label="Всего объявлений"
              delta={`▲ +${stats.new_listings_today} за сегодня`}
            />
            <StatCard
              icon="📡"
              value={stats.active_sources}
              label="Активных источников"
            />
            <StatCard
              icon="👥"
              value={stats.total_users}
              label="Пользователей"
              delta={`▲ +${stats.new_users_week} за неделю`}
            />
            <StatCard
              icon="🚫"
              value={stats.hidden_listings}
              label="Скрытых объявлений"
            />
          </div>

          <div className="chart-block">
            <div className="chart-header">
              <div className="chart-title">Новые объявления по дням</div>
              <div className="chart-period">
                {[7, 30, 90].map((d) => (
                  <button
                    key={d}
                    className={`cp-btn ${period === d ? 'active' : ''}`}
                    onClick={() => setPeriod(d)}
                  >
                    {d === 7 ? '7 дн' : d === 30 ? '30 дн' : '3 мес'}
                  </button>
                ))}
              </div>
            </div>
            {byDay.length === 0 ? (
              <div
                style={{ padding: '30px 0', color: 'var(--c-muted)', textAlign: 'center' }}
              >
                Данных пока нет
              </div>
            ) : (
              <>
                <div className="chart-area">
                  {byDay.map((d, i) => (
                    <div
                      key={d.date}
                      className={`bar ${i === byDay.length - 1 ? 'today' : ''}`}
                      style={{ height: `${(d.count / maxByDay) * 100}%` }}
                      title={`${d.date}: ${d.count}`}
                    />
                  ))}
                </div>
                <div className="chart-labels">
                  {byDay.map((d, i) => (
                    <div
                      key={d.date}
                      className="chart-label"
                      style={i === byDay.length - 1 ? { color: 'var(--c-text)' } : {}}
                    >
                      {formatDateShort(d.date)}
                    </div>
                  ))}
                </div>
              </>
            )}
          </div>

          <div className="two-charts">
            <div className="chart-block" style={{ marginBottom: 0 }}>
              <div className="chart-header">
                <div className="chart-title">Топ брендов</div>
              </div>
              {brands.length === 0 ? (
                <div style={{ color: 'var(--c-muted)', fontSize: 13 }}>Пока нет данных</div>
              ) : (
                brands.map((b) => (
                  <div className="bar-h-row" key={b.brand}>
                    <div className="bar-h-label">{b.brand}</div>
                    <div className="bar-h-track">
                      <div className="bar-h-fill" style={{ width: `${(b.count / maxBrand) * 100}%` }} />
                    </div>
                    <div className="bar-h-val">{formatNumber(b.count)}</div>
                  </div>
                ))
              )}
            </div>
            <div className="chart-block" style={{ marginBottom: 0 }}>
              <div className="chart-header">
                <div className="chart-title">Топ городов</div>
              </div>
              {cities.length === 0 ? (
                <div style={{ color: 'var(--c-muted)', fontSize: 13 }}>Пока нет данных</div>
              ) : (
                cities.map((c) => (
                  <div className="bar-h-row" key={c.city}>
                    <div className="bar-h-label">{c.city}</div>
                    <div className="bar-h-track">
                      <div className="bar-h-fill" style={{ width: `${(c.count / maxCity) * 100}%` }} />
                    </div>
                    <div className="bar-h-val">{formatNumber(c.count)}</div>
                  </div>
                ))
              )}
            </div>
          </div>
        </>
      ) : null}
    </>
  );
}

function StatCard({ icon, value, label, delta }) {
  return (
    <div className="stat-card">
      <div className="stat-card-icon">{icon}</div>
      <div className="stat-card-val">{formatNumber(value)}</div>
      <div className="stat-card-label">{label}</div>
      {delta && <div className="stat-card-delta">{delta}</div>}
    </div>
  );
}
