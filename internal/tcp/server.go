package tcp

import (
	"fmt"
	"net"

	. "github.com/eastside-eng/redis-proxy/internal/log"
)

type Server struct {
}

func (s *Server) process(tcpConn net.Conn) {
	// Parse incoming bytes into Redis commands
}

func (s *Server) Run(port int) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))

	if err != nil {
		panic("Error binding!")
	}

	for {
		tcpConn, err := listener.Accept()
		if err != nil {
			Logger.Warnf("Error accepting new connection! %v", err)
		} else {
			go s.process(tcpConn)
		}
	}
}
