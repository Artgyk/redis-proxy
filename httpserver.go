package main

import (
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"net/http"
	"time"
)

type HttpServer struct {
	Router      *gin.Engine
	RedisClient *redis.Client
	Cache       *LRUCache
}

type HttpServerConfig struct {
	RedisAddr     string
	CacheTtl      time.Duration
	CacheCapacity int
}

func NewHttpServer(config HttpServerConfig) *HttpServer {
	redisClient := redis.NewClient(&redis.Options{
		Addr: config.RedisAddr,
	})

	lruCache := NewLRUCache(config.CacheCapacity, config.CacheTtl)

	router := gin.Default()

	server := &HttpServer{Router: router, RedisClient: redisClient, Cache: lruCache}

	router.GET("/get/:key", server.GetKey)

	return server
}

func (httpServer *HttpServer) GetKey(c *gin.Context) {
	key := c.Param("key")

	if value, ok := httpServer.Cache.Get(key); ok {
		c.String(http.StatusOK, value.(string))
		return
	}

	value, err := httpServer.RedisClient.Get(key).Result()
	if err == redis.Nil {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	if err != nil {
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	httpServer.Cache.Add(key, value)

	c.String(http.StatusOK, value)
}

func (httpServer *HttpServer) Close() {
	httpServer.RedisClient.Close()
}
