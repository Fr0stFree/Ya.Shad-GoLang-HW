# Makefile для удобного управления Docker окружением для лекций
#
# Использование:
#   make help      - показать справку
#   make start     - запустить презентационный сервер
#   make stop      - остановить сервер
#   make restart   - перезапустить сервер
#   make logs      - смотреть логи
#   make build     - пересобрать Docker образ

.PHONY: help start stop restart logs build clean rebuild status exec

# Цвета для вывода
GREEN  := \033[0;32m
YELLOW := \033[0;33m
RED    := \033[0;31m
RESET  := \033[0m

# Переменные
COMPOSE := docker-compose
SERVICE := present
PORT    := 8080

help: ## Показать эту справку
	@echo "$(GREEN)Доступные команды:$(RESET)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  $(YELLOW)%-15s$(RESET) %s\n", $$1, $$2}'

start: ## Запустить презентационный сервер
	@echo "$(GREEN)Запуск present сервера...$(RESET)"
	$(COMPOSE) up -d
	@echo "$(GREEN)Сервер запущен: http://localhost:$(PORT)$(RESET)"

stop: ## Остановить сервер
	@echo "$(YELLOW)Остановка сервера...$(RESET)"
	$(COMPOSE) down

restart: ## Перезапустить сервер
	@echo "$(YELLOW)Перезапуск сервера...$(RESET)"
	$(COMPOSE) restart

logs: ## Смотреть логи сервера
	$(COMPOSE) logs -f $(SERVICE)

build: ## Пересобрать Docker образ
	@echo "$(GREEN)Сборка Docker образа...$(RESET)"
	$(COMPOSE) build

rebuild: ## Пересобрать образ с нуля (без кэша)
	@echo "$(GREEN)Пересборка Docker образа без кэша...$(RESET)"
	$(COMPOSE) build --no-cache

clean: ## Остановить и удалить все контейнеры, сети, volumes
	@echo "$(RED)Удаление всех контейнеров и volumes...$(RESET)"
	$(COMPOSE) down -v

status: ## Показать статус контейнеров
	@echo "$(GREEN)Статус контейнеров:$(RESET)"
	$(COMPOSE) ps

exec: ## Открыть shell в контейнере
	@echo "$(GREEN)Подключение к контейнеру...$(RESET)"
	$(COMPOSE) exec $(SERVICE) sh

health: ## Проверить здоровье сервиса
	@echo "$(GREEN)Проверка здоровья сервиса...$(RESET)"
	@curl -s -o /dev/null -w "HTTP Status: %{http_code}\n" http://localhost:$(PORT) || \
		echo "$(RED)Сервис недоступен!$(RESET)"

open: ## Открыть презентации в браузере
	@echo "$(GREEN)Открытие браузера...$(RESET)"
	@which open > /dev/null && open http://localhost:$(PORT) || \
		which xdg-open > /dev/null && xdg-open http://localhost:$(PORT) || \
		echo "Откройте вручную: http://localhost:$(PORT)"

dev: start ## Запустить в режиме разработки (алиас для start)
	@echo "$(GREEN)Режим разработки активен. Изменения в файлах отслеживаются автоматически.$(RESET)"

# Дополнительные утилиты

lint-docker: ## Проверить Dockerfile линтером
	@which hadolint > /dev/null && hadolint Dockerfile || \
		echo "$(YELLOW)hadolint не установлен. Установите: brew install hadolint$(RESET)"

validate-compose: ## Проверить docker-compose.yaml
	@echo "$(GREEN)Валидация docker-compose.yaml...$(RESET)"
	$(COMPOSE) config > /dev/null && echo "$(GREEN)✓ Конфигурация корректна$(RESET)"

ps: status ## Алиас для status

up: start ## Алиас для start

down: stop ## Алиас для stop

tail: logs ## Алиас для logs

# Информация о системе

info: ## Показать информацию о Docker окружении
	@echo "$(GREEN)Docker информация:$(RESET)"
	@echo "Docker версия:"
	@docker --version
	@echo "\nDocker Compose версия:"
	@docker-compose --version
	@echo "\nЗапущенные контейнеры:"
	@docker ps
	@echo "\nИспользование ресурсов:"
	@docker stats --no-stream

# Очистка

prune: ## Очистить неиспользуемые Docker ресурсы
	@echo "$(YELLOW)Очистка неиспользуемых Docker ресурсов...$(RESET)"
	docker system prune -f
	@echo "$(GREEN)Очистка завершена$(RESET)"

# По умолчанию показываем справку
.DEFAULT_GOAL := help
