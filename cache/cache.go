package cache

import (
    "container/list"
    "errors"
    "sync"
    "sync/atomic"
)

// Stats holds cache performance metrics
type Stats struct {
    Hits   uint64
    Misses uint64
    Size   int
}

// Cache represents an in-memory key-value cache with LRU eviction
type Cache struct {
    mutex     sync.RWMutex
    items     map[string]string
    lruList   *list.List
    lruMap    map[string]*list.Element
    maxItems  int
    stats     atomic.Value // Use atomic.Value for stats to reduce lock contention
}

// Add these constants
const (
    maxKeyLength   = 256
    maxValueLength = 256
)

// NewCache creates a new Cache with the specified maximum number of items
func NewCache(maxItems int) *Cache {
    c := &Cache{
        items:    make(map[string]string, maxItems), // Pre-allocate map capacity
        lruList:  list.New(),
        lruMap:   make(map[string]*list.Element, maxItems),
        maxItems: maxItems,
    }
    c.stats.Store(&Stats{})
    return c
}

// Put adds or updates a key-value pair in the cache
func (c *Cache) Put(key, value string) error {
    if len(key) == 0 || len(key) > maxKeyLength {
        return errors.New("key length must be between 1 and 256 characters")
    }
    if len(value) > maxValueLength {
        return errors.New("value length must not exceed 256 characters")
    }

    c.mutex.Lock()
    defer c.mutex.Unlock()

    // If key already exists, update its value
    if _, exists := c.items[key]; exists {
        c.items[key] = value
        if elem, ok := c.lruMap[key]; ok {
            c.lruList.MoveToFront(elem)
        }
        return nil
    }

    // Check capacity before adding
    if c.lruList.Len() >= c.maxItems {
        c.evictLRU()
    }

    // Add new key-value pair
    c.items[key] = value
    elem := c.lruList.PushFront(key)
    c.lruMap[key] = elem
    return nil
}

// Get retrieves a value from the cache by key
func (c *Cache) Get(key string) (string, bool) {
    if len(key) > maxKeyLength {
        return "", false
    }

    c.mutex.RLock()
    value, exists := c.items[key]
    if !exists {
        c.mutex.RUnlock()
        stats := c.stats.Load().(*Stats)
        atomic.AddUint64(&stats.Misses, 1)
        return "", false
    }

    // Update LRU under the same lock to prevent race conditions
    if elem, ok := c.lruMap[key]; ok {
        c.lruList.MoveToFront(elem)
    }
    c.mutex.RUnlock()

    stats := c.stats.Load().(*Stats)
    atomic.AddUint64(&stats.Hits, 1)
    return value, true
}

// GetStats returns statistics about the cache
func (c *Cache) GetStats() Stats {
    stats := c.stats.Load().(*Stats)
    return Stats{
        Hits:   atomic.LoadUint64(&stats.Hits),
        Misses: atomic.LoadUint64(&stats.Misses),
        Size:   len(c.items),
    }
}

// evictLRU removes the least recently used item from the cache
func (c *Cache) evictLRU() {
    // Get the oldest element (back of the list)
    elem := c.lruList.Back()
    if elem == nil {
        return
    }

    // Remove it from the list and maps
    key := elem.Value.(string)
    c.lruList.Remove(elem)
    delete(c.lruMap, key)
    delete(c.items, key)
}

// Add batch operations for better throughput
func (c *Cache) PutBatch(pairs map[string]string) {
    c.mutex.Lock()
    defer c.mutex.Unlock()

    for key, value := range pairs {
        if _, exists := c.items[key]; exists {
            c.items[key] = value
            c.lruList.MoveToFront(c.lruMap[key])
            continue
        }

        c.items[key] = value
        elem := c.lruList.PushFront(key)
        c.lruMap[key] = elem

        if c.lruList.Len() > c.maxItems {
            c.evictLRU()
        }
    }
}
