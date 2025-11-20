// Copyright © 2016 Alan A. A. Donovan & Brian W. Kernighan.
// License: https://creativecommons.org/licenses/by-nc-sa/4.0/

// See page 272.

//!+

// Package memo provides a concurrency-unsafe
// memoization of a function of type Func.
//
// ПРОБЛЕМА: Этот код содержит DATA RACE и не должен использоваться
// в многопоточной среде!
//
// Мемоизация - это техника кеширования результатов дорогих вычислений.
// Вместо того чтобы каждый раз вычислять результат для одного и того же
// ключа, мы сохраняем его при первом вызове и возвращаем из кеша при
// последующих обращениях.
//
// Пример использования:
//
//	func expensiveComputation(key string) (interface{}, error) {
//	    time.Sleep(2 * time.Second) // имитация долгой работы
//	    return fmt.Sprintf("result for %s", key), nil
//	}
//
//	memo := &Memo{
//	    f:     expensiveComputation,
//	    cache: make(map[string]result),
//	}
//
//	// Первый вызов займет 2 секунды
//	v1, _ := memo.Get("key1")
//
//	// Второй вызов с тем же ключом вернется мгновенно из кеша
//	v2, _ := memo.Get("key1")
//
// ВНИМАНИЕ: Если два горутина одновременно вызовут Get с одним и тем же
// ключом, произойдет data race при чтении и записи в map без
// синхронизации!
package memo

// A Memo caches the results of calling a Func.
// Memo кеширует результаты вызова функции Func.
//
// ВАЖНО: Эта структура НЕ безопасна для одновременного использования
// из нескольких горутин. Обращение к map без синхронизации приведет
// к data race.
type Memo struct {
	f     Func              // функция для мемоизации
	cache map[string]result // кеш: ключ -> результат
}

// Func is the type of the function to memoize.
// Func - тип функции, которую мы кешируем.
// Функция принимает строковый ключ и возвращает произвольное значение
// и ошибку (если вычисление не удалось).
type Func func(key string) (interface{}, error)

// result хранит кешированный результат вызова функции.
type result struct {
	value interface{} // возвращаемое значение
	err   error       // ошибка, если она произошла
}

// Get возвращает кешированный результат для данного ключа или вычисляет
// его при первом обращении.
//
// NOTE: not concurrency-safe!
// ВНИМАНИЕ: Эта функция НЕ безопасна для конкурентного использования!
//
// Проблема: Если несколько горутин одновременно вызовут Get():
//  1. DATA RACE при чтении map (res, ok := memo.cache[key])
//  2. DATA RACE при записи в map (memo.cache[key] = res)
//  3. Программа может упасть с "concurrent map read and write"
//
// Пример проблемы:
//
//	// Горутина 1                    // Горутина 2
//	memo.Get("x")                    memo.Get("x")
//	res, ok := memo.cache["x"]       res, ok := memo.cache["x"]
//	// ok == false                   // ok == false
//	res.value, res.err = memo.f("x") res.value, res.err = memo.f("x")
//	memo.cache["x"] = res            memo.cache["x"] = res
//	// DATA RACE! ^^^                // DATA RACE! ^^^
//
// Результат: паника "concurrent map writes" или повреждение данных.
//
// Смотрите memo2, memo3, memo4 для решений этой проблемы.
func (memo *Memo) Get(key string) (interface{}, error) {
	res, ok := memo.cache[key]
	if !ok {
		// Ключа нет в кеше - вычисляем значение
		res.value, res.err = memo.f(key)
		// Сохраняем результат в кеш для последующих вызовов
		memo.cache[key] = res
	}
	return res.value, res.err
}

// OMIT
