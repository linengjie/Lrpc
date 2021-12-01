package Codec

import (
	"bufio"
	"encoding/gob"
	"io"
)

type GobCodec struct {
	conn io.ReadWriteCloser
	buf  *bufio.Writer
	enc  *gob.Encoder
	dec  *gob.Decoder
}

var _ Codec = (*GobCodec)(nil)

//构造函数
func NewGobCodec(conn io.ReadWriteCloser) Codec {
	buf := bufio.NewWriter(conn)
	return &GobCodec{
		conn: conn,
		buf:  buf,
		enc:  gob.NewEncoder(buf),
		dec:  gob.NewDecoder(conn),
	}
}

func (g *GobCodec) ReadHeader(h *Header) error {
	if err := g.dec.Decode(h); err != nil {
		return err
	}
	return nil
}
func (g *GobCodec) ReadBody(body interface{}) error {
	if err := g.dec.Decode(body); err != nil {
		return err
	}
	return nil
}
func (g *GobCodec) Write(h *Header, body interface{}) (err error) {
	defer func() {
		g.buf.Flush()
		if err != nil {
			g.Close()

		}
	}()
	if err := g.enc.Encode(h); err != nil {
		return err
	}
	if err := g.enc.Encode(body); err != nil {
		return err
	}
	return nil
}
func (g *GobCodec) Close() error {
	return g.conn.Close()
}
