package proxy

import (
	"bufio"
	"errors"
	"fmt"
	"io/ioutil"
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
	reader := bufio.NewReader(tcpConn)
	writer := bufio.NewWriter(tcpConn)
	bytes, err := ioutil.ReadAll(reader)
	if err != nil {
		Logger.Warnw("Failed to parse connection bytes")
		return
	}
	command, err := parseCommand(bytes)
	if err != nil {
		Logger.Warnw("Failed to parse command", "command", command)
		return
	}

	resp, err := s.processCommand(command)
	if err != nil {
		Logger.Errorw("Failed to process command", "command", command)
	}

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
		if err != nil {
			Logger.Warnf("Error accepting new connection! %v", err)
		} else {
			go s.process(tcpConn)
		}
	}
}

// Handler ...
type handler func(cache *cache.DecayingLRUCache, redisClient *redis.Client, command *Command) ([]byte, error)

var getHandler handler = func(cache *cache.DecayingLRUCache, redisClient *redis.Client, command *Command) ([]byte, error) {
	key := command.Args[0]
	resp, exists := cache.Get(key)
	if !exists {
		resp := redisClient.Get(key)
		val, err := resp.Bytes()
		if err != nil {
			return nil, err
		}
		cache.Add(key, val)
		return val, nil
	}
	return resp.([]byte), nil
}

// A map of supported handlers for Redis commands.
var handlers = map[string]handler{
	"GET": getHandler,
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
