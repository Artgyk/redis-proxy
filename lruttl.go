package main

//I didn't find simple thread-safe lru cache with ttl.
//Based on https://github.com/dualinventive/go-lruttl

import (
	"container/list"
	"sync"
	"time"
)

var now = time.Now

type LRUCache struct {
	// MaxEntries is the maximum number of cache entries before
	// an item is evicted. Zero means no limit.
	MaxEntries int

	ll     *list.List
	cache  map[interface{}]*list.Element
	expiry time.Duration

	lock sync.Mutex
}

// A Key may be any value that is comparable. See http://golang.org/ref/spec#Comparison_operators
type Key interface{}

type entry struct {
	key   Key
	ttl   time.Time
	value interface{}
}

// NewLRUCache creates a new Cache.
// If maxEntries is zero, the cache has no limit and it's assumed
// that eviction is done by the caller.
func NewLRUCache(maxEntries int, expiry time.Duration) *LRUCache {
	return &LRUCache{
		MaxEntries: maxEntries,
		ll:         list.New(),
		cache:      make(map[interface{}]*list.Element),
		expiry:     expiry,
	}
}

// Add adds a value to the cache.
func (c *LRUCache) Add(key Key, value interface{}) {
	c.lock.Lock()

	defer c.lock.Unlock()

	if ee, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ee)
		ee.Value.(*entry).ttl = now().Add(c.expiry)
		ee.Value.(*entry).value = value
		return
	}
	ele := c.ll.PushFront(&entry{key, now().Add(c.expiry), value})
	c.cache[key] = ele
	if c.MaxEntries != 0 && c.ll.Len() > c.MaxEntries {
		c.removeOldest()
	}
}

// Get looks up a key's value from the cache.
func (c *LRUCache) Get(key Key) (value interface{}, ok bool) {
	c.lock.Lock()
	defer c.lock.Unlock()

	if ele, hit := c.cache[key]; hit {
		c.ll.MoveToFront(ele)
		if now().After(ele.Value.(*entry).ttl) {
			c.remove(key)
			return
		}
		return ele.Value.(*entry).value, true
	}
	return
}

func (c *LRUCache) remove(key Key) {
	if ele, hit := c.cache[key]; hit {
		c.removeElement(ele)
	}
}

func (c *LRUCache) removeOldest() {
	ele := c.ll.Back()
	if ele != nil {
		c.removeElement(ele)
	}
}

func (c *LRUCache) removeElement(e *list.Element) {
	c.ll.Remove(e)
	kv := e.Value.(*entry)
	delete(c.cache, kv.key)
}
