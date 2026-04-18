import { useEffect, useState, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import { adminApi } from '../../api/admin';
import { useToast } from '../../contexts/ToastContext';
import Spinner from '../../components/common/Spinner';
import EmptyState from '../../components/common/EmptyState';
import Pagination from '../../components/common/Pagination';
import EditTextModal from '../../components/admin/EditTextModal';
import PhotosModal from '../../components/admin/PhotosModal';
import { formatPrice, formatRelative, platformLabel } from '../../utils/format';

const LIMIT = 20;

export default function AdminListingsPage() {
  const navigate = useNavigate();
  const toast = useToast();

  const [items, setItems] = useState([]);
  const [total, setTotal] = useState(0);
  const [offset, setOffset] = useState(0);
  const [status, setStatus] = useState('');
  const [q, setQ] = useState('');
  const [qDraft, setQDraft] = useState('');
  const [loading, setLoading] = useState(true);

  const [editTextFor, setEditTextFor] = useState(null);
  const [photosFor, setPhotosFor] = useState(null);

  const load = useCallback(async () => {
    setLoading(true);
    try {
      const res = await adminApi.listings({
        status: status || undefined,
        q: q || undefined,
        limit: LIMIT,
        offset,
      });
      setItems(res.items || []);
      setTotal(res.total || 0);
    } catch {
      setItems([]);
    } finally {
      setLoading(false);
    }
  }, [status, q, offset]);

  useEffect(() => {
    load();
  }, [load]);

  const submitSearch = (e) => {
    e.preventDefault();
    setQ(qDraft.trim());
    setOffset(0);
  };

  const onStatus = (s) => {
    setStatus(s);
    setOffset(0);
  };

  const onVisibility = async (listing) => {
    try {
      await adminApi.setVisibility(listing.id, !listing.is_hidden);
      toast.success(listing.is_hidden ? 'Объявление восстановлено' : 'Объявление скрыто');
      load();
    } catch {
      // toast fired
    }
  };

  const onEditText = async (listing) => {
    try {
      const full = await adminApi.getListing(listing.id);
      setEditTextFor({
        id: listing.id,
        title: `${full.brand || ''} ${full.model || ''}`.trim(),
        text: full.original_text || '',
      });
    } catch {
      // toast fired
    }
  };

  const onManagePhotos = async (listing) => {
    try {
      const full = await adminApi.getListing(listing.id);
      setPhotosFor({
        id: listing.id,
        title: `${full.brand || ''} ${full.model || ''}`.trim(),
        photos: full.photos || [],
      });
    } catch {
      // toast fired
    }
  };

  return (
    <>
      <div className="admin-topbar">
        <div className="admin-page-title">Объявления</div>
      </div>
      <div className="table-block">
        <div className="table-header">
          <div className="table-title">Все объявления</div>
          <div className="table-controls">
            <form onSubmit={submitSearch}>
              <input
                className="tbl-search"
                type="text"
                placeholder="🔍 Поиск..."
                value={qDraft}
                onChange={(e) => setQDraft(e.target.value)}
              />
            </form>
            <button
              type="button"
              className={`filter-chip ${status === '' ? 'active' : ''}`}
              onClick={() => onStatus('')}
            >
              Все
            </button>
            <button
              type="button"
              className={`filter-chip ${status === 'active' ? 'active' : ''}`}
              onClick={() => onStatus('active')}
            >
              Активные
            </button>
            <button
              type="button"
              className={`filter-chip ${status === 'hidden' ? 'active' : ''}`}
              onClick={() => onStatus('hidden')}
            >
              Скрытые
            </button>
          </div>
        </div>

        {loading ? (
          <Spinner />
        ) : items.length === 0 ? (
          <EmptyState icon="📭" title="Объявлений не найдено" />
        ) : (
          <div className="tbl-wrap">
            <table className="admin-table">
              <thead>
                <tr>
                  <th>Товар</th>
                  <th>Источник</th>
                  <th>Цена</th>
                  <th>Статус</th>
                  <th>Дата</th>
                  <th>Действия</th>
                </tr>
              </thead>
              <tbody>
                {items.map((it) => (
                  <tr key={it.id}>
                    <td>
                      <div style={{ display: 'flex', gap: 9, alignItems: 'center' }}>
                        <div className="tbl-thumb">
                          {it.cover_photo_url ? (
                            <img src={it.cover_photo_url} alt="" />
                          ) : (
                            '🖼'
                          )}
                        </div>
                        <div>
                          <div className="tbl-name">
                            {(it.brand || '—') + ' · ' + (it.model || 'Без названия')}
                          </div>
                          <div className="tbl-sub">
                            {it.city || <span style={{ fontStyle: 'italic' }}>город не указан</span>}
                            {it.size_eu?.[0] ? ` · ${it.size_eu[0]} EU` : ''}
                          </div>
                        </div>
                      </div>
                    </td>
                    <td>
                      <span className={it.platform === 'vk' ? 'source-vk' : 'source-tg'}>
                        {platformLabel(it.platform)} {it.source_title || ''}
                      </span>
                    </td>
                    <td style={{ fontWeight: 600 }}>
                      {it.price !== null && it.price !== undefined ? (
                        formatPrice(it.price)
                      ) : (
                        <span style={{ color: 'var(--c-muted)', fontStyle: 'italic', fontSize: 12 }}>
                          —
                        </span>
                      )}
                    </td>
                    <td>
                      {it.is_hidden ? (
                        <span className="badge badge-hidden">Скрыто</span>
                      ) : (
                        <span className="badge badge-active">Активно</span>
                      )}
                    </td>
                    <td style={{ color: 'var(--c-muted)', fontSize: 12 }}>
                      {formatRelative(it.created_at)}
                    </td>
                    <td>
                      <div className="tbl-actions">
                        <button
                          type="button"
                          className="tbl-btn"
                          onClick={() => navigate(`/listings/${it.id}`)}
                        >
                          👁 Смотреть
                        </button>
                        <button
                          type="button"
                          className="tbl-btn edit"
                          onClick={() => onEditText(it)}
                        >
                          ✏️ Текст
                        </button>
                        <button
                          type="button"
                          className="tbl-btn edit"
                          onClick={() => onManagePhotos(it)}
                        >
                          🖼 Фото
                        </button>
                        {it.is_hidden ? (
                          <button
                            type="button"
                            className="tbl-btn restore"
                            onClick={() => onVisibility(it)}
                          >
                            ↩ Восстановить
                          </button>
                        ) : (
                          <button
                            type="button"
                            className="tbl-btn danger"
                            onClick={() => onVisibility(it)}
                          >
                            🚫 Скрыть
                          </button>
                        )}
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}

        {!loading && items.length > 0 && (
          <Pagination total={total} limit={LIMIT} offset={offset} onChange={setOffset} />
        )}
      </div>

      {editTextFor && (
        <EditTextModal
          listingId={editTextFor.id}
          initialText={editTextFor.text}
          title={editTextFor.title}
          onClose={() => setEditTextFor(null)}
          onSaved={() => {
            setEditTextFor(null);
            toast.success('Текст сохранён');
          }}
        />
      )}
      {photosFor && (
        <PhotosModal
          listingId={photosFor.id}
          initialPhotos={photosFor.photos}
          title={photosFor.title}
          onClose={() => {
            setPhotosFor(null);
            load();
          }}
        />
      )}
    </>
  );
}
