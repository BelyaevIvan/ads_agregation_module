import { useState } from 'react';
import { adminApi } from '../../api/admin';
import { useToast } from '../../contexts/ToastContext';

export default function PhotosModal({ listingId, initialPhotos, title, onClose }) {
  const toast = useToast();
  const [photos, setPhotos] = useState(initialPhotos || []);
  const [pending, setPending] = useState(null);

  const deletePhoto = async (photo) => {
    if (pending) return;
    setPending(photo.id);
    try {
      const res = await adminApi.deletePhoto(listingId, photo.id);
      setPhotos((list) => {
        const next = list.filter((p) => p.id !== photo.id);
        if (res?.new_cover_id) {
          return next.map((p) => (p.id === res.new_cover_id ? { ...p, is_cover: true } : p));
        }
        return next;
      });
      toast.success('Фото удалено');
    } catch {
      // toast fired
    } finally {
      setPending(null);
    }
  };

  return (
    <div className="modal-overlay" onClick={(e) => e.target === e.currentTarget && onClose()}>
      <div className="modal">
        <button type="button" className="modal-close" onClick={onClose}>
          ×
        </button>
        <div className="modal-title">Управление фотографиями</div>
        {title && <div className="modal-sub">{title}</div>}

        {photos.length === 0 ? (
          <div
            style={{
              textAlign: 'center',
              padding: 28,
              color: 'var(--c-muted)',
              fontSize: 13,
            }}
          >
            Нет фотографий — будет показана заглушка 🖼
          </div>
        ) : (
          <div className="photos-grid">
            {photos.map((p) => (
              <div key={p.id} className={`photo-item ${p.is_cover ? 'is-cover' : ''}`}>
                <img src={p.url} alt="" />
                {p.is_cover && <div className="photo-cover-badge">Обложка</div>}
                <button
                  type="button"
                  className="photo-del-btn"
                  onClick={() => deletePhoto(p)}
                  disabled={pending === p.id}
                  aria-label="Удалить"
                >
                  ×
                </button>
              </div>
            ))}
          </div>
        )}

        <div className="photo-hint">
          Обложка — фото с синей рамкой. При удалении обложки следующее фото становится обложкой
          автоматически. При удалении всех фото отображается заглушка.
        </div>
        <div className="modal-actions">
          <button type="button" className="modal-btn-cancel" onClick={onClose}>
            Закрыть
          </button>
        </div>
      </div>
    </div>
  );
}
