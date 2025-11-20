// Package cancelation демонстрирует использование context для отмены операций.
//
// Context - это стандартный механизм в Go для:
//  1. Отмены длительных операций
//  2. Установки таймаутов
//  3. Передачи deadlines
//
// Основные паттерны:
//   - context.WithCancel() - ручная отмена
//   - context.WithTimeout() - автоматическая отмена по таймауту
//   - context.WithDeadline() - отмена в конкретное время
//   - context.WithCancelCause() - отмена с указанием причины
//
// Типичные use cases:
//   - HTTP сервер: отменяем обработку если клиент разорвал соединение
//   - Database запросы: ограничиваем время выполнения запроса
//   - Background задачи: graceful shutdown при остановке сервера
//   - API клиенты: таймаут на внешние запросы
package cancelation

import (
	"context"
	"time"
)

// SimpleCancelation демонстрирует ручную отмену контекста.
//
// context.WithCancel создает контекст который можно отменить вручную
// вызовом функции cancel().
//
// Паттерн использования:
//  1. Создаем контекст с cancel функцией
//  2. Передаем контекст в длительную операцию
//  3. В любой момент можем вызвать cancel() для отмены
//  4. Операция должна периодически проверять ctx.Done()
//
// Пример из реальной жизни - отмена обработки HTTP запроса:
//
//	func handleRequest(w http.ResponseWriter, r *http.Request) {
//	    ctx := r.Context() // контекст отменится если клиент разорвет соединение
//
//	    result, err := processLongRunningTask(ctx)
//	    if err != nil {
//	        if errors.Is(err, context.Canceled) {
//	            log.Println("Client disconnected")
//	            return
//	        }
//	        http.Error(w, err.Error(), 500)
//	        return
//	    }
//	    json.NewEncoder(w).Encode(result)
//	}
func SimpleCancelation() {
	// WithCancel возвращает:
	//   - новый контекст (child от Background)
	//   - функцию cancel для отмены контекста
	ctx, cancel := context.WithCancel(context.Background())

	// ВАЖНО: Всегда вызывайте cancel(), даже если контекст не был отменен!
	// defer гарантирует что ресурсы будут освобождены.
	// Если не вызвать cancel, произойдет memory leak.
	defer cancel()

	// Запускаем горутину которая отменит контекст через 5 секунд
	go func() {
		time.Sleep(5 * time.Second)
		cancel() // Отменяем контекст
		// После этого ctx.Done() закроется и doSlowJob завершится
	}()

	// Запускаем долгую операцию с отменяемым контекстом
	if err := doSlowJob(ctx); err != nil {
		// err будет context.Canceled
		panic(err)
	}
}

// OMIT

// SimpleTimeout демонстрирует автоматическую отмену по таймауту.
//
// context.WithTimeout - это удобная обертка над WithDeadline.
// Контекст автоматически отменится через указанное время.
//
// Эквивалентный код:
//   deadline := time.Now().Add(5 * time.Second)
//   ctx, cancel := context.WithDeadline(ctx, deadline)
//
// Пример из реальной жизни - запрос к внешнему API с таймаутом:
//
//	func fetchUserData(userID string) (*User, error) {
//	    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
//	    defer cancel()
//
//	    req, _ := http.NewRequestWithContext(ctx, "GET", apiURL+userID, nil)
//	    resp, err := http.DefaultClient.Do(req)
//	    if err != nil {
//	        if errors.Is(err, context.DeadlineExceeded) {
//	            return nil, fmt.Errorf("API timeout after 2s")
//	        }
//	        return nil, err
//	    }
//	    defer resp.Body.Close()
//
//	    var user User
//	    json.NewDecoder(resp.Body).Decode(&user)
//	    return &user, nil
//	}
func SimpleTimeout() {
	// WithTimeout автоматически отменит контекст через 5 секунд
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	// ВАЖНО: Всегда defer cancel() даже с таймаутом!
	// Причины:
	//   1. Если операция завершится раньше таймаута - освободим ресурсы сразу
	//   2. Предотвращаем memory leak
	//   3. Good practice - единообразный код
	defer cancel()

	// Запускаем долгую операцию с таймаутом
	if err := doSlowJob(ctx); err != nil {
		// err будет context.DeadlineExceeded если истек таймаут
		panic(err)
	}
}

