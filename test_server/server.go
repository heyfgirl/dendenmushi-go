package main

import (
	"fmt"
	"log"
	"net"
	"net/rpc"

	codec "github.com/heyfgirl/dendenmushi-go/codec"
)

const HelloServiceName = "path/to/pkg.HelloService"

type HelloServiceInterface = interface {
	Hello(request Params, reply *string) error
}

func RegisterHelloService(svc HelloServiceInterface) error {
	return rpc.RegisterName(HelloServiceName, svc)
}

type A struct{}

// 传的参数
type Params struct {
	Username string `msgpack:"username"`
	Password string `msgpack:"password"`
}

func (a *A) Hello(params Params, reply *string) error {
	*reply = "hello:" + params.Username
	return nil
}
func main() {

	err := RegisterHelloService(&A{})
	fmt.Println(err)

	listener, err := net.Listen("tcp", ":8889")
	if err != nil {
		log.Fatal("ListenTCP error:", err)
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal("Accept error:", err)
		}

		go rpc.ServeCodec(codec.NewServerCodec(conn))
	}

}
