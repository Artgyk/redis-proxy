package main

import (
	"github.com/go-redis/redis"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"net"
	"testing"
)

type TcpServerTestSuite struct {
	suite.Suite
	RedisClient *redis.Client
	TcpServer   *TcpServer
	Listener    net.Listener
	Addr        string
}

func (suite *TcpServerTestSuite) SetupTest() {
	config := HttpServerConfigFromEnv()

	suite.Addr = ":7379"

	redisClient := redis.NewClient(&redis.Options{Addr: config.RedisAddr})
	cache := NewLRUCache(config.CacheCapacity, config.CacheTtl)

	server := NewTcpServer(redisClient, cache)
	suite.Listener = server.Serve(suite.Addr)
	suite.RedisClient = redisClient

	redisClient.FlushAll()
}

func (suite *TcpServerTestSuite) TearDownTest() {
	suite.Listener.Close()
	suite.RedisClient.Close()
}

func (suite *TcpServerTestSuite) TestPing() {
	t := suite.T()
	redisClient := redis.NewClient(&redis.Options{Addr: suite.Addr})

	actual, err := redisClient.Ping().Result()
	if assert.NoError(t, err) {
		assert.Equal(t, "PONG", actual)
	}
}

func (suite *TcpServerTestSuite) TestGetCommand_KeyExists() {
	t := suite.T()
	err := suite.RedisClient.Set("key10", "val1", 0).Err()
	if !assert.NoError(t, err) {
		return
	}

	redisClient := redis.NewClient(&redis.Options{Addr: suite.Addr})

	actual, err := redisClient.Get("key10").Result()
	if assert.NoError(t, err) {
		assert.Equal(t, "val1", actual)
	}
}

func (suite *TcpServerTestSuite) TestGetCommand_KeyNotExists() {
	t := suite.T()

	redisClient := redis.NewClient(&redis.Options{Addr: suite.Addr})

	_, err := redisClient.Get("key11").Result()
	assert.Equal(t, redis.Nil, err)
}

func (suite *TcpServerTestSuite) TestGetCommand_GetFromCache() {
	t := suite.T()
	redisClient := redis.NewClient(&redis.Options{Addr: suite.Addr})

	err := suite.RedisClient.Set("key10", "val1", 0).Err()
	if !assert.NoError(t, err) {
		return
	}
	_, err = redisClient.Get("key10").Result()
	if !assert.NoError(t, err) {
		return
	}

	suite.RedisClient.FlushAll()

	actual, err := redisClient.Get("key10").Result()
	if assert.NoError(t, err) {
		assert.Equal(t, "val1", actual)
	}
}

func (suite *TcpServerTestSuite) TestGetCommand_KeyHasWhitespace() {
	t := suite.T()
	redisClient := redis.NewClient(&redis.Options{Addr: suite.Addr})

	err := suite.RedisClient.Set("key\n10", "val1", 0).Err()
	if !assert.NoError(t, err) {
		return
	}

	actual, err := redisClient.Get("key\n10").Result()
	if assert.NoError(t, err) {
		assert.Equal(t, "val1", actual)
	}
}

func (suite *TcpServerTestSuite) TestGetCommand_NilValue() {
	t := suite.T()
	redisClient := redis.NewClient(&redis.Options{Addr: suite.Addr})

	err := suite.RedisClient.Set("key1", nil, 0).Err()
	if !assert.NoError(t, err) {
		return
	}

	actual, err := redisClient.Get("key1").Result()
	if assert.NoError(t, err) {
		// redis driver return empty string :(
		assert.Equal(t, "", actual)
	}
}

func (suite *TcpServerTestSuite) TestGetCommand_NotImplementedCommand() () {
	t := suite.T()
	redisClient := redis.NewClient(&redis.Options{Addr: suite.Addr})

	err := redisClient.Set("k1", "val", 0).Err()

	assert.Error(t, err)
}

func TestTcpServerSuite(t *testing.T) {
	if env("DOCKER", "false") == "true" {
		suite.Run(t, new(TcpServerTestSuite))
	} else {
		t.Skip("Should be inside docker environment")
	}
}
