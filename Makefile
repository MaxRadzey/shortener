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
