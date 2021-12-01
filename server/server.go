package server

import (
	"Lrpc/Codec"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"reflect"
	"sync"
)

var MagicNumber = 0x3bef5c

type Option struct {
	MagicNumber int
	CodecType   Codec.Type
}

var DefaultOption = &Option{
	MagicNumber: MagicNumber,
	CodecType:   Codec.GobType,
}

type Request struct {
	h            *Codec.Header
	argv, replyv reflect.Value
}

type Server struct {
}

func (s *Server) Accept(network, address string) {
	lis, _ := net.Listen(network, address)
	// Info is a Logger with LogLevel INF
	var lg = log.New(os.Stderr, "INF: ", log.LstdFlags|log.Llongfile|log.Lmsgprefix)
	for {
		lg.Printf("listern in address %s", address)
		conn, err := lis.Accept()
		if err != nil {
			lg.Println("rpc server: accept error:", err)
			return
		}
		go s.ServerConn(conn, *lg)
	}

}

//解析option
func (s *Server) ServerConn(c io.ReadWriteCloser, logger log.Logger) error {
	//解析option
	var opt Option
	if err := json.NewDecoder(c).Decode(&opt); err != nil {
		logger.Fatalf("format error %s", opt.CodecType)
		return err
	}
	logger.Printf("%s", opt.CodecType)
	//解析data header+
	// Codec.CodecMap[DefaultOption.CodecType]
	f := Codec.CodecMap[opt.CodecType]
	s.ServerCode(f(c), logger)
	return nil
}

//解析header body
func (s *Server) ServerCode(cc Codec.Codec, logger log.Logger) {
	sending := new(sync.Mutex) // make sure to send a complete response
	wg := new(sync.WaitGroup)  // wait until all request are handled
	for {
		request, err := s.ReadRequest(cc)
		if err != nil {
			// s.Response(cc, sending, err)
		}
		wg.Add(1)
		go s.HandleRequest(cc, request, sending, wg)
	}
}

func (s *Server) ReadRequest(cc Codec.Codec) (*Request, error) {
	var header Codec.Header
	if err := cc.ReadHeader(&header); err != nil {
		log.Fatal(err)
	}
	var body []byte
	if err := cc.ReadBody(body); err != nil {
		log.Fatal(err)
	}
	if header.ServiceMethod == "" {
		return nil, fmt.Errorf("method can't be nil")
	}
	return &Request{
		h: &header,
	}, nil
}

func (s *Server) HandleRequest(cc Codec.Codec, req *Request, l *sync.Mutex, wg *sync.WaitGroup) {
	defer wg.Done()
	//反射获取方法
	log.Println(req.h, req.argv.Elem())
	req.replyv = reflect.ValueOf(fmt.Sprintf("geerpc resp %d", req.h.Seq))
	s.Response(cc, req.h, req.replyv.Interface(), l)
}

func (s *Server) Response(cc Codec.Codec, h *Codec.Header, body interface{}, l *sync.Mutex) {
	l.Lock()
	cc.Write(h, body)
	defer l.Unlock()

}
