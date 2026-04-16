import json
import os

# PostgreSQL
DB_DSN = (
    f"host={os.getenv('POSTGRES_HOST', 'localhost')} "
    f"port={os.getenv('POSTGRES_PORT', '5432')} "
    f"dbname={os.getenv('POSTGRES_DB')} "
    f"user={os.getenv('POSTGRES_USER')} "
    f"password={os.getenv('POSTGRES_PASSWORD')}"
)

# MinIO (дефолт — localhost для локального запуска; Docker переопределяет через environment)
MINIO_ENDPOINT = os.getenv("MINIO_ENDPOINT", "localhost:9000")
MINIO_ACCESS_KEY = os.getenv("MINIO_ROOT_USER")
MINIO_SECRET_KEY = os.getenv("MINIO_ROOT_PASSWORD")
MINIO_BUCKET = os.getenv("MINIO_BUCKET", "brandhunt-photos")
MINIO_SECURE = os.getenv("MINIO_SECURE", "false").lower() == "true"

# Ollama (дефолт — localhost для локального запуска; Docker переопределяет через environment)
OLLAMA_HOST = os.getenv("OLLAMA_HOST", "http://localhost:11434")
OLLAMA_MODEL = os.getenv("OLLAMA_MODEL", "gemma3:4b")

# Telegram
TG_API_ID = int(os.getenv("TG_API_ID", "0"))
TG_API_HASH = os.getenv("TG_API_HASH", "")
TG_SESSION_PATH = os.getenv("TG_SESSION_PATH", "telethon_session")

# Telegram прокси (опционально — если TG_PROXY_HOST пустой, прокси не используется)
TG_PROXY_HOST = os.getenv("TG_PROXY_HOST", "")
TG_PROXY_PORT = int(os.getenv("TG_PROXY_PORT", "0"))
TG_PROXY_USER = os.getenv("TG_PROXY_USER", "")
TG_PROXY_PASS = os.getenv("TG_PROXY_PASS", "")

# Telegram каналы для мониторинга (через запятую, например: -100123456,-100789012)
TG_WATCHED_CHANNELS: list[int] = [
    int(x.strip())
    for x in os.getenv("TG_WATCHED_CHANNELS", "").split(",")
    if x.strip()
]

# VK
VK_ACCESS_TOKEN = os.getenv("VK_ACCESS_TOKEN", "")
VK_API_VERSION = os.getenv("VK_API_VERSION", "5.131")
VK_CHECK_INTERVAL = int(os.getenv("VK_CHECK_INTERVAL", "5"))
VK_STATE_FILE = os.getenv("VK_STATE_FILE", "vk_state.json")

# VK группы JSON-массивом: '[{"id": -225463035, "name": "ГрузИк"}]'
VK_GROUPS: list[dict] = json.loads(os.getenv("VK_GROUPS", "[]"))
