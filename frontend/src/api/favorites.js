import { api } from './client';

export const favoritesApi = {
  list: (params) => api.get('/users/me/favorites', { params, auth: true }),
  add: (listing_id) => api.post('/users/me/favorites', { listing_id }, { auth: true }),
  remove: (listing_id) => api.delete(`/users/me/favorites/${listing_id}`, { auth: true }),
};
