# database.md — Схема базы данных BrandHunt

СУБД: **PostgreSQL 15+**

---

## Обзор таблиц

| Таблица | Назначение |
|---|---|
| `users` | Пользователи системы |
| `sources` | Источники мониторинга (TG-каналы, VK-группы) |
| `listings` | Объявления о продаже товаров |
| `listing_photos` | Фотографии объявлений |
| `favorites` | Избранные объявления пользователей |
| `monitoring_log` | Технический лог обработки сообщений |

---

## Схема связей

```
sources ──────────< listings >──────────< listing_photos
                       │
                       ├──────< favorites >────── users
                       │
                       └── monitoring_log
```

| Связь | Тип | ON DELETE |
|---|---|---|
| `sources` → `listings` | 1 : N | SET NULL (объявления остаются) |
| `listings` → `listing_photos` | 1 : N | CASCADE (фото удаляются) |
| `listings` → `favorites` | 1 : N | CASCADE |
| `users` → `favorites` | 1 : N | CASCADE |
| `sources` → `monitoring_log` | 1 : N | SET NULL |

---

## Таблицы

### `users` — пользователи

```sql
CREATE TABLE users (
    id            UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    email         VARCHAR(255) NOT NULL UNIQUE,
    password_hash TEXT         NOT NULL,
    full_name     VARCHAR(255),
    phone         VARCHAR(30),
    tg_link       VARCHAR(255),
    vk_link       VARCHAR(255),
    role          VARCHAR(10)  NOT NULL DEFAULT 'user', -- 'user' | 'admin'
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);
```

| Колонка | Тип | Nullable | Описание |
|---|---|---|---|
| `id` | UUID | NO | PK, генерируется БД |
| `email` | VARCHAR(255) | NO | Уникальный логин |
| `password_hash` | TEXT | NO | bcrypt-хэш пароля |
| `full_name` | VARCHAR(255) | YES | ФИО |
| `phone` | VARCHAR(30) | YES | Номер телефона |
| `tg_link` | VARCHAR(255) | YES | Ссылка/юзернейм Telegram |
| `vk_link` | VARCHAR(255) | YES | Ссылка ВКонтакте |
| `role` | VARCHAR(10) | NO | `'user'` или `'admin'` |
| `created_at` | TIMESTAMPTZ | NO | Дата регистрации (UTC) |

**Решения:** UUID вместо SERIAL — нет утечки количества пользователей. `role` — строка, не ENUM, чтобы не делать миграцию при добавлении роли.

---

### `sources` — источники мониторинга

```sql
CREATE TABLE sources (
    id          UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    platform    VARCHAR(10)  NOT NULL,       -- 'telegram' | 'vk'
    external_id VARCHAR(255) NOT NULL,       -- chat_id для TG, group_id для VK
    title       VARCHAR(255),                -- название канала/группы
    is_active   BOOLEAN      NOT NULL DEFAULT TRUE,
    added_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),

    UNIQUE (platform, external_id)
);
```

| Колонка | Тип | Nullable | Описание |
|---|---|---|---|
| `id` | UUID | NO | PK |
| `platform` | VARCHAR(10) | NO | `'telegram'` или `'vk'` |
| `external_id` | VARCHAR(255) | NO | ID канала во внешней системе |
| `title` | VARCHAR(255) | YES | Человекочитаемое название |
| `is_active` | BOOLEAN | NO | Включён ли мониторинг |
| `added_at` | TIMESTAMPTZ | NO | Дата добавления |

**Решения:** Составной `UNIQUE (platform, external_id)` — один источник нельзя добавить дважды. `is_active` управляется администратором через UI без изменения кода.

---

### `listings` — объявления

```sql
CREATE TABLE listings (
    id            UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    source_id     UUID         REFERENCES sources(id) ON DELETE SET NULL,

    -- оригинальные данные из поста
    original_text TEXT,
    post_url      TEXT         NOT NULL,
    posted_at     TIMESTAMPTZ,

    -- атрибуты, извлечённые LLM
    brand         VARCHAR(100),
    model         VARCHAR(255),
    category      VARCHAR(100),  -- 'sneakers' | 'clothing' | 'electronics' | 'accessories'
    color         VARCHAR(100),
    price         NUMERIC(12,2),
    city          VARCHAR(100),
    condition     VARCHAR(50),   -- 'new' | 'used'

    -- размеры в трёх системах (все допускают NULL)
    size_rus      TEXT[],        -- {"42","43","44"} / {"S","M"}
    size_us       TEXT[],        -- {"8.5","9","9.5"} / {"S","M"}
    size_eu       TEXT[],        -- {"40","41","42"}

    -- служебные
    is_hidden     BOOLEAN        NOT NULL DEFAULT FALSE,
    llm_raw       JSONB,         -- сырой JSON-ответ LLM (для отладки и переразбора)
    created_at    TIMESTAMPTZ    NOT NULL DEFAULT NOW()
);
```

