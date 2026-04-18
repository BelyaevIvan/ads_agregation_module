import { Link, useNavigate } from 'react-router-dom';
import { useAuth } from '../../contexts/AuthContext';
import { initialsFrom } from '../../utils/format';

export default function Navbar() {
  const { user, isAdmin } = useAuth();
  const navigate = useNavigate();

  return (
    <nav className="app-nav">
      <div className="logo" onClick={() => navigate('/')}>
        BrandHunt
      </div>
      <div className="nav-links">
        {user ? (
          <>
            {isAdmin && (
              <button className="nav-btn ghost" onClick={() => navigate('/admin')}>
                Админка
              </button>
            )}
            <button className="nav-btn ghost" onClick={() => navigate('/profile')}>
              Кабинет
            </button>
            <div
              className="admin-avatar"
              style={{ marginLeft: 4 }}
              title={user.email}
              onClick={() => navigate('/profile')}
            >
              {initialsFrom(user.full_name, user.email)}
            </div>
          </>
        ) : (
          <>
            <Link to="/auth" className="nav-btn ghost">
              Войти
            </Link>
            <Link to="/auth?mode=register" className="nav-btn primary">
              Регистрация
            </Link>
          </>
        )}
      </div>
    </nav>
  );
}
