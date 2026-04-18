import { api } from './client';

export const listingsApi = {
  search: (params) => api.get('/listings', { params }),
  // silentError: страница сама показывает 404-экран, тост был бы шумом
  getById: (id) => api.get(`/listings/${id}`, { silentError: true }),
};
