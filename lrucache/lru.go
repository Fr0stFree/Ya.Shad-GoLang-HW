//go:build !solution

package lrucache

import "container/list"

type LRUCache struct {
	capacity int
	storage  map[int]*list.Element
	order    *list.List
}

type CacheItem struct {
	key   int
	value int
}

func (c *LRUCache) Set(key, value int) {
	if c.capacity == 0 {
		return
	}
	if element, exists := c.storage[key]; exists {
		element.Value = CacheItem{key: key, value: value}
		c.order.MoveToFront(element)
		return
	}
	if c.order.Len() >= c.capacity {
		last := c.order.Back()
		if last != nil {
			lastItem := last.Value.(CacheItem)
			delete(c.storage, lastItem.key)
			c.order.Remove(last)
		}
	}
	item := CacheItem{key: key, value: value}
	element := c.order.PushFront(item)
	c.storage[key] = element
}

func (c *LRUCache) Get(key int) (int, bool) {
	element, exists := c.storage[key]
	if !exists {
		return 0, false
	}
	c.order.MoveToFront(element)
	return element.Value.(CacheItem).value, true
}

func (c *LRUCache) Clear() {
	c.storage = make(map[int]*list.Element, c.capacity)
	c.order.Init()
}

func (c *LRUCache) Range(f func(key, value int) bool) {
	for element := c.order.Back(); element != nil; element = element.Prev() {
		item := element.Value.(CacheItem)
		if !f(item.key, item.value) {
			return
		}
	}
}

func New(cap int) Cache {
	return &LRUCache{
		capacity: cap,
		storage:  make(map[int]*list.Element),
		order:    list.New(),
	}
}
