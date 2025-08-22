# Subscriptions API — Makefile
# Локальная разработка + Docker Compose сценарии
# =========================

# --- Параметры/алиасы ---
APP_NAME            ?= subscriptions-api

# Docker Compose alias (V2 встроен в docker)
COMPOSE             ?= docker compose

# Имена сервисов в docker-compose.yml
DB_SERVICE          ?= db
APP_SERVICE         ?= app
MIGRATOR_SERVICE    ?= migrator

# DSN/путь для мигратора (должны совпадать с docker-compose.yml)
MIGRATE_PATH        ?= /migrations
MIGRATE_DSN         ?= postgres://postgres:postgres@db:5432/subscriptions?sslmode=disable

# Порт локального запуска приложения (go run)
PORT                ?= 8080

# --- Цель по умолчанию ---
.DEFAULT_GOAL := help

# --- Хелп ---
.PHONY: help
help: ## Показать доступные команды
	@echo ""
	@echo "Make targets for $(APP_NAME):"
	@echo ""
	@echo "  Локальная разработка:"
	@echo "    make tidy              - go mod tidy"
	@echo "    make fmt               - go fmt ./..."
	@echo "    make vet               - go vet ./..."
	@echo "    make build             - локальная сборка бинарника в ./bin/server"
	@echo "    make run               - локальный запуск (APP_PORT=$(PORT))"
	@echo "    make test              - go test ./... -race"
	@echo ""
	@echo "  Docker Compose:"
	@echo "    make dc-up-db          - поднять только БД (в фоне)"
	@echo "    make dc-logs-db        - смотреть логи БД"
	@echo "    make dc-ps             - статус контейнеров проекта"
	@echo "    make dc-migrate        - применить миграции (up)"
	@echo "    make dc-migrate-down-1 - откатить одну миграцию (down 1)"
	@echo "    make dc-migrate-version-raw - показать версию миграций (через migrate CLI)"
	@echo "    make dc-migrate-force-1 - принудительно выставить версию 1 (сброс dirty)"
	@echo "    make dc-up-app         - собрать образ и поднять приложение (в фоне)"
	@echo "    make dc-logs           - смотреть логи приложения"
	@echo "    make dc-restart-app    - перезапустить только приложение"
	@echo "    make dc-stop-app       - остановить только приложение"
	@echo "    make dc-down           - остановить все сервисы (без удаления томов)"
	@echo "    make dc-down-v         - остановить все и удалить тома (ВНИМАНИЕ: удалит данные!)"
	@echo ""
	@echo "  Утилиты БД:"
	@echo "    make dc-psql           - интерактивный psql к БД"
	@echo "    make db-backup         - создать SQL-дамп в ./backup.sql"
	@echo "    make db-restore        - восстановить из ./backup.sql"
	@echo ""
	@echo "  Swagger документация:"
	@echo "    make swagger           - создаст swagger документацию"

# --- Локальная разработка (без Docker) ---
.PHONY: tidy fmt vet build run test
tidy:  ## go mod tidy
	go mod tidy

fmt:   ## go fmt ./...
	go fmt ./...

vet:   ## go vet ./...
	go vet ./...

build: ## локальная сборка бинарника в ./bin/server
	mkdir -p bin
	go build -o bin/server ./cmd/server

run:   ## локальный запуск (читает configs/config.yaml и ENV)
	APP_ENV=dev APP_PORT=$(PORT) go run ./cmd/server

test:  ## запуск тестов
	go test ./... -race -count=1

# --- Docker Compose сценарии ---
.PHONY: dc-up-db dc-logs-db dc-ps dc-migrate dc-migrate-down-1 dc-migrate-version-raw dc-migrate-force-1 dc-up-app dc-logs dc-restart-app dc-stop-app dc-down dc-down-v
dc-up-db: ## поднять только БД (в фоне)
	$(COMPOSE) up -d $(DB_SERVICE)

dc-logs-db: ## логи БД
	$(COMPOSE) logs -f $(DB_SERVICE)

dc-ps: ## статус контейнеров проекта
	$(COMPOSE) ps

# У мигратора в compose по умолчанию 'command: up'
dc-migrate: ## применить миграции (up)
	$(COMPOSE) run --rm $(MIGRATOR_SERVICE)

dc-migrate-down-1: ## откатить одну миграцию (down 1)
	$(COMPOSE) run --rm --entrypoint "" $(MIGRATOR_SERVICE) \
	migrate -path=$(MIGRATE_PATH) -database=$(MIGRATE_DSN) down 1

dc-migrate-version-raw: ## показать текущую версию миграций
	$(COMPOSE) run --rm --entrypoint "" $(MIGRATOR_SERVICE) \
	migrate -path=$(MIGRATE_PATH) -database=$(MIGRATE_DSN) version || true

dc-migrate-force-1: ## принудительно выставить версию 1 (сброс dirty)
	$(COMPOSE) run --rm --entrypoint "" $(MIGRATOR_SERVICE) \
	migrate -path=$(MIGRATE_PATH) -database=$(MIGRATE_DSN) force 1

dc-up-app: ## собрать образ и поднять приложение (в фоне)
	$(COMPOSE) up --build -d $(APP_SERVICE)

dc-logs: ## логи приложения
	$(COMPOSE) logs -f $(APP_SERVICE)

dc-restart-app: ## перезапустить приложение
	$(COMPOSE) restart $(APP_SERVICE)

dc-stop-app: ## остановить приложение
	$(COMPOSE) stop $(APP_SERVICE)

dc-down: ## остановить все сервисы (оставить тома)
	$(COMPOSE) down

dc-down-v: ## остановить все сервисы и удалить тома (ОПАСНО: удалит данные БД!)
	$(COMPOSE) down -v

# --- Утилиты БД --- 
.PHONY: dc-psql db-backup db-restore
dc-psql: ## интерактивный psql в контейнере БД
	$(COMPOSE) exec $(DB_SERVICE) psql -U postgres -d subscriptions

db-backup: ## создать SQL-дамп в ./backup.sql
	$(COMPOSE) exec $(DB_SERVICE) pg_dump -U postgres -d subscriptions > backup.sql
	@echo "Дамп сохранён в ./backup.sql"

db-restore: ## восстановить из ./backup.sql
	test -f backup.sql
	$(COMPOSE) exec -T $(DB_SERVICE) psql -U postgres -d subscriptions < backup.sql
	@echo "Дамп из ./backup.sql восстановлен"

.PHONY: swagger
swagger:
	swag init -g cmd/server/main.go -o docs