import { useNavigate } from 'react-router-dom';

const TAGS = ['Nike', 'Jordan', 'New Balance', 'iPhone', 'Stone Island'];

export default function NotFoundPage() {
  const navigate = useNavigate();
  const go = (q) => navigate(`/search?q=${encodeURIComponent(q)}`);
  return (
    <div className="error-page">
      <div className="error-glow" />
      <div className="error-num">404</div>
      <div className="error-title">Страница не найдена</div>
      <div className="error-desc">
        Похоже, это объявление уже продано или ссылка устарела. Попробуйте поискать что-то другое.
      </div>
      <div className="error-actions">
        <button className="error-btn-primary" onClick={() => navigate('/')}>
          ← На главную
        </button>
        <button className="error-btn-ghost" onClick={() => navigate('/search')}>
          Смотреть объявления
        </button>
      </div>
      <div style={{ marginTop: 40, position: 'relative', zIndex: 1 }}>
        <div style={{ fontSize: 12, color: 'var(--c-muted)', marginBottom: 12 }}>
          Возможно, тебя заинтересует
        </div>
        <div style={{ display: 'flex', gap: 8, flexWrap: 'wrap', justifyContent: 'center' }}>
          {TAGS.map((t) => (
            <span key={t} className="tag" onClick={() => go(t)}>
              {t}
            </span>
          ))}
        </div>
      </div>
    </div>
  );
}
