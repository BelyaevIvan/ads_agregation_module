import { useState, useEffect } from 'react';

const SIZE_RUS = ['40', '41', '42', '43', '44', '45'];
const SIZE_EU = ['40', '41', '42', '43', '44', '45'];
const SIZE_US = ['7', '7.5', '8', '8.5', '9', '9.5', '10'];

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

export default function FiltersPanel({ value, onApply, isOpen, onToggleOpen }) {
  const [local, setLocal] = useState(value);

  useEffect(() => {
    setLocal(value);
  }, [value]);

  const patch = (p) => setLocal((v) => ({ ...v, ...p }));

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

          <div
            style={{
              fontSize: 11,
              color: 'var(--c-muted)',
              marginBottom: 8,
              textTransform: 'uppercase',
              letterSpacing: 0.5,
            }}
          >
            RUS
          </div>
          <div className="filter-options">
            {SIZE_RUS.map((s) => (
              <Checkbox
                key={`r${s}`}
                checked={local.size_rus.includes(s)}
                onChange={() => patch({ size_rus: toggleInArray(local.size_rus, s) })}
                label={s}
              />
            ))}
          </div>

          <div
            style={{
              fontSize: 11,
              color: 'var(--c-muted)',
              margin: '10px 0 8px',
              textTransform: 'uppercase',
              letterSpacing: 0.5,
            }}
          >
            EU
          </div>
          <div className="filter-options">
            {SIZE_EU.map((s) => (
              <Checkbox
                key={`e${s}`}
                checked={local.size_eu.includes(s)}
                onChange={() => patch({ size_eu: toggleInArray(local.size_eu, s) })}
                label={s}
              />
            ))}
          </div>

          <div
            style={{
              fontSize: 11,
              color: 'var(--c-muted)',
              margin: '10px 0 8px',
              textTransform: 'uppercase',
              letterSpacing: 0.5,
            }}
          >
            US
          </div>
          <div className="filter-options">
            {SIZE_US.map((s) => (
              <Checkbox
                key={`u${s}`}
                checked={local.size_us.includes(s)}
                onChange={() => patch({ size_us: toggleInArray(local.size_us, s) })}
                label={s}
              />
            ))}
          </div>

          <NullCheckbox
            checked={local.include_no_size}
            onChange={(c) => patch({ include_no_size: c })}
            label="+ показывать без указания размера"
          />
        </div>

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
