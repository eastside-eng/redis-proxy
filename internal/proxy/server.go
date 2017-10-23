package proxy

import (
	"bufio"
	"errors"
	"fmt"
	"net"

	"github.com/eastside-eng/redis-proxy/internal/cache"
	. "github.com/eastside-eng/redis-proxy/internal/log"
	"github.com/go-redis/redis"
)

// Server is a stateful container that processes incoming TCP connections and
// responds to RESP (Redis Serialization Protocol) commands. The server keeps
// a stateful cache and delegates calls to the redis-go library.
type Server struct {
	cache       *cache.DecayingLRUCache
	redisClient *redis.Client
}

// NewServer returns a new Server instance.
func NewServer(cache *cache.DecayingLRUCache, redisClient *redis.Client) *Server {
	server := &Server{cache, redisClient}
	return server
}

func (s *Server) process(tcpConn net.Conn) {
	defer tcpConn.Close()

	bytes := make([]byte, 1024)
	_, err := tcpConn.Read(bytes)

	Logger.Infow("Processing connection", "bytes", len(bytes), "err", err)
	command, err := parseCommand(bytes)
	if err != nil {
		Logger.Warnw("Failed to parse command", "command", command)
		return
	}

	resp, err := s.processCommand(command)
	if err != nil {
		Logger.Errorw("Failed to process command", "command", command, "err", err)
	}

	writer := bufio.NewWriter(tcpConn)
	writer.Write(resp)
	writer.Flush()
}

// Run spawns a TCP server on the port given and begins accepting incoming
// connections.
func (s *Server) Run(port int) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))

	if err != nil {
		panic("Error binding!")
	}

	defer s.cache.Stop()
	s.cache.Start()

	for {
		tcpConn, err := listener.Accept()
		Logger.Infow("Accepted new connection!")
		if err != nil {
			Logger.Warnf("Error accepting new connection! %v", err)
		} else {
			go s.process(tcpConn)
		}
	}
}

var RespNIL = []byte("$-1\r\n")

func RespEncodeString(str string) []byte {
	res := fmt.Sprintf("$%d\r\n%s\r\n", len(str), str)
	return []byte(res)
}

func RespEncodeInteger(i int) []byte {
	res := fmt.Sprintf(":%d\r\n", i)
	return []byte(res)
}

// Handler ...
type handler func(cache *cache.DecayingLRUCache, redisClient *redis.Client, command *Command) ([]byte, error)

var getHandler handler = func(cache *cache.DecayingLRUCache, redisClient *redis.Client, command *Command) ([]byte, error) {
	key := command.Args[0]
	resp, exists := cache.Get(key)
	Logger.Infow("Invoking GET on cache", "cache-entry", resp)
	if !exists {
		resp := redisClient.Get(key)
		val, err := resp.Result()
		Logger.Infow("Invoking GET on backing Redis", "redis-entry", val)
		if err != nil {
			return RespNIL, nil
		}
		bytes := RespEncodeString(val)
		cache.Add(key, bytes)
		return bytes, nil
	}
	return resp.([]byte), nil
}

var pingHandler handler = func(cache *cache.DecayingLRUCache, redisClient *redis.Client, command *Command) ([]byte, error) {
	return RespEncodeString("PONG"), nil
}

// A map of supported handlers for Redis commands.
var handlers = map[string]handler{
	"GET":  getHandler,
	"PING": pingHandler,
}

// ProcessCommand ..
func (s *Server) processCommand(command *Command) ([]byte, error) {
	handler, exists := handlers[command.Name]
	if exists {
		resp, err := handler(s.cache, s.redisClient, command)
		if err != nil {
			Logger.Infow("Error handling command!", "command", command)
			return nil, err
		}
		return resp, nil
	}
	return nil, errors.New("No handler for command")
}
