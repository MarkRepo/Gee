// Package codec 实现 rpc 编解码功能
package codec

import "io"

// Header 请求头
type Header struct {
	ServiceMethod string // format "Service.Method"
	Seq           uint64 // sequence number chosen by client
	Error         string
}

// Codec rpc 编解码器接口，client、server 共用
type Codec interface {
	io.Closer
	// ReadHeader 读请求头
	ReadHeader(*Header) error
	// ReadBody 读请求体
	ReadBody(interface{}) error
	// Write 写 Header 和 Body 到conn
	Write(*Header, interface{}) error
}

// NewCodecFunc Codec 创建函数类型
type NewCodecFunc func(closer io.ReadWriteCloser) Codec

// Type 协议类型名称
type Type string

const (
	GobType  Type = "application/gob"
	JsonType Type = "application/json" // not implemented
)

var NewCodecFuncMap map[Type]NewCodecFunc

func init() {
	NewCodecFuncMap = make(map[Type]NewCodecFunc)
	NewCodecFuncMap[GobType] = NewGobCodec
}
