package main

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

func env(key string, defValue string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	if defValue == "" {
		panic(fmt.Errorf("%q not configurated", key))
	}
	return defValue
}

func envInt(key string, defValue string) int {
	strVal := env(key, defValue)

	val, err := strconv.Atoi(strVal)
	if err != nil {
		panic(err)
	}
	return val
}

func envDuration(key string, defValue string) time.Duration {
	duration := envInt(key, defValue)
	return time.Duration(duration) * time.Second
}

func HttpServerConfigFromEnv() HttpServerConfig {
	return HttpServerConfig{
		env("REDIS_ADDR", ""),
		envDuration("CACHE_TTL", "60"),
		envInt("CACHE_CAPACITY", "10"),
	}
}
