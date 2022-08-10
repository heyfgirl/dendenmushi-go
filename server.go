package dendenmushi

import (
	"log"
	"net"
	"net/rpc"

	codec "github.com/heyfgirl/dendenmushi-go/codec"
)

// Server rpc server based on net/rpc implementation
type Server struct {
	*rpc.Server
}

// NewServer Create a new rpc server
func NewServer() *Server {
	return &Server{&rpc.Server{}}
}

// Register register rpc function
func (s *Server) Register(rcvr interface{}) error {
	return s.Server.Register(rcvr)
}

// RegisterName register the rpc function with the specified name
func (s *Server) RegisterName(name string, rcvr interface{}) error {
	return s.Server.RegisterName(name, rcvr)
}

// Serve start service
func (s *Server) Serve(lis net.Listener) {
	log.Printf("tinyrpc started on: %s", lis.Addr().String())
	for {
		conn, err := lis.Accept()
		if err != nil {
			continue
		}
		go s.Server.ServeCodec(codec.NewServerCodec(conn))
	}
}
