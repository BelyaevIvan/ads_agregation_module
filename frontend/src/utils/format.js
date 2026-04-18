export function formatPrice(price) {
  if (price === null || price === undefined) return null;
  const n = Number(price);
  if (!Number.isFinite(n)) return null;
  return n.toLocaleString('ru-RU', { maximumFractionDigits: 0 }) + ' ₽';
}

export function formatDate(iso) {
  if (!iso) return '';
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return '';
  return d.toLocaleDateString('ru-RU', { day: '2-digit', month: '2-digit', year: 'numeric' });
}

export function formatDateShort(iso) {
  if (!iso) return '';
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return '';
  return d.toLocaleDateString('ru-RU', { day: '2-digit', month: '2-digit' });
}

export function formatRelative(iso) {
  if (!iso) return '';
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return '';
  const diff = Date.now() - d.getTime();
  const mins = Math.floor(diff / 60000);
  if (mins < 1) return 'только что';
  if (mins < 60) return `${mins} мин назад`;
  const hours = Math.floor(mins / 60);
  if (hours < 24) return `${hours} ч назад`;
  const days = Math.floor(hours / 24);
  if (days < 7) return `${days} дн назад`;
  return formatDate(iso);
}

export function platformLabel(platform) {
  if (platform === 'telegram') return 'TG';
  if (platform === 'vk') return 'VK';
  return platform || '';
}

export function conditionLabel(condition) {
  if (condition === 'new') return 'Новое';
  if (condition === 'used') return 'Б/у';
  return null;
}

export function initialsFrom(fullName, email) {
  if (fullName && fullName.trim()) {
    const parts = fullName.trim().split(/\s+/).slice(0, 2);
    return parts.map((p) => p[0].toUpperCase()).join('');
  }
  if (email) return email[0].toUpperCase();
  return 'U';
}

export function formatNumber(n) {
  if (n === null || n === undefined) return '0';
  return Number(n).toLocaleString('ru-RU');
}
