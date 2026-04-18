import { useState } from 'react';
import { adminApi } from '../../api/admin';

export default function AddSourceModal({ onClose, onCreated }) {
  const [platform, setPlatform] = useState('telegram');
  const [externalId, setExternalId] = useState('');
  const [title, setTitle] = useState('');
  const [err, setErr] = useState('');
  const [saving, setSaving] = useState(false);

  const save = async (e) => {
    e.preventDefault();
    setErr('');
    if (!externalId.trim()) {
      setErr('Укажите external_id канала или группы');
      return;
    }
    setSaving(true);
    try {
      await adminApi.addSource({
        platform,
        external_id: externalId.trim(),
        title: title.trim() || undefined,
      });
      onCreated?.();
    } catch (e2) {
      if (e2.status === 409) setErr('Такой источник уже добавлен');
      else if (e2.data?.message) setErr(e2.data.message);
      else setErr('Не удалось добавить источник');
    } finally {
      setSaving(false);
    }
  };

  return (
    <div className="modal-overlay" onClick={(e) => e.target === e.currentTarget && onClose()}>
      <div className="modal">
        <button type="button" className="modal-close" onClick={onClose}>
          ×
        </button>
        <div className="modal-title">Новый источник</div>
        <div className="modal-sub">Добавьте Telegram-канал или VK-группу для мониторинга</div>

        <form onSubmit={save} className="add-source-form">
          <div className="form-group" style={{ margin: 0 }}>
            <label className="form-label">Платформа</label>
            <div style={{ display: 'flex', gap: 8 }}>
              <button
                type="button"
                className={`filter-chip ${platform === 'telegram' ? 'active' : ''}`}
                onClick={() => setPlatform('telegram')}
              >
                Telegram
              </button>
              <button
                type="button"
                className={`filter-chip ${platform === 'vk' ? 'active' : ''}`}
                onClick={() => setPlatform('vk')}
              >
                ВКонтакте
              </button>
            </div>
          </div>

          <div className="form-group" style={{ margin: 0 }}>
            <label className="form-label">External ID</label>
            <input
              className="form-input"
              type="text"
              placeholder={platform === 'telegram' ? '-1001234567890' : '-12345678'}
              value={externalId}
              onChange={(e) => setExternalId(e.target.value)}
              required
            />
          </div>

          <div className="form-group" style={{ margin: 0 }}>
            <label className="form-label">Название (опционально)</label>
            <input
              className="form-input"
              type="text"
              placeholder={platform === 'telegram' ? '@channel_name' : 'Название группы'}
              value={title}
              onChange={(e) => setTitle(e.target.value)}
            />
          </div>

          {err && <div className="field-error">{err}</div>}

          <div className="modal-actions">
            <button type="button" className="modal-btn-cancel" onClick={onClose}>
              Отмена
            </button>
            <button type="submit" className="modal-btn-save" disabled={saving}>
              {saving ? 'Сохранение…' : 'Добавить'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
