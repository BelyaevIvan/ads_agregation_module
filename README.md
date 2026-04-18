# BrandHunt — Агрегатор объявлений о брендовых товарах

Система мониторинга Telegram-каналов и групп ВКонтакте: ловит новые посты, парсит объявления через LLM (Ollama / gemma3:4b), сохраняет товары и фото в PostgreSQL + MinIO. Предоставляет REST API для поиска, фильтрации и управления объявлениями.

---

## Требования

- [Docker](https://docs.docker.com/get-docker/) + [Docker Compose](https://docs.docker.com/compose/) (v2)
- Доступ к Telegram API (api_id / api_hash)
- Доступ к VK API (access_token)

---

## Первый запуск (с нуля)

### 1. Заполнить `.env`

Скопировать шаблон и вставить реальные значения:

```bash
cp .env.example .env
# отредактировать .env: вставить TG_API_ID, TG_API_HASH, VK_ACCESS_TOKEN и пр.
```

### 2. Поднять инфраструктуру

```bash
docker compose up -d postgres minio ollama
```

Дождаться, пока все три контейнера перейдут в статус `healthy`:

```bash
docker compose ps
```

Миграции (создание таблиц) применяются автоматически при первом старте PostgreSQL.

### 3. Скачать LLM-модель (один раз)

```bash
docker exec -it brandhunt_ollama ollama pull gemma3:4b
```

Скачивается ~3 ГБ. Дождаться сообщения `success`.

### 4. Авторизоваться в Telegram (один раз)

```bash
docker compose run --rm --no-deps tg-monitor python auth_tg.py
```

Скрипт попросит номер телефона и код подтверждения. Файл сессии сохраняется в Docker volume `tg_sessions` и больше не требует повторной авторизации.

### 5. Запустить API-сервис

```bash
docker compose up -d api
```

API будет доступен на порту из `.env` (`API_PORT`, по умолчанию 3000).
Swagger UI: **http://localhost:3000/swagger/**

### 6. Запустить мониторы

```bash
docker compose up -d tg-monitor vk-monitor
```

### 7. Подготовить и запустить фронтенд

Один раз — поставить зависимости на хост (нужно для `package-lock.json` и для IDE):

```bash
cd frontend
npm install --registry=https://registry.npmmirror.com
cd ..
```

Потом собрать и запустить nginx-образ:

```bash
docker compose build frontend
docker compose up -d frontend
```

Открыть в браузере: **http://localhost:8080** (порт настраивается через `FRONTEND_PORT` в `.env`).

Подробнее про режимы работы (dev-сервер vs prod-nginx) — в разделе [Фронтенд](#фронтенд) ниже.

---

## Обычный запуск (после первого раза)

Все данные (БД, MinIO, сессия Telegram, состояние VK) хранятся в Docker volumes и не сбрасываются при перезапуске.

```bash
docker compose up -d
```

Эта команда поднимает весь стек. Миграции повторно не применяются (PostgreSQL видит, что data-volume уже инициализирован).

---

## Остановка

```bash
docker compose down        # остановить контейнеры (данные сохраняются)
docker compose down -v     # остановить и удалить все volumes (данные удалятся!)
```

---

## Проверка работоспособности

```bash
# Статус всех контейнеров
docker compose ps

# Логи API-сервиса
docker compose logs -f api

# Логи мониторов в реальном времени
docker compose logs -f tg-monitor
docker compose logs -f vk-monitor

# Health check API
curl http://localhost:3000/health

# Подключиться к PostgreSQL и проверить данные
docker exec -it brandhunt_postgres psql -U brandhunt -d brandhunt -c "SELECT COUNT(*) FROM listings;"

# Проверить MinIO (веб-консоль)
# Открыть в браузере: http://localhost:9001
# Логин/пароль — значения MINIO_ROOT_USER / MINIO_ROOT_PASSWORD из .env
```

---

## API-сервис

REST API на Go (стандартная библиотека `net/http`). Порт настраивается через `API_PORT` в `.env`.

### Запуск только бэкенда (без мониторинга)

```bash
docker compose up -d postgres minio api
```

### Swagger UI

Открыть в браузере: **http://localhost:3000/swagger/**

Все методы можно протестировать прямо из интерфейса. После вызова `/auth/login` токен подставляется автоматически.

### Авторизация

JWT без Redis. Токен живёт 30 дней от момента логина. Содержит `user_id` и `role`, подписан секретом `JWT_SECRET` из `.env`.

Создание администратора — только вручную через БД:
```bash
docker exec -it brandhunt_postgres psql -U brandhunt -d brandhunt -c "UPDATE users SET role = 'admin' WHERE email = 'your@email.com';"
```

### Эндпоинты

| Метод | Путь | Доступ | Описание |
|---|---|---|---|
| POST | `/api/v1/auth/register` | Public | Регистрация |
| POST | `/api/v1/auth/login` | Public | Вход, получение JWT |
| POST | `/api/v1/auth/logout` | Auth | Выход |
| GET | `/api/v1/listings` | Public | Поиск и фильтрация объявлений |
| GET | `/api/v1/listings/{id}` | Public | Карточка объявления |
| GET | `/api/v1/users/me` | Auth | Профиль пользователя |
| PUT | `/api/v1/users/me` | Auth | Обновить профиль |
| GET | `/api/v1/users/me/favorites` | Auth | Список избранного |
| POST | `/api/v1/users/me/favorites` | Auth | Добавить в избранное |
| DELETE | `/api/v1/users/me/favorites/{id}` | Auth | Убрать из избранного |
| GET | `/api/v1/admin/listings` | Admin | Все объявления (вкл. скрытые) |
| PATCH | `/api/v1/admin/listings/{id}/visibility` | Admin | Скрыть / восстановить |
| PATCH | `/api/v1/admin/listings/{id}/text` | Admin | Редактировать текст |
| DELETE | `/api/v1/admin/listings/{id}/photos/{photo_id}` | Admin | Удалить фото |
| GET | `/api/v1/admin/stats` | Admin | Сводная статистика |
| GET | `/api/v1/admin/stats/listings-by-day` | Admin | Динамика по дням |
| GET | `/api/v1/admin/stats/top-brands` | Admin | Топ брендов |
| GET | `/api/v1/admin/stats/top-cities` | Admin | Топ городов |
| GET | `/api/v1/admin/sources` | Admin | Список источников |
| POST | `/api/v1/admin/sources` | Admin | Добавить источник |
| PATCH | `/api/v1/admin/sources/{id}/toggle` | Admin | Вкл/выкл мониторинг |

---

## Ручное добавление каналов/групп

**Telegram** — добавить ID канала в `TG_WATCHED_CHANNELS` в `.env` (через запятую).

**VK** — добавить объект в `VK_GROUPS` в `.env`:
```
VK_GROUPS=[{"id": -12345678, "name": "Название группы"}, ...]
```

После изменения `.env` перезапустить соответствующий монитор:
```bash
docker compose up -d --force-recreate tg-monitor
docker compose up -d --force-recreate vk-monitor
```

---

## Структура проекта

```
ads_agregation_module/
├── docker-compose.yml
├── .env                        # секреты (не коммитится)
├── .env.example                # шаблон
├── backend/
│   ├── migrations/             # SQL-схема БД (применяется автоматически)
│   ├── api-service/            # Go REST API
│   │   ├── cmd/                # точка входа + Swagger
│   │   ├── config/             # конфигурация из env
│   │   ├── internal/
│   │   │   ├── handler/        # HTTP-хэндлеры
│   │   │   ├── service/        # бизнес-логика
│   │   │   ├── repository/     # SQL-запросы
│   │   │   ├── model/          # структуры данных
│   │   │   └── middleware/     # JWT, CORS, обработчик ошибок
│   │   └── Dockerfile
│   └── monitor-service/        # Python-воркеры TG и VK
│       ├── tg_monitor.py
│       ├── vk_monitor.py
│       ├── llm_parser.py
│       ├── db.py
│       ├── storage.py
│       ├── config.py
│       └── auth_tg.py
├── frontend/                   # SPA на React + Vite
│   ├── Dockerfile              # multi-stage: node builder → nginx
│   ├── nginx.conf
│   └── src/
│       ├── api/                # HTTP-клиент (fetch + JWT + тосты)
│       ├── contexts/           # AuthContext, ToastContext
│       ├── components/         # UI-компоненты (common/listings/admin)
│       ├── pages/              # экраны (Home, Search, Product, Auth, Profile, admin/)
│       └── styles/             # глобальные стили
└── claude-instructions/        # контекст для разработки
```

---

## Фронтенд

SPA на **React 18 + Vite + React Router**, без TypeScript. В продакшене раздаётся `nginx`, который также проксирует `/api/*` на `api`-сервис внутри docker-сети (один origin → без CORS).

### Первый запуск: установка зависимостей

Один раз надо выполнить локально на хосте:

```bash
cd frontend
npm install --registry=https://registry.npmmirror.com
```

Это создаст:
- `node_modules/` — библиотеки (в git не коммитится)
- `package-lock.json` — фиксация точных версий (**коммитится**, нужен для детерминированных сборок в Docker и CI)

Из РФ `registry.npmjs.org` часто лагает — флаг `--registry` переключает на зеркало, ставится за ~1 минуту.

> `node_modules` нужен локально для подсветки кода, автокомплита и ESLint в IDE — поэтому даже если ты планируешь работать только через Docker, один `npm install` на хосте обязателен.

### Вариант А — dev-сервер Vite (для разработки)

```bash
cd frontend
npm run dev
```

Открыть: **http://localhost:5173**

Что работает:
- Hot-reload — правки в коде применяются мгновенно, без пересборки
- Прокси `/api/*` → `http://localhost:3000` (адрес бэка), поэтому браузер видит всё как один origin
- Читаемые ошибки прямо в браузере

API при этом должен крутиться в Docker:
```bash
docker compose up -d postgres minio api
```

### Вариант Б — «как на проде» (nginx в Docker)

```bash
docker compose build frontend
docker compose up -d frontend
```

Открыть: **http://localhost:8080** (порт настраивается через `FRONTEND_PORT` в `.env`)

Первая сборка образа — пара минут (скачивание пакетов). Последующие — секунды, пока не меняется `package-lock.json` (Docker переиспользует закэшированный слой).

### URL-ы для проверки работы

| Что | URL | Описание |
|---|---|---|
| Фронт (prod) | http://localhost:8080 | nginx раздаёт SPA |
| Фронт (dev) | http://localhost:5173 | Vite dev server |
| API health | http://localhost:3000/health | должен ответить `{"status":"ok"}` |
| Swagger UI | http://localhost:3000/swagger/ | интерактивный тест всех эндпоинтов |
| MinIO Console | http://localhost:9001 | веб-интерфейс хранилища фото |

Быстрый sanity-check после запуска:
```bash
# все контейнеры живы и healthy
docker compose ps

# бэк отвечает
curl http://localhost:3000/health

# в браузере открыть http://localhost:8080 — видим главную BrandHunt
```

### Обработка ошибок

При любой неудачной операции (не-2xx от API) в правом верхнем углу всплывает тост **«Ошибка. Не удалось выполнить запрос»**. Вторым текстом — сообщение сервера. Удобно открыть DevTools → Network и сразу увидеть, какой запрос упал.

Исключения — логин, регистрация, сохранение профиля: ошибки показываются inline в самой форме.
