import asyncio
import logging
import time
from collections import defaultdict
from datetime import timezone

from telethon import TelegramClient, events

import db
import storage
from llm_parser import parse_with_llm
from config import (
    TG_API_ID, TG_API_HASH, TG_SESSION_PATH,
    TG_PROXY_HOST, TG_PROXY_PORT, TG_PROXY_USER, TG_PROXY_PASS,
    TG_WATCHED_CHANNELS,
)

logging.basicConfig(
    format="%(asctime)s [%(name)s] %(levelname)s: %(message)s",
    level=logging.INFO,
)
logger = logging.getLogger(__name__)

# Буфер для сборки альбомов
album_buffer = defaultdict(list)
album_tasks = {}
ALBUM_WAIT = 1.5

# ── Прокси ────────────────────────────────────────────────────────────────
proxy = None
if TG_PROXY_HOST:
    proxy = {
        "proxy_type": "socks5",
        "addr": TG_PROXY_HOST,
        "port": TG_PROXY_PORT,
        "username": TG_PROXY_USER,
        "password": TG_PROXY_PASS,
        "rdns": True,
    }

# ── Клиент ────────────────────────────────────────────────────────────────
client = TelegramClient(
    TG_SESSION_PATH,
    TG_API_ID,
    TG_API_HASH,
    proxy=proxy,
    timeout=15,
    connection_retries=10,
    retry_delay=3,
    auto_reconnect=True,
)


# ── Вспомогательные функции ───────────────────────────────────────────────

def get_message_link(chat, msg):
    username = getattr(chat, "username", None)
    if username:
        return f"https://t.me/{username}/{msg.id}"
    else:
        chat_id = chat.id
        if chat_id < 0:
            chat_id = str(chat_id).replace("-100", "")
        return f"https://t.me/c/{chat_id}/{msg.id}"


def _sync_save(source_id, original_text, post_url, posted_at, items, photos):
    """Синхронное сохранение в БД + MinIO. Вызывается через asyncio.to_thread."""
    conn = db.get_connection()
    try:
        for item in items:
            listing_id = db.save_listing(conn, source_id, original_text, post_url, posted_at, item)
            if not listing_id:
                continue
            for idx, photo_bytes in enumerate(photos):
                try:
                    url = storage.upload_photo(photo_bytes, listing_id, idx)
                    db.save_photo(conn, listing_id, url, idx, is_cover=(idx == 0))
                except Exception as e:
                    logger.error("Ошибка загрузки фото для listing %s: %s", listing_id, e)
        db.save_log(conn, source_id, post_url.split("/")[-1], "parsed")
        conn.commit()
    except Exception:
        conn.rollback()
        raise
    finally:
        conn.close()


def _sync_upsert_source(platform, external_id, title):
    conn = db.get_connection()
    try:
        source_id = db.upsert_source(conn, platform, external_id, title)
        conn.commit()
        return source_id
    finally:
        conn.close()


def _sync_log(source_id, message_id, status, error_msg=None):
    conn = db.get_connection()
    try:
        db.save_log(conn, source_id, message_id, status, error_msg)
        conn.commit()
    finally:
        conn.close()


# ── Обработка поста ───────────────────────────────────────────────────────

async def process_post(chat, messages, photos):
    first = messages[0]
    text = next((m.text for m in messages if m.text), "")
    link = get_message_link(chat, first)
    title = getattr(chat, "title", None) or getattr(chat, "first_name", "Unknown")

    source_id = await asyncio.to_thread(_sync_upsert_source, "telegram", str(first.chat_id), title)

    items = await asyncio.to_thread(parse_with_llm, text)

    if not items:
        await asyncio.to_thread(_sync_log, source_id, str(first.id), "skipped")
        logger.info("🤖 Пропущен: %s", link)
        return

    posted_at = first.date
    if first.date and first.date.tzinfo is None:
        posted_at = first.date.replace(tzinfo=timezone.utc)

    try:
        await asyncio.to_thread(_sync_save, source_id, text, link, posted_at, items, photos)
        logger.info("✅ Сохранено %d товар(ов): %s", len(items), link)
    except Exception as e:
        await asyncio.to_thread(_sync_log, source_id, str(first.id), "error", str(e))
        logger.error("❌ Ошибка сохранения %s: %s", link, e)


async def flush_album(grouped_id):
    """Ждёт сборки всех частей альбома, затем обрабатывает."""
    await asyncio.sleep(ALBUM_WAIT)
    messages = album_buffer.pop(grouped_id, [])
    album_tasks.pop(grouped_id, None)

    if not messages:
        return

    first = messages[0]
    chat = await first.get_chat()
    title = getattr(chat, "title", None) or getattr(chat, "first_name", "Unknown")
    link = get_message_link(chat, first)
    text = next((m.text for m in messages if m.text), "")

    logger.info("=" * 40)
    logger.info("📷 Альбом (%d медиа)", len(messages))
    logger.info("Канал/чат: %s | ID: %s", title, first.chat_id)
    logger.info("🔗 Ссылка: %s", link)
    logger.info("Текст: %s", text)

    photos = []
    for i, msg in enumerate(messages):
        if msg.photo or (msg.document and getattr(msg.document, "mime_type", "").startswith("image/")):
            try:
                photo_bytes = await client.download_media(msg, bytes)
                photos.append(photo_bytes)
                logger.info("  📷 Фото %d: %d байт", i + 1, len(photo_bytes))
            except Exception as e:
                logger.error("Ошибка скачивания фото: %s", e)

    await process_post(chat, messages, photos)


@client.on(events.NewMessage)
async def handler(event):
    if event.chat_id not in TG_WATCHED_CHANNELS:
        return

    grouped_id = event.message.grouped_id

    if grouped_id:
        album_buffer[grouped_id].append(event.message)
        if grouped_id not in album_tasks:
            album_tasks[grouped_id] = asyncio.create_task(flush_album(grouped_id))
        return

    # Одиночное сообщение
    chat = await event.get_chat()
    title = getattr(chat, "title", None) or getattr(chat, "first_name", "Unknown")
    link = get_message_link(chat, event.message)

    logger.info("=" * 40)
    logger.info("🟢 Новый пост пойман")
    logger.info("Канал/чат: %s | ID: %s", title, event.chat_id)
    logger.info("🔗 Ссылка: %s", link)
    logger.info("📅 Дата: %s UTC", event.message.date)
    logger.info("Текст: %s", event.text)

    photos = []
    if event.message.photo:
        try:
            photo_bytes = await client.download_media(event.message, bytes)
            photos.append(photo_bytes)
            logger.info("📷 Фото: %d байт", len(photo_bytes))
        except Exception as e:
            logger.error("Ошибка скачивания фото: %s", e)

    await process_post(chat, [event.message], photos)


# ── Точка входа ───────────────────────────────────────────────────────────

def wait_for_db(retries=12, delay=5):
    for i in range(retries):
        try:
            conn = db.get_connection()
            conn.close()
            return
        except Exception as e:
            logger.warning("БД не готова (%d/%d): %s", i + 1, retries, e)
            time.sleep(delay)
    raise RuntimeError("Не удалось подключиться к БД")


async def main():
    wait_for_db()
    await client.start()
    logger.info("🚀 Слушаем каналы...")
    logger.info("Отслеживаемые ID: %s", TG_WATCHED_CHANNELS)
    await client.run_until_disconnected()


with client:
    client.loop.run_until_complete(main())
