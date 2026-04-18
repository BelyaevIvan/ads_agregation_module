import { useNavigate } from 'react-router-dom';
import { useState } from 'react';
import { formatPrice, platformLabel } from '../../utils/format';
import { useAuth } from '../../contexts/AuthContext';
import { favoritesApi } from '../../api/favorites';

export default function ListingCard({ listing, isFavorite = false, onToggleFavorite }) {
  const navigate = useNavigate();
  const { user } = useAuth();
  const [fav, setFav] = useState(isFavorite);
  const [pending, setPending] = useState(false);

  const hasPrice = listing.price !== null && listing.price !== undefined;
  const hasCity = listing.city && listing.city.trim();
  const sizeLabel = listing.size_rus?.[0] || listing.size_eu?.[0] || listing.size_us?.[0] || null;

  async function handleToggleFav(e) {
    e.stopPropagation();
    if (!user) {
      navigate('/auth');
      return;
    }
    if (pending) return;
    setPending(true);
    try {
      if (fav) {
        await favoritesApi.remove(listing.id);
        setFav(false);
        onToggleFavorite?.(listing.id, false);
      } else {
        await favoritesApi.add(listing.id);
        setFav(true);
        onToggleFavorite?.(listing.id, true);
      }
    } catch {
      // toast already fired by client
    } finally {
      setPending(false);
    }
  }

  return (
    <div className="card" onClick={() => navigate(`/listings/${listing.id}`)}>
      <div className="card-source-badge">{platformLabel(listing.platform)}</div>
      {listing.cover_photo_url ? (
        <img className="card-img" src={listing.cover_photo_url} alt={listing.model || 'товар'} loading="lazy" />
      ) : (
        <div className="card-img-placeholder">🖼 нет фото</div>
      )}
      <div className="card-body">
        <div className="card-brand">{listing.brand || '—'}</div>
        <div className="card-name">{listing.model || 'Без названия'}</div>
        <div className="card-meta">
          {sizeLabel ? (
            <span className="card-tag-sm">Р. {sizeLabel}</span>
          ) : (
            <span className="card-tag-null">размер не указан</span>
          )}
          {hasCity ? (
            <span className="card-tag-sm">{listing.city}</span>
          ) : (
            <span className="card-tag-null">город не указан</span>
          )}
        </div>
        <div className="card-footer">
          {hasPrice ? (
            <div className="card-price">{formatPrice(listing.price)}</div>
          ) : (
            <div className="card-price no-price">цена не указана</div>
          )}
          <button
            type="button"
            className={`card-fav ${fav ? 'active' : ''}`}
            onClick={handleToggleFav}
            disabled={pending}
            aria-label={fav ? 'Удалить из избранного' : 'В избранное'}
          >
            {fav ? '♥' : '♡'}
          </button>
        </div>
      </div>
    </div>
  );
}
