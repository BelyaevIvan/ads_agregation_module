import { useState, useEffect } from 'react';
import { useSizeFilters } from '../../contexts/FiltersContext';

function toggleInArray(arr, value) {
  return arr.includes(value) ? arr.filter((v) => v !== value) : [...arr, value];
}

function Checkbox({ checked, onChange, label }) {
  return (
    <label className="filter-option">
      <span
        className={`filter-check ${checked ? 'checked' : ''}`}
        onClick={(e) => {
          e.preventDefault();
          onChange(!checked);
        }}
        role="checkbox"
        aria-checked={checked}
      >
        {checked ? '✓' : ''}
      </span>
      <span className="filter-option-text" onClick={() => onChange(!checked)}>
        {label}
      </span>
    </label>
  );
}

function NullCheckbox({ checked, onChange, label }) {
  return (
    <label className="filter-null-option" onClick={(e) => e.preventDefault()}>
      <span
        className={`filter-null-check ${checked ? 'checked' : ''}`}
        onClick={() => onChange(!checked)}
        role="checkbox"
        aria-checked={checked}
      >
        {checked ? '✓' : ''}
      </span>
      <span className="filter-null-text" onClick={() => onChange(!checked)}>
        {label}
      </span>
    </label>
  );
}

const SIZE_SCALES = [
  { key: 'size_rus', label: 'RUS' },
  { key: 'size_eu', label: 'EU' },
  { key: 'size_us', label: 'US' },
];
const SIZES_PREVIEW = 4; // сколько размеров показывать в свёрнутом виде

