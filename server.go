package dendenmushi

import (
	"bufio"
	"io"
	"net/rpc"
	"strconv"
	"sync"

	"github.com/vmihailenco/msgpack/v5"
)

type upServerCodec struct {
	r     io.Reader
	w     io.Writer
	c     io.Closer
	mutex sync.Mutex // protects pending
	// pending map[uint64]string // map request id to method name
	encBuf *bufio.Writer
	body   *msgpack.RawMessage
}

// NewServerCodec Create a new client codec
func NewServerCodec(conn io.ReadWriteCloser) rpc.ServerCodec {
	return &upServerCodec{
		r: bufio.NewReader(conn),
		w: bufio.NewWriter(conn),
		c: conn,
		// pending: make(map[uint64]string),
	}
}
func (s *upServerCodec) ReadRequestHeader(r *rpc.Request) error {
	// 解析body 按 upyun rpc的格式化解析
	type upRequestFormat struct {
		ID          string
		Type        string
		ServiceName string
		MethodName  string
		Body        *msgpack.RawMessage
		Header      *msgpack.RawMessage
	}
	result := upRequestFormat{}
	err := read(s.r, &result)
	// err = msgpack.Unmarshal(b, &result)
	if err != nil {
		return err
	}
	s.body = result.Body
	// 处理请求头信息
	ridint, err := strconv.Atoi(result.ID)
	if err != nil {
		return err
	}
	rid := uint64(ridint)

	method := result.ServiceName + "." + result.MethodName

	s.mutex.Lock()
	r.Seq = rid
	r.ServiceMethod = method
	// delete(c.pending, r.Seq)
	s.mutex.Unlock()

	return nil

}

func (s *upServerCodec) ReadRequestBody(body interface{}) error {
	if body == nil {
		return nil
	}
	return msgpack.Unmarshal(*s.body, body)
}

func (s *upServerCodec) WriteResponse(r *rpc.Response, body interface{}) (err error) {
	replayMessage := []interface{}{}
	replayMessage = append(replayMessage, r.Seq)
	if r.Error != "" {
		replayMessage = append(replayMessage, "error", r.Error)
	} else {
		replayMessage = append(replayMessage, "reply", body)
	}
	b, err := msgpack.Marshal(&replayMessage)
	if err != nil {
		return err
	}
	return write(s.w, b)

}

func (s *upServerCodec) Close() error {
	return s.c.Close()
}
