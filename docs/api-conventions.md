# Конвенции HTTP API

Источник правды для формы ответов. Реализация: `internal/core/transport/http/response/handler.go`,
клиент — `web/src/api/client.ts`.

## Базовые правила

- Все маршруты живут под `/api/{version}`, сейчас `/api/v1` (`RegisterAPIRoutes("v1", ...)` в `cmd/friday/main.go`).
  Версия задаётся на уровне сервера — в хендлерах её не хардкодят.
- Тело — всегда JSON, поля в `snake_case` (`created_at`, `why_you`).
- Списки отдаются top-level массивом, без обёртки вида `{"data": [...]}`.
- Аутентификация (E1): JWT access + refresh, заголовок `Authorization: Bearer <access>`.
  Не cookie-only сессии — один контракт для веба и будущих мобильных клиентов.

## Когда возвращать тело

Тело ответа есть тогда и только тогда, когда оно сообщает клиенту то, чего он ещё не знает.
`{"id": N}` в ответ на запрос, где этот же `id` был в пути, телом не считается — это шум.

| Операция | Статус | Тело |
|----------|--------|------|
| Создание сущности | 201 | Полное представление + заголовок `Location` |
| Принято к фоновому выполнению | 202 | Представление задачи + заголовок `Location` |
| Чтение | 200 | Представление сущности или массив |
| Изменение с новым состоянием | 200 | Полное представление после изменения |
| Мутация без новой информации | 204 | Нет |

## Location

На 201 указывается канонический URL созданного ресурса. Путь строится от текущего запроса,
чтобы не дублировать префикс версии:

```go
location := path.Join(r.URL.Path, strconv.FormatInt(entity.ID, 10))
responseHandler.Created(location, dtoFromDomain(entity))
```

## Хелперы ответа

Хендлеры не пишут в `http.ResponseWriter` напрямую, а идут через `response.Handler`:

| Метод | Результат |
|-------|-----------|
| `OK(data)` | 200 + JSON |
| `Created(location, body)` | 201 + JSON, `Location` при непустой строке |
| `Accepted(location, body)` | 202 + JSON, `Location` при непустой строке |
| `NoContent()` | 204 без тела |
| `JSON(body, status)` | произвольный статус |
| `Error(err, msg)` | статус по типу ошибки + JSON ошибки |

## Ошибки

Форма: `{"message": "...", "error": "..."}`. Поле `message` — текст для пользователя,
`error` — техническая причина; на 5xx `error` не отдаётся, чтобы не светить внутренности наружу.

Статус выводится из sentinel-ошибок `internal/core/errors`:

| Ошибка | Статус |
|--------|--------|
| `ErrInvalidArgument` | 400 |
| `ErrUnauthorized` | 401 |
| `ErrForbidden` | 403 |
| `ErrNotFound` | 404 |
| `ErrConflict` | 409 |
| `ErrTimeout` | 504 |
| прочее | 500 |

Ошибки валидации запроса (`request.DecodeAndValidateRequest`) приходят как `ErrInvalidArgument` → 400.

## DTO

У каждой фичи свои DTO и мапперы в `transport/http/dto.go` и `mapper.go` — фичи не шарят DTO между
собой. Доменные структуры наружу не отдаются.

На фронте DTO живут в `web/src/api/*.ts` и мапятся в типы из `web/src/types/`.

## Текущая карта эндпоинтов

| Метод и путь | Доступ | Ответ |
|--------------|--------|-------|
| `GET /health` | публичный | 200 `{status}` |

Дальше по Phase 0: auth, profile, graph, radar — см. [TASKS.md](TASKS.md) и [openapi.yaml](openapi.yaml).
