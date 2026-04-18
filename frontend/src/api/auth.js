import { api } from './client';

export const authApi = {
  register: (email, password) =>
    api.post('/auth/register', { email, password }, { silentError: true }),

  login: (email, password) =>
    api.post('/auth/login', { email, password }, { silentError: true }),

  logout: () => api.post('/auth/logout', undefined, { auth: true, silentError: true }),

  me: () => api.get('/users/me', { auth: true, silentError: true }),

  updateMe: (data) => api.put('/users/me', data, { auth: true }),
};
