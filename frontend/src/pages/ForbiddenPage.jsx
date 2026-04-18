import { useNavigate } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';

export default function ForbiddenPage() {
  const navigate = useNavigate();
  const { user } = useAuth();
  return (
    <div className="error-page">
      <div className="error-glow danger" />
      <div className="error-icon">🔒</div>
      <div className="error-num danger">403</div>
      <div className="error-title">Доступ запрещён</div>
      <div className="error-desc">
        У вас нет прав для просмотра этой страницы. Обратитесь к администратору или вернитесь на
        главную.
      </div>
      <div className="error-actions">
        {!user ? (
          <button className="error-btn-primary" onClick={() => navigate('/auth')}>
            Войти в аккаунт
          </button>
        ) : null}
        <button className="error-btn-ghost" onClick={() => navigate('/')}>
          На главную
        </button>
      </div>
    </div>
  );
}
