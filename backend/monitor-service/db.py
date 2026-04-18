import logging

import psycopg2
import psycopg2.extras

from config import DB_DSN

logger = logging.getLogger(__name__)


def get_connection():
    return psycopg2.connect(DB_DSN)


def upsert_source(conn, platform: str, external_id: str, title: str | None = None) -> str:
    """Создаёт источник если не существует, возвращает его UUID."""
    with conn.cursor() as cur:
        cur.execute(
            """
            INSERT INTO sources (platform, external_id, title)
            VALUES (%s, %s, %s)
            ON CONFLICT (platform, external_id)
            DO UPDATE SET title = COALESCE(EXCLUDED.title, sources.title)
            RETURNING id
            """,
            (platform, str(external_id), title),
        )
        return str(cur.fetchone()[0])


def save_listing(
    conn,
    source_id: str,
    original_text: str,
    post_url: str,
    posted_at,
    item: dict,
) -> str | None:
    """Сохраняет объявление и возвращает его UUID."""
    with conn.cursor() as cur:
        cur.execute(
            """
            INSERT INTO listings (
                source_id, original_text, post_url, posted_at,
                brand, model, category, color, price, city, condition,
                size_rus, size_us, size_eu, llm_raw
            ) VALUES (
                %s, %s, %s, %s,
                %s, %s, %s, %s, %s, %s, %s,
                %s, %s, %s, %s
            )
            RETURNING id
            """,
            (
                source_id,
                original_text,
                post_url,
                posted_at,
                item.get("brand"),
                item.get("model"),
                item.get("category"),
                item.get("color"),
                item.get("price"),
                item.get("city"),
                item.get("condition"),
                item.get("size_rus"),
                item.get("size_us"),
                item.get("size_eu"),
                psycopg2.extras.Json(item),
            ),
        )
        row = cur.fetchone()
        return str(row[0]) if row else None


def save_photo(
    conn,
    listing_id: str,
    photo_url: str,
    sort_order: int,
    is_cover: bool = False,
):
    with conn.cursor() as cur:
        cur.execute(
            """
            INSERT INTO listing_photos (listing_id, photo_url, sort_order, is_cover)
            VALUES (%s, %s, %s, %s)
            ON CONFLICT (listing_id, sort_order) DO NOTHING
            """,
            (listing_id, photo_url, sort_order, is_cover),
        )


def get_active_sources(conn, platform: str) -> set[str]:
    """Возвращает множество external_id активных источников для платформы."""
    with conn.cursor() as cur:
        cur.execute(
            "SELECT external_id FROM sources WHERE platform = %s AND is_active = TRUE",
            (platform,),
        )
        return {row[0] for row in cur.fetchall()}


def get_active_sources_full(conn, platform: str) -> list[tuple[str, str | None]]:
    """Возвращает [(external_id, title), ...] для всех активных источников платформы.
    Используется мониторами, чтобы получать актуальный список наблюдаемых групп
    напрямую из БД — включая добавленные через админский UI, а не только из .env."""
    with conn.cursor() as cur:
        cur.execute(
            "SELECT external_id, title FROM sources "
            "WHERE platform = %s AND is_active = TRUE",
            (platform,),
        )
        return list(cur.fetchall())


def save_log(
    conn,
    source_id: str | None,
    message_id: str | None,
    status: str,
    error_msg: str | None = None,
):
    with conn.cursor() as cur:
        cur.execute(
            """
            INSERT INTO monitoring_log (source_id, message_id, status, error_msg)
            VALUES (%s, %s, %s, %s)
            """,
            (source_id, message_id, status, error_msg),
        )
