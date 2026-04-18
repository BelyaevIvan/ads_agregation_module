import { useEffect, useState, useCallback } from 'react';
import { adminApi } from '../../api/admin';
import { useToast } from '../../contexts/ToastContext';
import Spinner from '../../components/common/Spinner';
import EmptyState from '../../components/common/EmptyState';
import Pagination from '../../components/common/Pagination';
import AddSourceModal from '../../components/admin/AddSourceModal';
import { formatNumber } from '../../utils/format';

const LIMIT = 50;

export default function SourcesPage() {
  const toast = useToast();
  const [items, setItems] = useState([]);
  const [total, setTotal] = useState(0);
  const [offset, setOffset] = useState(0);
  const [loading, setLoading] = useState(true);
  const [adding, setAdding] = useState(false);

  const load = useCallback(async () => {
    setLoading(true);
    try {
      const res = await adminApi.sources({ limit: LIMIT, offset });
      setItems(res.items || []);
      setTotal(res.total || 0);
    } catch {
      setItems([]);
    } finally {
      setLoading(false);
    }
  }, [offset]);

  useEffect(() => {
    load();
  }, [load]);

  const toggle = async (source) => {
    try {
      await adminApi.toggleSource(source.id, !source.is_active);
      setItems((list) =>
        list.map((s) => (s.id === source.id ? { ...s, is_active: !s.is_active } : s))
      );
      toast.success(source.is_active ? 'Источник отключён' : 'Источник включён');
    } catch {
      // toast fired
    }
  };

  return (
    <>
      <div className="admin-topbar">
        <div className="admin-page-title">Источники мониторинга</div>
        <button
          type="button"
          className="nav-btn primary"
          style={{ fontSize: 12, padding: '8px 16px' }}
          onClick={() => setAdding(true)}
        >
          + Добавить
        </button>
      </div>

      {loading ? (
        <Spinner />
      ) : items.length === 0 ? (
        <EmptyState
          icon="📡"
          title="Источников пока нет"
          subtitle="Добавьте первый канал Telegram или группу ВКонтакте"
        />
      ) : (
        <>
          <div className="sources-grid">
            {items.map((s) => (
              <div key={s.id} className="source-card">
                <div className="source-icon">{s.platform === 'vk' ? '📘' : '✈️'}</div>
                <div className="source-info">
                  <div className="source-name">{s.title || s.external_id}</div>
                  <div className="source-meta">
                    {s.platform === 'vk' ? 'ВКонтакте' : 'Telegram'} ·{' '}
                    {formatNumber(s.listings_count)} объявлений
                    {!s.is_active && ' · Отключён'}
                  </div>
                </div>
                <button
                  type="button"
                  className={`source-toggle ${s.is_active ? '' : 'off'}`}
                  onClick={() => toggle(s)}
                  aria-label={s.is_active ? 'Отключить' : 'Включить'}
                />
              </div>
            ))}
          </div>
          <Pagination total={total} limit={LIMIT} offset={offset} onChange={setOffset} />
        </>
      )}

      {adding && (
        <AddSourceModal
          onClose={() => setAdding(false)}
          onCreated={() => {
            setAdding(false);
            toast.success('Источник добавлен');
            load();
          }}
        />
      )}
    </>
  );
}