// OMIT

// doSlowJob имитирует долгую операцию которая может быть отменена.
//
// Правильный паттерн для cancelable операций:
//  1. В цикле периодически проверяем ctx.Done()
//  2. При получении сигнала отмены - немедленно завершаемся
//  3. Возвращаем ошибку из context.Cause() или ctx.Err()
//
// select с default позволяет:
//   - Проверить отмену БЕЗ блокировки
//   - Продолжить работу если отмены нет
//
// Альтернативный паттерн (blocking):
//
//	for {
//	    select {
//	    case <-ctx.Done():
//	        return ctx.Err()
//	    case result := <-workChannel:
//	        process(result)
//	    }
//	}
//
// Пример из реальной жизни - обработка очереди задач:
//
//	func processQueue(ctx context.Context, queue <-chan Task) error {
//	    for {
//	        select {
//	        case <-ctx.Done():
//	            log.Println("Shutting down queue processor")
//	            return ctx.Err()
//	        case task := <-queue:
//	            if err := processTask(ctx, task); err != nil {
//	                log.Printf("Task failed: %v", err)
//	            }
//	        }
//	    }
//	}
func doSlowJob(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			// Канал ctx.Done() закрывается когда контекст отменяется
			// Это происходит при:
			//   - Вызове cancel() (для WithCancel)
			//   - Истечении timeout (для WithTimeout)
			//   - Наступлении deadline (для WithDeadline)
			//   - Отмене родительского контекста

			// context.Cause() возвращает причину отмены (Go 1.20+)
			// Возможные значения:
			//   - context.Canceled - контекст отменен вызовом cancel()
			//   - context.DeadlineExceeded - истек таймаут/deadline
			//   - custom error - если использовался WithCancelCause
			return context.Cause(ctx)

			// До Go 1.20 использовался ctx.Err():
			// return ctx.Err()

		default:
			// Если контекст не отменен - продолжаем работу

			// perform a portion of slow job
			// Выполняем часть медленной работы
			time.Sleep(1 * time.Second)

			// В реальном коде здесь может быть:
			//   - Обработка одного элемента из очереди
			//   - Один запрос к database
			//   - Одна итерация сложных вычислений
			//   - и т.д.

			// Важно: между итерациями всегда проверяем ctx.Done()
			// чтобы отреагировать на отмену как можно быстрее
		}
	}
}

// OMIT

// Пример более сложного сценария - параллельная работа с отменой:
//
//	func fetchDataFromMultipleSources(ctx context.Context) ([]Data, error) {
//	    // Создаем контекст с таймаутом для всей операции
//	    ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
//	    defer cancel()
//
//	    // Запускаем параллельные запросы к разным источникам
//	    var wg sync.WaitGroup
//	    results := make(chan Data, 3)
//	    errors := make(chan error, 3)
//
//	    sources := []string{"source1", "source2", "source3"}
//	    for _, source := range sources {
//	        wg.Add(1)
//	        go func(s string) {
//	            defer wg.Done()
//	            data, err := fetchFromSource(ctx, s)
//	            if err != nil {
//	                errors <- err
//	                return
//	            }
//	            results <- data
//	        }(source)
//	    }
//
//	    // Закрываем каналы после завершения всех горутин
//	    go func() {
//	        wg.Wait()
//	        close(results)
//	        close(errors)
//	    }()
//
//	    // Собираем результаты или возвращаем первую ошибку
//	    var data []Data
//	    for {
//	        select {
//	        case <-ctx.Done():
//	            // Таймаут или ручная отмена - все горутины тоже остановятся
//	            return nil, ctx.Err()
//	        case d, ok := <-results:
//	            if !ok {
//	                return data, nil // Все успешно завершено
//	            }
//	            data = append(data, d)
//	        case err := <-errors:
//	            cancel() // Отменяем остальные запросы
//	            return nil, err
//	        }
//	    }
//	}
