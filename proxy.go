package main

import (
	"errors"
	"github.com/go-redis/redis"
)

var ErrKeyNotFound = errors.New("redis-proxy: key not found")

type Proxy struct {
	redisClient *redis.Client
	lruCache    *LRUCache
}

func NewProxy(redisClient *redis.Client, lruCache *LRUCache) *Proxy {
	return &Proxy{redisClient: redisClient, lruCache: lruCache}
}

func (proxy *Proxy) Get(key string) (string, error) {

	if value, ok := proxy.lruCache.Get(key); ok {
		return value.(string), nil
	}

	value, err := proxy.redisClient.Get(key).Result()
	if err == redis.Nil {
		return "", ErrKeyNotFound
	}
	if err != nil {
		return "", err
	}

	proxy.lruCache.Add(key, value)

	return value, err
}
