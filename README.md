# Simple Go Microservice + CI

Минимальный учебный проект: маленький Go-сервис и базовый CI в GitHub Actions.

## Что здесь происходит

1. Приложение поднимает HTTP-сервер на порту `8080` (или на `PORT`, если переменная окружения задана).
2. Есть базовый UI и API endpoint'ы:
- `GET /` - простая веб-страница для smoke-check ручек.
- `GET /healthz` - проверка, что сервис жив.
- `GET /hello?name=dev` - простой ответ `hello`.
3. В CI запускаются 4 независимые проверки:
- `fmt` (форматирование Go-кода),
- `lint` (`go vet`),
- `test` (тесты),
- `build` (сборка бинарника).

## Структура проекта

```text
.
├── .github/workflows/ci.yml      # pipeline для GitHub Actions
├── cmd/microservice/main.go      # HTTP-сервис
├── cmd/microservice/main_test.go # тесты для handler'ов
├── Makefile                      # локальные команды и команды для CI
└── go.mod
```

## Как запустить локально

Требование: Go `1.22+`.

```bash
make run
```

Сервис стартует на `http://localhost:8080`.

Проверка:

```bash
open http://localhost:8080/
curl http://localhost:8080/healthz
curl "http://localhost:8080/hello?name=Dave"
curl http://localhost:8080/hello
```

Пример ответа `/healthz`:

```json
{"status":"ok","time":"2026-02-08T18:00:00Z"}
```

Пример ответа `/hello?name=Dave`:

```json
{"message":"hello, Dave"}
```

## Команды Makefile

```bash
make fmt     # fail, если есть неотформатированные .go файлы
make format  # применить gofmt -w ко всем .go файлам
make lint    # go vet ./...
make test    # go test ./...
make build   # сборка в bin/microservice
make run     # запуск сервиса
make ci      # fmt + lint + test + build
```

`make ci` повторяет логику GitHub Actions локально.

## Как работает CI

Файл: `.github/workflows/ci.yml`.

Триггеры:
- `pull_request`,
- `push` в ветки `main` и `master`.

Джобы:
1. `fmt` -> `make fmt`
2. `lint` -> `make lint`
3. `test` -> `make test`
4. `build` -> `make build`

Каждая job:
1. Checkout кода (`actions/checkout@v4`)
2. Установка Go (`actions/setup-go@v5`)
3. Запуск соответствующей команды

Дополнительно:
- `permissions: contents: read` (минимальные права),
- `concurrency` с `cancel-in-progress: true` (отмена старых раннов на той же ветке).

## Какие проверки ставить в Branch Protection

Если включаешь обязательные проверки в GitHub, добавь:

- `fmt`
- `lint`
- `test`
- `build`

## Что дальше

Этот каркас готов для следующего этапа:
1. Добавить Dockerfile и публикацию образа.
2. Добавить `deploy.yml` (staging/prod) с environment approvals.
3. Подключить AWS (ECR/ECS/EKS) после того, как будут входные данные по инфраструктуре.
