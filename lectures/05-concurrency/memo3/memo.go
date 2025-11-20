// Copyright © 2016 Alan A. A. Donovan & Brian W. Kernighan.
// License: https://creativecommons.org/licenses/by-nc-sa/4.0/

// See page 276.

// Package memo provides a concurrency-safe memoization a function of
// type Func.  Requests for different keys run concurrently.
// Concurrent requests for the same key result in duplicate work.
//
// РЕШЕНИЕ №2: Освобождаем мьютекс на время вычисления.
//
// ПЛЮСЫ:
//   - Безопасно для concurrent использования
//   - Запросы к РАЗНЫМ ключам выполняются параллельно!
//   - Намного быстрее, чем memo2
//
// МИНУСЫ:
//   - Запросы к ОДНОМУ ключу приводят к дублированию работы
//   - Функция f() может быть вызвана несколько раз для одного ключа
//   - Неэффективно, если вычисления дорогие
//
// Пример проблемы:
//
//	memo := New(func(key string) (interface{}, error) {
//	    fmt.Printf("Computing %s...\n", key)
//	    time.Sleep(2 * time.Second)
//	    return fmt.Sprintf("result: %s", key), nil
//	})
//
//	// Три горутины одновременно запрашивают один и тот же ключ:
//	go memo.Get("expensive")
//	go memo.Get("expensive")
//	go memo.Get("expensive")
//
//	// Вывод:
//	// Computing expensive...
//	// Computing expensive...  <- дублирование!
//	// Computing expensive...  <- дублирование!
//
// Смотрите memo4 для оптимального решения без дублирования.
package memo

import "sync"

// Memo - потокобезопасный кеш с поддержкой параллельных вычислений
// для разных ключей.
type Memo struct {
	f     Func
	mu    sync.Mutex        // guards cache
	cache map[string]result // кеш результатов
}

// Func - тип функции для мемоизации.
type Func func(string) (interface{}, error)

// result хранит результат вычисления.
type result struct {
	value interface{} // возвращаемое значение
	err   error       // ошибка вычисления
}

// New создает новый Memo с заданной функцией.
func New(f Func) *Memo {
	return &Memo{f: f, cache: make(map[string]result)}
}

// Get возвращает кешированный результат или вычисляет его.
//
// УЛУЧШЕНИЕ по сравнению с memo2:
// Мьютекс захватывается только на время работы с map, а НЕ на время
// вычисления f(). Это позволяет разным ключам обрабатываться параллельно.
//
// Алгоритм:
//  1. [С мьютексом] Проверяем, есть ли ключ в кеше
//  2. [БЕЗ мьютекса] Если нет - вычисляем f(key) параллельно!
//  3. [С мьютексом] Сохраняем результат в кеш
//
// ПРОБЛЕМА: Duplicate work (дублирование работы)
//
// Если несколько горутин одновременно запросят один ключ:
//
//	// Время ->   0s        1s        2s        3s
//	// Горутина 1:
//	Get("x")
//	mu.Lock()
//	res, ok := cache["x"]  // ok == false
//	mu.Unlock()
//	f("x") [2 sec] ------>
//	mu.Lock()
//	cache["x"] = res                         ✓
//	mu.Unlock()
//
//	// Горутина 2 (запустилась одновременно с горутиной 1):
//	Get("x")
//	mu.Lock()
//	res, ok := cache["x"]  // ok == false (горутина 1 еще не записала!)
//	mu.Unlock()
//	f("x") [2 sec] ------>  // ДУБЛИРОВАНИЕ РАБОТЫ!
//	mu.Lock()
//	cache["x"] = res                         ✓ (перезапись)
//	mu.Unlock()
//
// Результат:
//   - Функция f("x") вызвана ДВА РАЗА вместо одного
//   - Тратится двойное время CPU и ресурсов
//   - Если f() делает HTTP запрос или обращение к БД - это особенно плохо
//
// Решение: смотрите memo4, где используется канал для синхронизации
// ожидающих горутин.
func (memo *Memo) Get(key string) (value interface{}, err error) {
	// КРИТИЧЕСКАЯ СЕКЦИЯ 1: Проверяем кеш
	memo.mu.Lock()
	res, ok := memo.cache[key]
	memo.mu.Unlock()

	if !ok {
		// ВНИМАНИЕ: Здесь мьютекс НЕ захвачен!
		// Несколько горутин могут одновременно выполнять f() для одного ключа.
		res.value, res.err = memo.f(key)

		// Between the two critical sections, several goroutines
		// may race to compute f(key) and update the map.
		//
		// КРИТИЧЕСКАЯ СЕКЦИЯ 2: Сохраняем результат
		// ПРОБЛЕМА: К этому моменту другая горутина могла уже
		// вычислить и сохранить результат для того же ключа!
		memo.mu.Lock()
		memo.cache[key] = res
		memo.mu.Unlock()
	}
	return res.value, res.err
}

// OMIT
