package main

import (
	"fmt"
	"io"
	"net"
	"net/rpc"
)

//Client 客户端
type Client struct {
	Client   *rpc.Client
	Shutdown bool
	// ctx      context.Context
}

// Dial connects to a JSON-RPC server at the specified network address.
func Dial(network, address string) (*Client, error) {
	c := &Client{}
	// var d net.Dialer
	// c.ctx = context.Background()
	// conn, err := d.DialContext(c.ctx, network, address)
	// 创建tcp连接
	conn, err := net.Dial(network, address)
	if err != nil {
		return nil, err
	}
	shutdown := make(chan error)
	c.Client = NewClient(conn, shutdown)
	c.Shutdown = false
	go func() {
		select {
		case <-shutdown:
			c.Shutdown = true
		}
	}()
	return c, err
}

// NewClient returns a new rpc.Client to handle requests to the
// set of services at the other end of the connection.
func NewClient(conn io.ReadWriteCloser, shutdown chan error) *rpc.Client {
	return rpc.NewClientWithCodec(codec.NewClientCodec(conn, shutdown))
}

type body struct {
	Method string                 `json:"method"`
	Params map[string]interface{} `json:"params"`
}

func main() {
	c, err := Dial("tcp", "127.0.0.1:8888")
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	body2 := &body{
		Method: "contact.update",
		Params: map[string]interface{}{"account_name": "yousri", "cellphone": "15581502447"},
	}
	var reply map[string]interface{}
	err = c.Client.Call(body2.Method, body2.Params, &reply)
	fmt.Println(reply)
}
