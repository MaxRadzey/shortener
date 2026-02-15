# Генерация Swagger-доки
swag:
	go run github.com/swaggo/swag/cmd/swag@latest init -g cmd/shortener/main.go -d . --parseInternal -o docs

# Форматирование кода + сортировка импортов
fmt:
	go run golang.org/x/tools/cmd/goimports@latest -w .

# Запуск приложения локально через docker-compose в фоновом режиме
local:
	docker-compose -f docker-compose.yml up -d

# Запуск приложения локально в dev-режиме (pprof на /debug/pprof)
run:
	go run cmd/shortener/main.go -dev

# Запуск всех тестов одной командой
test:
	go test ./... -v

# Нагрузка POST (wrk): из profiles/urls.txt
load-post:
	wrk -t4 -c100 -d1m -s profiles/load_post.lua http://localhost:8080/

# Нагрузка GET (wrk): из profiles/short_paths.txt
load-get:
	wrk -t4 -c100 -d1m -s profiles/load_get.lua http://localhost:8080/

# Нагрузка GET /api/user/urls с кукой из profiles/cookies.txt (сначала make gen-cookies и load-post)
load-get-user:
	wrk -t4 -c100 -d30s -s profiles/load_get_user.lua http://localhost:8080/
