# Переменные для базы данных (обновлены под ваш docker-compose)
DB_HOST ?= localhost
DB_PORT ?= 5442
DB_USER ?= user
DB_PASSWORD ?= password
DB_NAME ?= db
DB_URL = postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable

# Переменные для миграций
MIGRATIONS_PATH = ./migrations
MIGRATION_NAME ?= migration

.PHONY: help
help: ## Показать справку
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

.PHONY: db-up
db-up: ## Запустить базу данных
	docker-compose up -d postgres

.PHONY: db-down
db-down: ## Остановить базу данных
	docker-compose down

.PHONY: migrate-create
migrate-create: ## Создать новую миграцию
	docker run --rm -v $(PWD)/migrations:/migrations migrate/migrate:v4.18.1 \
		create -ext sql -dir /migrations -seq $(MIGRATION_NAME)

.PHONY: migrate-up
migrate-up: ## Применить все миграции
	docker run --rm -v $(PWD)/migrations:/migrations --network host migrate/migrate:v4.18.1 \
		-path=/migrations -database="$(DB_URL)" up

.PHONY: migrate-down
migrate-down: ## Откатить одну миграцию
	docker run --rm -v $(PWD)/migrations:/migrations --network host migrate/migrate:v4.18.1 \
		-path=/migrations -database="$(DB_URL)" down 1

.PHONY: migrate-status
migrate-status: ## Показать статус миграций
	docker run --rm -v $(PWD)/migrations:/migrations --network host migrate/migrate:v4.18.1 \
		-path=/migrations -database="$(DB_URL)" version

.PHONY: build
build: ## Собрать приложение
	go build -o bin/fixdrive cmd/main.go

.PHONY: run
run: ## Запустить приложение
	go run cmd/main.go