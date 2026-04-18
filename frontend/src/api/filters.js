import { api } from './client';

export const filtersApi = {
  // Возвращает { size_rus, size_eu, size_us } — списки доступных размеров
  // из активных объявлений. На бэке кэшируется на 5 минут.
  getSizes: () => api.get('/filters/sizes'),
};
