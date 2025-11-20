# Изменения в лекции 05-modules (Modules)

## Обзор

Лекция "Modules" была значительно расширена практическими примерами, объяснениями современных фич (workspaces, retract) и best practices.

## Статистика изменений

- **lectures.slide**: +313 строк (было 503, стало ~816)
- **Новые слайды**: ~18 новых слайдов
- **Новые темы**: Workspaces, Retract, Private modules, Troubleshooting

## Добавленные темы

### 1. Вводный слайд "Зачем нужны модули?"

**Новый слайд**: Объяснение проблем до Go modules

Добавлено:
- Проблемы GOPATH и старого подхода
- Преимущества Go modules
- Мотивация для изучения

### 2. Go Workspaces (Go 1.18+)

**Новые слайды** (2):
- "Go Workspaces (Go 1.18+)" — введение в go.work
- "Go Workspaces: Пример" — практический пример

Содержание:
- Что такое go.work файл
- Как работать с несколькими модулями одновременно
- Преимущества над replace директивами
- Структура monorepo с workspaces

Пример:
```go
# go work init ./hello ./world
# cat go.work
go 1.22.0

use (
    ./hello
    ./world
)
```

### 3. Retracted Versions

**Новый слайд**: "Retracted Versions"

Объяснение механизма отзыва версий:
- Как пометить версию как retracted
- Что происходит при попытке использовать retracted версию
- Use case: уязвимости, ошибки в релизе

Пример:
```go
retract (
    v1.0.5 // Содержит критическую уязвимость CVE-2023-xxxxx
    v1.0.4 // Сломана сборка на Windows
    [v1.0.0, v1.0.3] // Несовместимое API
)
```

### 4. Private Modules

**Новые слайды** (2):
- "Private Modules: GOPRIVATE" — настройка приватных репозиториев
- "Private Modules: Полный пример" — детальный пример с CI/CD

Содержание:
- Переменная окружения GOPRIVATE
- Настройка git credentials
- OAuth токены для GitHub/GitLab
- SSH keys
- Примеры для CI/CD (GitLab CI, GitHub Actions)

Практические команды:
```bash
# Настроить GOPRIVATE
export GOPRIVATE=github.com/mycompany/*

# Git credentials
git config --global url."https://oauth2:${TOKEN}@github.com/".insteadOf "https://github.com/"
```

### 5. Dependency Management Best Practices

**Новый слайд**: "Dependency Management Best Practices"

5 ключевых правил:
1. Всегда коммитьте go.sum
2. Запускайте go mod tidy перед коммитом
3. Фиксируйте версии
4. Обновляйте регулярно
5. Проверяйте лицензии

Команды для обслуживания:
```bash
# Устаревшие зависимости
go list -u -m all

# Обновить зависимость
go get github.com/pkg/errors@v0.9.1

# Найти уязвимости
govulncheck ./...
```

### 6. Semantic Versioning - детально

**Новый слайд**: "Версионирование: Semantic Versioning"

Подробное объяснение:
- Формат версий (MAJOR.MINOR.PATCH)
- Когда менять каждую часть
- Специальные версии (v0, v1, v2+)
- Псевдо-версии
- Визуализация формата

Примеры:
```
v1.0.1  — bug fix (patch)
v1.1.0  — new feature (minor)
v2.0.0  — breaking change (major)

Псевдо-версия:
v0.0.0-20240101120000-abcdef123456
│      │                  └─ commit hash
│      └─ timestamp
└─ базовая версия
```

### 7. go mod commands - Шпаргалка

**Новый слайд**: "go mod commands: Шпаргалка"

Категории команд:
- **Основные**: init, tidy, download, verify, vendor
- **Информация**: list, graph, why
- **Редактирование**: edit с различными флагами

Удобная справочная таблица для быстрого поиска нужной команды.

### 8. Troubleshooting

**Новые слайды** (2):
- "Troubleshooting: Распространенные проблемы"
- "Troubleshooting: Очистка и reset"

Решение типичных проблем:

**Проблема 1**: `no required module provides package`
```bash
# Решения:
go list -m example.com/pkg
go get -u example.com/pkg
go clean -modcache
```

**Проблема 2**: `checksum mismatch`
```bash
# Решения:
rm go.sum && go mod tidy
go mod verify
```

**Очистка кеша**:
```bash
go clean -modcache    # модули
go clean -cache       # build cache
go clean -testcache   # test cache
```

### 9. Практические примеры

**Новые слайды** (2):
- "Практические примеры: Monorepo"
- "Практические примеры: CI/CD Pipeline"

