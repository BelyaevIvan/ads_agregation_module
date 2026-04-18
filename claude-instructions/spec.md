# spec.md — Спецификация API BrandHunt (api-service)

Язык реализации: **Go**, без фреймворков (стандартная библиотека `net/http`).

---

## Общие соглашения

### Base URL
```
/api/v1
```

### Аутентификация
Защищённые эндпоинты требуют заголовок:
```
Authorization: Bearer <jwt_token>
```
JWT содержит: `user_id` (UUID), `role` (string: `"user"` | `"admin"`), `exp` (Unix timestamp).

### Роли доступа
| Маркер | Значение |
|---|---|
| 🌐 Public | доступно всем без токена |
| 🔐 Auth | требуется валидный JWT (любая роль) |
| 🛡 Admin | требуется JWT с `role = "admin"` |

### Пагинация
Все списковые эндпоинты поддерживают:
| Query-параметр | Тип | По умолчанию | Описание |
|---|---|---|---|
| `limit` | int | 20 | Кол-во записей на странице (макс. 100) |
| `offset` | int | 0 | Смещение |

Ответ всегда включает:
```json
{
  "items": [...],
  "total": 248,
  "limit": 20,
  "offset": 0
}
```

### Формат ошибок (универсальный обработчик)
При любой ошибке возвращается единый формат:
```json
{
  "statusCode": 400,
  "url": "/api/v1/listings?q=nike",
  "message": "параметр 'limit' должен быть положительным числом",
  "date": "2026-04-17T14:32:00Z"
}
```
Для непредвиденных ошибок (500): `"message": "внутренняя ошибка сервера, попробуйте позже"`.

Обработчик реализуется в `internal/middleware/error.go` и оборачивает все хэндлеры.

---

## Группы эндпоинтов