| Колонка | Тип | Nullable | Описание |
|---|---|---|---|
| `id` | UUID | NO | PK |
| `source_id` | UUID | YES | FK → `sources.id` |
| `original_text` | TEXT | YES | Исходный текст поста |
| `post_url` | TEXT | NO | Ссылка на оригинальный пост в TG/VK |
| `posted_at` | TIMESTAMPTZ | YES | Дата публикации поста |
| `brand` | VARCHAR(100) | YES | Бренд (Nike, Stone Island...) |
| `model` | VARCHAR(255) | YES | Модель (Air Max 90...) |
| `category` | VARCHAR(100) | YES | Категория товара |
| `color` | VARCHAR(100) | YES | Цвет |
| `price` | NUMERIC(12,2) | YES | Цена в рублях |
| `city` | VARCHAR(100) | YES | Город продавца |
| `condition` | VARCHAR(50) | YES | Состояние: `'new'` / `'used'` |
| `size_rus` | TEXT[] | YES | Размеры в российской системе |
| `size_us` | TEXT[] | YES | Размеры в американской системе |
| `size_eu` | TEXT[] | YES | Размеры в европейской системе |
| `is_hidden` | BOOLEAN | NO | Скрыто администратором |
| `llm_raw` | JSONB | YES | Сырой ответ LLM для отладки |
| `created_at` | TIMESTAMPTZ | NO | Дата создания записи |

**Решения:**
- `size_rus/us/eu` — массивы `TEXT[]`, все три допускают NULL. Массив нужен потому что в одном объявлении может быть несколько размеров. Пример хранимого значения: `{"42","43","44"}`. LLM извлекает только те системы, что упомянуты в тексте
- `NUMERIC(12,2)` для цены — никогда не FLOAT (погрешности при арифметике)
- `llm_raw JSONB` — позволяет переразобрать объявления при изменении логики парсинга без повторного вызова API

---

### `listing_photos` — фотографии объявлений

```sql
CREATE TABLE listing_photos (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    listing_id  UUID        NOT NULL REFERENCES listings(id) ON DELETE CASCADE,
    photo_url   TEXT        NOT NULL,
    sort_order  SMALLINT    NOT NULL DEFAULT 0, -- порядок в галерее (0 = первая)
    is_cover    BOOLEAN     NOT NULL DEFAULT FALSE,
    uploaded_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    UNIQUE (listing_id, sort_order)
);
```

| Колонка | Тип | Nullable | Описание |
|---|---|---|---|
| `id` | UUID | NO | PK |
| `listing_id` | UUID | NO | FK → `listings.id` |
| `photo_url` | TEXT | NO | URL в MinIO |
| `sort_order` | SMALLINT | NO | Порядок в галерее, начиная с 0 |
| `is_cover` | BOOLEAN | NO | Признак обложки (титульное фото) |
| `uploaded_at` | TIMESTAMPTZ | NO | Дата загрузки |

**Решения:**
- `sort_order` — позволяет менять порядок фотографий без удаления и повторной загрузки
- `UNIQUE (listing_id, sort_order)` — у одного объявления не может быть двух фото с одинаковой позицией
- `ON DELETE CASCADE` — при удалении объявления фото удаляются из БД автоматически. Удаление файлов из MinIO — ответственность сервисного слоя

---

### `favorites` — избранное

```sql
CREATE TABLE favorites (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    listing_id UUID        NOT NULL REFERENCES listings(id) ON DELETE CASCADE,
    saved_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    UNIQUE (user_id, listing_id)
);
```

| Колонка | Тип | Nullable | Описание |
|---|---|---|---|
| `id` | UUID | NO | PK |
| `user_id` | UUID | NO | FK → `users.id` |
| `listing_id` | UUID | NO | FK → `listings.id` |
| `saved_at` | TIMESTAMPTZ | NO | Дата добавления в избранное |

**Решения:**
- `UNIQUE (user_id, listing_id)` — нельзя добавить одно объявление дважды; эндпоинт «добавить в избранное» идемпотентен
- `ON DELETE CASCADE` с обеих сторон — удалился пользователь или объявление, строка в `favorites` удаляется автоматически

---

### `monitoring_log` — лог обработки сообщений

```sql
CREATE TABLE monitoring_log (
    id           BIGSERIAL    PRIMARY KEY,
    source_id    UUID         REFERENCES sources(id) ON DELETE SET NULL,
    message_id   VARCHAR(255),
    status       VARCHAR(20)  NOT NULL, -- 'parsed' | 'skipped' | 'error'
    error_msg    TEXT,
    processed_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);
```

| Колонка | Тип | Nullable | Описание |
|---|---|---|---|
| `id` | BIGSERIAL | NO | PK (автоинкремент) |
| `source_id` | UUID | YES | FK → `sources.id` |
| `message_id` | VARCHAR(255) | YES | ID сообщения во внешней системе |
| `status` | VARCHAR(20) | NO | `'parsed'` / `'skipped'` / `'error'` |
| `error_msg` | TEXT | YES | Текст ошибки |
| `processed_at` | TIMESTAMPTZ | NO | Время обработки |

**Значения `status`:**
- `'parsed'` — сообщение распознано как объявление, атрибуты извлечены, запись создана в `listings`
- `'skipped'` — сообщение не является объявлением (по оценке LLM), запись в `listings` не создавалась
- `'error'` — произошла ошибка при обращении к LLM API или при сохранении

**Решения:** `BIGSERIAL` вместо UUID — технический лог, строк будет очень много, скорость вставки важнее. Используется в админке для статистики и мониторинга здоровья воркера.
