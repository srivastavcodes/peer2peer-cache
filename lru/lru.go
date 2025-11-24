package lru

import "container/list"

// LruCache is not safe for concurrent access.
type LruCache struct {
	// MaxEntries is the number of cache entries before an item is evicted.
	// Zero means no limit.
	MaxEntries int

	// OnEvicted specifies a callback function to be executed when an entry
	// is removed from cache.
	OnEvicted func(key Key, value any)

	// dll is a doubly linked list.
	dll *list.List

	// cache holds key-value pairs of list elements in memory until evicted.
	cache map[any]*list.Element
}

// Key may be a value that is comparable.
type Key any

type entry struct {
	key   Key
	value any
}

// NewLru creates a new LruCache. If maxEntries is zero then cache has no
// limit, and it's assumed the caller will handle eviction.
func NewLru(maxEntries int) *LruCache {
	return &LruCache{
		MaxEntries: maxEntries,
		dll:        list.New(),
		cache:      make(map[any]*list.Element),
	}
}

// Add adds a value to the cache.
func (lru *LruCache) Add(key Key, value any) {
	if lru.cache == nil {
		lru.cache = make(map[any]*list.Element)
		lru.dll = list.New()
	}
	if elem, ok := lru.cache[key]; ok {
		lru.dll.MoveToFront(elem)
		elem.Value.(*entry).value = value
		return
	}
	elem := lru.dll.PushFront(&entry{key, value})
	lru.cache[key] = elem
	if lru.MaxEntries != 0 && lru.dll.Len() > lru.MaxEntries {
		lru.RemoveOldest()
	}
}

// Get returns a key's value if exists.
func (lru *LruCache) Get(key Key) (val any, ok bool) {
	if lru.cache == nil {
		return
	}
	if elem, hit := lru.cache[key]; hit {
		lru.dll.MoveToFront(elem)
		val, ok = elem.Value.(*entry).value, true
	}
	return val, ok
}

// Remove removes the provided key from the cache.
func (lru *LruCache) Remove(key Key) {
	if lru.cache == nil {
		return
	}
	if elem, hit := lru.cache[key]; hit {
		lru.removeElement(elem)
	}
}

// RemoveOldest removes the oldest item from the cache.
func (lru *LruCache) RemoveOldest() {
	if lru.cache == nil {
		return
	}
	elem := lru.dll.Back()
	if elem != nil {
		lru.removeElement(elem)
	}
}

// removeElement removes the provided element from dll and cache.
// Calls OnEvicted if provided.
func (lru *LruCache) removeElement(elem *list.Element) {
	lru.dll.Remove(elem)
	ent := elem.Value.(*entry)

	delete(lru.cache, ent.key)
	if lru.OnEvicted != nil {
		lru.OnEvicted(ent.key, ent.value)
	}
}

// Len returns the count of elements in the cache.
func (lru *LruCache) Len() int {
	if lru.cache == nil {
		return 0
	}
	return lru.dll.Len()
}

// Clear purges the cache and calls OnEvicted on every cache entry.
func (lru *LruCache) Clear() {
	if lru.OnEvicted != nil {
		for _, elem := range lru.cache {
			ent := elem.Value.(*entry)
			lru.OnEvicted(ent.key, ent.value)
		}
	}
	lru.dll = nil
	lru.cache = nil
}
