# Subscriptions API

REST‑сервис для агрегации данных об онлайн‑подписках пользователей.

**Сделано по ТЗ:** CRUDL, суммирование за период с фильтрами, PostgreSQL + миграции, конфиг через YAML/ENV, логирование, Swagger UI, запуск через Docker Compose.

> Требования: Docker Desktop (Compose V2). Для локального запуска — Go 1.23. Swagger-спека уже сгенерирована и лежит в `docs/`.

---

## 🚀 Быстрый старт (Docker)

```bash
# 1) Поднять БД
docker compose up -d db

# 2) Применить миграции
docker compose run --rm migrator

# 3) Собрать и запустить приложение
docker compose up --build -d app

# 4) Проверить
curl -s http://localhost:8080/subscriptions || true
```
### Swagger UI
Открыть: <http://localhost:8080/swagger/index.html>

Проверка JSON-спеки:
```bash
curl -s http://localhost:8080/swagger/doc.json | head
```

---

## 🧰 Makefile (удобные цели)

```bash
# docker
make dc-up-db             # поднять БД
make dc-migrate           # применить миграции
make dc-up-app            # собрать и поднять приложение
make dc-logs              # логи приложения (Ctrl+C для выхода)
make dc-ps                # статус контейнеров
make dc-down              # остановить всё (данные сохранятся)
make dc-down-v            # остановить всё и удалить тома (данные БД будут удалены!)

# локально
make tidy fmt vet build   # go инструменты
make run                  # локальный запуск (порт 8080)
make test                 # тесты (плейсхолдер)
make swagger              # пересобрать swagger (если менялись аннотации)
```

---

## ⚙️ Конфигурация

Конфиг читается из `configs/config.yaml` + ENV (Viper). В Compose переменные заданы в `docker-compose.yml`:

```yaml
APP_ENV: dev
APP_PORT: 8080
APP_DB_HOST: db
APP_DB_PORT: 5432
APP_DB_USER: postgres
APP_DB_PASSWORD: postgres
APP_DB_NAME: subscriptions
APP_DB_SSLMODE: disable
```

DSN:
```
postgres://postgres:postgres@db:5432/subscriptions?sslmode=disable
```

---

## 🗃️ Миграции

Файлы в `migrations/`. Применение:

```bash
docker compose run --rm migrator   # up
docker compose run --rm --entrypoint "" migrator   migrate -path=/migrations -database=postgres://postgres:postgres@db:5432/subscriptions?sslmode=disable version
docker compose run --rm --entrypoint "" migrator   migrate -path=/migrations -database=postgres://postgres:postgres@db:5432/subscriptions?sslmode=disable down 1
```

Проверка состояния:
```bash
docker compose exec db psql -U postgres -d subscriptions -c "SELECT version, dirty FROM schema_migrations;"
```

---

## 🔌 API кратко

База пути: `/`

- `POST /subscriptions` — создать
- `GET /subscriptions/{id}` — получить по ID
- `PUT /subscriptions/{id}` — обновить
- `DELETE /subscriptions/{id}` — удалить
- `GET /subscriptions` — список (фильтры: `user_id`, `service_name`, пагинация: `limit`, `offset`)
- `GET /subscriptions/summary?from=MM-YYYY&to=MM-YYYY&user_id=&service_name=` — суммирование стоимости за период

Примеры:
```bash
# Создать
curl -X POST http://localhost:8080/subscriptions   -H "Content-Type: application/json"   -d '{"service_name":"Netflix","price":999,"user_id":"60601fee-2bf1-4721-ae6f-7636e79a0cba","start_date":"07-2025"}'

# Сумма за период
curl "http://localhost:8080/subscriptions/summary?from=07-2025&to=10-2025&user_id=60601fee-2bf1-4721-ae6f-7636e79a0cba&service_name=Netflix"
```

---

## 🗂️ Структура проекта

```
cmd/server/             # main.go — точка входа
internal/
  config/               # Viper + конфиг YAML/ENV
  handler/              # HTTP-ручки (chi)
  model/                # доменные модели и payload
  storage/              # Postgres (pgxpool), репозиторий
migrations/             # SQL миграции
docs/                   # Swagger (сгенерированные файлы)
configs/                # config.yaml
docker-compose.yml
Dockerfile
Makefile
```

---

## 🔧 Технологии

- Go 1.23, `net/http` + `chi`
- PostgreSQL 15 (Docker), `pgx/v5`
- Миграции: `migrate/migrate`
- Конфиг: YAML + ENV (Viper)
- Логи: `log/slog`
- Swagger: `swaggo/swag` + `http-swagger`

---

## ✅ Траблшутинг

- **`connection refused localhost:5432`** — БД не запущена → `make dc-up-db` и дождитесь `healthy`.
- **`relation "subscriptions" does not exist`** — не применены миграции → `make dc-migrate`.
- **Swagger 404** — перегенерируйте `docs/` (`make swagger`) и проверьте маршрут `/swagger/*` в `main.go`.

---

## 🗺️ Roadmap (v1.1)

- Оборачиваемые ошибки по всему проекту (`fmt.Errorf("%s: %w", ...)`) + единый маппинг в хендлерах.
- Больше структурных логов (детали запроса и поле‑по‑полю ошибки валидации).
- Интеграционные тесты и GitHub Actions.
- Улучшённая пагинация/сортировка.
