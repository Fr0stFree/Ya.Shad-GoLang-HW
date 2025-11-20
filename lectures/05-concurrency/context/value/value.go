// Package main демонстрирует правильное использование context.Value
// для передачи request-scoped данных.
//
// Context.Value позволяет передавать данные через цепочку вызовов функций
// без явной передачи через параметры. Это полезно для:
//   - Request ID для трассировки
//   - User ID для авторизации
//   - Trace spans для observability
//   - Локализации (язык пользователя)
//
// ВАЖНО: НЕ используйте context.Value для:
//   - Передачи обязательных параметров функции
//   - Передачи опций конфигурации
//   - Любых данных, критичных для бизнес-логики
//
// Правила работы с context.Value:
//  1. Используйте приватный тип для ключа (не string!)
//  2. Предоставьте type-safe функции для чтения/записи
//  3. Документируйте что хранится в контексте
//  4. Не храните изменяемые данные
//
// Пример правильного использования - трассировка запросов:
//
//	func handleRequest(w http.ResponseWriter, r *http.Request) {
//	    ctx := r.Context()
//	    ctx = WithUser(ctx, getUserFromAuth(r))
//
//	    processOrder(ctx, orderID)
//	}
//
//	func processOrder(ctx context.Context, orderID string) {
//	    user, ok := ContextUser(ctx)
//	    if ok {
//	        log.Printf("User %s is processing order %s", user, orderID)
//	    }
//	    // ...
//	}
package main

import (
	"context"
	"fmt"
)

// myKey - приватный тип для ключа контекста.
//
// ВАЖНО: Используем struct{}, а НЕ string!
//
// Почему не string:
//   - Разные пакеты могут использовать одинаковые строки как ключи
//   - Это приведет к коллизиям и перезаписи значений
//
// Пример проблемы со string:
//   package auth:   ctx = context.WithValue(ctx, "user", "alice")
//   package logging: ctx = context.WithValue(ctx, "user", "logger_user")
//   // Второй вызов перезапишет первый! ❌
//
// С приватным типом:
//   package auth:   type authKey struct{}
//                   ctx = context.WithValue(ctx, authKey{}, "alice")
//   package logging: type logKey struct{}
//                   ctx = context.WithValue(ctx, logKey{}, "logger_user")
//   // Нет коллизий! ✓
//
// Приватный тип гарантирует что только этот пакет может создавать
// такие ключи, предотвращая случайные коллизии.
//
// use private type to restrict access to this package
type myKey struct{}

// WithUser создает новый контекст с информацией о пользователе.
//
// Эта функция НЕ мутирует существующий контекст - она создает новый!
// Контексты в Go immutable (неизменяемые).
//
// Пример:
//
//	ctx1 := context.Background()
//	ctx2 := WithUser(ctx1, "alice")
//	ctx3 := WithUser(ctx2, "bob")
//
//	// ctx1 - пустой (без user)
//	// ctx2 - содержит user="alice"
//	// ctx3 - содержит user="bob"
func WithUser(ctx context.Context, user string) context.Context {
	return context.WithValue(ctx, myKey{}, user)
}

// ContextUser извлекает имя пользователя из контекста (если оно есть).
//
// Export type-safe interface for users of this value.
// Экспортируем type-safe интерфейс для безопасного получения значения.
//
// Возвращает:
//   - (user, true) если пользователь найден в контексте
//   - ("", false) если пользователя нет
//
// Type-safety:
// Вместо того чтобы возвращать interface{} и заставлять вызывающий код
// делать type assertion, мы уже проверяем тип здесь и возвращаем string.
//
// Пример использования:
//
//	func processRequest(ctx context.Context) {
//	    user, ok := ContextUser(ctx)
//	    if !ok {
//	        log.Println("Anonymous request")
//	        return
//	    }
//	    log.Printf("Request from user: %s", user)
//	}
func ContextUser(ctx context.Context) (string, bool) {
	// Context.Value возвращает interface{} или nil
	v := ctx.Value(myKey{})

	// Type assertion для проверки что значение действительно string
	s, ok := v.(string)
	return s, ok
}

// OMIT

func main() {
	// Background() возвращает пустой контекст без значений
	ctx := context.Background()

	// Попытка получить пользователя из пустого контекста
	user, ok := ContextUser(ctx)
	fmt.Println(ok, user) // Output: false ""

	// Добавляем пользователя в контекст
	// ВАЖНО: WithUser создает НОВЫЙ контекст, не модифицируя старый
	ctx = WithUser(ctx, "petya")

	// Теперь пользователь есть в контексте
	user, ok = ContextUser(ctx)
	fmt.Println(ok, user) // Output: true petya
}

// Пример более сложного использования:
//
//	func handleHTTPRequest(w http.ResponseWriter, r *http.Request) {
//	    // Получаем контекст из HTTP запроса
//	    ctx := r.Context()
//
//	    // Добавляем информацию о пользователе из токена авторизации
//	    token := r.Header.Get("Authorization")
//	    if userID := validateToken(token); userID != "" {
//	        ctx = WithUser(ctx, userID)
//	    }
//
//	    // Передаем контекст дальше по цепочке вызовов
//	    // Все функции смогут получить user через ContextUser(ctx)
//	    result, err := processBusinessLogic(ctx, requestData)
//	    if err != nil {
//	        http.Error(w, err.Error(), http.StatusInternalServerError)
//	        return
//	    }
//	    json.NewEncoder(w).Encode(result)
//	}
//
//	func processBusinessLogic(ctx context.Context, data RequestData) (*Result, error) {
//	    // Извлекаем пользователя для логирования/аудита
//	    user, ok := ContextUser(ctx)
//	    if !ok {
//	        return nil, errors.New("unauthorized")
//	    }
//
//	    log.Printf("User %s is processing request", user)
//
//	    // Передаем контекст дальше в database запрос
//	    return db.QueryContext(ctx, "SELECT ...", user)
//	}
