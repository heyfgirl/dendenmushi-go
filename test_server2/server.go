package main

import (
	"log"
	"net"
	"net/rpc"

	"github.com/heyfgirl/dendenmushi-go"
)

const HelloServiceName = "path/to/pkg.HelloService"

type HelloServiceInterface = interface {
	Hello(request Params, reply *string) error
}

func RegisterHelloService(svc HelloServiceInterface) error {
	return rpc.RegisterName(HelloServiceName, svc)
}

type A struct{}

func (a *A) Hello(params Params, reply *string) error {
	*reply = "hello:" + params.Username
	return nil
}

// 传的参数
type Params struct {
	Username string `msgpack:"username"`
	Password string `msgpack:"password"`
}

func main() {
	lis, err := net.Listen("tcp", ":8889")
	if err != nil {
		log.Fatal(err)
	}
	server := dendenmushi.NewServer()

	var p HelloServiceInterface = &A{}
	server.RegisterName(HelloServiceName, p)
	server.Serve(lis)
}
