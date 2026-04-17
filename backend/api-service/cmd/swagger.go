package main

import (
	"net/http"
)

func serveSwaggerJSON(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write([]byte(swaggerSpec))
}

func serveSwaggerUI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(swaggerHTML))
}

const swaggerHTML = `<!DOCTYPE html>
<html lang="ru">
<head>
  <meta charset="UTF-8">
  <title>BrandHunt API — Swagger UI</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css">
  <style>body { margin: 0; }</style>
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
  <script>
    const ui = SwaggerUIBundle({
      url: '/api/v1/swagger.json',
      dom_id: '#swagger-ui',
      presets: [SwaggerUIBundle.presets.apis, SwaggerUIBundle.SwaggerUIStandalonePreset],
      layout: 'BaseLayout',
      responseInterceptor: function(response) {
        if (response.url && response.url.includes('/auth/login') && response.ok) {
          try {
            var body = typeof response.body === 'string' ? JSON.parse(response.body) : response.body;
            if (body && body.access_token) {
              ui.preauthorizeApiKey('BearerAuth', body.access_token);
              console.log('Token auto-applied from /auth/login');
            }
          } catch(e) {}
        }
        return response;
      }
    });
  </script>
</body>
</html>`

const swaggerSpec = `{
  "openapi": "3.0.3",
  "info": {
    "title": "BrandHunt API",
    "description": "API агрегатора объявлений о брендовых товарах",
    "version": "1.0.0"
  },
  "servers": [{ "url": "/api/v1" }],
  "components": {
    "securitySchemes": {
      "BearerAuth": {
        "type": "http",
        "scheme": "bearer",
        "bearerFormat": "JWT"
      }
    },
    "schemas": {
      "Error": {
        "type": "object",
        "properties": {
          "statusCode": { "type": "integer" },
          "url": { "type": "string" },
          "message": { "type": "string" },
          "date": { "type": "string", "format": "date-time" }
        }
      },
      "Paginated": {
        "type": "object",
        "properties": {
          "items": { "type": "array", "items": {} },
          "total": { "type": "integer" },
          "limit": { "type": "integer" },
          "offset": { "type": "integer" }
        }
      }
    }
  },
  "paths": {
    "/auth/register": {
      "post": {
        "tags": ["Auth"],
        "summary": "Регистрация",
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {
                "type": "object",
                "required": ["email", "password"],
                "properties": {
                  "email": { "type": "string", "format": "email", "example": "user@example.com" },
                  "password": { "type": "string", "minLength": 8, "example": "strongpassword123" }
                }
              }
            }
          }
        },
        "responses": {
          "201": { "description": "Пользователь создан" },
          "400": { "description": "Ошибка валидации" },
          "409": { "description": "Email уже занят" }
        }
      }
    },
    "/auth/login": {
      "post": {
        "tags": ["Auth"],
        "summary": "Вход, получение JWT",
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {
                "type": "object",
                "required": ["email", "password"],
                "properties": {
                  "email": { "type": "string", "example": "user@example.com" },
                  "password": { "type": "string", "example": "strongpassword123" }
                }
              }
            }
          }
        },
        "responses": {
          "200": { "description": "Токен получен" },
          "401": { "description": "Неверные учётные данные" }
        }
      }
    },
    "/auth/logout": {
      "post": {
        "tags": ["Auth"],
        "summary": "Выход",
        "security": [{ "BearerAuth": [] }],
        "responses": {
          "200": { "description": "Выход выполнен" },
          "401": { "description": "Не авторизован" }
        }
      }
    },
    "/listings": {
      "get": {
        "tags": ["Listings"],
        "summary": "Поиск и фильтрация объявлений",
        "parameters": [
          { "name": "q", "in": "query", "schema": { "type": "string" }, "description": "Полнотекстовый поиск" },
          { "name": "brand", "in": "query", "schema": { "type": "array", "items": { "type": "string" } }, "description": "Фильтр по бренду" },
          { "name": "category", "in": "query", "schema": { "type": "array", "items": { "type": "string" } }, "description": "Фильтр по категории" },
          { "name": "city", "in": "query", "schema": { "type": "array", "items": { "type": "string" } }, "description": "Фильтр по городу" },
          { "name": "condition", "in": "query", "schema": { "type": "string", "enum": ["new", "used"] } },
          { "name": "size_rus", "in": "query", "schema": { "type": "array", "items": { "type": "string" } } },
          { "name": "size_eu", "in": "query", "schema": { "type": "array", "items": { "type": "string" } } },
          { "name": "size_us", "in": "query", "schema": { "type": "array", "items": { "type": "string" } } },
          { "name": "price_min", "in": "query", "schema": { "type": "number" } },
          { "name": "price_max", "in": "query", "schema": { "type": "number" } },
          { "name": "include_no_size", "in": "query", "schema": { "type": "boolean", "default": true } },
          { "name": "include_no_price", "in": "query", "schema": { "type": "boolean", "default": true } },
          { "name": "include_no_city", "in": "query", "schema": { "type": "boolean", "default": true } },
          { "name": "platform", "in": "query", "schema": { "type": "array", "items": { "type": "string", "enum": ["telegram", "vk"] } } },
          { "name": "sort", "in": "query", "schema": { "type": "string", "enum": ["date_desc", "price_asc", "price_desc"], "default": "date_desc" } },
          { "name": "limit", "in": "query", "schema": { "type": "integer", "default": 20, "maximum": 100 } },
          { "name": "offset", "in": "query", "schema": { "type": "integer", "default": 0 } }
        ],
        "responses": {
          "200": { "description": "Список объявлений с пагинацией" },
          "400": { "description": "Ошибка валидации параметров" }
        }
      }
    },
    "/listings/{id}": {
      "get": {
        "tags": ["Listings"],
        "summary": "Карточка объявления",
        "parameters": [
          { "name": "id", "in": "path", "required": true, "schema": { "type": "string", "format": "uuid" } }
        ],
        "responses": {
          "200": { "description": "Полная карточка объявления" },
          "400": { "description": "Некорректный UUID" },
          "404": { "description": "Объявление не найдено" }
        }
      }
    },
    "/users/me": {
      "get": {
        "tags": ["User"],
        "summary": "Профиль текущего пользователя",
        "security": [{ "BearerAuth": [] }],
        "responses": {
          "200": { "description": "Данные профиля" },
          "401": { "description": "Не авторизован" }
        }
      },
      "put": {
        "tags": ["User"],
        "summary": "Обновить профиль",
        "security": [{ "BearerAuth": [] }],
        "requestBody": {
          "content": {
            "application/json": {
              "schema": {
                "type": "object",
                "properties": {
                  "full_name": { "type": "string" },
                  "phone": { "type": "string" },
                  "tg_link": { "type": "string" },
                  "vk_link": { "type": "string" }
                }
              }
            }
          }
        },
        "responses": {
          "200": { "description": "Профиль обновлён" },
          "401": { "description": "Не авторизован" }
        }
      }
    },
    "/users/me/favorites": {
      "get": {
        "tags": ["Favorites"],
        "summary": "Список избранного",
        "security": [{ "BearerAuth": [] }],
        "parameters": [
          { "name": "limit", "in": "query", "schema": { "type": "integer", "default": 20 } },
          { "name": "offset", "in": "query", "schema": { "type": "integer", "default": 0 } }
        ],
        "responses": {
          "200": { "description": "Список избранных объявлений" },
          "401": { "description": "Не авторизован" }
        }
      },
      "post": {
        "tags": ["Favorites"],
        "summary": "Добавить в избранное",
        "security": [{ "BearerAuth": [] }],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {
                "type": "object",
                "required": ["listing_id"],
                "properties": {
                  "listing_id": { "type": "string", "format": "uuid" }
                }
              }
            }
          }
        },
        "responses": {
          "201": { "description": "Добавлено" },
          "200": { "description": "Уже было в избранном" },
          "404": { "description": "Объявление не найдено" }
        }
      }
    },
    "/users/me/favorites/{listing_id}": {
      "delete": {
        "tags": ["Favorites"],
        "summary": "Убрать из избранного",
        "security": [{ "BearerAuth": [] }],
        "parameters": [
          { "name": "listing_id", "in": "path", "required": true, "schema": { "type": "string", "format": "uuid" } }
        ],
        "responses": {
          "204": { "description": "Удалено" },
          "404": { "description": "Не найдено в избранном" }
        }
      }
    },
    "/admin/listings": {
      "get": {
        "tags": ["Admin"],
        "summary": "Все объявления (включая скрытые)",
        "security": [{ "BearerAuth": [] }],
        "parameters": [
          { "name": "q", "in": "query", "schema": { "type": "string" } },
          { "name": "status", "in": "query", "schema": { "type": "string", "enum": ["active", "hidden", ""] } },
          { "name": "platform", "in": "query", "schema": { "type": "array", "items": { "type": "string" } } },
          { "name": "sort", "in": "query", "schema": { "type": "string", "enum": ["date_desc", "price_asc", "price_desc"] } },
          { "name": "limit", "in": "query", "schema": { "type": "integer", "default": 20 } },
          { "name": "offset", "in": "query", "schema": { "type": "integer", "default": 0 } }
        ],
        "responses": {
          "200": { "description": "Список объявлений" },
          "403": { "description": "Доступ запрещён" }
        }
      }
    },
    "/admin/listings/{id}/visibility": {
      "patch": {
        "tags": ["Admin"],
        "summary": "Скрыть / восстановить объявление",
        "security": [{ "BearerAuth": [] }],
        "parameters": [
          { "name": "id", "in": "path", "required": true, "schema": { "type": "string", "format": "uuid" } }
        ],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {
                "type": "object",
                "required": ["is_hidden"],
                "properties": {
                  "is_hidden": { "type": "boolean" }
                }
              }
            }
          }
        },
        "responses": {
          "200": { "description": "Видимость изменена" },
          "404": { "description": "Объявление не найдено" }
        }
      }
    },
    "/admin/listings/{id}/text": {
      "patch": {
        "tags": ["Admin"],
        "summary": "Редактировать текст объявления",
        "security": [{ "BearerAuth": [] }],
        "parameters": [
          { "name": "id", "in": "path", "required": true, "schema": { "type": "string", "format": "uuid" } }
        ],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {
                "type": "object",
                "required": ["original_text"],
                "properties": {
                  "original_text": { "type": "string" }
                }
              }
            }
          }
        },
        "responses": {
          "200": { "description": "Текст обновлён" },
          "404": { "description": "Объявление не найдено" }
        }
      }
    },
    "/admin/listings/{id}/photos/{photo_id}": {
      "delete": {
        "tags": ["Admin"],
        "summary": "Удалить фотографию",
        "security": [{ "BearerAuth": [] }],
        "parameters": [
          { "name": "id", "in": "path", "required": true, "schema": { "type": "string", "format": "uuid" } },
          { "name": "photo_id", "in": "path", "required": true, "schema": { "type": "string", "format": "uuid" } }
        ],
        "responses": {
          "200": { "description": "Фото удалено" },
          "404": { "description": "Фото не найдено" }
        }
      }
    },
    "/admin/stats": {
      "get": {
        "tags": ["Admin — Stats"],
        "summary": "Сводная статистика",
        "security": [{ "BearerAuth": [] }],
        "responses": {
          "200": { "description": "Статистика" }
        }
      }
    },
    "/admin/stats/listings-by-day": {
      "get": {
        "tags": ["Admin — Stats"],
        "summary": "Динамика объявлений по дням",
        "security": [{ "BearerAuth": [] }],
        "parameters": [
          { "name": "days", "in": "query", "schema": { "type": "integer", "enum": [7, 30, 90], "default": 30 } }
        ],
        "responses": {
          "200": { "description": "Данные по дням" }
        }
      }
    },
    "/admin/stats/top-brands": {
      "get": {
        "tags": ["Admin — Stats"],
        "summary": "Топ брендов",
        "security": [{ "BearerAuth": [] }],
        "parameters": [
          { "name": "limit", "in": "query", "schema": { "type": "integer", "default": 10 } }
        ],
        "responses": {
          "200": { "description": "Топ брендов" }
        }
      }
    },
    "/admin/stats/top-cities": {
      "get": {
        "tags": ["Admin — Stats"],
        "summary": "Топ городов",
        "security": [{ "BearerAuth": [] }],
        "parameters": [
          { "name": "limit", "in": "query", "schema": { "type": "integer", "default": 10 } }
        ],
        "responses": {
          "200": { "description": "Топ городов" }
        }
      }
    },
    "/admin/sources": {
      "get": {
        "tags": ["Admin — Sources"],
        "summary": "Список источников",
        "security": [{ "BearerAuth": [] }],
        "parameters": [
          { "name": "limit", "in": "query", "schema": { "type": "integer", "default": 50 } },
          { "name": "offset", "in": "query", "schema": { "type": "integer", "default": 0 } }
        ],
        "responses": {
          "200": { "description": "Список источников" }
        }
      },
      "post": {
        "tags": ["Admin — Sources"],
        "summary": "Добавить источник",
        "security": [{ "BearerAuth": [] }],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {
                "type": "object",
                "required": ["platform", "external_id"],
                "properties": {
                  "platform": { "type": "string", "enum": ["telegram", "vk"] },
                  "external_id": { "type": "string" },
                  "title": { "type": "string" }
                }
              }
            }
          }
        },
        "responses": {
          "201": { "description": "Источник создан" },
          "409": { "description": "Уже существует" }
        }
      }
    },
    "/admin/sources/{id}/toggle": {
      "patch": {
        "tags": ["Admin — Sources"],
        "summary": "Вкл/выкл мониторинг",
        "security": [{ "BearerAuth": [] }],
        "parameters": [
          { "name": "id", "in": "path", "required": true, "schema": { "type": "string", "format": "uuid" } }
        ],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {
                "type": "object",
                "required": ["is_active"],
                "properties": {
                  "is_active": { "type": "boolean" }
                }
              }
            }
          }
        },
        "responses": {
          "200": { "description": "Статус обновлён" },
          "404": { "description": "Источник не найден" }
        }
      }
    }
  }
}`
