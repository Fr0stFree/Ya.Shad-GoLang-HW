// Copyright © 2016 Alan A. A. Donovan & Brian W. Kernighan.
// License: https://creativecommons.org/licenses/by-nc-sa/4.0/

// See page 275.

// Package memo provides a concurrency-safe memoization a function of
// type Func.  Concurrent requests are serialized by a Mutex.
//
// РЕШЕНИЕ №1: Защищаем весь кеш мьютексом.
//
// ПЛЮСЫ:
//   - Простое и понятное решение
//   - Полностью безопасно для concurrent использования
//   - Нет data race
//
// МИНУСЫ:
//   - Низкая производительность: все запросы выполняются последовательно
//   - Даже запросы к разным ключам блокируют друг друга
//   - Если функция f() медленная, все горутины будут ждать
//
// Пример:
//
//	memo := New(func(key string) (interface{}, error) {
//	    time.Sleep(2 * time.Second) // долгая операция
//	    return fmt.Sprintf("result: %s", key), nil
//	})
//
//	// Эти вызовы будут выполняться ПОСЛЕДОВАТЕЛЬНО:
//	go memo.Get("key1") // займет мьютекс на 2+ секунды
//	go memo.Get("key2") // будет ждать освобождения мьютекса
//	go memo.Get("key3") // тоже будет ждать
//
//	// Итого: 6+ секунд вместо возможных 2 секунд при параллельном выполнении
//
// Смотрите memo3 для улучшенного решения.
package memo

import "sync"

// Func is the type of the function to memoize.
// Func - тип функции для мемоизации.
type Func func(string) (interface{}, error)

// result хранит результат вычисления функции.
type result struct {
	value interface{} // возвращаемое значение
	err   error       // ошибка вычисления
}

// New создает новый Memo с заданной функцией для мемоизации.
func New(f Func) *Memo {
	return &Memo{f: f, cache: make(map[string]result)}
}

//!+

// Memo - потокобезопасный кеш для мемоизации.
type Memo struct {
	f     Func
	mu    sync.Mutex        // защищает cache от одновременного доступа
	cache map[string]result // кеш результатов
}

// Get возвращает кешированный результат или вычисляет его.
//
// Get is concurrency-safe.
// Эта версия безопасна для concurrent использования, НО неэффективна.
//
// Проблема производительности:
//
// Мьютекс захватывается на ВСЁ время работы функции:
//  1. Захватываем мьютекс
//  2. Проверяем кеш
//  3. Если нет в кеше - ВЫЗЫВАЕМ МЕДЛЕННУЮ ФУНКЦИЮ f() ПОД МЬЮТЕКСОМ
//  4. Сохраняем результат
//  5. Освобождаем мьютекс
//
// Это означает, что даже запросы к РАЗНЫМ ключам будут блокировать
// друг друга:
//
//	// Горутина 1         // Горутина 2         // Горутина 3
//	memo.Get("a")        memo.Get("b")        memo.Get("c")
//	mu.Lock() ✓
//	f("a") [2 sec]       mu.Lock() ⏳        mu.Lock() ⏳
//	mu.Unlock()          (ждет...)            (ждет...)
//	                     mu.Lock() ✓
//	                     f("b") [2 sec]       mu.Lock() ⏳
//	                     mu.Unlock()          (ждет...)
//	                                          mu.Lock() ✓
//	                                          f("c") [2 sec]
//	                                          mu.Unlock()
//
// Итого: 6 секунд вместо возможных 2 секунд параллельно!
//
// Решение: смотрите memo3, где мьютекс удерживается только на время
// работы с map, а вычисления выполняются параллельно.
func (memo *Memo) Get(key string) (value interface{}, err error) {
	memo.mu.Lock()
	res, ok := memo.cache[key]
	if !ok {
		// ПРОБЛЕМА: Вызываем долгую функцию f() под мьютексом!
		// Все остальные горутины будут заблокированы на это время.
		res.value, res.err = memo.f(key)
		memo.cache[key] = res
	}
	memo.mu.Unlock()
	return res.value, res.err
}

// OMIT
