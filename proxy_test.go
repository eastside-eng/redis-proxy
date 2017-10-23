package main

import (
	"fmt"
	"testing"
	"time"

	"github.com/go-redis/redis"
	"github.com/stretchr/testify/assert"
)

func TestEndToEnd(t *testing.T) {
	proxy := redis.NewClient(&redis.Options{
		Addr:     "localhost:8001",
		Password: "",
		DB:       0,
	})

	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	// Ensure both are up before starting
	assert.Equal(t, "PONG", proxy.Ping().Val())
	assert.Equal(t, "PONG", client.Ping().Val())

	// Write to Redis
	assert.Equal(t, "OK", client.Set("test", "123", time.Minute*10).Val())

	// Read from Cache
	assert.Equal(t, "123", proxy.Get("test").Val())

	// Update Redis
	assert.Equal(t, "OK", client.Set("test", "321", time.Minute*10).Val())

	// Read Stale data from Cache
	assert.Equal(t, "123", proxy.Get("test").Val())
}

func clientWrapper(client *redis.Client, key string, out chan string) {
	for i := 0; i < 1000; i++ {
		out <- client.Get(key).Val()
	}
}

func TestMultipleClients(t *testing.T) {
	proxy := redis.NewClient(&redis.Options{
		Addr:     "localhost:8001",
		Password: "",
		DB:       0,
	})

	proxy2 := redis.NewClient(&redis.Options{
		Addr:     "localhost:8001",
		Password: "",
		DB:       0,
	})

	proxy3 := redis.NewClient(&redis.Options{
		Addr:     "localhost:8001",
		Password: "",
		DB:       0,
	})

	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	// Ensure both are up before starting
	assert.Equal(t, "PONG", proxy.Ping().Val())
	assert.Equal(t, "PONG", proxy2.Ping().Val())
	assert.Equal(t, "PONG", client.Ping().Val())

	// Write to Redis
	assert.Equal(t, "OK", client.Set("test", "123", time.Minute*10).Val())

	// Read data from Cache
	out := make(chan string)
	go clientWrapper(proxy, "test", out)
	go clientWrapper(proxy2, "test", out)
	go clientWrapper(proxy3, "test", out)
	for i := 0; i < 3000; i++ {
		assert.Equal(t, "123", <-out)
	}
}

func TestLRU(t *testing.T) {
	proxy := redis.NewClient(&redis.Options{
		Addr:     "localhost:8001",
		Password: "",
		DB:       0,
	})

	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	// Ensure both are up before starting
	assert.Equal(t, "PONG", proxy.Ping().Val())
	assert.Equal(t, "PONG", client.Ping().Val())

	// Write to Redis
	for i := 0; i < 2048; i++ {
		key := fmt.Sprintf("test%d", i)
		assert.Equal(t, "OK", client.Set(key, "123", time.Minute*10).Val())
		// Read from Cache
		assert.Equal(t, "123", proxy.Get(key).Val())
		// De;ete from backing Redis
		assert.Equal(t, int64(1), client.Del(key).Val())
	}

	// Read data from cache. Default size is 1024 so half will be cache misses.
	for i := 0; i < 2048; i++ {
		key := fmt.Sprintf("test%d", i)
		// Read from Cache
		if i >= 1024 {
			assert.Equal(t, "123", proxy.Get(key).Val())
		} else {
			assert.Equal(t, "", proxy.Get(key).Val())
		}
	}
}
