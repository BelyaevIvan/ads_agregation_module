"""
Скрипт для авторизации Telegram — создаёт файл сессии.

Запуск локально:
    python auth_tg.py

Запуск в Docker (первый раз):
    docker compose run --rm --no-deps tg-monitor python auth_tg.py
"""
import asyncio
from telethon import TelegramClient
from config import TG_API_ID, TG_API_HASH, TG_SESSION_PATH, TG_PROXY_HOST, TG_PROXY_PORT, TG_PROXY_USER, TG_PROXY_PASS

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


async def main():
    async with TelegramClient(TG_SESSION_PATH, TG_API_ID, TG_API_HASH, proxy=proxy, timeout=15, connection_retries=10, retry_delay=3, auto_reconnect=True) as client:
        await client.start()
        print(f"\n✅ Авторизация успешна!")
        print(f"Сессия сохранена: {TG_SESSION_PATH}.session\n")


asyncio.run(main())
