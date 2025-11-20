// Copyright © 2016 Alan A. A. Donovan & Brian W. Kernighan.
// License: https://creativecommons.org/licenses/by-nc-sa/4.0/

// See page 276.

// Package memo provides a concurrency-safe memoization a function of
// a function.  Requests for different keys proceed in parallel.
// Concurrent requests for the same key block until the first completes.
// This implementation uses a Mutex.
//
// РЕШЕНИЕ №3 (ОПТИМАЛЬНОЕ): Используем канал для координации.
//
// ПЛЮСЫ:
//   - Безопасно для concurrent использования
//   - Запросы к разным ключам выполняются параллельно
//   - Запросы к одному ключу НЕ дублируются - первая горутина вычисляет,
//     остальные ждут результата через канал
//   - Максимальная эффективность: каждое значение вычисляется ровно один раз
//
// МИНУСЫ:
//   - Более сложная реализация
//
// Ключевая идея:
// Вместо хранения result напрямую, мы храним *entry, которая содержит:
//   - result - результат вычисления
//   - ready chan struct{} - канал для синхронизации
//
// Когда канал ready закрывается, это сигнал что результат готов.
//
// Пример работы:
//
//	memo := New(func(key string) (interface{}, error) {
//	    fmt.Printf("Computing %s...\n", key)
//	    time.Sleep(2 * time.Second)
//	    return fmt.Sprintf("result: %s", key), nil
//	})
//
//	// Три горутины одновременно запрашивают один ключ:
//	go memo.Get("expensive") // первая - вычисляет
//	go memo.Get("expensive") // вторая - ждет
//	go memo.Get("expensive") // третья - ждет
//
//	// Вывод:
//	// Computing expensive...  <- только ОДИН раз!
//	// (через 2 секунды все три горутины получат результат)
package memo

import "sync"

// Func is the type of the function to memoize.
// Func - тип функции для мемоизации.
type Func func(string) (interface{}, error)

// result хранит результат вычисления.
type result struct {
	value interface{} // возвращаемое значение
	err   error       // ошибка вычисления
}

//!+

// New создает новый Memo с заданной функцией.
func New(f Func) *Memo {
	return &Memo{f: f, cache: make(map[string]*entry)}
}

// Memo - оптимальный потокобезопасный кеш для мемоизации.
type Memo struct {
	f     Func
	mu    sync.Mutex       // защищает cache
	cache map[string]*entry // кеш: ключ -> entry (не result!)
}

// entry представляет собой вычисление в процессе или завершенное.
// Канал ready используется для координации между горутинами:
//   - Первая горутина создает entry, вычисляет результат и закрывает ready
//   - Остальные горутины ждут закрытия ready, чтобы получить результат
type entry struct {
	res   result
	ready chan struct{} // закрывается когда res готов к чтению
}

// Get возвращает кешированный результат или вычисляет его.
// Если несколько горутин одновременно запрашивают один ключ, только
// первая будет вычислять значение, остальные дождутся результата.
//
// Алгоритм:
//
// Случай 1: Первый запрос для данного ключа
//  1. [С мьютексом] Создаем entry с открытым каналом ready
//  2. [С мьютексом] Сохраняем entry в cache
//  3. [БЕЗ мьютекса] Вычисляем f(key) - долго!
//  4. [БЕЗ мьютекса] Закрываем ready - сигнал для ожидающих горутин
//
// Случай 2: Повторный запрос для того же ключа (пока первый еще вычисляется)
//  1. [С мьютексом] Находим существующий entry
//  2. [БЕЗ мьютекса] Ждем закрытия entry.ready
//  3. [БЕЗ мьютекса] Читаем готовый результат
//
// Пример с временной шкалой:
//
//	// Время ->   0s           1s           2s           3s
//	// Горутина 1:
//	Get("x")
//	mu.Lock()
//	e := cache["x"]        // nil
//	e = &entry{ready: ...}
//	cache["x"] = e
//	mu.Unlock()
//	f("x") [2 sec] ----------------------->
//	close(e.ready)                                      ✓
//	return result
//
//	// Горутина 2 (запустилась почти одновременно):
//	Get("x")
//	mu.Lock()
//	e := cache["x"]        // нашли entry от горутины 1!
//	mu.Unlock()
//	<-e.ready              // ждем... ----->           ✓
//	return result                                       ✓
//
//	// Горутина 3 (запустилась после завершения горутины 1):
//	Get("x")
//	mu.Lock()
//	e := cache["x"]        // нашли entry с закрытым ready
//	mu.Unlock()
//	<-e.ready              // вернется сразу (канал закрыт) ✓
//	return result
//
// Результат: f("x") вызвана только ОДИН раз, все горутины получили результат!
func (memo *Memo) Get(key string) (value interface{}, err error) {
	memo.mu.Lock()
	e := memo.cache[key]
	if e == nil {
		// This is the first request for this key.
		// This goroutine becomes responsible for computing
		// the value and broadcasting the ready condition.
		//
		// Это первый запрос для данного ключа.
		// Эта горутина становится ответственной за вычисление значения
		// и оповещение всех ожидающих через закрытие канала ready.
		e = &entry{ready: make(chan struct{})}
		memo.cache[key] = e
		memo.mu.Unlock()

		// Вычисляем результат БЕЗ мьютекса
		// Другие горутины могут параллельно обрабатывать другие ключи!
		e.res.value, e.res.err = memo.f(key)

		// Оповещаем всех ожидающих горутин о готовности результата
		// Закрытие канала - это broadcast: все, кто ждет на <-e.ready,
		// немедленно разблокируются
		close(e.ready) // broadcast ready condition
	} else {
		// This is a repeat request for this key.
		//
		// Это повторный запрос для того же ключа.
		// Entry уже существует - значит кто-то уже вычисляет или вычислил.
		memo.mu.Unlock()

		// Ждем, пока первая горутина закроет канал ready.
		// Если ready уже закрыт - чтение вернется мгновенно.
		// Если еще нет - блокируемся до закрытия.
		<-e.ready // wait for ready condition
	}
	return e.res.value, e.res.err
}

// OMIT
