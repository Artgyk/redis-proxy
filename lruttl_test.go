package main

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestLRUCache_Get_Exists(t *testing.T) {
	lru := NewLRUCache(10, time.Hour)
	lru.Add("key1", "val1")

	actual, ok := lru.Get("key1")

	assert.True(t, ok, "Key not exists in cache")
	assert.Equal(t, "val1", actual)
}

func TestLRUCache_Get_NotExists(t *testing.T) {
	lru := NewLRUCache(10, time.Hour)
	lru.Add("key1", "val1")

	_, exists := lru.Get("key2")

	assert.False(t, exists, "Key shouldn't exists in cache")
}

func TestLRUCache_MaxEntries(t *testing.T) {
	lru := NewLRUCache(10, time.Hour)

	keyFunc := func(i int) string { return fmt.Sprintf("key%d", i) }
	valFunc := func(i int) string { return fmt.Sprintf("val%d", i) }
	// Act
	for i := 0; i < 15; i++ {
		lru.Add(keyFunc(i), valFunc(i))
	}

	//Assert

	for i := 0; i < 5; i++ {
		key := keyFunc(i)
		_, ok := lru.Get(key)
		assert.False(t, ok, "Key %s should be evicted", key)
	}

	for i := 5; i < 15; i++ {
		key := keyFunc(i)
		val, ok := lru.Get(key)
		assert.True(t, ok, "Key %s should exists", key)
		assert.Equal(t, valFunc(i), val)
	}
}

func TestLRUCache_TTL(t *testing.T) {
	currentTime := time.Now()
	freezeTime(currentTime)
	defer resetTime()

	lru := NewLRUCache(10, time.Hour)

	lru.Add("key1", "val1")

	freezeTime(currentTime.Add(30 * time.Minute))

	actualVal, ok := lru.Get("key1")
	assert.True(t, ok, "Key shouldn'b be expired")
	assert.Equal(t, "val1", actualVal)

	freezeTime(currentTime.Add(61 * time.Minute))

	_, ok = lru.Get("key1")
	assert.False(t, ok, "Key should be expired")
}

func freezeTime(t time.Time) {
	now = func() time.Time {
		return t
	}
}

func resetTime() {
	now = time.Now
}
