import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';
import { useToast } from '../contexts/ToastContext';
import { authApi } from '../api/auth';
import { favoritesApi } from '../api/favorites';
import ListingCard from '../components/listings/ListingCard';
import EmptyState from '../components/common/EmptyState';
import Spinner from '../components/common/Spinner';
import { initialsFrom } from '../utils/format';

export default function ProfilePage() {
  const { user, updateUser, logout } = useAuth();
  const toast = useToast();
  const navigate = useNavigate();

  const [tab, setTab] = useState('data');
  const [form, setForm] = useState({
    full_name: user?.full_name || '',
    phone: user?.phone || '',
    tg_link: user?.tg_link || '',
    vk_link: user?.vk_link || '',
  });
  const [saving, setSaving] = useState(false);

  const [favs, setFavs] = useState([]);
  const [favsLoading, setFavsLoading] = useState(false);
  const [favsTotal, setFavsTotal] = useState(0);

  useEffect(() => {
    if (user) {
      setForm({
        full_name: user.full_name || '',
        phone: user.phone || '',
        tg_link: user.tg_link || '',
        vk_link: user.vk_link || '',
      });
    }
  }, [user]);

  const loadFavs = async () => {
    setFavsLoading(true);
    try {
      const res = await favoritesApi.list({ limit: 100 });
      setFavs(res.items || []);
      setFavsTotal(res.total || 0);
    } catch {
      setFavs([]);
    } finally {
      setFavsLoading(false);
    }
  };

  useEffect(() => {
    if (tab === 'favs') loadFavs();
  }, [tab]);

  const save = async (e) => {
    e.preventDefault();
    setSaving(true);
    try {
      const updated = await authApi.updateMe(form);
      updateUser(updated);
      toast.success('Изменения сохранены');
    } catch {
      // toast fired by client
    } finally {
      setSaving(false);
    }
  };

  const handleLogout = async () => {
    await logout();
    navigate('/');
  };

  const handleRemoveFav = async (listingId) => {
    setFavs((list) => list.filter((f) => f.listing.id !== listingId));
    setFavsTotal((t) => Math.max(0, t - 1));
  };

  if (!user) return null;

  return (
    <div className="profile-layout">
      <aside className="profile-sidebar">
        <div className="profile-card">
          <div className="avatar">{initialsFrom(user.full_name, user.email)}</div>
          <div className="profile-name">{user.full_name || 'Пользователь'}</div>
          <div className="profile-email">{user.email}</div>
          <div className="profile-links">
            {user.tg_link && (
              <div className="profile-link-item">
                <span>✈️</span> {user.tg_link}
              </div>
            )}
            {user.vk_link && (
              <div className="profile-link-item">
                <span>📘</span> {user.vk_link}
              </div>
            )}
            {user.phone && (
              <div className="profile-link-item">
                <span>📞</span> {user.phone}
              </div>
            )}
          </div>
        </div>
        <div className="sidebar-menu">
          <button
            type="button"
            className={`menu-item ${tab === 'data' ? 'active' : ''}`}
            onClick={() => setTab('data')}
          >
            👤 Мои данные
          </button>
          <button
            type="button"
            className={`menu-item ${tab === 'favs' ? 'active' : ''}`}
            onClick={() => setTab('favs')}
          >
            ♥ Избранное
            {favsTotal > 0 && <span className="menu-badge">{favsTotal}</span>}
          </button>
          {user.role === 'admin' && (
            <button type="button" className="menu-item" onClick={() => navigate('/admin')}>
              ⚙️ Админка
            </button>
          )}
          <button type="button" className="menu-item danger" onClick={handleLogout}>
            🚪 Выйти
          </button>
        </div>
      </aside>

      <main className="profile-main">
        {tab === 'data' ? (
          <div>
            <div className="section-heading">Личные данные</div>
            <form className="profile-form" onSubmit={save}>
              <div className="form-group">
                <label className="form-label">Полное имя</label>
                <input
                  className="form-input"
                  type="text"
                  value={form.full_name}
                  onChange={(e) => setForm({ ...form, full_name: e.target.value })}
                  placeholder="Иван Иванов"
                />
              </div>
              <div className="form-group">
                <label className="form-label">Email</label>
                <input className="form-input" type="email" value={user.email} disabled />
              </div>
              <div className="form-row">
                <div className="form-group">
                  <label className="form-label">Телефон</label>
                  <input
                    className="form-input"
                    type="text"
                    value={form.phone}
                    onChange={(e) => setForm({ ...form, phone: e.target.value })}
                    placeholder="+7 999 123-45-67"
                  />
                </div>
                <div className="form-group">
                  <label className="form-label">Telegram</label>
                  <input
                    className="form-input"
                    type="text"
                    value={form.tg_link}
                    onChange={(e) => setForm({ ...form, tg_link: e.target.value })}
                    placeholder="@username"
                  />
                </div>
              </div>
              <div className="form-group">
                <label className="form-label">ВКонтакте</label>
                <input
                  className="form-input"
                  type="text"
                  value={form.vk_link}
                  onChange={(e) => setForm({ ...form, vk_link: e.target.value })}
                  placeholder="vk.com/username"
                />
              </div>
              <button type="submit" className="save-btn" disabled={saving}>
                {saving ? 'Сохранение…' : 'Сохранить изменения'}
              </button>
            </form>
          </div>
        ) : (
          <div>
            <div className="fav-header">
              <div className="section-heading" style={{ marginBottom: 0 }}>
                Избранное
              </div>
              <div className="fav-count">
                {favsTotal} {favsTotal === 1 ? 'объявление' : 'объявлений'}
              </div>
            </div>
            {favsLoading ? (
              <Spinner />
            ) : favs.length === 0 ? (
              <EmptyState
                icon="♡"
                title="Пока пусто"
                subtitle="Сохраняйте понравившиеся объявления — они появятся здесь"
              />
            ) : (
              <div className="fav-grid">
                {favs.map((f) => (
                  <ListingCard
                    key={f.id}
                    listing={f.listing}
                    isFavorite
                    onToggleFavorite={(id, nowFav) => {
                      if (!nowFav) handleRemoveFav(id);
                    }}
                  />
                ))}
              </div>
            )}
          </div>
        )}
      </main>
    </div>
  );
}
