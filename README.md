# Redis proxy

Http and TCP web server implementation on top of Redis instance. 
Server implements two protocol HTTP REST API with one command GET.
And tcp redis protocol for GET and PING commands.

Server has in memory LRU cache with ttl. 
Server first try to server request from in memory cache. 
If in memory cache is empty server makes request to Redis and 
cache result into memory cache.

### Complexity
LRU cache implementation has O(1) complexity 
Redis get has O(1) complexity
Overall web server GET command complexity O(1)

## Testing

Service has unit tests and end-to end tests.

Tests implemented using docker and docker-compose.

To run all tests:
```
make test

```


To run only unit tests without running docker compose:

Required golang installation

```
go mod download
go test ./... -v

```

## Run service

#### Configuration 

| KEY              | Description                                     |
|------------------|-------------------------------------------------|
| REDIS_ADDR       | address of backing redis server                 |
| HTTP_LISTEN_PORT | listen port for HTTP REST API                   |
| TCP_LISTEN_PORT  | listen port for TCP redis api                   |
| CACHE_TTL        | in memory cache ttl in seconds                  |
| CACHE_CAPACITY   | in memory cache capacity, number of stored keys |


#### Run using docker

docker compose file contains configuration to start up local redis and redis proxy server

```
    docker-compose up --build -d redis proxy
    
    //add something inside backing redis
    docker-compose exec redis redis-cli SET key1 val1
    
    //test http api
    curl http://0.0.0.0:8383/get/key1
    
    //test tcp api
    docker-compose exec redis redis-cli -h proxy -p 6379 GET key1

    //shutdown
    docker-compose down
```


### Features implementation

* Implemented features:
* Http and tcp api
* LRU cache with ttl
* Parallel concurrent processing

---
* Http server structure (including docker) ~ 1h 30m
* Lru ttl cache ~ 30 min
* Main logic and tests ~ 2h
* Tcp server ~ 2-3h