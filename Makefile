include .env
export

export PROJECT_ROOT=$(shell pwd)
GOEXE := $(shell go env GOEXE)

env-up:
	@docker compose up -d friday-postgres friday-redis

env-down:
	@docker compose down friday-postgres friday-redis

env-cleanup:
	@read -p "Очистить все volume файлы окружения? Опасность утери данных! [y/N]: " ans; \
	if [ "$$ans" = "y" ]; then \
		docker compose down friday-postgres friday-redis && \
		rm -rf out/pgdata out/redisdata && \
		echo "Файлы окружения очищены"; \
		else \
			echo "Очистка окружения отменена"; \
	fi

env-port-forward:
	@docker compose up -d port-forwarder-postgres port-forwarder-redis

env-port-close:
	@docker compose down port-forwarder-postgres port-forwarder-redis

migrate-create:
	@if [ -z "$(seq)" ]; then \
		echo "Отсутствует необходимый параметр seq. Пример: make migrate-create seq=init"; \
		exit 1; \
	fi; \
	MSYS_NO_PATHCONV=1 docker compose run --rm friday-postgres-migrate \
	create \
	-ext sql \
	-dir /migrations \
	-seq "$(seq)"

migrate-up:
	@make migrate-action action=up

migrate-down:
	@make migrate-action action=down

migrate-action:
	@if [ -z "$(action)" ]; then \
		echo "Отсутствует необходимый параметр action. Пример: make migrate-action action=up"; \
		exit 1; \
	fi; \
	MSYS_NO_PATHCONV=1 docker compose run --rm friday-postgres-migrate \
		-path /migrations \
		-database postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@friday-postgres:5432/${POSTGRES_DB}?sslmode=disable \
		"$(action)"

run:
	@mkdir -p out/bin && \
	export POSTGRES_HOST=localhost && \
	export REDIS_HOST=localhost && \
	go tool air \
		-build.cmd "go build -o ./out/bin/friday$(GOEXE) ./cmd/friday" \
		-build.bin "./out/bin/friday$(GOEXE)"

run-web:
	@cd web && npm run dev

dev:
	@$(MAKE) -j2 run run-web

test:
	@go test $$(go list ./... | grep -v '/web/')
	@cd web && npm run test

lint:
	@golangci-lint run
	@cd web && npm run lint
