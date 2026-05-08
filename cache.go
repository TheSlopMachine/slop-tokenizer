package tokenizer

import "container/list"

// lruCache implements a simple LRU cache for BPE merge results
type lruCache struct {
	capacity int
	items    map[string]*list.Element
	list     *list.List
}

// cacheEntry represents a cached encoding result
type cacheEntry struct {
	key    string
	tokens []int
}

// newLRUCache creates a new LRU cache with the given capacity
func newLRUCache(capacity int) *lruCache {
	return &lruCache{
		capacity: capacity,
		items:    make(map[string]*list.Element, capacity),
		list:     list.New(),
	}
}

// get retrieves a value from the cache
func (c *lruCache) get(key string) ([]int, bool) {
	if elem, ok := c.items[key]; ok {
		c.list.MoveToFront(elem)
		return elem.Value.(*cacheEntry).tokens, true
	}
	return nil, false
}

// put adds a value to the cache
func (c *lruCache) put(key string, tokens []int) {
	// If key exists, update and move to front
	if elem, ok := c.items[key]; ok {
		c.list.MoveToFront(elem)
		elem.Value.(*cacheEntry).tokens = tokens
		return
	}

	// If at capacity, remove oldest
	if c.list.Len() >= c.capacity {
		oldest := c.list.Back()
		if oldest != nil {
			c.list.Remove(oldest)
			delete(c.items, oldest.Value.(*cacheEntry).key)
		}
	}

	// Add new entry
	entry := &cacheEntry{key: key, tokens: tokens}
	elem := c.list.PushFront(entry)
	c.items[key] = elem
}
