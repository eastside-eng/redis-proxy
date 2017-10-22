package redis

import (
	"github.com/go-redis/redis"

	. "github.com/eastside-eng/redis-proxy/internal/log"
)

type Callback func(command *Command, val []byte)
type Handler func(redisClient *redis.Client) ([]byte, error)

// A map of supported handlers for Redis commands.
var handlers = map[string]Handler{}

// Command is a parsed version of the Redis RESP protocol. Each Command
// should implement a Handler function that takes a RedisClient and processes
// the request. The Handler should also accept a Callback function for callees
// to inject logic into the processing step.
type Command struct {
	Name    string
	Args    []string
	Handler Handler
}

func (c *Command) Process(redisClient *redis.Client, callback Callback) {
	resp, err := c.Handler(redisClient)
	if err != nil {
		Logger.Infow("Error handling command!", "command", c)
	}
	if callback != nil {
		callback(c, resp)
	}
}

func ParseCommand(raw []byte) *Command {
	return nil
}
