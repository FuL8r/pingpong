# pingpong
Стандартный пингпонг сервис, получает на post запрос с параметром "NotBad"
Если параметр равен true(только в нижнем регистре), возвращает ReallyNotBad. При любом другом вводе возвращается forbidden

## Contract

Сервис слушает порт `8089` (go server) и реализует единственное правило:

```bash
curl -X POST -H "NotBad: true" https://localhost:8089/
# -> ReallyNotBad
```

| Запрос | Ответ |
|---------------------------|--------------------------------------|
| `POST /` + заголовок `NotBad: true` | `200`, тело `ReallyNotBad` |
| `POST /` без валидного заголовка | `403 Forbidden` (generic) |
| не-`POST` метод на `/` | `405`, заголовок `Allow: POST` |
| другой путь | `404` |
| тело больше лимита | `413` |
| паника внутри | `500` (без утечки деталей) |  string(name: 'CUSTOM_DOMAIN', defaultValue: '', description: 'Custom domain (empty = free *.awsapprunner.com)')
    string(name: 'HOSTED_ZONE_ID', defaultValue: '', description: 'Route53 hosted zone id (for custom domain DNS)')

go mod tidy
echo $?

## Start

Требуется Go, Terraform, Docker

## Конфигурация (terraform, 12-factor)

| Переменная | Дефолт | Назначение |
|---|---|---|
| `PINGPONG_ADDR` | `:8089` | адрес прослушивания |
| `TLS_CERT` / `TLS_KEY` | (пусто) | пути к сертификату/ключу; оба заданы → HTTPS |
| `PINGPONG_MAX_BODY_BYTES` | `4096` | лимит тела запроса |
| `PINGPONG_READ_HEADER_TIMEOUT` | `5s` | таймаут чтения заголовков |
| `PINGPONG_READ_TIMEOUT` | `10s` | таймаут чтения запроса |
| `PINGPONG_WRITE_TIMEOUT` | `10s` | таймаут записи ответа |
| `PINGPONG_IDLE_TIMEOUT` | `60s` | таймаут keep-alive |
| `PINGPONG_SHUTDOWN_TIMEOUT` | `10s` | таймаут graceful shutdown |
| `PINGPONG_MAX_INFLIGHT` | `256` | лимит одновременных запросов |
| `PINGPONG_LOG_LEVEL` | `info` | уровень логирования (`debug`/`info`/`warn`/`error`) |

## Конфигурация Jenkins
Требуется установить credentials для AWS user - aws-pingpong-ci

### 1. Локальный запуск через `make`
Используется для разработки, локального тестирования и отладки

* **Команда:** `make run`
* **Требования:** Наличие установленной утилиты `make`

### 2. Запуск через Jenkins + Terraform
Используется для автоматизации, сборки релизов и запуска на удаленных серверах

* **Где запускать:** Jenkins jobs
* **Параметры:** Запуск происходит вручную с указанием параметров ветки

## Comments
Порт не 8089, т.к. App Runner работает только на 443. Я не смог купить (не продали) домен AWS в Route 53, а полностью автоматизировать можно было с ним (чтобы не получать cert от внешних провайдеров, не менять записи на DNS серверах и т.д.)
Используется distroless контейнер, как максимально легкое решение
Сервер написан без внешних зависимостей (только стандартная библиотека)
