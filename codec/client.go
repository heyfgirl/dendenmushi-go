package codec

import (
	"bufio"
	"io"
	"net/rpc"
	"strings"
	"sync"

	"github.com/vmihailenco/msgpack/v5"
)

type clientCodec struct {
	r        io.Reader
	w        io.Writer
	c        io.Closer
	mutex    sync.Mutex        // protects pending
	pending  map[uint64]string // map request id to method name
	shutdown chan error
	body     *msgpack.RawMessage
}

// NewClientCodec Create a new client codec
func NewClientCodec(conn io.ReadWriteCloser, shutdown chan error) rpc.ClientCodec {
	return &clientCodec{
		r:        bufio.NewReader(conn),
		w:        bufio.NewWriter(conn),
		c:        conn,
		pending:  make(map[uint64]string),
		shutdown: shutdown,
	}
}

// 阻塞获取 socket头数据信息
func (c *clientCodec) ReadResponseHeader(r *rpc.Response) error {
	// // 读取 header
	// 解析body 按 upyun rpc的格式化解析
	type upResponseFormat struct {
		ID   uint64              `msgpack:"id"`
		Type string              `msgpack:"type"`
		Body *msgpack.RawMessage `msgpack:"result"`
	}
	result := upResponseFormat{}
	err := read(c.r, &result)
	if err != nil {
		// 网络断开
		if err == io.EOF {
			c.shutdown <- err
		}
		return err
	}
	c.body = result.Body
	replyErr := ""
	if result.Type == "error" {
		err := msgpack.Unmarshal(*c.body, &replyErr)
		if err != nil {
			return err
		}
	}
	// 设置返回头信息
	c.mutex.Lock()
	r.Seq = result.ID
	if replyErr != "" {
		r.Error = replyErr
	}
	r.ServiceMethod = c.pending[r.Seq]
	delete(c.pending, r.Seq)
	c.mutex.Unlock()

	return nil
}

// 获取头成功之后的 body获取，
func (c *clientCodec) ReadResponseBody(reply interface{}) error {
	if reply == nil {
		return nil
	}
	return msgpack.Unmarshal(*c.body, reply)
}

// 写数据方法
func (c *clientCodec) WriteRequest(r *rpc.Request, param interface{}) error {
	c.mutex.Lock()
	c.pending[r.Seq] = r.ServiceMethod
	c.mutex.Unlock()
	var Method = r.ServiceMethod

	pointIndex := strings.Index(Method, ".")
	var ServiceName string = Method[0:pointIndex]
	var MethodName string = Method[pointIndex+1:]
	var Params = param
	var ID = r.Seq
	// 对数据进行格式化
	// 写入头长度
	params := make([]interface{}, 0)
	params = append(params, ID, "call", ServiceName, MethodName, Params)
	b, err := msgpack.Marshal(&params)
	if err != nil {
		return err
	}
	return write(c.w, b)
}

// 关闭code
func (c *clientCodec) Close() error {
	return c.c.Close()
}
