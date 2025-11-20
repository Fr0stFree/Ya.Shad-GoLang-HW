# Dockerfile для запуска Go present - инструмента для презентаций
#
# Present - это инструмент из golang.org/x/tools для создания и показа
# презентаций в формате .slide (аналог reveal.js, но для Go)
#
# Использование:
#   docker build -t go-lectures .
#   docker run -p 8080:8080 go-lectures
#   Откройте браузер: http://localhost:8080
#
# Для разработки с live reload:
#   docker-compose up
#   (файлы монтируются через volume)

# Базовый образ: Go 1.25 на Alpine Linux
# Alpine выбран для минимального размера образа (~300MB vs ~800MB для debian)
FROM golang:1.25-alpine

# Устанавливаем present - инструмент для показа презентаций Go
# CGO_ENABLED=0 для статической компиляции (не требует libc в runtime)
# @v0.6.0 - фиксированная версия для воспроизводимых сборок
RUN CGO_ENABLED=0 go install golang.org/x/tools/cmd/present@v0.38.0

# Копируем все файлы лекций в контейнер
# Включает: *.slide файлы, примеры кода (*.go), картинки
COPY . /lectures

# Устанавливаем рабочую директорию
# Present будет искать .slide файлы отсюда
WORKDIR /lectures

# Запускаем present сервер
# Параметры:
#   -base <path>        : путь к темплейтам present (для стилей)
#   -http :8080         : слушаем на порту 8080
#   -orighost <domain>  : оригинальный хост для генерации ссылок
#   -use_playground     : использовать Go playground для запуска кода
#   -play=0             : отключить интеграцию с play.golang.org
#                         (код выполняется локально, не на внешнем сервере)
ENTRYPOINT ["/go/bin/present", \
    "-base", "/go/pkg/mod/golang.org/x/tools@v0.38.0/cmd/present", \
    "-http", ":8080", \
    "-orighost", "p.go.manytask.org", \
    "-use_playground", \
    "-play=0"]

# Открываем порт для доступа к веб-интерфейсу
EXPOSE 8080

# Метаданные образа
LABEL maintainer="ITMO/ManYtask" \
    description="Go lectures presentation server" \
    version="1.0"
