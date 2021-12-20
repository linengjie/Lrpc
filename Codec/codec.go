package Codec

import "io"

type Header struct {
	ServiceMethod string
	Seq           uint64
	Err           string
}

//编码解码接口
type Codec interface {
	io.Closer
	ReadHeader(*Header) error
	ReadBody(interface{}) error
	Write(*Header, interface{}) error
}

//构造函数类型
type NewCodecFunc func(io.ReadWriteCloser) Codec

type Type string

const (
	GobType  Type = "application/gob"
	JsonType Type = "application/json"
)

var CodecMap map[Type]NewCodecFunc

func init() {
	CodecMap := make(map[Type]NewCodecFunc)
	CodecMap[GobType] = NewGobCodec
}
