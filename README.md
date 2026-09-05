# FRIDAY

Your personal information system. It knows what matters to you, watches the internet and data sources, and tells you what you need to know.

Глобальная персональная площадка: пользователь один раз задаёт, кто он и что ему важно. Система сама решает, какие слои и данные смотреть, и кладёт готовую картину на стол. Карьера, новости, календарь, знания, развлечения — равные слои. Какие загорятся, решает профиль.

- План продукта: [docs/PLAN.md](docs/PLAN.md)
- Задачи Phase 0: [docs/TASKS.md](docs/TASKS.md)
- API-конвенции: [docs/api-conventions.md](docs/api-conventions.md)
- OpenAPI: [docs/openapi.yaml](docs/openapi.yaml)

## Локальный запуск

Требования: Go 1.26+, Node 22+/24+, Docker, `golangci-lint` (для `make lint`).

```bash
cp .env.example .env
# set AUTH_JWT_SECRET to a long random value before first run
make env-up
make env-port-forward
make migrate-up   # includes users + refresh_tokens
make dev          # API (air) + Vue (vite)
```

- API: `http://localhost:8080` — `GET /api/v1/health`, auth under `/api/v1/auth/*`, `GET /api/v1/me`
- Web: Vite dev-server (proxy `/api` → `:8080`); open `/register` or `/login`
- Postgres на хосте: `127.0.0.1:5433` (внутри сети Compose — `friday-postgres:5432`; `5432` часто занят другими проектами)
- Redis на хосте: `127.0.0.1:6379`

Отдельно:

| Цель | Что делает |
|------|------------|
| `make run` | только backend с hot-reload |
| `make run-web` | только frontend |
| `make test` | `go test` + vitest |
| `make lint` | golangci-lint + eslint/oxlint |
| `make env-down` | остановить Postgres и Redis |
| `make env-cleanup` | снести volumes (`out/pgdata`, `out/redisdata`) |

## Структура

```
cmd/friday/          # точка входа API
internal/core/       # logger, postgres pool, HTTP server
internal/features/   # вертикальные слайсы (health, …)
migrations/          # golang-migrate
web/                 # Vue 3 + Vite + Pinia + Tailwind
docs/                # план, задачи, OpenAPI
```
