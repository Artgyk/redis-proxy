package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type ApiHttpTestSuite struct {
	suite.Suite
	Router      *gin.Engine
	RedisClient *redis.Client
	HttpServer  *HttpServer
}

func (suite *ApiHttpTestSuite) SetupTest() {
	server := NewHttpServer(HttpServerConfigFromEnv())
	suite.HttpServer = server
	suite.Router = server.Router
	suite.RedisClient = server.RedisClient
	server.RedisClient.FlushAll()
}

func (suite *ApiHttpTestSuite) TearDownTest() {
	suite.HttpServer.Close()
}

func (suite *ApiHttpTestSuite) TestHttpServer_GetKey_KeyNotExists_Return_404() {
	t := suite.T()
	router := suite.Router

	code, _ := request("/get/k1", router)

	assert.Equal(t, 404, code)
}

func (suite *ApiHttpTestSuite) TestHttpServer_GetKey_KeyExistsInRedis_Return_Value() {
	t := suite.T()
	router := suite.Router
	err := suite.RedisClient.Set("k2", "val", 0).Err()
	if !assert.NoError(t, err) {
		return
	}

	code, body := request("/get/k2", router)

	assert.Equal(t, 200, code)
	assert.Equal(t, "val", body)
}

func (suite *ApiHttpTestSuite) TestHttpServer_GetKey_ReturnValue_From_Cache() {
	t := suite.T()
	router := suite.Router
	err := suite.RedisClient.Set("k3", "val", 0).Err()
	if !assert.NoError(t, err) {
		return
	}

	code, body := request("/get/k3", router)

	assert.Equal(t, 200, code)
	assert.Equal(t, "val", body)

	suite.RedisClient.FlushAll()

	code, body = request("/get/k3", router)

	assert.Equal(t, 200, code)
	assert.Equal(t, "val", body)
}

func (suite *ApiHttpTestSuite) TestHttpServer_GetKey_TTLExpired_Return_404() {
	defer resetTime()

	t := suite.T()
	router := suite.Router
	err := suite.RedisClient.Set("k4", "val", 0).Err()
	if !assert.NoError(t, err) {
		return
	}

	code, body := request("/get/k4", router)

	assert.Equal(t, 200, code)
	assert.Equal(t, "val", body)

	suite.RedisClient.FlushAll()

	freezeTime(time.Now().Add(time.Hour))
	code, body = request("/get/k4", router)

	assert.Equal(t, 404, code)
}

func (suite *ApiHttpTestSuite) TestHttpServer_GetKey_CacheCapacity() {
	defer resetTime()

	t := suite.T()
	router := suite.Router

	for i := 0; i < 15; i++ {
		key := fmt.Sprintf("key%d", i)
		err := suite.RedisClient.Set(key, "val", 0).Err()
		if !assert.NoError(t, err) {
			return
		}
		code, _ := request(fmt.Sprintf("/get/%s", key), router)
		assert.Equal(t, 200, code)
	}
	suite.RedisClient.FlushAll()

	code, _ := request("/get/key1", router)
	assert.Equal(t, 404, code)

	suite.RedisClient.FlushAll()

	code, body := request("/get/key11", router)

	assert.Equal(t, 200, code)
	assert.Equal(t, "val", body)
}

func request(url string, router *gin.Engine) (code int, body string) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", url, nil)
	router.ServeHTTP(w, req)
	code = w.Code
	body = w.Body.String()
	return code, body
}

func TestHttpServerSuite(t *testing.T) {
	if env("DOCKER", "false") == "true" {
		suite.Run(t, new(ApiHttpTestSuite))
	} else {
		t.Skip("Should be inside docker environment")
	}
}
