# Запуск приложения локально через docker-compose в фоновом режиме
local:
	docker-compose -f docker-compose.yml up -d

# Запуск всех тестов одной командой
test:
	go test ./... -v