1. [Объявления (Listings)](#1-объявления)
2. [Фотографии объявлений (Photos)](#2-фотографии-объявлений)
3. [Авторизация (Auth)](#3-авторизация)
4. [Профиль пользователя (User)](#4-профиль-пользователя)
5. [Избранное (Favorites)](#5-избранное)
6. [Административные методы (Admin)](#6-административные-методы)

---

## 1. Объявления

### `GET /listings` 🌐
**Поиск и фильтрация объявлений.**

#### Query-параметры
| Параметр | Тип | Обязательный | Описание |
|---|---|---|---|
| `q` | string | нет | Полнотекстовый поисковый запрос (по `brand`, `model`, `original_text`) |
| `brand` | string[] | нет | Фильтр по бренду. Несколько значений: `brand=Nike&brand=Adidas` |
| `category` | string[] | нет | Фильтр по категории |
| `city` | string[] | нет | Фильтр по городу |
| `condition` | string | нет | `"new"` / `"used"` |
| `size_rus` | string[] | нет | Фильтр по размеру RUS. Несколько: `size_rus=42&size_rus=43` |
| `size_eu` | string[] | нет | Фильтр по размеру EU |
| `size_us` | string[] | нет | Фильтр по размеру US |
| `price_min` | float | нет | Минимальная цена |
| `price_max` | float | нет | Максимальная цена |
| `include_no_size` | bool | нет | `true` — включать объявления без размера (default: `true`) |
| `include_no_price` | bool | нет | `true` — включать объявления без цены (default: `true`) |
| `include_no_city` | bool | нет | `true` — включать объявления без города (default: `true`) |
| `platform` | string[] | нет | `"telegram"` / `"vk"` |
| `sort` | string | нет | `"date_desc"` (default) / `"price_asc"` / `"price_desc"` |
| `limit` | int | нет | default: 20 |
| `offset` | int | нет | default: 0 |

#### Тело ответа `200 OK`
```json
{
  "items": [
    {
      "id": "uuid",
      "brand": "Nike",
      "model": "Air Max 90",
      "category": "sneakers",
      "color": "White",
      "price": 8500.00,
      "city": "Москва",
      "condition": "new",
      "size_rus": ["42", "43"],
      "size_eu": ["42"],
      "size_us": ["8.5"],
      "cover_photo_url": "https://minio.../photo.jpg",
      "platform": "telegram",
      "posted_at": "2026-04-17T12:00:00Z",
      "created_at": "2026-04-17T12:05:00Z"
    }
  ],
  "total": 248,
  "limit": 20,
  "offset": 0
}
```

#### Алгоритм
1. Валидировать query-параметры (`limit` ≥ 1, `limit` ≤ 100, `offset` ≥ 0, `sort` из допустимых значений).
2. Собрать SQL-запрос к таблице `listings` с условием `is_hidden = FALSE`.
3. Если передан `q` — применить полнотекстовый поиск: `fts @@ plainto_tsquery('russian', $q)`.
4. Для каждого из фильтров-массивов (`brand`, `category`, `city`, `platform`, `size_rus`, `size_eu`, `size_us`) — если параметр передан, добавить условие `AND field = ANY($arr)`. Для полей-массивов в БД (`size_rus`, `size_eu`, `size_us`): `AND $arr && size_rus` (пересечение массивов).
5. Обработать `include_no_size/price/city`: если `true` — добавить `OR field IS NULL` к соответствующему фильтру.
6. Обработать `price_min` / `price_max`: `AND price >= $min AND price <= $max`. Если `include_no_price = true` — добавить `OR price IS NULL`.
7. Применить сортировку и `LIMIT` / `OFFSET`.
8. Параллельно выполнить `COUNT(*)` с теми же условиями для поля `total`.
9. Для каждого объявления подтянуть `cover_photo_url` из `listing_photos WHERE listing_id = id AND is_cover = TRUE LIMIT 1`.
10. Вернуть результат.

---

### `GET /listings/{id}` 🌐
**Получить полную карточку объявления.**

#### Path-параметры
| Параметр | Тип | Описание |
|---|---|---|
| `id` | UUID | ID объявления |

#### Тело ответа `200 OK`
```json
{
  "id": "uuid",
  "source": {
    "platform": "telegram",
    "title": "@sneakers_moscow",
    "external_id": "-1001234567890"
  },
  "original_text": "Продаю Nike Air Max 90...",
  "post_url": "https://t.me/sneakers_moscow/123",
  "posted_at": "2026-04-17T12:00:00Z",
  "brand": "Nike",
  "model": "Air Max 90",
  "category": "sneakers",
  "color": "White",
  "price": 8500.00,
  "city": "Москва",
  "condition": "new",
  "size_rus": ["42", "43"],
  "size_eu": ["42"],
  "size_us": ["8.5"],
  "photos": [
    { "url": "https://minio.../photo1.jpg", "is_cover": true, "sort_order": 0 },
    { "url": "https://minio.../photo2.jpg", "is_cover": false, "sort_order": 1 }
  ],
  "created_at": "2026-04-17T12:05:00Z"
}
```

#### Алгоритм
1. Валидировать `id` как UUID.
2. Запросить запись из `listings` по `id`, где `is_hidden = FALSE`.
3. Если запись не найдена — вернуть `404`.
4. JOIN с таблицей `sources` для получения данных об источнике.
5. Запросить все фотографии из `listing_photos WHERE listing_id = id ORDER BY sort_order ASC`.
6. Собрать и вернуть ответ.

---

## 2. Фотографии объявлений

### `DELETE /admin/listings/{id}/photos/{photo_id}` 🛡
*(описан в разделе Admin — см. ниже)*

---

## 3. Авторизация

### `POST /auth/register` 🌐
**Регистрация нового пользователя.**

#### Тело запроса (JSON)
```json
{
  "email": "user@example.com",
  "password": "strongpassword123"
}
```

| Поле | Тип | Обязательное | Описание |
|---|---|---|---|
| `email` | string | да | Валидный email, уникальный |
| `password` | string | да | Минимум 8 символов |

#### Тело ответа `201 Created`
```json
{
  "id": "uuid",
  "email": "user@example.com",
  "role": "user",
  "created_at": "2026-04-17T12:00:00Z"
}
```

#### Алгоритм
1. Валидировать формат email и длину пароля (≥ 8 символов).
2. Проверить, что пользователь с таким email не существует (`SELECT id FROM users WHERE email = $email`). Если существует — `409 Conflict`.
3. Хэшировать пароль через `bcrypt` (cost 12).
4. Вставить новую запись в `users` с `role = "user"`.
5. Вернуть созданного пользователя без `password_hash`.

---

### `POST /auth/login` 🌐
**Аутентификация пользователя, получение JWT.**

#### Тело запроса (JSON)
```json
{
  "email": "user@example.com",
  "password": "strongpassword123"
}
```

#### Тело ответа `200 OK`
```json
{
  "access_token": "eyJhbGci...",
  "token_type": "Bearer",
  "expires_at": "2026-05-17T12:00:00Z"
}
```

#### Алгоритм
1. Найти пользователя по email. Если не найден — `401 Unauthorized` (намеренно не уточняем, что именно неверно).
2. Сравнить пароль с `password_hash` через `bcrypt.CompareHashAndPassword`. Если не совпадает — `401`.
3. Сгенерировать JWT с payload: `{ user_id, role, exp: now + 30 days }`, подписать секретом из конфига.
4. Вернуть токен.

---

### `POST /auth/logout` 🔐
**Выход из системы (клиентская инвалидация токена).**

#### Тело ответа `200 OK`
```json
{ "message": "выход выполнен успешно" }
```

#### Алгоритм
Поскольку JWT stateless и Redis не используется, реальная инвалидация не происходит на сервере. Эндпоинт существует для семантической корректности клиентского кода: клиент удаляет токен из локального хранилища. Сервер валидирует наличие токена в заголовке и возвращает `200`.

---

## 4. Профиль пользователя

### `GET /users/me` 🔐
**Получить профиль текущего пользователя.**

#### Тело ответа `200 OK`
```json
{
  "id": "uuid",
  "email": "user@example.com",
  "full_name": "Андрей Иванов",
  "phone": "+7 999 123-45-67",
  "tg_link": "@andrey_tg",
  "vk_link": "vk.com/andrey",
  "role": "user",
  "created_at": "2026-04-17T12:00:00Z"
}
```

#### Алгоритм
1. Извлечь `user_id` из JWT.
2. Запросить пользователя из `users` по `id`. Если не найден — `404` (токен валиден, но запись удалена).
3. Вернуть данные без `password_hash`.

---

### `PUT /users/me` 🔐
**Обновить профиль текущего пользователя.**

#### Тело запроса (JSON)
```json
{
  "full_name": "Андрей Иванов",
  "phone": "+7 999 123-45-67",
  "tg_link": "@andrey_tg",
  "vk_link": "vk.com/andrey"
}
```

| Поле | Тип | Обязательное | Описание |
|---|---|---|---|
| `full_name` | string | нет | ФИО |
| `phone` | string | нет | Номер телефона |
| `tg_link` | string | нет | Ссылка/юзернейм Telegram |
| `vk_link` | string | нет | Ссылка ВКонтакте |

#### Тело ответа `200 OK`
Обновлённый объект пользователя (тот же формат, что `GET /users/me`).

#### Алгоритм
1. Извлечь `user_id` из JWT.
2. Валидировать входные поля (длина, формат).
3. Обновить только переданные (ненулевые) поля через `UPDATE users SET ... WHERE id = $user_id`.
4. Вернуть обновлённую запись.

---

## 5. Избранное

### `GET /users/me/favorites` 🔐
**Получить список избранных объявлений текущего пользователя.**

#### Query-параметры
| Параметр | Тип | Описание |
|---|---|---|
| `limit` | int | default: 20 |
| `offset` | int | default: 0 |

#### Тело ответа `200 OK`
```json
{
  "items": [
    {
      "id": "uuid",
      "listing": {
        "id": "uuid",
        "brand": "Nike",
        "model": "Air Max 90",
        "price": 8500.00,
        "city": "Москва",
        "size_rus": ["42"],
        "size_eu": ["42"],
        "size_us": ["8.5"],
        "cover_photo_url": "https://minio.../photo.jpg",
        "platform": "telegram",
        "is_hidden": false
      },
      "saved_at": "2026-04-17T12:00:00Z"
    }
  ],
  "total": 12,
  "limit": 20,
  "offset": 0
}
```

#### Алгоритм
1. Извлечь `user_id` из JWT.
2. Запросить `favorites` с JOIN на `listings` и `listing_photos` (для обложки).
3. Включать в результат объявления независимо от `is_hidden` — пользователь должен видеть, что объявление скрыто, а не терять его из избранного.
4. Применить `LIMIT` / `OFFSET`, вернуть с `total`.

---

### `POST /users/me/favorites` 🔐
**Добавить объявление в избранное.**

#### Тело запроса (JSON)
```json
{
  "listing_id": "uuid"
}
```

#### Тело ответа `201 Created`
```json
{
  "id": "uuid",
  "listing_id": "uuid",
  "saved_at": "2026-04-17T12:00:00Z"
}
```

#### Алгоритм
1. Извлечь `user_id` из JWT.
2. Проверить, что объявление с `listing_id` существует в `listings`. Если нет — `404`.
3. Попытаться вставить запись в `favorites`. Использовать `INSERT ... ON CONFLICT (user_id, listing_id) DO NOTHING` — эндпоинт идемпотентен.
4. Если запись уже существовала (конфликт) — вернуть `200 OK` с существующей записью.
5. Если создана новая — вернуть `201 Created`.

---

### `DELETE /users/me/favorites/{listing_id}` 🔐
**Удалить объявление из избранного.**

#### Path-параметры
| Параметр | Тип | Описание |
|---|---|---|
| `listing_id` | UUID | ID объявления |

#### Тело ответа `204 No Content`
Пустое тело.

#### Алгоритм
1. Извлечь `user_id` из JWT.
2. Выполнить `DELETE FROM favorites WHERE user_id = $user_id AND listing_id = $listing_id`.
3. Если ни одна строка не удалена — вернуть `404`.
4. Вернуть `204`.

---

## 6. Административные методы

> Все эндпоинты этой группы требуют `role = "admin"` в JWT. Middleware проверяет роль до вызова хэндлера и возвращает `403 Forbidden` при несоответствии.

---

### `GET /admin/listings` 🛡
**Получить список всех объявлений (включая скрытые) для панели администратора.**

#### Query-параметры
| Параметр | Тип | Описание |
|---|---|---|
| `q` | string | Полнотекстовый поиск |
| `status` | string | `"active"` / `"hidden"` / `""` (все, default) |
| `platform` | string[] | `"telegram"` / `"vk"` |
| `sort` | string | `"date_desc"` (default) / `"price_asc"` / `"price_desc"` |
| `limit` | int | default: 20 |
| `offset` | int | default: 0 |

#### Тело ответа `200 OK`
```json
{
  "items": [
    {
      "id": "uuid",
      "brand": "Nike",
      "model": "Air Max 90",
      "price": 8500.00,
      "city": "Москва",
      "size_rus": ["42"],
      "size_eu": ["42"],
      "size_us": ["8.5"],
      "platform": "telegram",
      "source_title": "@sneakers_moscow",
      "cover_photo_url": "https://minio.../photo.jpg",
      "is_hidden": false,
      "posted_at": "2026-04-17T12:00:00Z",
      "created_at": "2026-04-17T12:05:00Z"
    }
  ],
  "total": 12847,
  "limit": 20,
  "offset": 0
}
```

#### Алгоритм
1. Валидировать параметры.
2. Собрать SQL без фильтра по `is_hidden` (в отличие от публичного `/listings`).
3. Если `status = "active"` — добавить `WHERE is_hidden = FALSE`; если `"hidden"` — `WHERE is_hidden = TRUE`.
4. JOIN с `sources` для `source_title`, подтянуть `cover_photo_url`.
5. Вернуть с пагинацией.

---

### `GET /admin/listings/{id}` 🛡
**Получить полную карточку объявления (включая скрытое).**

Отличается от публичного `GET /listings/{id}` только тем, что не фильтрует по `is_hidden`: админ должен иметь возможность посмотреть карточку и содержимое скрытого объявления (например, чтобы отредактировать текст или удалить фото).

#### Path-параметры
| Параметр | Тип | Описание |
|---|---|---|
| `id` | UUID | ID объявления |

#### Тело ответа `200 OK`
Тот же формат, что у `GET /listings/{id}`, но поле `is_hidden` присутствует всегда.

#### Алгоритм
1. Валидировать `id` как UUID.
2. Запросить запись из `listings` по `id` **без фильтра** `is_hidden`.
3. Если не найдена — `404`.
4. JOIN с `sources`, подтянуть фото из `listing_photos`.
5. Вернуть ответ.

---

### `PATCH /admin/listings/{id}/visibility` 🛡
**Скрыть или восстановить объявление.**

#### Path-параметры
| Параметр | Тип | Описание |
|---|---|---|
| `id` | UUID | ID объявления |

#### Тело запроса (JSON)
```json
{
  "is_hidden": true
}
```

#### Тело ответа `200 OK`
```json
{
  "id": "uuid",
  "is_hidden": true
}
```

#### Алгоритм
1. Валидировать `id` и наличие поля `is_hidden` в теле.
2. Проверить существование объявления. Если нет — `404`.
3. Выполнить `UPDATE listings SET is_hidden = $is_hidden WHERE id = $id`.
4. Вернуть обновлённые поля.

---

### `PATCH /admin/listings/{id}/text` 🛡
**Отредактировать оригинальный текст объявления.**

#### Path-параметры
| Параметр | Тип | Описание |
|---|---|---|
| `id` | UUID | ID объявления |

#### Тело запроса (JSON)
```json
{
  "original_text": "Обновлённый текст объявления..."
}
```

| Поле | Тип | Обязательное | Описание |
|---|---|---|---|
| `original_text` | string | да | Новый текст. Не может быть пустой строкой |

#### Тело ответа `200 OK`
```json
{
  "id": "uuid",
  "original_text": "Обновлённый текст объявления..."
}
```

#### Алгоритм
1. Валидировать `id` и непустоту `original_text`.
2. Проверить существование объявления. Если нет — `404`.
3. Выполнить `UPDATE listings SET original_text = $text WHERE id = $id`.
4. Вернуть обновлённые поля.

---

### `DELETE /admin/listings/{id}/photos/{photo_id}` 🛡
**Удалить фотографию объявления.**

#### Path-параметры
| Параметр | Тип | Описание |
|---|---|---|
| `id` | UUID | ID объявления |
| `photo_id` | UUID | ID фотографии |

#### Тело ответа `200 OK`
```json
{
  "deleted_photo_id": "uuid",
  "new_cover_id": "uuid или null"
}
```

#### Алгоритм
1. Валидировать оба UUID.
2. Получить запись из `listing_photos` по `photo_id` и `listing_id = id`. Если нет — `404`.
3. Запомнить `was_cover = photo.is_cover` и `photo_url`.
4. Удалить запись: `DELETE FROM listing_photos WHERE id = $photo_id`.
5. Если `was_cover = true`:
   - Найти следующую по `sort_order` фотографию: `SELECT id FROM listing_photos WHERE listing_id = $id ORDER BY sort_order ASC LIMIT 1`.
   - Если найдена — обновить `UPDATE listing_photos SET is_cover = TRUE WHERE id = $next_id`, вернуть `new_cover_id`.
   - Если фотографий не осталось — `new_cover_id = null`.
6. Если `was_cover = false` — `new_cover_id = null`.
7. Удалить файл из MinIO по `photo_url` (вызов S3 API через storage-слой). Если MinIO вернул ошибку — залогировать, не возвращать ошибку клиенту (файл удалён из БД, консистентность БД приоритетнее).
8. Вернуть ответ.

---

### `GET /admin/stats` 🛡
**Получить сводную статистику для дашборда.**

#### Тело ответа `200 OK`
```json
{
  "total_listings": 12847,
  "active_listings": 12780,
  "hidden_listings": 67,
  "total_users": 1284,
  "active_sources": 340,
  "new_listings_today": 143,
  "new_users_week": 28
}
```

#### Алгоритм
Выполнить набор агрегирующих запросов (можно параллельно через горутины):
1. `SELECT COUNT(*) FROM listings` → `total_listings`
2. `SELECT COUNT(*) FROM listings WHERE is_hidden = FALSE` → `active_listings`
3. `SELECT COUNT(*) FROM listings WHERE is_hidden = TRUE` → `hidden_listings`
4. `SELECT COUNT(*) FROM users` → `total_users`
5. `SELECT COUNT(*) FROM sources WHERE is_active = TRUE` → `active_sources`
6. `SELECT COUNT(*) FROM listings WHERE created_at >= NOW() - INTERVAL '1 day'` → `new_listings_today`
7. `SELECT COUNT(*) FROM users WHERE created_at >= NOW() - INTERVAL '7 days'` → `new_users_week`

---

### `GET /admin/stats/listings-by-day` 🛡
**Получить динамику поступления объявлений по дням.**

#### Query-параметры
| Параметр | Тип | По умолчанию | Описание |
|---|---|---|---|
| `days` | int | 30 | Период в днях (7, 30, 90) |

#### Тело ответа `200 OK`
```json
{
  "items": [
    { "date": "2026-04-17", "count": 143 },
    { "date": "2026-04-16", "count": 127 }
  ]
}
```

#### Алгоритм
1. Валидировать `days` — допустимые значения: 7, 30, 90.
2. Выполнить:
```sql
SELECT DATE(created_at) AS date, COUNT(*) AS count
FROM listings
WHERE created_at >= NOW() - INTERVAL '$days days'
GROUP BY DATE(created_at)
ORDER BY date DESC
```
3. Вернуть результат.

---

### `GET /admin/stats/top-brands` 🛡
**Топ брендов по числу объявлений.**

#### Query-параметры
| Параметр | Тип | По умолчанию | Описание |
|---|---|---|---|
| `limit` | int | 10 | Количество брендов в топе |

#### Тело ответа `200 OK`
```json
{
  "items": [
    { "brand": "Nike", "count": 3241 },
    { "brand": "New Balance", "count": 2347 }
  ]
}
```

#### Алгоритм
```sql
SELECT brand, COUNT(*) AS count
FROM listings
WHERE brand IS NOT NULL AND is_hidden = FALSE
GROUP BY brand
ORDER BY count DESC
LIMIT $limit
```

---

### `GET /admin/stats/top-cities` 🛡
**Топ городов по числу объявлений.**

#### Query-параметры
| Параметр | Тип | По умолчанию | Описание |
|---|---|---|---|
| `limit` | int | 10 | Количество городов в топе |

#### Тело ответа `200 OK`
```json
{
  "items": [
    { "city": "Москва", "count": 5610 },
    { "city": "Санкт-Петербург", "count": 3320 }
  ]
}
```

#### Алгоритм
```sql
SELECT city, COUNT(*) AS count
FROM listings
WHERE city IS NOT NULL AND is_hidden = FALSE
GROUP BY city
ORDER BY count DESC
LIMIT $limit
```

---

### `GET /admin/sources` 🛡
**Получить список всех источников мониторинга.**

#### Query-параметры
| Параметр | Тип | Описание |
|---|---|---|
| `limit` | int | default: 50 |
| `offset` | int | default: 0 |

#### Тело ответа `200 OK`
```json
{
  "items": [
    {
      "id": "uuid",
      "platform": "telegram",
      "external_id": "-1001234567890",
      "title": "@sneakers_moscow",
      "is_active": true,
      "listings_count": 1243,
      "added_at": "2026-01-01T00:00:00Z"
    }
  ],
  "total": 340,
  "limit": 50,
  "offset": 0
}
```

#### Алгоритм
1. JOIN `sources` с `COUNT(listings.id)` через LEFT JOIN для подсчёта объявлений каждого источника.
2. Вернуть с пагинацией.

---

### `POST /admin/sources` 🛡
**Добавить новый источник мониторинга.**

#### Тело запроса (JSON)
```json
{
  "platform": "telegram",
  "external_id": "-1001234567890",
  "title": "@sneakers_moscow"
}
```

| Поле | Тип | Обязательное | Описание |
|---|---|---|---|
| `platform` | string | да | `"telegram"` / `"vk"` |
| `external_id` | string | да | ID канала/группы во внешней системе |
| `title` | string | нет | Отображаемое название |

#### Тело ответа `201 Created`
Объект созданного источника (тот же формат, что в списке).

#### Алгоритм
1. Валидировать `platform` (только `"telegram"` или `"vk"`).
2. Проверить уникальность `(platform, external_id)`. Если уже существует — `409 Conflict`.
3. Вставить в `sources` с `is_active = TRUE`.
4. Вернуть созданную запись.

---

### `PATCH /admin/sources/{id}/toggle` 🛡
**Включить или отключить мониторинг источника.**

#### Path-параметры
| Параметр | Тип | Описание |
|---|---|---|
| `id` | UUID | ID источника |

#### Тело запроса (JSON)
```json
{
  "is_active": false
}
```

#### Тело ответа `200 OK`
```json
{
  "id": "uuid",
  "is_active": false
}
```

#### Алгоритм
1. Проверить существование источника. Если нет — `404`.
2. `UPDATE sources SET is_active = $is_active WHERE id = $id`.
3. Вернуть обновлённые поля.

---

## Сводная таблица эндпоинтов

| Метод | Путь | Доступ | Описание |
|---|---|---|---|
| GET | `/api/v1/listings` | 🌐 | Поиск и фильтрация объявлений |
| GET | `/api/v1/listings/{id}` | 🌐 | Карточка объявления |
| POST | `/api/v1/auth/register` | 🌐 | Регистрация |
| POST | `/api/v1/auth/login` | 🌐 | Вход, получение JWT |
| POST | `/api/v1/auth/logout` | 🔐 | Выход |
| GET | `/api/v1/users/me` | 🔐 | Профиль текущего пользователя |
| PUT | `/api/v1/users/me` | 🔐 | Обновить профиль |
| GET | `/api/v1/users/me/favorites` | 🔐 | Список избранного |
| POST | `/api/v1/users/me/favorites` | 🔐 | Добавить в избранное |
| DELETE | `/api/v1/users/me/favorites/{listing_id}` | 🔐 | Убрать из избранного |
| GET | `/api/v1/admin/listings` | 🛡 | Все объявления (включая скрытые) |
| GET | `/api/v1/admin/listings/{id}` | 🛡 | Карточка объявления (включая скрытые) |
| PATCH | `/api/v1/admin/listings/{id}/visibility` | 🛡 | Скрыть / восстановить |
| PATCH | `/api/v1/admin/listings/{id}/text` | 🛡 | Редактировать текст |
| DELETE | `/api/v1/admin/listings/{id}/photos/{photo_id}` | 🛡 | Удалить фото |
| GET | `/api/v1/admin/stats` | 🛡 | Сводная статистика |
| GET | `/api/v1/admin/stats/listings-by-day` | 🛡 | Динамика по дням |
| GET | `/api/v1/admin/stats/top-brands` | 🛡 | Топ брендов |
| GET | `/api/v1/admin/stats/top-cities` | 🛡 | Топ городов |
| GET | `/api/v1/admin/sources` | 🛡 | Список источников |
| POST | `/api/v1/admin/sources` | 🛡 | Добавить источник |
| PATCH | `/api/v1/admin/sources/{id}/toggle` | 🛡 | Вкл/выкл мониторинг |
