# Go Лекции - Презентационный сервер

Этот репозиторий содержит лекции по программированию на Go для курса ITMO/Yandex ШАД.

## Быстрый старт

### Использование Docker Compose (рекомендуется для разработки)

```bash
# Запустить презентационный сервер
docker-compose up

# Или в фоновом режиме
docker-compose up -d

# Смотреть логи
docker-compose logs -f

# Остановить
docker-compose down
```

После запуска откройте браузер: http://localhost:8080

### Использование Docker напрямую

```bash
# Собрать образ
docker build -t go-lectures .

# Запустить контейнер
docker run -p 8080:8080 go-lectures

# Или с пробросом файлов для live reload
docker run -p 8080:8080 -v $(pwd):/lectures:ro go-lectures
```

## Структура репозитория

```
lectures/
├── 01-basics/           # Основы Go
├── 02-interfaces/       # Интерфейсы
├── 03-goroutines/       # Горутины и каналы
├── 04-testing/          # Тестирование
├── 05-concurrency/      # Конкурентность с shared memory
│   ├── context/         # Примеры работы с context
│   ├── memo1-4/         # Эволюция concurrent cache
│   ├── synconce/        # sync.Map + sync.Once
│   └── lecture.slide    # Слайды лекции
├── ...
├── Dockerfile           # Docker образ для present
├── docker-compose.yaml  # Docker Compose конфигурация
└── README.md            # Этот файл
```

## О Present

[Present](https://pkg.go.dev/golang.org/x/tools/cmd/present) - это инструмент из `golang.org/x/tools` для создания и показа презентаций.

### Формат .slide файлов

Презентации пишутся в текстовом формате `.slide`:

```
Название презентации
Подзаголовок
Дата

Автор
email@example.com

* Первый слайд

Текст слайда

    // Код с подсветкой синтаксиса
    func main() {
        fmt.Println("Hello, World!")
    }

* Второй слайд

- Маркированный список
- Еще один пункт
```

### Встраивание исполняемого кода

Present поддерживает выполнение Go кода прямо в презентации:

```
.play example.go /START OMIT/,/END OMIT/
.code example.go /func main/,/^}/
```

Директивы:
- `.play` - код можно запустить через кнопку "Run"
- `.code` - просто показать код
- `/START/,/END/` - показать код между маркерами

## Разработка

### Live Reload

При использовании `docker-compose up` файлы монтируются в контейнер через volume.
Изменения в `.slide` и `.go` файлах будут видны сразу после обновления страницы.

### Добавление новой лекции

1. Создайте директорию: `mkdir lectures/XX-topic/`
2. Создайте файл презентации: `touch lectures/XX-topic/lecture.slide`
3. Добавьте примеры кода в `.go` файлы
4. Перезагрузите страницу в браузере

### Проверка синтаксиса

```bash
# Запустить present локально (требуется Go)
go install golang.org/x/tools/cmd/present@latest
present
```

## Полезные команды Docker

```bash
# Пересобрать образ
docker-compose build

# Пересобрать образ без кэша
docker-compose build --no-cache

# Посмотреть запущенные контейнеры
docker-compose ps

# Выполнить команду в контейнере
docker-compose exec present sh

# Удалить все (контейнеры, сети, volumes)
docker-compose down -v

# Посмотреть использование ресурсов
docker stats
```

## Troubleshooting

### Порт 8080 уже занят

```bash
# Изменить порт в docker-compose.yaml:
ports:
  - "127.0.0.1:3000:8080"  # Теперь доступно на :3000
```

### Изменения не видны в браузере

1. Проверьте что используется volume mount: `docker-compose ps`
2. Очистите кэш браузера (Ctrl+F5)
3. Проверьте что файл действительно изменился в контейнере:
   ```bash
   docker-compose exec present ls -la /lectures/05-concurrency/
   ```

### Present не запускается

```bash
# Посмотреть логи
docker-compose logs present

# Проверить образ
docker images | grep lectures

# Пересобрать с нуля
docker-compose down
docker-compose build --no-cache
docker-compose up
```

### Health check failed

Если контейнер помечен как unhealthy:

```bash
# Проверить доступность внутри контейнера
docker-compose exec present wget -qO- http://localhost:8080

# Отключить health check (в docker-compose.yaml):
# Закомментируйте секцию healthcheck
```

## CI/CD

### GitLab CI

Пример `.gitlab-ci.yml` для автоматической публикации:

```yaml
deploy:
  stage: deploy
  image: docker:latest
  services:
    - docker:dind
  script:
    - cd lectures
    - docker build -t $CI_REGISTRY_IMAGE/lectures:latest .
    - docker push $CI_REGISTRY_IMAGE/lectures:latest
  only:
    - master
```

## Production Deployment

Для production рекомендуется:

1. Использовать reverse proxy (nginx/traefik)
2. Настроить HTTPS
3. Добавить аутентификацию если нужно
4. Использовать docker-compose с внешними сетями

Пример с nginx:

```yaml
# docker-compose.prod.yaml
version: "3.8"
services:
  present:
    image: go-lectures:latest
    restart: always
    networks:
      - web
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.lectures.rule=Host(`lectures.example.com`)"

  traefik:
    image: traefik:v2.5
    ports:
      - "80:80"
      - "443:443"
    # ... traefik config
```

## Полезные ссылки

- [Present documentation](https://pkg.go.dev/golang.org/x/tools/present)
- [Example presentations](https://talks.golang.org/)
- [Go Blog: Go Present](https://go.dev/blog/present)
- [Present source code](https://github.com/golang/tools/tree/master/cmd/present)

## Контакты

- Вопросы по курсу: через GitLab Issues
- Документация курса: `docs/` директория
- ManYtask: https://go.manytask.org

## Лицензия

См. LICENSE файл в корне репозитория.
