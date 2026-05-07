package cache

import (
	"container/list"
	"sync"
)

// entry represents a key-value pair stored in the LRU cache.
type entry[K comparable, V any] struct {
	key   K
	value V
}

// LRUCache is a generic, thread-safe LRU cache.
type LRUCache[K comparable, V any] struct {
	mu       sync.RWMutex
	capacity int
	items    map[K]*list.Element
	order    *list.List
}

// NewLRU creates a new LRUCache with the specified capacity.
func NewLRU[K comparable, V any](capacity int) *LRUCache[K, V] {
	return &LRUCache[K, V]{
		capacity: capacity,
		items:    make(map[K]*list.Element),
		order:    list.New(),
	}
}

// Get retrieves a value from the cache. The second return value indicates
// whether the key was found.
func (c *LRUCache[K, V]) Get(key K) (V, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, ok := c.items[key]; ok {
		c.order.MoveToFront(elem)
		return elem.Value.(*entry[K, V]).value, true
	}

	var zero V
	return zero, false
}

// Put inserts or updates a key-value pair in the cache.
func (c *LRUCache[K, V]) Put(key K, value V) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.capacity <= 0 {
		return
	}

	if elem, ok := c.items[key]; ok {
		// Update existing entry and move to front
		elem.Value.(*entry[K, V]).value = value
		c.order.MoveToFront(elem)
		return
	}

	// Add new entry
	e := &entry[K, V]{key: key, value: value}
	elem := c.order.PushFront(e)
	c.items[key] = elem

	// Evict oldest if over capacity
	if c.order.Len() > c.capacity {
		oldest := c.order.Back()
		if oldest != nil {
			c.order.Remove(oldest)
			oldEntry := oldest.Value.(*entry[K, V])
			delete(c.items, oldEntry.key)
		}
	}
}

// Remove deletes a key from the cache.
func (c *LRUCache[K, V]) Remove(key K) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, ok := c.items[key]; ok {
		c.order.Remove(elem)
		delete(c.items, key)
	}
}

// Len returns the current number of items in the cache.
func (c *LRUCache[K, V]) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.order.Len()
}
