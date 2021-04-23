package codec

import (
	"bufio"
	"encoding/gob"
	"io"
	"log"
)

// GobCodec gob 协议编解码器
type GobCodec struct {
	conn io.ReadWriteCloser // conn 连接
	buf  *bufio.Writer      // buf 是为了防止阻塞而创建的带缓冲的 Writer
	dec  *gob.Decoder
	enc  *gob.Encoder
}

// NewGobCodec 创建 gob 协议编解码器
func NewGobCodec(conn io.ReadWriteCloser) Codec {
	buf := bufio.NewWriter(conn)
	return &GobCodec{
		conn: conn,
		buf:  buf,
		dec:  gob.NewDecoder(conn),
		enc:  gob.NewEncoder(buf),
	}
}

// ReadHeader 实现 Codec.ReadHeader 接口
func (c *GobCodec) ReadHeader(h *Header) error {
	return c.dec.Decode(h)
}

// ReadBody 实现 Codec.ReadBody 接口
func (c *GobCodec) ReadBody(body interface{}) error {
	return c.dec.Decode(body)
}

// Write 实现 Codec.Write 接口
func (c *GobCodec) Write(h *Header, body interface{}) (err error) {
	defer func() {
		_ = c.buf.Flush()
		if err != nil {
			_ = c.Close()
		}
	}()
	if err := c.enc.Encode(h); err != nil {
		log.Println("rpc codec: gob error encoding header:", err)
		return err
	}
	if err := c.enc.Encode(body); err != nil {
		log.Println("rpc codec: gob error encoding body:", err)
		return err
	}
	return nil
}

// Close 实现 Codec.Closer 接口
func (c *GobCodec) Close() error {
	return c.conn.Close()
}
