-- BrandHunt — начальная схема БД
-- PostgreSQL 15+

-- ---------------------------------------------------------------------------
-- users
-- ---------------------------------------------------------------------------
CREATE TABLE users (
    id            UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    email         VARCHAR(255) NOT NULL UNIQUE,
    password_hash TEXT         NOT NULL,
    full_name     VARCHAR(255),
    phone         VARCHAR(30),
    tg_link       VARCHAR(255),
    vk_link       VARCHAR(255),
    role          VARCHAR(10)  NOT NULL DEFAULT 'user',
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- ---------------------------------------------------------------------------
-- sources
-- ---------------------------------------------------------------------------
CREATE TABLE sources (
    id          UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    platform    VARCHAR(10)  NOT NULL,
    external_id VARCHAR(255) NOT NULL,
    title       VARCHAR(255),
    is_active   BOOLEAN      NOT NULL DEFAULT TRUE,
    added_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),

    UNIQUE (platform, external_id)
);

-- ---------------------------------------------------------------------------
-- listings
-- ---------------------------------------------------------------------------
CREATE TABLE listings (
    id            UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    source_id     UUID         REFERENCES sources(id) ON DELETE SET NULL,

    original_text TEXT,
    post_url      TEXT         NOT NULL,
    posted_at     TIMESTAMPTZ,

    brand         VARCHAR(100),
    model         VARCHAR(255),
    category      VARCHAR(100),
    color         VARCHAR(100),
    price         NUMERIC(12,2),
    city          VARCHAR(100),
    condition     VARCHAR(50),

    size_rus      TEXT[],
    size_us       TEXT[],
    size_eu       TEXT[],

    is_hidden     BOOLEAN      NOT NULL DEFAULT FALSE,
    llm_raw       JSONB,
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- ---------------------------------------------------------------------------
-- listing_photos
-- ---------------------------------------------------------------------------
CREATE TABLE listing_photos (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    listing_id  UUID        NOT NULL REFERENCES listings(id) ON DELETE CASCADE,
    photo_url   TEXT        NOT NULL,
    sort_order  SMALLINT    NOT NULL DEFAULT 0,
    is_cover    BOOLEAN     NOT NULL DEFAULT FALSE,
    uploaded_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    UNIQUE (listing_id, sort_order)
);

-- ---------------------------------------------------------------------------
-- favorites
-- ---------------------------------------------------------------------------
CREATE TABLE favorites (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    listing_id UUID        NOT NULL REFERENCES listings(id) ON DELETE CASCADE,
    saved_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    UNIQUE (user_id, listing_id)
);

-- ---------------------------------------------------------------------------
-- monitoring_log
-- ---------------------------------------------------------------------------
CREATE TABLE monitoring_log (
    id           BIGSERIAL    PRIMARY KEY,
    source_id    UUID         REFERENCES sources(id) ON DELETE SET NULL,
    message_id   VARCHAR(255),
    status       VARCHAR(20)  NOT NULL,
    error_msg    TEXT,
    processed_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- ---------------------------------------------------------------------------
-- Индексы
-- ---------------------------------------------------------------------------

-- listings: самые частые фильтры в поиске
CREATE INDEX idx_listings_brand      ON listings (brand);
CREATE INDEX idx_listings_category   ON listings (category);
CREATE INDEX idx_listings_city       ON listings (city);
CREATE INDEX idx_listings_price      ON listings (price);
CREATE INDEX idx_listings_condition  ON listings (condition);
CREATE INDEX idx_listings_posted_at  ON listings (posted_at DESC);
CREATE INDEX idx_listings_source_id  ON listings (source_id);
CREATE INDEX idx_listings_is_hidden  ON listings (is_hidden);

-- listing_photos: быстрый поиск фото по объявлению
CREATE INDEX idx_photos_listing_id   ON listing_photos (listing_id);

-- favorites: быстрый поиск избранного пользователя
CREATE INDEX idx_favorites_user_id   ON favorites (user_id);

-- monitoring_log: фильтрация по источнику и статусу в админке
CREATE INDEX idx_monlog_source_id    ON monitoring_log (source_id);
CREATE INDEX idx_monlog_status       ON monitoring_log (status);
CREATE INDEX idx_monlog_processed_at ON monitoring_log (processed_at DESC);
