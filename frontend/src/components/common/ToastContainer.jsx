import { useToast } from '../../contexts/ToastContext';

export default function ToastContainer() {
  const { toasts, remove } = useToast();
  if (toasts.length === 0) return null;
  return (
    <div className="toast-container">
      {toasts.map((t) => (
        <div key={t.id} className={`toast ${t.type === 'success' ? 'success' : ''}`}>
          <div className="toast-icon">{t.type === 'success' ? '✓' : '⚠'}</div>
          <div className="toast-content">
            <div className="toast-title">{t.title}</div>
            {t.sub && <div className="toast-sub">{t.sub}</div>}
          </div>
          <button className="toast-close" onClick={() => remove(t.id)} aria-label="Закрыть">
            ×
          </button>
        </div>
      ))}
    </div>
  );
}
