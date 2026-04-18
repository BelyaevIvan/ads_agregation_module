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
brandhunt/
│
├── CLAUDE.md
├── docker-compose.yml              # поднимает весь стек одной командой
├── .env.example                    # шаблон переменных окружения
├── .env
│
├── claude-instructions/            # контекст для Claude
│   ├── database.md                 # схема БД
│   ├── spec.md                     # спецификация API (все методы)
│   └── ...                         # другие файлы по мере появления
│
├── backend/
│   │
│   ├── docs/                       # документация API (AsciiDoc)
│   │   └── {endpoint-name}/        # одна папка на один метод
│   │       ├── {name}.adoc
│   │       ├── {name}-request.adoc
│   │       ├── {name}-response.adoc
│   │       ├── {name}-errors.adoc
│   │       └── {name}.puml
│   │
│   ├── migrations/                 # SQL-миграции (общие для обоих сервисов)
│   │   ├── 001_init.sql
│   │   └── ...
│   │
│   ├── api-service/                # основной сервис (Go, без фреймворков)
│   │   ├── cmd/
│   │   │   └── main.go             # точка входа
│   │   ├── internal/
│   │   │   ├── handler/            # HTTP-хэндлеры (приём/отправка запросов)
│   │   │   ├── service/            # бизнес-логика
│   │   │   ├── repository/         # работа с PostgreSQL
│   │   │   ├── model/              # структуры данных (entities)
│   │   │   └── middleware/         # JWT, CORS, логирование, обработчик ошибок
│   │   │       ├── auth.go         # проверка JWT, извлечение user_id и role
│   │   │       ├── cors.go         # CORS-заголовки
│   │   │       └── error.go        # универсальный обработчик ошибок
│   │   ├── config/
│   │   │   └── config.go
│   │   ├── Dockerfile
│   │   └── go.mod
│   │
│   └── monitor-service/            # сервис мониторинга TG/VK (Python)
│       ├── tg_monitor.py           # воркер Telegram (asyncio + Telethon)
│       ├── vk_monitor.py           # воркер ВКонтакте (polling)
│       ├── llm_parser.py           # парсинг через Ollama (gemma3:4b)
│       ├── db.py                   # запись в PostgreSQL
│       ├── storage.py              # загрузка фото в MinIO
│       ├── config.py               # конфиг из env-переменных
│       ├── auth_tg.py              # генерация строковой сессии Telegram
│       ├── Dockerfile
│       └── requirements.txt
│
└── frontend/                       # SPA на React + Vite
    ├── Dockerfile                  # multi-stage: node builder → nginx
    ├── nginx.conf                  # proxy /api → api:8080, SPA fallback
    ├── package.json
    ├── vite.config.js
    ├── index.html
    └── src/
        ├── main.jsx                # точка входа, монтирует App в #root
        ├── App.jsx                 # корневой роутер (react-router)
        ├── api/                    # HTTP-клиент (слой доступа к бэку)
        │   ├── client.js           # fetch-обёртка: JWT, тост при ошибке
        │   ├── auth.js             # register/login/logout/me/updateMe
        │   ├── listings.js         # search, getById
        │   ├── favorites.js        # list/add/remove
        │   └── admin.js            # все админские методы
        ├── contexts/               # глобальное состояние через Context API
        │   ├── AuthContext.jsx     # user + токен в localStorage
        │   └── ToastContext.jsx    # очередь тостов (ошибки/успех)
        ├── components/             # переиспользуемые UI-блоки
        │   ├── common/             # Navbar, Toast, Spinner, Pagination,
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

### Архитектура фронтенда

**3-слойная структура:** `api/` → `contexts/` → `pages/` + `components/`.
- **api/** — единственная точка выхода в сеть. `client.js` сам подставляет JWT, при любой HTTP-ошибке вызывает глобальный коллбэк (который показывает тост). Методы с `silentError: true` не триггерят тост (напр. логин — ошибку 401 показываем inline в форме).
- **contexts/** — `AuthContext` хранит `user` и токен (в `localStorage` + в памяти), вызывает `authApi` под капотом. `ToastContext` — очередь уведомлений с авто-закрытием.
- **components/** — «глупые» компоненты: UI + минимум логики. **pages/** — композиция: получает данные через API/контексты, собирает компоненты.
- **Роутинг:** `react-router-dom` v6. `ProtectedRoute` редиректит на `/auth` если не залогинен. `AdminRoute` ещё и на `/403` если роль не `admin`.

### Обработка ошибок на фронте
При любом ответе не-2xx из API в правом верхнем углу появляется тост **«Ошибка. Не удалось выполнить запрос»** (длительность 5с). Сообщение от сервера (`data.message`) показывается вторичным текстом. Это позволяет пользователю понять, что запрос не удался, а разработчику — сразу открыть DevTools → Network. Исключения: логин/регистрация/профиль показывают inline-ошибку в форме (флаг `silentError`).
 
### Ключевые архитектурные решения
 
**Два сервиса на разных языках** — `api-service` (Go) и `monitor-service` (Python) полностью независимы: у каждого своё окружение, свой `Dockerfile` и свой процесс. Общее между ними только одно — PostgreSQL. `monitor-service` пишет в таблицы `listings`, `api-service` из неё читает. Прямых вызовов между сервисами нет.
 
**Миграции вынесены в `backend/migrations/`** — они не принадлежат ни одному из сервисов, так как оба работают с одной схемой БД. Применяются один раз при деплое или запуске.
 
**`docker-compose.yml` в корне** — поднимает PostgreSQL, MinIO, Ollama, `tg-monitor` и `vk-monitor` одной командой `docker-compose up`. Сервис `ollama-init` автоматически скачивает модель `gemma3:4b` при первом запуске.
 
**Авторизация** — JWT без Redis. Токен живёт 30 дней, содержит `user_id`, `role`, `exp`. Проверяется в `middleware/auth.go`. Администратор создаётся только вручную через БД (смена поля `role`).
 
**Универсальный обработчик ошибок** — реализован в `middleware/error.go`. При любой ошибке возвращает JSON:
```json
{
  "statusCode": 400,
  "url": "/api/v1/listings",
  "message": "описание ошибки",
  "date": "2026-04-17T14:32:00Z"
}
```
 
**Авторизация Telegram** — используется строковая сессия (`TG_SESSION_STRING` в `.env`). Для генерации: `docker compose run --rm --no-deps tg-monitor python auth_tg.py`.
 
**VK-состояние** (`/app/data/vk_state.json`) хранит ID последних обработанных постов по каждой группе. Персистируется в Docker volume `vk_data`.
 
---
 
## Документация
 
| Файл | Содержимое |
|---|---|
| [`claude-instructions/database.md`](claude-instructions/database.md) | Полная схема БД: все таблицы, типы данных, связи, обоснование решений |
| [`claude-instructions/spec.md`](claude-instructions/spec.md) | Спецификация API: все методы, параметры, тела запросов/ответов, алгоритмы |