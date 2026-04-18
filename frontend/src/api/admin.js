import { api } from './client';

export const adminApi = {
  listings: (params) => api.get('/admin/listings', { params, auth: true }),
  getListing: (id, opts = {}) => api.get(`/admin/listings/${id}`, { auth: true, ...opts }),
  setVisibility: (id, is_hidden) =>
    api.patch(`/admin/listings/${id}/visibility`, { is_hidden }, { auth: true }),
  updateText: (id, original_text) =>
    api.patch(`/admin/listings/${id}/text`, { original_text }, { auth: true }),
  deletePhoto: (listingId, photoId) =>
    api.delete(`/admin/listings/${listingId}/photos/${photoId}`, { auth: true }),

  stats: () => api.get('/admin/stats', { auth: true }),
  listingsByDay: (days = 30) =>
    api.get('/admin/stats/listings-by-day', { params: { days }, auth: true }),
  topBrands: (limit = 5) => api.get('/admin/stats/top-brands', { params: { limit }, auth: true }),
  topCities: (limit = 5) => api.get('/admin/stats/top-cities', { params: { limit }, auth: true }),

  sources: (params) => api.get('/admin/sources', { params, auth: true }),
  addSource: (data) => api.post('/admin/sources', data, { auth: true }),
  toggleSource: (id, is_active) =>
    api.patch(`/admin/sources/${id}/toggle`, { is_active }, { auth: true }),
};
