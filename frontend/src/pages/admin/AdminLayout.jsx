import { NavLink, Outlet, useNavigate } from 'react-router-dom';
import { useAuth } from '../../contexts/AuthContext';
import { initialsFrom } from '../../utils/format';

export default function AdminLayout() {
  const { user } = useAuth();
  const navigate = useNavigate();

  return (
    <div className="admin-layout">
      <aside className="admin-sidebar">
        <div className="admin-logo">
          Brand<span>Hunt</span> Admin
        </div>
        <NavLink
          to="/admin"
          end
          className={({ isActive }) => `admin-nav-item ${isActive ? 'active' : ''}`}
        >
          <span>📊</span> Дашборд
        </NavLink>
        <NavLink
          to="/admin/listings"
          className={({ isActive }) => `admin-nav-item ${isActive ? 'active' : ''}`}
        >
          <span>📋</span> Объявления
        </NavLink>
        <NavLink
          to="/admin/sources"
          className={({ isActive }) => `admin-nav-item ${isActive ? 'active' : ''}`}
        >
          <span>📡</span> Источники
        </NavLink>
        <button
          type="button"
          className="admin-nav-item"
          onClick={() => navigate('/')}
          style={{ cursor: 'pointer' }}
        >
          <span>🏠</span> На сайт
        </button>
      </aside>
      <main className="admin-main">
        <Outlet context={{ user, avatar: initialsFrom(user?.full_name, user?.email) }} />
      </main>
    </div>
  );
}
