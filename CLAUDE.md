# CLAUDE.md — Контекст проекта BrandHunt

Этот файл читается Claude перед началом любой работы над проектом.

---

## Что это за проект

**BrandHunt** — веб-система агрегации объявлений о продаже брендовых товаров (одежда, обувь, электроника) с рук.

Система в режиме реального времени мониторит Telegram-каналы и группы ВКонтакте, извлекает из постов структурированные атрибуты товаров с помощью LLM (бренд, модель, размер, цена, город и т.д.), сохраняет результат в БД и предоставляет пользователям поисковый веб-интерфейс.

**Три ключевых процесса:**
1. **Мониторинг** — фоновый воркер слушает TG/VK и ловит новые посты
2. **Парсинг** — текст поста отправляется в LLM, возвращаются структурированные атрибуты
3. **Поиск** — пользователь ищет товары через веб-интерфейс с фильтрами

**Роли пользователей:** Гость, Зарегистрированный пользователь, Администратор.

---

## Структура репозитория

```
ads_agregation_module/
│
├── CLAUDE.md                       # этот файл
├── README.md                       # инструкция по запуску проекта
├── docker-compose.yml              # поднимает весь стек одной командой
├── Caddyfile                       # конфиг Caddy (TLS + reverse-proxy)
├── .env.example                    # шаблон переменных окружения
├── .env                            # реальные секреты (в git не коммитится)
│
├── claude-instructions/            # контекст для Claude
│   ├── database.md                 # полная схема БД
│   └── spec.md                     # спецификация API (все методы)
│
├── backend/
│   │
│   ├── docs/                       # документация API (AsciiDoc)
│   │   ├── index.adoc              # корневой список всех методов
│   │   └── {endpoint-name}/        # одна папка на один метод
│   │       ├── {endpoint-name}.adoc      # главный файл (разделы 1–4)
│   │       ├── request.adoc              # пример запроса
│   │       ├── response-success.adoc     # пример успешного ответа
│   │       ├── response-errors.adoc      # все возможные ошибки
│   │       └── diagram.puml              # PlantUML-диаграмма алгоритма
│   │
│   ├── migrations/                 # SQL-миграции (общие для обоих сервисов)
│   │   ├── 001_init.sql
│   │   └── ...
│   │
│   ├── api-service/                # основной сервис (Go, без фреймворков)
│   │   ├── cmd/
│   │   │   ├── main.go             # точка входа, регистрация роутов
│   │   │   └── swagger.go          # Swagger UI + встроенная OpenAPI-спека
│   │   ├── config/
│   │   │   └── config.go           # чтение env-переменных в структуру
│   │   ├── internal/
│   │   │   ├── handler/            # HTTP-хэндлеры (приём/отправка запросов)
│   │   │   │   ├── auth.go
│   │   │   │   ├── user.go
│   │   │   │   ├── listing.go
│   │   │   │   ├── favorite.go
│   │   │   │   ├── admin.go
│   │   │   │   └── filters.go
│   │   │   ├── service/            # бизнес-логика, валидация
│   │   │   │   ├── auth.go
│   │   │   │   ├── user.go
│   │   │   │   ├── listing.go
│   │   │   │   ├── favorite.go
│   │   │   │   ├── source.go
│   │   │   │   ├── stats.go
│   │   │   │   ├── photo.go
│   │   │   │   └── filters.go      # + in-memory кэш (TTL 5 мин)
│   │   │   ├── repository/         # SQL-запросы к PostgreSQL
│   │   │   │   ├── user.go
│   │   │   │   ├── listing.go
│   │   │   │   ├── favorite.go
│   │   │   │   ├── source.go
│   │   │   │   ├── stats.go
│   │   │   │   ├── photo.go
│   │   │   │   ├── filters.go
│   │   │   │   └── photo_url.go    # переписывание URL фото minio:9000 → /minio/...
│   │   │   ├── model/              # структуры данных (entities, DTO)
│   │   │   │   ├── common.go
│   │   │   │   ├── user.go
│   │   │   │   ├── listing.go
│   │   │   │   ├── favorite.go
│   │   │   │   ├── source.go
│   │   │   │   ├── stats.go
│   │   │   │   └── filters.go
│   │   │   └── middleware/
│   │   │       ├── auth.go         # проверка JWT, извлечение user_id и role
│   │   │       ├── cors.go         # CORS-заголовки
│   │   │       └── error.go        # универсальный обработчик ошибок + JSON-хелпер
│   │   ├── Dockerfile
│   │   ├── go.mod
│   │   └── go.sum
│   │
│   └── monitor-service/            # сервис мониторинга TG/VK (Python)
│       ├── tg_monitor.py           # воркер Telegram (asyncio + Telethon)
│       ├── vk_monitor.py           # воркер ВКонтакте (polling)
│       ├── llm_parser.py           # парсинг через Ollama (gemma3:4b)
│       ├── db.py                   # запись в PostgreSQL
│       ├── storage.py              # загрузка фото в MinIO + public-read политика
│       ├── config.py               # конфиг из env-переменных
│       ├── auth_tg.py              # интерактивная авторизация Telegram (один раз)
│       ├── Dockerfile
│       └── requirements.txt
│
└── frontend/                       # SPA на React + Vite
    ├── Dockerfile                  # multi-stage: node builder → nginx
    ├── nginx.conf                  # proxy /api → api:8080, /minio → minio:9000, SPA fallback
    ├── package.json
    ├── package-lock.json           # коммитится, нужен для npm ci в Docker-сборке
    ├── vite.config.js
    ├── .npmrc                      # зеркало npm (registry.npmmirror.com)
    ├── index.html
    └── src/
        ├── main.jsx                # точка входа, монтирует App в #root
        ├── App.jsx                 # корневой роутер (react-router)
        ├── api/                    # HTTP-клиент (слой доступа к бэку)
        │   ├── client.js           # fetch-обёртка: JWT, тост при ошибке
        │   ├── auth.js             # register/login/logout/me/updateMe
        │   ├── listings.js         # search, getById
        │   ├── favorites.js        # list/add/remove
        │   ├── admin.js            # все админские методы
        │   └── filters.js          # метаданные для UI-фильтров (размеры и т.д.)
        ├── contexts/               # глобальное состояние через Context API
        │   ├── AuthContext.jsx     # user + токен в localStorage
        │   ├── ToastContext.jsx    # очередь тостов (ошибки/успех)
        │   └── FiltersContext.jsx  # кэш метаданных для UI-фильтров
        ├── components/             # переиспользуемые UI-блоки
        │   ├── common/             # Navbar, ToastContainer, Spinner, Pagination,
        │   │                       # ProtectedRoute, AdminRoute, EmptyState
        │   ├── listings/           # ListingCard, FiltersPanel
        │   └── admin/              # EditTextModal, PhotosModal, AddSourceModal
        ├── pages/                  # одна папка = один роут
        │   ├── HomePage.jsx
        │   ├── SearchResultsPage.jsx
        │   ├── ListingDetailPage.jsx
        │   ├── AuthPage.jsx        # единый экран с табами вход/регистрация
        │   ├── ProfilePage.jsx     # данные + избранное
        │   ├── ForbiddenPage.jsx   # 403
        │   ├── NotFoundPage.jsx    # 404 (fallback-роут)
        │   └── admin/              # админ-зона (защищена AdminRoute)
        │       ├── AdminLayout.jsx
        │       ├── DashboardPage.jsx
        │       ├── AdminListingsPage.jsx
        │       └── SourcesPage.jsx
        ├── utils/
        │   └── format.js           # формат цен, дат, инициалов
        └── styles/
            └── index.css           # глобальные стили (перенесены из прототипа)
```

