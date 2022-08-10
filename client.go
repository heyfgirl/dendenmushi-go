package dendenmushi

import (
	"context"
	"net"
	"net/rpc"

	codec "github.com/heyfgirl/dendenmushi-go/codec"
	"github.com/heyfgirl/dendenmushi-go/pool"
)

// stater client 客户端
type stater struct {
	client   *(rpc.Client)
	shutdown bool
}

// dial connects to a  server at the specified network address.
func createConn(network, address string) (*stater, error) {
	d := &net.Dialer{}
	ctx := context.Background()
	conn, err := d.DialContext(ctx, network, address)
	if err != nil {
		return nil, err
	}
	shutdown := make(chan error)
	c := &stater{
		shutdown: false,
		client:   rpc.NewClientWithCodec(codec.NewClientCodec(conn, shutdown)),
	}
	go func() {
		select {
		case <-shutdown:
			c.shutdown = true
		}
	}()
	return c, err
}

//PoolClient 连接池客户端 dendenmushi
type PoolClient struct {
	pool pool.Pool
}

// Call 方法
func (c *PoolClient) Call(serviceMethod string, args interface{}, reply interface{}) error {
	v, err := c.pool.Get()
	if err != nil {
		// 获取连接失败
		return err
	}
	client := v.(*stater)
	err = client.client.Call(serviceMethod, args, &reply)
	if err != nil {
		c.pool.Close(v)
	}
	// 用完返还
	c.pool.Put(v)
	return err
}

// NewClient 创建客户端
func NewClient(network, address string, config *pool.Config) (*PoolClient, error) {
	if config == nil {
		config = &pool.Config{
			InitialCap: 0,
			MaxIdle:    5,
			MaxCap:     30,
		}
	}
	poolConfig := &pool.Config{
		InitialCap: config.InitialCap,
		MaxIdle:    config.MaxIdle,
		MaxCap:     config.MaxCap,
		Factory: func() (interface{}, error) {
			c, err := createConn(network, address)
			return c, err
		},
		Close: func(v interface{}) error {
			return v.(*stater).client.Close()
		},
		IdleTimeout: config.IdleTimeout,
		Ping: func(v interface{}) error { //检查是否还处于连接状态
			client := v.(*stater)
			if client.shutdown {
				return rpc.ErrShutdown
			}
			return nil
		},
	}
	p, err := pool.NewChannelPool(poolConfig)
	if err != nil {
		return nil, err
	}
	poolclient := &PoolClient{
		pool: p,
	}
	return poolclient, nil
}
