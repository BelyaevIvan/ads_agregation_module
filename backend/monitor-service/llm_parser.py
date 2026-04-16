import json
import logging

from ollama import Client as OllamaClient

from config import OLLAMA_HOST, OLLAMA_MODEL

logger = logging.getLogger(__name__)

ollama = OllamaClient(host=OLLAMA_HOST)

SYSTEM_PROMPT = """Ты парсер объявлений о продаже товаров.

ПЕРВЫМ ДЕЛОМ определи: является ли текст объявлением о продаже товара?
Признаки объявления: указана цена, описан конкретный товар для продажи, упоминается размер/состояние/доставка.
НЕ объявления: новости, анонсы событий, спортивные трансляции, мемы, обсуждения и любые другие сообщения.

Если текст НЕ является объявлением о продаже — верни ТОЛЬКО:
[{"message": "Это не объявление"}]

Если текст ЯВЛЯЕТСЯ объявлением, для каждого товара извлеки:
- brand: бренд (Nike, Stone Island, Adidas и т.д.)
- model: модель (Air Max 90, Jordan 1 и т.д.)
- category: категория товара — ТОЛЬКО одно из: "Одежда", "Обувь", "Электроника", "Аксессуары" или null
- color: цвет
- price: цена в рублях (только число, без валюты)
- city: город продавца
- condition: состояние — ТОЛЬКО "new" или "used" или null
- size_rus: массив размеров в российской системе (например ["42", "43"]), или null
- size_us: массив размеров в американской системе (например ["9", "10"]), или null
- size_eu: массив размеров в европейской системе (например ["42", "43"]), или null

ПРАВИЛА:
- Если информация отсутствует в тексте — ставь null
- Размеры — это ВСЕГДА массив строк, даже если размер один: ["42"], не "42"
- НЕ конвертируй размеры между системами. Если указан только EU размер — заполни только size_eu, остальные null
- Если в посте НЕСКОЛЬКО товаров — верни массив объектов по каждому
- Если в посте ОДИН товар — всё равно верни массив с одним объектом
- Отвечай ТОЛЬКО валидным JSON-массивом, БЕЗ пояснений, БЕЗ markdown-разметки"""


def parse_with_llm(text: str) -> list[dict] | None:
    """Отправляет текст в LLM, возвращает список распарсенных товаров или None."""
    if not text or not text.strip():
        return None

    try:
        response = ollama.chat(
            model=OLLAMA_MODEL,
            messages=[
                {"role": "system", "content": SYSTEM_PROMPT},
                {"role": "user", "content": text},
            ],
        )

        # Поддержка обоих форматов SDK: dict (< 0.4) и object (>= 0.4)
        msg = response.message if hasattr(response, "message") else response["message"]
        raw = msg.content if hasattr(msg, "content") else msg["content"]

        clean = raw.strip()
        if clean.startswith("```"):
            clean = clean.split("\n", 1)[1] if "\n" in clean else clean[3:]
            if clean.endswith("```"):
                clean = clean[:-3]
            clean = clean.strip()

        parsed = json.loads(clean)

        if isinstance(parsed, dict):
            parsed = [parsed]

        if len(parsed) == 1 and "message" in parsed[0]:
            logger.debug("LLM: не объявление — %s", parsed[0]["message"])
            return None

        fields = ["brand", "model", "category", "price"]
        filtered = [item for item in parsed if any(item.get(f) is not None for f in fields)]

        return filtered if filtered else None

    except json.JSONDecodeError as e:
        logger.error("LLM вернул невалидный JSON: %s", e)
        return None
    except Exception as e:
        logger.error("Ошибка LLM-парсинга: %s", e)
        return None
