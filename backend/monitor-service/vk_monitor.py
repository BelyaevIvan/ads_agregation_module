import json
import logging
import os
import time
from datetime import datetime, timezone

import requests

import db
import storage
from llm_parser import parse_with_llm
from config import VK_ACCESS_TOKEN, VK_API_VERSION, VK_CHECK_INTERVAL, VK_GROUPS, VK_STATE_FILE

logging.basicConfig(
    format="%(asctime)s [%(name)s] %(levelname)s: %(message)s",
    level=logging.INFO,
)
logger = logging.getLogger(__name__)


# ── VK API ────────────────────────────────────────────────────────────────

def get_latest_posts(group_id):
    url = "https://api.vk.com/method/wall.get"
    params = {
        "owner_id": group_id,
        "count": 5,
        "access_token": VK_ACCESS_TOKEN,
        "v": VK_API_VERSION,
    }
    response = requests.get(url, params=params).json()
    if "error" in response:
        raise RuntimeError(response["error"])
    return response["response"]["items"]


def get_post_link(owner_id, post_id):
    return f"https://vk.com/wall{owner_id}_{post_id}"


def download_photos(post):
    """Скачивает все фото из поста, возвращает список bytes."""
    photos = []
    attachments = post.get("attachments", [])
    for att in attachments:
        if att["type"] != "photo":
            continue
        sizes = att["photo"].get("sizes", [])
        if not sizes:
            continue
        best = max(sizes, key=lambda s: s["width"] * s["height"])
        url = best["url"]
        try:
            resp = requests.get(url, timeout=15)
            resp.raise_for_status()
            photos.append(resp.content)
            logger.info("  📷 Фото: %d байт (%dx%d)", len(resp.content), best["width"], best["height"])
        except Exception as e:
            logger.error("  ⚠️ Не удалось скачать фото: %s", e)
    return photos


# ── Состояние ─────────────────────────────────────────────────────────────

def load_state():
    if not os.path.exists(VK_STATE_FILE):
        return {}
    with open(VK_STATE_FILE, "r", encoding="utf-8") as f:
        return json.load(f)


def save_state(state):
    state_dir = os.path.dirname(VK_STATE_FILE)
    if state_dir:
        os.makedirs(state_dir, exist_ok=True)
    with open(VK_STATE_FILE, "w", encoding="utf-8") as f:
        json.dump(state, f, ensure_ascii=False, indent=2)


# ── Обработка поста ───────────────────────────────────────────────────────

def process_post(conn, group_name, owner_id, post):
    post_url = get_post_link(owner_id, post["id"])
    text = post.get("text", "").strip()
    posted_at = datetime.fromtimestamp(post["date"], tz=timezone.utc)
    message_id = str(post["id"])

    logger.info("=" * 70)
    logger.info("🔔 Новый пост из: %s", group_name)
    logger.info("ID поста: %s", post["id"])
    logger.info("🔗 Ссылка: %s", post_url)
    logger.info("Текст:\n%s", text or "[без текста]")

    attachments = post.get("attachments", [])
    if attachments:
        logger.info("Вложения: %d", len(attachments))
        for a in attachments:
            logger.info(" - %s", a["type"])

    # Скачиваем фото
    photos = download_photos(post)

    # LLM-парсинг
    source_id = db.upsert_source(conn, "vk", str(owner_id), group_name)
    items = parse_with_llm(text)

    if not items:
        db.save_log(conn, source_id, message_id, "skipped")
        logger.info("🤖 LLM: не объявление")
        return

    logger.info("🤖 LLM: найдено товаров: %d", len(items))

    for item in items:
        listing_id = db.save_listing(conn, source_id, text, post_url, posted_at, item)
        if not listing_id:
            continue
        for idx, photo_bytes in enumerate(photos):
            try:
                url = storage.upload_photo(photo_bytes, listing_id, idx)
                db.save_photo(conn, listing_id, url, idx, is_cover=(idx == 0))
            except Exception as e:
                logger.error("Ошибка загрузки фото: %s", e)

    db.save_log(conn, source_id, message_id, "parsed")
    logger.info("✅ Сохранено %d товар(ов) из %s", len(items), post_url)


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


print("🔍 Мониторинг нескольких групп VK запущен\n")

wait_for_db()
last_posts = load_state()

while True:
    try:
        for group_id, group_name in {g["id"]: g.get("name", str(g["id"])) for g in VK_GROUPS}.items():
            posts = get_latest_posts(group_id)

            group_key = str(group_id)
            last_id = last_posts.get(group_key)

            # Первая инициализация группы
            if last_id is None:
                for post in posts:
                    if post.get("is_pinned"):
                        continue
                    last_posts[group_key] = post["id"]
                    print(f"ℹ️ Инициализация [{group_name}]: {post['id']}")
                    break
                continue

            new_posts = []
            for post in posts:
                if post.get("is_pinned"):
                    continue
                if post["id"] > last_id:
                    new_posts.append(post)

            if new_posts:
                new_posts.sort(key=lambda p: p["id"])

                conn = db.get_connection()
                try:
                    for post in new_posts:
                        try:
                            process_post(conn, group_name, group_id, post)
                            last_posts[group_key] = post["id"]
                            conn.commit()
                        except Exception as e:
                            conn.rollback()
                            logger.error("Ошибка обработки поста %s: %s", post["id"], e)
                            try:
                                db.save_log(conn, None, str(post["id"]), "error", str(e))
                                conn.commit()
                            except Exception:
                                conn.rollback()
                finally:
                    conn.close()
            else:
                print(f"🔴 [{group_name}] новых постов нет")

        save_state(last_posts)

    except Exception as e:
        print(f"❌ Ошибка: {e}")

    time.sleep(VK_CHECK_INTERVAL)
