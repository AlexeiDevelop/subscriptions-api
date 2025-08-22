# Subscriptions API (Go + PostgreSQL)

Учебный, но максимально приближенный к продакшену REST‑сервис учёта онлайн‑подписок.

## Быстрый старт (локально)

1) Установи Go 1.22+.
2) Создай `.env` по образцу из `.env.example` **или** поправь `configs/config.yaml`.
3) Создай БД `subscriptions` в PostgreSQL (или используй docker-compose позже).
4) Примени миграции (через golang-migrate CLI) или создай таблицу руками из `migrations/001_create_subscriptions.up.sql`.
5) Запусти сервер:
```bash
go run ./cmd/server
```

Сервер поднимется на `http://localhost:8080`.

## Эндпоинты

- `POST /subscriptions` — создать подписку
- `GET  /subscriptions/{id}` — получить по ID
- `PUT  /subscriptions/{id}` — обновить
- `DELETE /subscriptions/{id}` — удалить
- `GET  /subscriptions` — список (фильтры: user_id, service_name; пагинация: limit, offset)
- `GET  /subscriptions/summary` — сумма за период по месяцам (`from=MM-YYYY&to=MM-YYYY`, фильтры: user_id, service_name)

## Конфигурация

- По умолчанию читается `configs/config.yaml`.
- Переменные окружения перекрывают YAML (pref: `APP_`), см. `.env.example`.

## Миграции

CLI: https://github.com/golang-migrate/migrate

```bash
migrate -database "postgres://user:password@localhost:5432/subscriptions?sslmode=disable" -path ./migrations up
```

## Swagger (опционально, позже)

Добавлены комментарии‑заготовки. Сгенерировать доки можно через `swaggo/swag`:

```bash
go install github.com/swaggo/swag/cmd/swag@latest
swag init -g cmd/server/main.go
```

## Docker (позже)

В проекте есть `Dockerfile` и `docker-compose.yml`, можно запустить всё одной командой:

```bash
docker compose up --build
```

## Структура
```
cmd/server/main.go
internal/
  config/config.go
  handler/subscription.go
  model/subscription.go
  storage/postgres.go
  storage/repository.go
configs/config.yaml
migrations/001_create_subscriptions.up.sql
migrations/001_create_subscriptions.down.sql
.env.example
Dockerfile
docker-compose.yml
Makefile
```
