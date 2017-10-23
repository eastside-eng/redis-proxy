package cache

import (
	"testing"
	"time"

	log "github.com/eastside-eng/redis-proxy/log"
	"github.com/stretchr/testify/assert"
)

func init() {
	logger := log.NewLogger()
	log.SetLogger(logger)
}

func TestCacheConstructor(t *testing.T) {
	assert := assert.New(t)
	cache, err := NewDecayingLRUCache(2, -1, 1)

	assert.Nil(cache)
	assert.NotNil(err)
}

// TestCache tests basic happy-case behavior for TTL and LRU.
func TestCacheSimple(t *testing.T) {
	assert := assert.New(t)
	period := time.Duration(50 * time.Microsecond)
	ttl := time.Duration(1 * time.Second)
	cache, err := NewDecayingLRUCache(2, period, ttl)
	assert.NotNil(cache)
	assert.Nil(err)

	cache.Start()
	defer cache.Stop()

	cache.Add("1", "test1")
	cache.Add("2", "test2")

	res, exists := cache.Get("1")
	assert.Equal("test1", res)
	assert.True(exists)

	res, exists = cache.Get("2")
	assert.Equal("test2", res)
	assert.True(exists)

	res, exists = cache.Get("3")
	assert.Nil(res)
	assert.False(exists)

	cache.Remove("2")

	res, exists = cache.Get("2")
	assert.Nil(res)
	assert.False(exists)
}

func TestCacheLRU(t *testing.T) {
	assert := assert.New(t)
	period := time.Duration(50 * time.Microsecond)
	ttl := time.Duration(1 * time.Second)
	cache, err := NewDecayingLRUCache(2, period, ttl)
	assert.NotNil(cache)
	assert.Nil(err)

	cache.Start()
	defer cache.Stop()

	cache.Add("1", "test1")
	cache.Add("2", "test2")

	res, exists := cache.Get("1")
	assert.Equal("test1", res)
	assert.True(exists)

	res, exists = cache.Get("2")
	assert.Equal("test2", res)
	assert.True(exists)

	res, exists = cache.Get("3")
	assert.Nil(res)
	assert.False(exists)

	// Testing that 1 is evicted as the LRU goes over capacity.
	cache.Add("3", "test3")

	res, exists = cache.Get("1")
	assert.Nil(res)
	assert.False(exists)

	res, exists = cache.Get("2")
	assert.Equal("test2", res)
	assert.True(exists)

	res, exists = cache.Get("3")
	assert.Equal("test3", res)
	assert.True(exists)

	// Testing that LRU correctly updates 2 and evicts 3.
	cache.Get("2")
	cache.Add("1", "test1")

	res, exists = cache.Get("1")
	assert.Equal("test1", res)
	assert.True(exists)

	res, exists = cache.Get("2")
	assert.Equal("test2", res)
	assert.True(exists)

	res, exists = cache.Get("3")
	assert.Nil(res)
	assert.False(exists)
}

func TestCacheTTL(t *testing.T) {
	assert := assert.New(t)
	period := time.Duration(50 * time.Microsecond)
	ttl := time.Duration(1 * time.Second)
	cache, err := NewDecayingLRUCache(2, period, ttl)
	assert.NotNil(cache)
	assert.Nil(err)

	cache.Start()
	defer cache.Stop()

	cache.Add("1", "test1")
	cache.Add("2", "test2")

	res, exists := cache.Get("1")
	assert.Equal("test1", res)
	assert.True(exists)

	res, exists = cache.Get("2")
	assert.Equal("test2", res)
	assert.True(exists)

	// Testing the TTL functionality.
	time.Sleep(time.Second * 2)

	res, exists = cache.Get("1")
	assert.Nil(res)
	assert.False(exists)

	res, exists = cache.Get("2")
	assert.Nil(res)
	assert.False(exists)
}

// This test executes many writes for the same key and asserts that the older
// log entries do not affect newer key modifications.
func TestCacheTTLCorrectness(t *testing.T) {
	assert := assert.New(t)
	period := time.Duration(50 * time.Microsecond)
	ttl := time.Duration(1 * time.Second)
	cache, err := NewDecayingLRUCache(2, period, ttl)
	assert.NotNil(cache)
	assert.Nil(err)

	cache.Start()
	defer cache.Stop()

	for i := 0; i < 1000; i++ {
		cache.Add("1", "test1")
		res, exists := cache.Get("1")
		assert.Equal("test1", res)
		assert.True(exists)
	}
}
