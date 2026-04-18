import { useEffect, useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { listingsApi } from '../api/listings';
import { adminApi } from '../api/admin';
import { favoritesApi } from '../api/favorites';
import { useAuth } from '../contexts/AuthContext';
import Spinner from '../components/common/Spinner';
import {
  formatPrice,
  formatRelative,
  conditionLabel,
  platformLabel,
} from '../utils/format';

function AttrItem({ label, value }) {
  const empty = value === null || value === undefined || value === '' || (Array.isArray(value) && value.length === 0);
  return (
    <div className={`attr-item${empty ? ' no-value' : ''}`}>
      <div className="attr-key">{label}</div>
      {empty ? (
        <div className="attr-val null-val">не указан</div>
      ) : (
        <div className="attr-val">{Array.isArray(value) ? value.join(', ') : value}</div>
      )}
    </div>
  );
}

export default function ListingDetailPage() {
  const { id } = useParams();
  const navigate = useNavigate();
  const { user, isAdmin } = useAuth();

  const [listing, setListing] = useState(null);
  const [loading, setLoading] = useState(true);
  const [notFound, setNotFound] = useState(false);
  const [activePhoto, setActivePhoto] = useState(0);
  const [isFav, setIsFav] = useState(false);
  const [favPending, setFavPending] = useState(false);

  useEffect(() => {
    let cancelled = false;
    setLoading(true);
    setNotFound(false);
    (async () => {
      try {
        let data;
        try {
          data = await listingsApi.getById(id);
        } catch (e) {
          // Скрытое объявление недоступно публично, но админ должен его видеть
          if (e.status === 404 && isAdmin) {
            data = await adminApi.getListing(id, { silentError: true });
          } else {
            throw e;
          }
        }
        if (!cancelled) {
          setListing(data);
          const cover = (data.photos || []).findIndex((p) => p.is_cover);
          setActivePhoto(cover >= 0 ? cover : 0);
        }
      } catch (e) {
        if (!cancelled) {
          if (e.status === 404) setNotFound(true);
        }
      } finally {
        if (!cancelled) setLoading(false);
      }
    })();
    return () => {
      cancelled = true;
    };
  }, [id, isAdmin]);

  useEffect(() => {
    if (!user || !listing) return;
    let cancelled = false;
    (async () => {
      try {
        const res = await favoritesApi.list({ limit: 100 });
        if (!cancelled) {
          setIsFav((res.items || []).some((f) => f.listing.id === listing.id));
        }
      } catch {
        // silent
      }
    })();
    return () => {
      cancelled = true;
    };
  }, [user, listing]);

  const toggleFav = async () => {
    if (!user) {
      navigate('/auth');
      return;
    }
    if (favPending) return;
    setFavPending(true);
    try {
      if (isFav) {
        await favoritesApi.remove(listing.id);
        setIsFav(false);
      } else {
        await favoritesApi.add(listing.id);
        setIsFav(true);
      }
    } catch {
      // toast fired by client
    } finally {
      setFavPending(false);
    }
  };

  if (loading) return <Spinner />;
  if (notFound || !listing) {
    return (
      <div className="error-page">
        <div className="error-glow" />
        <div className="error-num">404</div>
        <div className="error-title">Объявление не найдено</div>
        <div className="error-desc">Возможно, объявление скрыто администратором или удалено.</div>
        <div className="error-actions">
          <button className="error-btn-primary" onClick={() => navigate('/')}>
            На главную
          </button>
          <button className="error-btn-ghost" onClick={() => navigate('/search')}>
            Смотреть объявления
          </button>
        </div>
      </div>
    );
  }

  const photos = listing.photos || [];
  const mainPhoto = photos[activePhoto];
  const hasPrice = listing.price !== null && listing.price !== undefined;

  return (
    <div className="product-layout">
      <div className="product-gallery">
        <div className="main-photo">
          {mainPhoto ? <img src={mainPhoto.url} alt={listing.model || 'фото'} /> : '🖼'}
        </div>
        {photos.length > 1 && (
          <div className="thumb-row">
            {photos.map((p, i) => (
              <div
                key={p.url}
                className={`thumb ${i === activePhoto ? 'active' : ''}`}
                onClick={() => setActivePhoto(i)}
              >
                <img src={p.url} alt="" />
              </div>
            ))}
          </div>
        )}
      </div>
      <div className="product-info">
        <div className="product-source">
          <div className="source-chip">
            {platformLabel(listing.source?.platform)} · {listing.source?.title || '—'}
          </div>
          <span style={{ fontSize: 12, color: 'var(--c-muted)' }}>
            {formatRelative(listing.posted_at || listing.created_at)}
          </span>
        </div>
        <div className="product-brand">{listing.brand || '—'}</div>
        <div className="product-title">{listing.model || 'Без названия'}</div>
        <div className="product-price-row">
          {hasPrice ? (
            <div className="product-price">{formatPrice(listing.price)}</div>
          ) : (
            <div className="product-price no-price">цена не указана</div>
          )}
          {conditionLabel(listing.condition) && (
            <div className="product-condition">{conditionLabel(listing.condition)}</div>
          )}
        </div>

        <div className="attrs-grid">
          <AttrItem label="Размер RUS" value={listing.size_rus} />
          <AttrItem label="Размер EU" value={listing.size_eu} />
          <AttrItem label="Размер US" value={listing.size_us} />
          <AttrItem label="Цвет" value={listing.color} />
          <AttrItem label="Город" value={listing.city} />
          <AttrItem label="Категория" value={listing.category} />
        </div>

        {listing.original_text && (
          <div className="original-text-block">
            <div className="ot-label">Оригинальный текст объявления</div>
            <div className="ot-text">{listing.original_text}</div>
          </div>
        )}

        <div className="product-actions">
          {listing.post_url ? (
            <a
              className="btn-tg"
              href={listing.post_url}
              target="_blank"
              rel="noopener noreferrer"
            >
              Перейти к посту →
            </a>
          ) : (
            <button className="btn-tg" disabled>
              Ссылка на пост недоступна
            </button>
          )}
          <button
            className={`btn-fav2 ${isFav ? 'active' : ''}`}
            onClick={toggleFav}
            disabled={favPending}
          >
            {isFav ? '♥' : '♡'}
          </button>
        </div>
      </div>
    </div>
  );
}
