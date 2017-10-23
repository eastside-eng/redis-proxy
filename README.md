# redis-proxy
A simple in-memory LRU cache that implements the Redis GET protocol.

## Usage

### Docker

A docker-compose file is setup for the project. Running `make run` will bring
a cluster up with `redis` and `redis-proxy` running. `redis-proxy` runs on port 8001
by default.

### Redis CLI

Very basic Redis CLI commands will work against redis-proxy.

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
The redis-proxy is very basic, at the core it's a TCP server that handles the RESP (Redis Serialization Protocol).

Assumptions:
* Simple string keys and values. Future improvements can be made to the server to extend support for other types.
* Only GET and PING commands supported. Supporting other commands requires more complex serialization of RESP types.
* Assuming _well_ formed input. The parser is fragile as is.

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
the underlying Redis instance. To query redis from the server, we actually use the `redis-go` library, as it supports meta-commands
that are required for the full Redis protocol.

The cache simply ignores Redis connectivity issues currently. All requests will be served by the cache in the event that the
backing redis becomes unreachable.

Improvements:

* [ ] Support Redis commands with multi-word names.
* [ ] Handle malformed input.
* [ ] Support more complex RESP types.
* [ ] Add a Redis healthcheck.
* [ ] Add instrumentation.