export default function FiltersPanel({ value, onApply, isOpen, onToggleOpen }) {
  const { sizes } = useSizeFilters();
  const [local, setLocal] = useState(value);
  const [expanded, setExpanded] = useState({}); // { size_rus: true, ... }

  useEffect(() => {
    setLocal(value);
  }, [value]);

  const patch = (p) => setLocal((v) => ({ ...v, ...p }));
  const toggleExpanded = (key) =>
    setExpanded((e) => ({ ...e, [key]: !e[key] }));

  // Показываем секцию размеров, только если уже подтянули данные с бэка
  // и хотя бы в одной сетке есть хоть один размер. До этого — не рендерим.
  const hasAnySizes =
    sizes && (sizes.size_rus.length || sizes.size_eu.length || sizes.size_us.length);

  const apply = () => {
    onApply(local);
    if (onToggleOpen && window.innerWidth <= 768) onToggleOpen();
  };

  return (
    <>
      <button className="mobile-filter-btn" onClick={onToggleOpen} type="button">
        {isOpen ? '✕ Закрыть фильтры' : '⚙️ Фильтры и параметры поиска'}
      </button>
      <aside className={`filters-panel ${isOpen ? 'open' : ''}`}>
        <div className="filters-title">Фильтры</div>

        <div className="filter-group">
          <div className="filter-label">Платформа</div>
          <div className="filter-options">
            <Checkbox
              checked={local.platform.includes('telegram')}
              onChange={() => patch({ platform: toggleInArray(local.platform, 'telegram') })}
              label="Telegram"
            />
            <Checkbox
              checked={local.platform.includes('vk')}
              onChange={() => patch({ platform: toggleInArray(local.platform, 'vk') })}
              label="ВКонтакте"
            />
          </div>
        </div>

        <div className="filter-group">
          <div className="filter-label">Состояние</div>
          <div className="filter-options">
            <Checkbox
              checked={local.condition === 'new'}
              onChange={(c) => patch({ condition: c ? 'new' : '' })}
              label="Новое"
            />
            <Checkbox
              checked={local.condition === 'used'}
              onChange={(c) => patch({ condition: c ? 'used' : '' })}
              label="Б/у"
            />
          </div>
        </div>

        {hasAnySizes && (
          <div className="filter-group">
            <div className="filter-label-row">
              <span className="filter-label" style={{ marginBottom: 0 }}>
                Размер
              </span>
              <div className="tooltip-wrap">
                <div className="tooltip-icon">?</div>
                <div className="tooltip-box">
                  <b>Размеры не конвертируются автоматически.</b> Если укажете только RUS 43,
                  объявления с EU 43 или US 9.5 не попадут в выборку. Для максимального охвата
                  выбирайте размеры во <b>всех трёх сетках</b>.
                </div>
              </div>
            </div>

            {SIZE_SCALES.map(({ key, label }) => {
              const list = sizes[key];
              if (!list.length) return null;

              // Автоматически раскрываем сетку, если в ней есть выбранный размер,
              // который не попал бы в превью — иначе галочка была бы скрыта.
              const selected = local[key];
              const hasSelectedBeyondPreview = selected.some(
                (s) => list.indexOf(s) >= SIZES_PREVIEW
              );
              const isOpen = expanded[key] || hasSelectedBeyondPreview;
              const visible = isOpen ? list : list.slice(0, SIZES_PREVIEW);
              const canToggle = list.length > SIZES_PREVIEW && !hasSelectedBeyondPreview;

              return (
                <div key={key}>
                  <div
                    style={{
                      fontSize: 11,
                      color: 'var(--c-muted)',
                      margin: '0 0 8px',
                      textTransform: 'uppercase',
                      letterSpacing: 0.5,
                    }}
                  >
                    {label}
                  </div>
                  <div className="filter-options">
                    {visible.map((s) => (
                      <Checkbox
                        key={`${key}-${s}`}
                        checked={selected.includes(s)}
                        onChange={() => patch({ [key]: toggleInArray(selected, s) })}
                        label={s}
                      />
                    ))}
                  </div>
                  {canToggle ? (
                    <button
                      type="button"
                      className="size-toggle"
                      onClick={() => toggleExpanded(key)}
                    >
                      {isOpen
                        ? '← Свернуть'
                        : `Показать все (${list.length}) →`}
                    </button>
                  ) : (
                    <div style={{ height: 10 }} />
                  )}
                </div>
              );
            })}

            <NullCheckbox
              checked={local.include_no_size}
              onChange={(c) => patch({ include_no_size: c })}
              label="+ показывать без указания размера"
            />
          </div>
        )}

        <div className="filter-group">
          <div className="filter-label">Цена, ₽</div>
          <div className="price-inputs">
            <input
              className="price-input"
              type="number"
              placeholder="от"
              value={local.price_min}
              onChange={(e) => patch({ price_min: e.target.value })}
            />
            <span className="price-sep">—</span>
            <input
              className="price-input"
              type="number"
              placeholder="до"
              value={local.price_max}
              onChange={(e) => patch({ price_max: e.target.value })}
            />
          </div>
          <NullCheckbox
            checked={local.include_no_price}
            onChange={(c) => patch({ include_no_price: c })}
            label="+ показывать без указания цены"
          />
        </div>

        <div className="filter-group">
          <div className="filter-label">Город</div>
          <input
            className="price-input"
            type="text"
            placeholder="Например, Москва"
            value={local.cityInput}
            onChange={(e) => patch({ cityInput: e.target.value })}
          />
          <NullCheckbox
            checked={local.include_no_city}
            onChange={(c) => patch({ include_no_city: c })}
            label="+ показывать без указания города"
          />
        </div>

        <div className="filter-group">
          <div className="filter-label">Бренд</div>
          <input
            className="price-input"
            type="text"
            placeholder="Например, Nike"
            value={local.brandInput}
            onChange={(e) => patch({ brandInput: e.target.value })}
          />
        </div>

        <button className="apply-btn" onClick={apply} type="button">
          Применить фильтры
        </button>
      </aside>
    </>
  );
}

export function defaultFilters() {
  return {
    platform: [],
    condition: '',
    size_rus: [],
    size_eu: [],
    size_us: [],
    include_no_size: true,
    include_no_price: true,
    include_no_city: true,
    price_min: '',
    price_max: '',
    cityInput: '',
    brandInput: '',
  };
}

export function filtersToParams(filters) {
  const p = {};
  if (filters.platform.length) p.platform = filters.platform;
  if (filters.condition) p.condition = filters.condition;
  if (filters.size_rus.length) p.size_rus = filters.size_rus;
  if (filters.size_eu.length) p.size_eu = filters.size_eu;
  if (filters.size_us.length) p.size_us = filters.size_us;
  if (filters.price_min) p.price_min = filters.price_min;
  if (filters.price_max) p.price_max = filters.price_max;
  if (filters.cityInput.trim()) p.city = filters.cityInput.trim();
  if (filters.brandInput.trim()) p.brand = filters.brandInput.trim();
  p.include_no_size = filters.include_no_size;
  p.include_no_price = filters.include_no_price;
  p.include_no_city = filters.include_no_city;
  return p;
}
