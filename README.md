# Shortener

## Описание проекта

Shortener — это веб-сервис для сокращения длинных URL-адресов. Сервис позволяет создавать короткие ссылки из длинных URL и перенаправлять пользователей на оригинальные адреса.

### Основные возможности:

- **Создание коротких URL** через POST-запросы (поддержка текстового и JSON форматов)
- **Редирект на оригинальный URL** по короткой ссылке
- **Хранение данных** в PostgreSQL или файловой системе
- **Проверка соединения с базой данных** через эндпоинт `/ping`
- **Сжатие ответов** с помощью Gzip
- **Логирование запросов и ответов**

### Технологический стек:

- **Go 1.24** — язык программирования
- **Gin** — веб-фреймворк
- **PostgreSQL** — база данных
- **Zap** — структурированное логирование
- **Docker & Docker Compose** — контейнеризация

## Локальный запуск

### Предварительные требования

- [Docker](https://www.docker.com/get-started) и [Docker Compose](https://docs.docker.com/compose/install/)
- [Go](https://go.dev/dl/) версии 1.24 или выше (для локальной разработки)

### Запуск через Docker Compose

1. **Клонируйте репозиторий:**
   ```bash
   git clone <repository-url>
   cd shortener
   ```

2. **Запустите сервисы:**
   ```bash
   docker-compose up -d
   ```

   Эта команда запустит:
   - PostgreSQL базу данных на порту `5432`
   - API сервер на порту `8080`

3. **Проверьте работу сервиса:**
   ```bash
   curl http://localhost:8080/ping
   ```

4. **Остановка сервисов:**
   ```bash
   docker-compose down
   ```

### Установка зависимостей для локальной разработки

Если вы хотите запускать проект локально без Docker:

1. **Установите зависимости Go:**
   ```bash
   go mod download
   ```

2. **Убедитесь, что PostgreSQL запущен и доступен**, или настройте переменные окружения для использования файлового хранилища.

3. **Запустите приложение:**
   ```bash
   go run cmd/shortener/main.go
   ```

### Переменные окружения

При необходимости вы можете настроить следующие переменные окружения:

- `SERVER_ADDRESS` — адрес и порт сервера (по умолчанию: `:8080`)
- `BASE_URL` — базовый URL для генерации коротких ссылок (по умолчанию: `http://localhost:8080`)
- `DATABASE_DSN` — строка подключения к PostgreSQL (например: `postgres://user:password@localhost:5432/shortener`)
- `LOG_LEVEL` — уровень логирования (по умолчанию: `info`)
- `FILE_PATH` — путь к файлу для хранения данных, если не используется БД (по умолчанию: `/tmp/data.json`)

### Режим разработки и pprof

**Swagger-спека** (`GET /swagger/doc.json`) доступны **только в режиме разработки (DevMode)**. В проде эти маршруты не регистрируются.

Для актуальной спеки после изменения хендлеров выполните: `make swag` (нужен установленный `swag`: `go install github.com/swaggo/swag/cmd/swag@latest`).

Включить DevMode можно одним из способов:

- **Переменная окружения:** `APP_ENV=dev` или `APP_ENV=development`
- **Флаг:** `-dev` при запуске (например: `go run cmd/shortener/main.go -dev`)

Пример запуска с pprof:

```bash
APP_ENV=dev go run cmd/shortener/main.go
# или
go run cmd/shortener/main.go -dev
```

Снимок heap-профиля: `curl -o profiles/heap.pprof "http://localhost:8080/debug/pprof/heap"`

### Примеры использования API

**Создание короткой ссылки (текстовый формат):**
```bash
curl -X POST http://localhost:8080/ \
  -H "Content-Type: text/plain" \
  -d "https://example.com/very/long/url"
```

**Создание короткой ссылки (JSON формат):**
```bash
curl -X POST http://localhost:8080/api/shorten \
  -H "Content-Type: application/json" \
  -d '{"url": "https://example.com/very/long/url"}'
```

**Получение оригинального URL (редирект):**
```bash
curl -L http://localhost:8080/<short_path>
```

**Проверка соединения с БД:**
```bash
curl http://localhost:8080/ping
```

## Обновление шаблона

Чтобы иметь возможность получать обновления автотестов и других частей шаблона, выполните команду:

```bash
git remote add -m main template https://github.com/Yandex-Practicum/go-musthave-shortener-tpl.git
```

Для обновления кода автотестов выполните команду:

```bash
git fetch template && git checkout template/main .github
```

Затем добавьте полученные изменения в свой репозиторий.

## Запуск автотестов

Для успешного запуска автотестов называйте ветки `iter<number>`, где `<number>` — порядковый номер инкремента. Например, в ветке с названием `iter4` запустятся автотесты для инкрементов с первого по четвёртый.

При мёрже ветки с инкрементом в основную ветку `main` будут запускаться все автотесты.

Подробнее про локальный и автоматический запуск читайте в [README автотестов](https://github.com/Yandex-Practicum/go-autotests).


## Результат профилирования


### Анализ оптимизации

- **Буферы HTTP** — экономия ~1.5 MB на чтение и ~514 KB на запись. Оптимизации в обработчиках и middleware ускорили обработку запросов, поэтому в момент снятия профиля было меньше активных соединений и меньше буферов в памяти.

- **Логирование** — экономия ~512 KB. Убрали лишние логи в горячих путях (там, где много запросов), уменьшили объём данных в логах.

- **Получение URL по короткой ссылке** — экономия ~512 KB. Оптимизировали работу с результатами запросов к БД: теперь меньше копирований данных и меньше памяти удерживается после обработки запроса.

- **Подключения к PostgreSQL** — экономия памяти на пул соединений и протокол обмена с БД.

### Сравнение профилей (pprof -top -diff_base)

```
maksimradzej@MacBook-Pro-Maksim shortener % go tool pprof -top -diff_base=profiles/base.pprof profiles/result.pprof
File: shortener
Build ID: 8b05be819f1574281205d4aaab8cef8ca0a582bf
Type: inuse_space
Time: 2026-02-15 16:25:29 +05
Showing nodes accounting for -3080.38kB, 42.77% of 7202.79kB total
Dropped 2 nodes (cum <= 36.01kB)
      flat  flat%   sum%        cum   cum%
-1542.01kB 21.41% 21.41% -1542.01kB 21.41%  bufio.NewReaderSize (inline)
    -514kB  7.14% 28.54%     -514kB  7.14%  bufio.NewWriterSize (inline)
 -512.50kB  7.12% 35.66%  -512.50kB  7.12%  go.uber.org/zap/internal/bufferpool.init.NewPool.func1
  512.14kB  7.11% 28.55%   512.14kB  7.11%  github.com/jackc/pgx/v5.(*Conn).getRows
 -512.01kB  7.11% 35.66%  -512.01kB  7.11%  github.com/MaxRadzey/shortener/internal/repository/postgres.(*PostgresRepository).Get
 -512.01kB  7.11% 42.77%  -512.01kB  7.11%  github.com/jackc/pgx/v5/pgproto3.(*AuthenticationSASL).Decode
         0     0% 42.77% -1542.01kB 21.41%  bufio.NewReader (inline)
         0     0% 42.77%  -512.01kB  7.11%  github.com/MaxRadzey/shortener/internal/handler.(*Handler).GetURL
         0     0% 42.77%   512.14kB  7.11%  github.com/MaxRadzey/shortener/internal/handler.(*Handler).GetUserURLs
         0     0% 42.77%   512.14kB  7.11%  github.com/MaxRadzey/shortener/internal/repository/postgres.(*PostgresRepository).GetByUserID
         0     0% 42.77%  -512.37kB  7.11%  github.com/MaxRadzey/shortener/internal/router.SetupRouter.RequestLogger.func2
         0     0% 42.77%  -512.37kB  7.11%  github.com/MaxRadzey/shortener/internal/router.SetupRouter.ResponseLogger.func3
         0     0% 42.77%  -512.01kB  7.11%  github.com/MaxRadzey/shortener/internal/service.(*Service).GetLongURL
         0     0% 42.77%   512.14kB  7.11%  github.com/MaxRadzey/shortener/internal/service.(*Service).GetUserURLs
         0     0% 42.77%  -512.37kB  7.11%  github.com/gin-gonic/gin.(*Context).Next (partial-inline)
         0     0% 42.77%  -512.37kB  7.11%  github.com/gin-gonic/gin.(*Engine).ServeHTTP
         0     0% 42.77%  -512.37kB  7.11%  github.com/gin-gonic/gin.(*Engine).handleHTTPRequest
         0     0% 42.77%  -512.37kB  7.11%  github.com/gin-gonic/gin.CustomRecoveryWithWriter.func1
         0     0% 42.77%  -512.37kB  7.11%  github.com/gin-gonic/gin.LoggerWithConfig.func1
         0     0% 42.77%   512.14kB  7.11%  github.com/jackc/pgx/v5.(*Conn).Query
         0     0% 42.77%  -512.01kB  7.11%  github.com/jackc/pgx/v5.ConnectConfig
         0     0% 42.77%  -512.01kB  7.11%  github.com/jackc/pgx/v5.connect
         0     0% 42.77%  -512.01kB  7.11%  github.com/jackc/pgx/v5/pgconn.(*PgConn).peekMessage
         0     0% 42.77%  -512.01kB  7.11%  github.com/jackc/pgx/v5/pgconn.(*PgConn).receiveMessage
         0     0% 42.77%  -512.01kB  7.11%  github.com/jackc/pgx/v5/pgconn.ConnectConfig
         0     0% 42.77%  -512.01kB  7.11%  github.com/jackc/pgx/v5/pgconn.connectOne
         0     0% 42.77%  -512.01kB  7.11%  github.com/jackc/pgx/v5/pgconn.connectPreferred
         0     0% 42.77%  -512.01kB  7.11%  github.com/jackc/pgx/v5/pgproto3.(*Frontend).Receive
         0     0% 42.77%   512.14kB  7.11%  github.com/jackc/pgx/v5/pgxpool.(*Conn).Query
         0     0% 42.77%   512.14kB  7.11%  github.com/jackc/pgx/v5/pgxpool.(*Pool).Query
         0     0% 42.77%  -512.01kB  7.11%  github.com/jackc/pgx/v5/pgxpool.NewWithConfig.func3
         0     0% 42.77%  -512.01kB  7.11%  github.com/jackc/puddle/v2.(*Pool[go.shape.*uint8]).initResourceValue.func1
         0     0% 42.77%  -512.50kB  7.12%  go.uber.org/zap.(*Logger).Info
         0     0% 42.77%  -512.50kB  7.12%  go.uber.org/zap/buffer.Pool.Get
         0     0% 42.77%  -512.50kB  7.12%  go.uber.org/zap/internal/bufferpool.init.NewPool.New[go.shape.*uint8].func2
         0     0% 42.77%  -512.50kB  7.12%  go.uber.org/zap/internal/pool.(*Pool[go.shape.*uint8]).Get (inline)
         0     0% 42.77%  -512.50kB  7.12%  go.uber.org/zap/zapcore.(*CheckedEntry).Write
         0     0% 42.77%  -512.50kB  7.12%  go.uber.org/zap/zapcore.(*ioCore).Write
         0     0% 42.77%  -512.50kB  7.12%  go.uber.org/zap/zapcore.EntryCaller.TrimmedPath
         0     0% 42.77%  -512.50kB  7.12%  go.uber.org/zap/zapcore.ShortCallerEncoder
         0     0% 42.77%  -512.50kB  7.12%  go.uber.org/zap/zapcore.consoleEncoder.EncodeEntry
         0     0% 42.77% -2568.38kB 35.66%  net/http.(*conn).serve
         0     0% 42.77% -1542.01kB 21.41%  net/http.newBufioReader
         0     0% 42.77%     -514kB  7.14%  net/http.newBufioWriterSize
         0     0% 42.77%  -512.37kB  7.11%  net/http.serverHandler.ServeHTTP
         0     0% 42.77%  -512.50kB  7.12%  sync.(*Pool).Get
```