**Monorepo структура:**
```
mycompany/
├── go.work
├── services/
│   ├── api/
│   └── worker/
└── libs/
    ├── database/
    └── logger/
```

**CI/CD примеры:**
- .gitlab-ci.yml с go mod download и verify
- GitHub Actions с govulncheck
- Проверка go.sum в CI

### 10. Расширенный раздел go mod tidy

**Обновлен слайд**: "go mod tidy: Детали работы"

Добавлено:
- Детальное объяснение всех действий команды
- Когда именно нужно запускать
- Best practice с pre-commit hook
- Пример скрипта для автоматизации

```bash
#!/bin/sh
go mod tidy
go mod verify
git diff --exit-code go.mod go.sum || exit 1
```

### 11. Итоговый слайд

**Новый слайд**: "Итоги"

Краткое резюме всех тем:
- Module system
- go.mod и go.sum
- Semantic versioning
- go mod tidy
- Workspaces
- Private modules
- Best practices

## Улучшения существующих слайдов

### Вводные слайды

- Добавлено объяснение "Зачем нужны модули?"
- Расширены определения ключевых терминов
- Добавлены аналогии с другими языками (package.json, requirements.txt)

### Ссылки

Добавлены ссылки на:
- go.dev/blog/using-go-modules
- go.dev/ref/mod
- go.dev/doc/modules/managing-dependencies
- pkg.go.dev/golang.org/x/vuln/cmd/govulncheck

## Педагогические улучшения

### 1. Прогрессия обучения

Лекция теперь структурирована:
1. **Зачем** нужны модули (мотивация)
2. **Основы** (go.mod, go.sum, команды)
3. **Продвинутые** фичи (workspaces, retract, private)
4. **Практика** (best practices, troubleshooting)
5. **Реальные** примеры (monorepo, CI/CD)

### 2. Практическая направленность

Все новые концепции проиллюстрированы:
- Реальными use cases (CI/CD, monorepo)
- Частыми проблемами и их решениями
- Командами готовыми к копированию

### 3. Современность

Добавлены темы из Go 1.18+:
- Workspaces (Go 1.18)
- Retract mechanism
- govulncheck для безопасности

### 4. Практичность

Добавлены:
- Шпаргалки команд
- Troubleshooting guide
- CI/CD примеры
- Pre-commit hooks

## Сравнение до/после

### Было (503 строки)

- Базовые команды go mod
- Простые примеры с hello world
- Module proxy
- Vanity imports

### Стало (~816 строк)

**Все вышеперечисленное ПЛЮС:**

- Go Workspaces
- Retracted versions
- Private modules (GOPRIVATE)
- Best practices для dependency management
- Подробный semantic versioning
- Шпаргалка go mod команд
- Troubleshooting guide (2 слайда)
- Monorepo примеры
- CI/CD pipeline примеры
- Расширенные ссылки

## Статистика контента

- **Оригинальные слайды**: ~30
- **Новые слайды**: ~18
- **Итого слайдов**: ~48
- **Новые разделы**: 6
- **Практические примеры**: 8+
- **Команды и сниппеты**: 25+

## Для преподавателя

### Ключевые моменты для демонстрации

1. **go mod tidy** — показать на реальном проекте
2. **Workspaces** — продемонстрировать с двумя модулями
3. **go.sum** — показать почему это важно для безопасности
4. **Troubleshooting** — разобрать пару реальных проблем

### Практические упражнения

1. Создать модуль и добавить зависимость
2. Работа с workspace для двух модулей
3. Настроить GOPRIVATE для приватного репозитория
4. Решить проблему с checksum mismatch

### Дополнительные темы (опционально)

- athens proxy для корпоративного использования
- vendoring и когда его использовать
- GOPROXY цепочки
- Minimal version selection algorithm

## Совместимость

- Обратная совместимость сохранена
- Все оригинальные примеры работают
- Новые фичи помечены версией Go (1.18+)
- Present корректно отображает все слайды

## Следующие шаги

Рекомендации для дальнейшего развития:

1. **Видео примеры** — записать демонстрацию команд
2. **Интерактивные упражнения** — add hands-on практику
3. **Athens proxy** — добавить слайд про корпоративный proxy
4. **MVS algorithm** — объяснить minimal version selection

## Полезные ресурсы

Добавленные в лекцию:
- Official Go modules documentation
- govulncheck tool
- Using Go Modules blog post
- Managing Dependencies guide

---

**Дата изменений**: 2025-10-30
**Автор**: Обновлено с помощью Claude Code
**Версия**: 2.0
**Добавлено строк**: +313 (~+62%)
