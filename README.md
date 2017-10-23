# redis-proxy

A simple in-memory LRU cache that implements the Redis RESP protocol.

## Usage

```
A simple in-memory Redis proxy that supports the RESP protocol.

Usage:
  redis-proxy [flags]

Flags:
      --cache_period int        The periodicity of the cache eviction thread, in milliseconds. (default 100)
      --cache_ttl int           A global TTL for cache entries, in milliseconds. (default 300000)
      --capacity int            The maximum number of entries to cache. (default 1024)
      --config string           config file
  -h, --help                    help for redis-proxy
      --port int                A open port used for listening. (default 8001)
      --redis_database int      The redis database to use. See https://redis.io/commands/select.
      --redis_hostname string   The hostname for the backing redis cache. (default "localhost:6379")
      --redis_password string   The password for the backing redis cache.
```

redis-proxy uses the Viper and Cobra libraries to provide configuration and CLI support. Environment variables and config files are supported,
see the Cobra documentation.

### Docker

A docker-compose file is setup for the project. Running `make run` will bring
a cluster up with `redis` and `redis-proxy` running. `redis-proxy` runs on port 8001
by default.

* `make run` aliases docker-compose, it will bring up a container for redis and redis-proxy
* `make test` will run all unit tests and end to end tests.
* `make dev-run` will build a new version of the docker image and run it via compose.

I'm new to Golang and haven't quite figured out how to build a docker image without cloning the repo via go get, so building will not reflect local changes.

### Redis CLI

Very basic Redis CLI commands will work against redis-proxy. As of this time `PING` and `GET` are supported.

```
00:47 $ redis-cli -p 8001 # or you can use docker to launch the cli
127.0.0.1:8001> ping
"PONG"
127.0.0.1:8001> GET X
(nil)
127.0.0.1:8001> get x
"true" # x existed in the persisted Redis
```

# Design
The redis-proxy is very basic, at the core it's a TCP server that handles the RESP (Redis Serialization Protocol). The current implementation only supports GET and PING. GET is backed by a LRU cache with a global TTL mechanism; on cache misses the GET request will be delegated to the backing Redis and populated by the response.

The Server/Command/Handler handoff is supposed to encourage modularity when supporting multiple commands in some future context. I initially started with a callback mechanism but needed a way to inject logic into the pre-command-process step too. For simplicity I removed the callback, but
having some kind of pre/post hooks would probably work well if this was to be used as a flexible redis proxy for arbitrary business logic.

Assumptions:
* Simple string keys and values. Future improvements can be made to the server to extend support for other types.
* Only GET and PING commands supported. Supporting other commands requires more complex serialization of RESP types.
* Assuming _well_ formed input. The parser is fragile as is.
* For missing keys, we return a nil response.

I built this in about a day of work ~(8 hours). My time was spent about 2 hours reading up on RESP, implementing and testing the cache. 2 more hours familiarizing myself with Cobra, Viper, PFlags, and dockerizing Go binaries. 2-3 hours debugging/implementing a basic RESP parser and getting it working with the redis-go library.

## Cache
Under the covers is a LRU cache with a TTL mechanism driven by a 'redeemer' coroutine. The LRU is implemented
using a doubly linked list, a time-ordered queue and a hashmap; all guarded by a sync.Mutex.

The redeemer coroutine polls the time-ordered queue with a given periodicity. If an element is found to be expired,
it is removed from the queue and then atomicly removed from cache.

## Server
The server handles connections in parallel, each new connection being handled by a new Go routine.

The request is parsed into a ordered list of RESP BulkStrings (called a `Command`) and then checked against
a set of handlers. If a handler is available, the request is processed. Processing a request is dependent on the
command being executed, but is essentially a function to apply side effects to the cache and delegate behavior to
the underlying Redis instance. To query redis from the server, we actually use the `redis-go` library, as it supports meta-commands that are required for the full Redis protocol.

The cache simply ignores Redis connectivity issues currently. All requests will be served by the cache in the event that the
backing redis becomes unreachable.

Improvements:

* [ ] Support Redis commands with multi-word names. I didn't realize there were commands with multiple parts.
* [ ] Handle malformed input.
* [ ] Support more complex RESP types.
* [ ] Add a Redis healthcheck.
* [ ] Add instrumentation.


## Logging

I strongly believe in structured logging and decided to use Zap, it worked pretty well. There are still some places to clean up in the code that use `fmt`.

## Metrics

I'd like to implement Prometheus as a side-car container and emit timing and counter metrics but I've run out of time on this project.
