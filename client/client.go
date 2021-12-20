package client

import (
	"Lrpc/Codec"
	"Lrpc/server"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"sync"
)

/*

	创建与服务器得连接--构建消息--消息编码--写入数据--发送请求
	客户端构建
	|----
	消息体封装
	|----
	消息收发
	|----

*/

//消息体
type Call struct {
	Seq    uint64
	Method string
	Args   interface{}
	Reply  interface{}
	Error  error
	Done   chan *Call
}

func (call *Call) done() {
	call.Done <- call
}

type Client struct {
	address string

	opt *server.Option

	seq     uint64
	header  Codec.Header
	mu      sync.Mutex       //
	sending sync.Mutex       //保证请求的有序发送
	pending map[uint64]*Call //存储未处理完的请求，键是编号，值是 Call 实例

	cc       Codec.Codec
	closing  bool
	shutdown bool
}

//建立连接
func Dial(network, addr string) *Client {
	opt := server.Option{
		MagicNumber: 1234,
		CodecType:   Codec.GobType,
	}

	conn, err := net.Dial(network, addr)
	if err != nil {

	}
	f := Codec.CodecMap[opt.CodecType]
	if err := json.NewEncoder(conn).Encode(opt); err != nil {
		log.Println("rpc client: options error: ", err)
		_ = conn.Close()
	}
	client := &Client{
		opt:     &opt,
		address: addr,
		cc:      f(conn),
		pending: make(map[uint64]*Call),
	}
	go client.receive()
	return client
}

//调用链，包装消息体
func (client *Client) call(method string, args, reply interface{}) {
	call := &Call{
		Seq:    1234,
		Method: method,
		Args:   args,
		Reply:  reply,
		Done:   make(chan *Call),
	}

	client.send(call)

}

//消息发送-接受
func (client *Client) send(call *Call) {
	client.sending.Lock()
	defer client.sending.Unlock()

	client.registerCall(call)
	client.header.Seq = call.Seq
	client.header.ServiceMethod = call.Method
	client.header.Err = ""
	err := client.cc.Write(&client.header, nil)
	if err != nil {

	}
}

func (client *Client) receive() {
	var err error
	for err == nil {
		var h Codec.Header
		if err = client.cc.ReadHeader(&h); err != nil {
			break
		}
		call := client.removeCall(h.Seq)
		switch {
		case call == nil:
			// it usually means that Write partially failed
			// and call was already removed.
			err = client.cc.ReadBody(nil)
		case h.Err != "":
			call.Error = fmt.Errorf(h.Err)
			err = client.cc.ReadBody(nil)
			call.done()
		default:
			err = client.cc.ReadBody(call.Reply)
			if err != nil {
				call.Error = errors.New("reading body " + err.Error())
			}
			call.done()
		}
	}
	// error occurs, so terminateCalls pending calls
	client.terminateCalls(err)
}

func (client *Client) registerCall(call *Call) (uint64, error) {
	client.mu.Lock()
	defer client.mu.Unlock()
	client.pending[call.Seq] = call
	client.seq = call.Seq
	client.seq++
	return client.seq, nil
}

func (client *Client) removeCall(seq uint64) *Call {
	client.mu.Lock()
	defer client.mu.Unlock()
	call := client.pending[seq]
	delete(client.pending, seq)
	return call
}

func (client *Client) terminateCalls(err error) {
	client.sending.Lock()
	defer client.sending.Unlock()
	client.mu.Lock()
	defer client.mu.Unlock()
	client.shutdown = true
	for _, call := range client.pending {
		call.Error = err
		call.done()
	}
}
