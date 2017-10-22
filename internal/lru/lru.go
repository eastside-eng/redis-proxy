package dlru

import (
	"container/list"
	"errors"
	"sync"
	"time"
)

// cacheElement is a container type used by DecayingLRUCache.
type cacheElement struct {
	Key       string
	Val       interface{}
	Timestamp time.Time
}

// DecayingLRUCache is a LRU cache that also uses wall-clock time to expire
// elements based on a configurable TTL. To enable this functionality, the cache
// uses ~2x memory per key (note that's not per entry) as a typical LRU cache.
//
// The time-based expiry is driven by a periodic coroutine that checks in
// constant time if an element is expired. It keeps a time-ordered queue of
// elements and simply iterates from the front until it hits an element that has
// a valid TTL. The periodicity of the eviction routine is parameterized and
// given to the constructor.
type DecayingLRUCache struct {
	elements *list.List
	log      *list.List
	hashmap  map[string]*list.Element
	capacity int
	lock     sync.Mutex

	ticker     *time.Ticker
	stopTicker chan bool
	ttl        time.Duration
}

// NewDecayingLRUCache returns a new DecayingLRUCache with the given capacity,
// period and ttl.
func NewDecayingLRUCache(capacity int, period time.Duration, ttl time.Duration) (*DecayingLRUCache, error) {
	if period < 0 {
		return nil, errors.New("Period must be non-negative")
	}

	if ttl < 0 {
		return nil, errors.New("Expiry TTL must be non-negative")
	}

	cache := &DecayingLRUCache{
		list.New(),
		list.New(),
		make(map[string]*list.Element),
		capacity,
		sync.Mutex{},

		// For the redeemer
		time.NewTicker(period),
		make(chan bool),
		ttl,
	}
	return cache, nil
}

// Get returns the value of the key in the cache, iff it exists, and a boolean
// for checking existence. If a key has no entry, nil will be returned.
func (cache *DecayingLRUCache) Get(key string) (interface{}, bool) {
	ref, exists := cache.hashmap[key]
	if exists {
		return ref.Value.(*cacheElement).Val, exists
	}
	return nil, exists
}

// Add atomicly inserts the key and value into the cache, updating it's value,
// recency and timestamp.
func (cache *DecayingLRUCache) Add(key string, val interface{}) {
	cache.lock.Lock()
	defer cache.lock.Unlock()

	element := &cacheElement{key, val, time.Now()}

	// Append to our time-ordered log
	cache.log.PushBack(element)

	// Update both the hashmap and doubly linked list for the element.
	ref, exists := cache.hashmap[key]
	if exists {
		cache.elements.MoveToFront(ref)
		ref.Value = element
	} else {
		listElement := cache.elements.PushFront(element)
		cache.hashmap[key] = listElement
	}

	// Handle eviction. Could be more DRY.
	if cache.elements.Len() > cache.capacity {
		lru := cache.elements.Back()
		if lru != nil {
			lruKey := lru.Value.(*cacheElement).Key
			cache.elements.Remove(lru)
			delete(cache.hashmap, lruKey)
		}
	}
}

// Remove atomicly removes the given key from the cache.
func (cache *DecayingLRUCache) Remove(key string) {
	cache.lock.Lock()
	defer cache.lock.Unlock()

	ref, exists := cache.hashmap[key]
	if exists {
		cache.elements.Remove(ref)
		delete(cache.hashmap, key)
	}
}

// RemoveIfExpired will atomicly remove the given key from the cache, iff the most recent
// timestamp is after the given time.
func (cache *DecayingLRUCache) RemoveIfExpired(key string, after time.Time) {
	cache.lock.Lock()
	defer cache.lock.Unlock()

	ref, exists := cache.hashmap[key]
	if exists {
		element := ref.Value.(*cacheElement)
		expiry := element.Timestamp.Add(cache.ttl)
		expired := expiry.After(after)
		if expired {
			cache.elements.Remove(ref)
			delete(cache.hashmap, key)
		}
	}
}

func (cache *DecayingLRUCache) redeemer() {
	for {
		select {
		case <-cache.ticker.C:
			cursor := cache.log.Front()
			now := time.Now()
			for cursor != nil {
				element := cursor.Value.(*cacheElement)
				expiry := element.Timestamp.Add(cache.ttl)
				expired := expiry.After(now)

				// Because the log is ordered by time, we can bail out once we hit a
				// non-expired entry.
				if !expired {
					break
				}

				// Because the TS is set on creation and not mutable, we don't need to
				// lock when reading TS above and can simply lock when removing from
				// the underly data structures.

				// One issue is that the log will contain multiple entries for a key,
				// we need to lock and check the map for the _real_ TS.

				// This is a bit of an optimization, we can check if
				// element.TS == hash[element.key].TS in constant time to determine
				// if we need to invoke remove.
				mapElement, exists := cache.hashmap[element.Key]
				if exists {
					mapElement := mapElement.Value.(*cacheElement)
					if mapElement.Timestamp == element.Timestamp {
						cache.RemoveIfExpired(element.Key, now)
					}
				}

				// Move cursor and remove last element.
				cursor = cursor.Next()
				cache.log.Remove(cursor.Prev())
			}
		case <-cache.stopTicker:
			cache.ticker.Stop()
			return
		}
	}
}

// Start will start the Redeemer coroutine. The callee must call #Stop() to
// allow GC to clean up the cache.
func (cache *DecayingLRUCache) Start() {
	go cache.redeemer()
}

// Stop will kill the Redeemer coroutine and allow GC to happen.
func (cache *DecayingLRUCache) Stop() {
	cache.stopTicker <- true
}
