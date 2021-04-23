package rpc

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net"
	"reflect"
	"strings"
	"sync"

	"github.com/MarkRepo/Gee/GeeRPC/rpc/codec"
)

const MagicNumber = 0x3bef5c // MagicNumber gee rpc 魔数

// Option 协议协商信息
type Option struct {
	MagicNumber int        // MagicNumber marks this is a gee rpc request
	CodecType   codec.Type // CodecType client may choose different Codec to encode body
}

var DefaultOption = &Option{
	MagicNumber: MagicNumber,
	CodecType:   codec.GobType,
}

// | Option{MagicNumber: xxx, CodecType: xxx} | Header{ServiceMethod ...} | Body interface{} |
// | <------      固定 JSON 编码      ------>  | <-------   编码方式由 CodeType 决定   ------->|

// 在一次连接中，Option 固定在报文的最开始，Header 和 Body 可以有多个，即报文可能是这样的:
// | Option | Header1 | Body1 | Header2 | Body2 | ...

// Server represents an RPC Server.
type Server struct {
	serviceMap sync.Map
}

// NewServer returns a new Server.
func NewServer() *Server {
	return &Server{}
}

// DefaultServer is the default instance of *Server.
var DefaultServer = NewServer()

// Accept accepts connections on the listener and serves requests for each incoming connection.
func (server *Server) Accept(lis net.Listener) {
	for {
		conn, err := lis.Accept()
		if err != nil {
			log.Println("rpc server: accept error:", err)
			return
		}
		go server.ServeConn(conn)
	}
}

// Accept accepts connections on the listener and serves requests for each incoming connection.
func Accept(lis net.Listener) { DefaultServer.Accept(lis) }

// ServeConn runs the server on a single connection
// ServeConn blocks, serving the connection until the client hangs up
func (server *Server) ServeConn(conn io.ReadWriteCloser) {
	defer func() {
		_ = conn.Close()
	}()

	var opt Option
	if err := json.NewDecoder(conn).Decode(&opt); err != nil {
		log.Println("rpc server: options error: ", err)
		return
	}

	if opt.MagicNumber != MagicNumber {
		log.Printf("rpc server: invalid magic number %x", opt.MagicNumber)
		return
	}

	f := codec.NewCodecFuncMap[opt.CodecType]
	if f == nil {
		log.Printf("rpc server: invalid code type %s", opt.CodecType)
		return
	}

	server.serveCodec(f(conn))
}

// invalidRequest is a placeHolder for response argv when error occurs
var invalidRequest = struct{}{}

func (server *Server) serveCodec(cc codec.Codec) {
	sending := new(sync.Mutex) // make sure to send a complete response
	wg := new(sync.WaitGroup)
	for {
		req, err := server.readRequest(cc)
		if err != nil {
			if req == nil {
				break // it's not possible to recover, so close the connection
			}
			req.h.Error = err.Error()
			server.sendResponse(cc, req.h, invalidRequest, sending)
			continue
		}
		wg.Add(1)
		go server.handleRequest(cc, req, sending, wg)
	}
	wg.Wait()
	_ = cc.Close()
}

// request stores all information of a call
type request struct {
	h          *codec.Header // h header of request
	arg, reply reflect.Value // arg and reply of request
	mt         *methodType   // mt the method to call
	svc        *service      // svc the service to handle the request
}

func (server *Server) readRequestHeader(cc codec.Codec) (*codec.Header, error) {
	var h codec.Header
	if err := cc.ReadHeader(&h); err != nil {
		if err != io.EOF && err != io.ErrUnexpectedEOF {
			log.Println("rpc server: read header error:", err)
		}
		return nil, err
	}
	return &h, nil
}

func (server *Server) readRequest(cc codec.Codec) (*request, error) {
	h, err := server.readRequestHeader(cc)
	if err != nil {
		return nil, err
	}

	req := &request{h: h}
	req.svc, req.mt, err = server.findService(h.ServiceMethod)
	if err != nil {
		return req, err
	}
	req.arg = req.mt.newArgv()
	req.reply = req.mt.newReply()
	//make sure that arg is a pointer, ReadBody need a pointer as parameter
	arg := req.arg.Interface()
	if req.arg.Type().Kind() != reflect.Ptr {
		arg = req.arg.Addr().Interface()
	}

	if err = cc.ReadBody(arg); err != nil {
		log.Println("rpc server: read body err:", err)
		return req, err
	}

	return req, nil
}

// sendResponse 按顺序发送每一个响应
func (server *Server) sendResponse(cc codec.Codec, h *codec.Header, body interface{}, sending *sync.Mutex) {
	sending.Lock()
	defer sending.Unlock()
	if err := cc.Write(h, body); err != nil {
		log.Println("rpc server: write response error:", err)
	}
}

func (server *Server) handleRequest(cc codec.Codec, req *request, sending *sync.Mutex, wg *sync.WaitGroup) {
	defer wg.Done()
	if err := req.svc.Call(req.mt, req.arg, req.reply); err != nil {
		req.h.Error = err.Error()
		server.sendResponse(cc, req.h, invalidRequest, sending)
		return
	}
	server.sendResponse(cc, req.h, req.reply.Interface(), sending)
}

// Register publishes in the server the set of methods
func (server *Server) Register(svr interface{}) error {
	s := newService(svr)
	if _, dup := server.serviceMap.LoadOrStore(s.name, s); dup {
		return errors.New("rpc: service already defined: " + s.name)
	}
	return nil
}

// Register publishes the receiver's methods in the DefaultServer.
func Register(svr interface{}) error {
	return DefaultServer.Register(svr)
}

func (server *Server) findService(serviceMethod string) (svc *service, mt *methodType, err error) {
	dot := strings.LastIndex(serviceMethod, ".")
	if dot < 0 {
		err = errors.New("rpc server: service/method request ill-formed:" + serviceMethod)
		return
	}
	serviceName, methodName := serviceMethod[:dot], serviceMethod[dot+1:]
	svcInterface, ok := server.serviceMap.Load(serviceName)
	if !ok {
		err = errors.New("rpc server: can't find service " + serviceName)
		return
	}
	svc = svcInterface.(*service)
	if mt = svc.method[methodName]; mt == nil {
		err = errors.New("rpc server: can't find method " + methodName)
	}
	return
}
