import { useState, useEffect } from 'react';
import { useNavigate, useSearchParams, useLocation } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';
import { useToast } from '../contexts/ToastContext';

export default function AuthPage() {
  const { user, login, register } = useAuth();
  const navigate = useNavigate();
  const location = useLocation();
  const [sp] = useSearchParams();
  const toast = useToast();

  const [mode, setMode] = useState(sp.get('mode') === 'register' ? 'register' : 'login');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [passwordConfirm, setPasswordConfirm] = useState('');
  const [fieldErr, setFieldErr] = useState('');
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    if (user) {
      const from = location.state?.from;
      navigate(from || (user.role === 'admin' ? '/admin' : '/profile'), { replace: true });
    }
  }, [user, navigate, location.state]);

  const submit = async (e) => {
    e.preventDefault();
    setFieldErr('');

    if (!email.trim() || !password) {
      setFieldErr('Заполните все поля');
      return;
    }
    if (password.length < 8) {
      setFieldErr('Пароль должен быть не короче 8 символов');
      return;
    }
    if (mode === 'register' && password !== passwordConfirm) {
      setFieldErr('Пароли не совпадают');
      return;
    }

    setSubmitting(true);
    try {
      if (mode === 'login') {
        await login(email.trim(), password);
      } else {
        await register(email.trim(), password);
        toast.success('Регистрация прошла успешно');
      }
    } catch (err) {
      if (err.status === 401) setFieldErr('Неверный email или пароль');
      else if (err.status === 409) setFieldErr('Пользователь с таким email уже существует');
      else if (err.data?.message) setFieldErr(err.data.message);
      else setFieldErr('Не удалось выполнить запрос');
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className="auth-wrap">
      <div className="auth-glow" />
      <div className="auth-card">
        <div className="auth-tabs">
          <button
            type="button"
            className={`auth-tab ${mode === 'login' ? 'active' : ''}`}
            onClick={() => {
              setMode('login');
              setFieldErr('');
            }}
          >
            Вход
          </button>
          <button
            type="button"
            className={`auth-tab ${mode === 'register' ? 'active' : ''}`}
            onClick={() => {
              setMode('register');
              setFieldErr('');
            }}
          >
            Регистрация
          </button>
        </div>
        <form onSubmit={submit}>
          <div className="auth-subtitle">
            {mode === 'login' ? 'С возвращением 👋' : 'Создать аккаунт'}
          </div>
          <div className="auth-desc">
            {mode === 'login'
              ? 'Войдите, чтобы управлять избранным'
              : 'Сохраняйте понравившиеся объявления'}
          </div>
          <div className="form-group">
            <label className="form-label">Email</label>
            <input
              className="form-input"
              type="email"
              placeholder="your@email.com"
              autoComplete="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              required
            />
          </div>
          <div className="form-group">
            <label className="form-label">Пароль</label>
            <input
              className="form-input"
              type="password"
              placeholder="••••••••"
              autoComplete={mode === 'login' ? 'current-password' : 'new-password'}
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              required
            />
          </div>
          {mode === 'register' && (
            <div className="form-group">
              <label className="form-label">Подтверждение пароля</label>
              <input
                className="form-input"
                type="password"
                placeholder="••••••••"
                autoComplete="new-password"
                value={passwordConfirm}
                onChange={(e) => setPasswordConfirm(e.target.value)}
                required
              />
            </div>
          )}
          {fieldErr && <div className="field-error">{fieldErr}</div>}
          <button type="submit" className="submit-btn" disabled={submitting}>
            {submitting
              ? 'Подождите…'
              : mode === 'login'
              ? 'Войти'
              : 'Зарегистрироваться'}
          </button>
          <div className="auth-divider">
            {mode === 'login' ? (
              <>
                Нет аккаунта?{' '}
                <button type="button" className="auth-link" onClick={() => setMode('register')}>
                  Зарегистрироваться
                </button>
              </>
            ) : (
              <>
                Уже есть аккаунт?{' '}
                <button type="button" className="auth-link" onClick={() => setMode('login')}>
                  Войти
                </button>
              </>
            )}
          </div>
        </form>
      </div>
    </div>
  );
}