---

## Архитектура бэкенда (api-service)

**3-слойная структура:** `repository/` → `service/` → `handler/`, обёрнуто middleware.

- **repository/** — SQL-запросы. Единственный слой, который знает о существовании PostgreSQL. Возвращает доменные структуры из `model/` (или `sql.ErrNoRows` как сигнал «не нашли»).
- **service/** — бизнес-логика и валидация: проверка UUID, длин полей, допустимых enum-значений, уникальности, правил видимости. Возвращает ошибки через `middleware.NewAppError(status, message)` — они подхватываются обработчиком ошибок.
- **handler/** — HTTP-обёртка: разбор query/path/body, вызов сервиса, сериализация в JSON. Ничего про SQL не знает.
- **middleware/** — сквозные заботы: CORS, проверка JWT (`auth.go`), проверка роли admin (`RequireAdmin`), универсальный обработчик ошибок (`error.go`) с JSON-хелпером.

**Паттерн `HandlerFunc`:** хэндлеры возвращают `error`, а не пишут ошибки напрямую. Функция-адаптер `ErrorHandler` превращает возвращённую ошибку в JSON-ответ стандартного формата. Это даёт единую точку обработки и избавляет хэндлеры от повторяющегося кода.

**Универсальный формат ошибок** (реализован в `middleware/error.go`):
```json
{
  "statusCode": 400,
  "url": "/api/v1/listings",
  "message": "описание ошибки",
  "date": "2026-04-17T14:32:00Z"
}
```
Для непредвиденных ошибок (500): `"message": "внутренняя ошибка сервера, попробуйте позже"`.

**Кэширование:** `FiltersService` держит in-memory кэш с TTL 5 минут под RWMutex — чтобы эндпоинт `/filters/sizes`, дёргаемый при каждом заходе в поиск, не долбил БД.

**Авторизация JWT без Redis.** Токен живёт 30 дней, содержит `user_id`, `role`, `exp`. Подписан секретом `JWT_SECRET`. Сервер полностью stateless — инвалидация токена возможна только истечением `exp`. Администратор создаётся только вручную через БД (смена поля `role`).

**Фото и MinIO.** URL в БД хранятся с внутренним хостом (`http://minio:9000/...`), но в ответах API репозиторий переписывает их в относительный путь `/minio/...` (функция `rewritePhotoURL` в `repository/photo_url.go`). На стороне nginx/Caddy `/minio/*` проксируется на контейнер MinIO. Это позволяет менять схему доставки фото, не трогая БД.

**Роутинг** — стандартный `http.ServeMux` без фреймворков. Path-параметры (UUID в URL) извлекаются вручную через `strings.TrimPrefix`. Дополнительный маршрутизатор `adminRouter` в `main.go` диспатчит запросы по суффиксу пути (`/visibility`, `/text`, `/photos/{photo_id}`).

---

## Архитектура монитор-сервиса

**Два независимых воркера**, запускаемых из одного образа по разным командам: `tg_monitor.py` (на Telethon, asyncio, event-driven) и `vk_monitor.py` (polling с интервалом в 5с).

**Общий слой:** `llm_parser.py` (обращение к Ollama, промпт на извлечение структуры), `db.py` (запись в Postgres), `storage.py` (загрузка в MinIO + установка public-read политики).

**Ключевая особенность парсинга.** LLM может вернуть **массив** товаров в одном посте (если продаются несколько). Для каждого товара создаётся отдельная запись в `listings` с общим `source_id`, `original_text`, `post_url`. Все фото поста копируются в каждую запись (с точки зрения `listing_photos` каждый товар имеет свою коллекцию фото, физически — одинаковые файлы в разных объектах MinIO).

**Фильтрация источников.** `.env` используется как **seed-набор**: при старте монитора все группы/каналы из конфига upsert-ятся в таблицу `sources`. Это гарантирует, что после чистой установки в админке сразу виден ожидаемый список. Но мониторим мы **весь активный набор из БД** (`SELECT external_id, title FROM sources WHERE platform=X AND is_active=TRUE`) — включая группы, **добавленные через админский UI** (`POST /admin/sources`), которых в `.env` может не быть. Это позволяет:
- включать/выключать источники «на лету» через UI (VK: сразу же, TG: в пределах 60с — фоновый refresh задачи);
- добавлять новые источники через UI без правки `.env` (VK — работает сразу; TG — работает только если аккаунт Telethon-сессии **уже состоит** в канале, Telethon принимает события только из чатов, куда юзер вступил).

---

## Архитектура фронтенда

**3-слойная структура:** `api/` → `contexts/` → `pages/` + `components/`.

- **api/** — единственная точка выхода в сеть. `client.js` сам подставляет JWT, при любой HTTP-ошибке вызывает глобальный коллбэк (который показывает тост). Методы с `silentError: true` не триггерят тост (напр. логин — ошибку 401 показываем inline в форме).
- **contexts/** — глобальные стейты: `AuthContext` хранит `user` и токен (в `localStorage` + в памяти), `ToastContext` — очередь уведомлений с авто-закрытием, `FiltersContext` — кэш метаданных для динамических фильтров (размеры из `/filters/sizes`, грузится один раз на старте).
- **components/** — «глупые» компоненты: UI + минимум логики. **pages/** — композиция: получает данные через API/контексты, собирает компоненты.
- **Роутинг:** `react-router-dom` v6. `ProtectedRoute` редиректит на `/auth` если не залогинен. `AdminRoute` ещё и на `/403` если роль не `admin`.

**Обработка ошибок.** При любом ответе не-2xx из API в правом верхнем углу появляется тост **«Ошибка. Не удалось выполнить запрос»** (длительность 5с). Сообщение от сервера (`data.message`) показывается вторичным текстом. Это позволяет пользователю понять, что запрос не удался, а разработчику — сразу открыть DevTools → Network. Исключения: логин/регистрация/профиль показывают inline-ошибку в форме (флаг `silentError`).

---

## Ключевые архитектурные решения

**Три сервиса на разных языках** — `api-service` (Go), `monitor-service` (Python), `frontend` (React/Node в билде, nginx в рантайме). Полностью независимы: у каждого своё окружение, свой `Dockerfile` и свой процесс. Общее между бэкенд-сервисами только одно — PostgreSQL. `monitor-service` пишет в таблицу `listings`, `api-service` из неё читает. Прямых вызовов между ними нет.

**Миграции вынесены в `backend/migrations/`** — они не принадлежат ни одному из сервисов, так как оба работают с одной схемой БД. Применяются один раз при первом старте PostgreSQL через механизм `/docker-entrypoint-initdb.d`.

**`docker-compose.yml` в корне** — поднимает `postgres`, `minio`, `ollama`, `api`, `frontend`, `tg-monitor`, `vk-monitor` (+ опциональный `caddy` для HTTPS) одной командой `docker compose up`. Модель `gemma3:4b` для Ollama скачивается один раз вручную командой `docker exec -it brandhunt_ollama ollama pull gemma3:4b`.

**Caddy для HTTPS.** Образ `caddy:2-alpine` слушает 80/443 и проксирует на `frontend:80` внутри docker-сети. На `.localhost`-домене использует self-signed cert (для локального тестирования), на публичном домене — автоматически выпускает и обновляет Let's Encrypt. Конфиг — один `Caddyfile` в корне, домен берётся из переменной `DOMAIN` в `.env`. Переключение local ↔ prod — одной строкой.

**Авторизация Telegram** — файловая сессия. В Docker-контейнере хранится по пути `/app/sessions/telethon_session.session` в volume `tg_sessions`. Генерация один раз: `docker compose run --rm --no-deps tg-monitor python auth_tg.py` (интерактивный ввод телефона и SMS-кода).

**VK-состояние** (`/app/data/vk_state.json`) хранит ID последних обработанных постов по каждой группе, чтобы не парсить их повторно. Персистируется в Docker volume `vk_data`.

**Фото с one-origin доступом.** MinIO не торчит наружу как отдельный сервис на своём домене. Вместо этого nginx фронта и Caddy проксируют `/minio/*` → `minio:9000`, чтобы всё приложение работало на одном origin (без CORS) и чтобы структура хранения была скрыта от клиента. Бакет `brandhunt-photos` публичен на чтение (политика ставится автоматически в `storage.py`).

---

## Документация

| Файл | Содержимое |
|---|---|
| [`README.md`](README.md) | Инструкция по запуску проекта (локально и на проде) |
| [`claude-instructions/database.md`](claude-instructions/database.md) | Полная схема БД: все таблицы, типы данных, связи, обоснование решений |
| [`claude-instructions/spec.md`](claude-instructions/spec.md) | Спецификация API: все методы, параметры, тела запросов/ответов, алгоритмы |
| [`backend/docs/index.adoc`](backend/docs/index.adoc) | Детальная AsciiDoc-документация по каждому эндпоинту (22 метода) |
