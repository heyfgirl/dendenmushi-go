package codec

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"io"

	"github.com/vmihailenco/msgpack/v5"
)

func read(r io.Reader, data interface{}) error {
	// 读取 header
	b := make([]byte, 4)
	_, err := io.ReadFull(r, b)
	if err != nil {
		return err
	}
	// 获取数据body的长度
	bytesBuffer := bytes.NewBuffer(b)
	var bodyLen int32
	binary.Read(bytesBuffer, binary.BigEndian, &bodyLen)
	// 读取body
	b = make([]byte, bodyLen)
	_, err = io.ReadFull(r, b)
	if err != nil {
		return err
	}
	err = msgpack.Unmarshal(b, data)
	return err
}

func write(w io.Writer, data []byte) error {
	headCode := int32(len(data))
	bytesBuffer := bytes.NewBuffer([]byte{})
	binary.Write(bytesBuffer, binary.BigEndian, headCode)
	_, err := w.Write(bytesBuffer.Bytes())
	if err != nil {
		return err
	}
	// 写入数据
	_, err = w.Write(data)
	if err != nil {
		return err
	}
	w.(*bufio.Writer).Flush()
	return nil
}
