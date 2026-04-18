import { useState } from 'react';
import { adminApi } from '../../api/admin';

export default function EditTextModal({ listingId, initialText, title, onClose, onSaved }) {
  const [text, setText] = useState(initialText || '');
  const [saving, setSaving] = useState(false);

  const save = async () => {
    if (!text.trim()) return;
    setSaving(true);
    try {
      await adminApi.updateText(listingId, text);
      onSaved?.();
    } catch {
      // toast fired
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
        <div className="modal-title">Редактировать текст объявления</div>
        {title && <div className="modal-sub">{title}</div>}
        <textarea
          className="modal-textarea"
          value={text}
          onChange={(e) => setText(e.target.value)}
          placeholder="Текст объявления…"
        />
        <div className="modal-actions">
          <button type="button" className="modal-btn-cancel" onClick={onClose}>
            Отмена
          </button>
          <button
            type="button"
            className="modal-btn-save"
            onClick={save}
            disabled={saving || !text.trim()}
          >
            {saving ? 'Сохранение…' : 'Сохранить'}
          </button>
        </div>
      </div>
    </div>
  );
}
