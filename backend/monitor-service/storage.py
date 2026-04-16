import io
import logging

from minio import Minio
from minio.error import S3Error

from config import MINIO_ENDPOINT, MINIO_ACCESS_KEY, MINIO_SECRET_KEY, MINIO_BUCKET, MINIO_SECURE

logger = logging.getLogger(__name__)

_client: Minio | None = None


def get_client() -> Minio:
    global _client
    if _client is None:
        _client = Minio(
            MINIO_ENDPOINT,
            access_key=MINIO_ACCESS_KEY,
            secret_key=MINIO_SECRET_KEY,
            secure=MINIO_SECURE,
        )
        if not _client.bucket_exists(MINIO_BUCKET):
            _client.make_bucket(MINIO_BUCKET)
            logger.info("Создан бакет MinIO: %s", MINIO_BUCKET)
    return _client


def upload_photo(photo_bytes: bytes, listing_id: str, index: int) -> str:
    """Загружает фото в MinIO, возвращает URL."""
    client = get_client()
    object_name = f"{listing_id}/{index}.jpg"
    try:
        client.put_object(
            MINIO_BUCKET,
            object_name,
            io.BytesIO(photo_bytes),
            length=len(photo_bytes),
            content_type="image/jpeg",
        )
        scheme = "https" if MINIO_SECURE else "http"
        return f"{scheme}://{MINIO_ENDPOINT}/{MINIO_BUCKET}/{object_name}"
    except S3Error as e:
        logger.error("Ошибка загрузки в MinIO: %s", e)
        raise
