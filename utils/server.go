package utils

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"net/rpc/jsonrpc"
	"reflect"
	"time"

	"github.com/cenk/rpc2"
	"go.uber.org/zap"
	"golang.org/x/net/websocket"

	_ "net/http/pprof"
)

type Server struct {
	rpcEnabled  bool
	httpEnabled bool
	birpcSrv    *rpc2.Server
}

func (s *Server) RPCRegister(rcvr interface{}) {
	if err := rpc.Register(rcvr); err != nil {
		Logger.Error("error registering RPC method: ", zap.Error(err))
	}
	s.rpcEnabled = true
}

func (s *Server) RPCRegisterName(name string, rcvr interface{}) {
	if err := rpc.RegisterName(name, rcvr); err != nil {
		Logger.Error("error registering RPC method: ", zap.Error(err))
	}
	s.rpcEnabled = true
}

func (s *Server) RegisterHTTPFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	http.HandleFunc(pattern, handler)
	s.httpEnabled = true
}

// Registers a new BiJsonRpc name
func (s *Server) BiRPCRegisterName(method string, handlerFunc interface{}) {
	if s.birpcSrv == nil {
		s.birpcSrv = rpc2.NewServer()
	}
	s.birpcSrv.Handle(method, handlerFunc)
}

func (s *Server) BiRPCRegister(rcvr interface{}) {
	if s.birpcSrv == nil {
		s.birpcSrv = rpc2.NewServer()
	}
	rcvType := reflect.TypeOf(rcvr)
	for i := 0; i < rcvType.NumMethod(); i++ {
		method := rcvType.Method(i)
		if method.Name != "Call" {
			s.birpcSrv.Handle("SMGenericV1."+method.Name, method.Func.Interface())
		}
	}
}

func (s *Server) ServeJSON(addr string) {
	if !s.rpcEnabled {
		return
	}
	lJSON, e := net.Listen("tcp", addr)
	if e != nil {
		log.Fatal("ServeJSON listen error:", e)
	}
	Logger.Info(fmt.Sprintf("Starting AccuRate JSON server at %s.", addr))
	errCnt := 0
	var lastErrorTime time.Time
	for {
		conn, err := lJSON.Accept()
		if err != nil {
			Logger.Error("<accuRate> JSON accept : ", zap.String("error", err.Error()))
			now := time.Now()
			if now.Sub(lastErrorTime) > 5*time.Second {
				errCnt = 0 // reset error count if last error was more than 5 seconds ago
			}
			lastErrorTime = time.Now()
			errCnt++
			if errCnt > 50 { // Too many errors in short interval, network buffer failure most probably
				break
			}
			continue
		}
		//utils.Logger.Info(fmt.Sprintf("<CGRServer> New incoming connection: %v", conn.RemoteAddr()))
		go jsonrpc.ServeConn(conn)
	}

}

func (s *Server) ServeGOB(addr string) {
	if !s.rpcEnabled {
		return
	}
	lGOB, e := net.Listen("tcp", addr)
	if e != nil {
		log.Fatal("ServeGOB listen error:", e)
	}
	Logger.Info(fmt.Sprintf("Starting AccuRate GOB server at %s.", addr))
	errCnt := 0
	var lastErrorTime time.Time
	for {
		conn, err := lGOB.Accept()
		if err != nil {
			Logger.Error("<accuRate> GOB accept ", zap.String("err", err.Error()))
			now := time.Now()
			if now.Sub(lastErrorTime) > 5*time.Second {
				errCnt = 0 // reset error count if last error was more than 5 seconds ago
			}
			lastErrorTime = time.Now()
			errCnt++
			if errCnt > 50 { // Too many errors in short interval, network buffer failure most probably
				break
			}
			continue
		}

		//utils.Logger.Info(fmt.Sprintf("<CGRServer> New incoming connection: %v", conn.RemoteAddr()))
		go rpc.ServeConn(conn)
	}
}

func (s *Server) ServeHTTP(addr string) {
	if s.rpcEnabled {
		http.HandleFunc("/jsonrpc", func(w http.ResponseWriter, req *http.Request) {
			defer req.Body.Close()
			w.Header().Set("Content-Type", "application/json")
			res := NewRPCRequest(req.Body).Call()
			io.Copy(w, res)
		})
		http.Handle("/ws", websocket.Handler(func(ws *websocket.Conn) {
			jsonrpc.ServeConn(ws)
		}))
		s.httpEnabled = true
	}
	if !s.httpEnabled {
		return
	}
	Logger.Info(fmt.Sprintf("Starting AccuRate HTTP server at %s.", addr))
	if err := http.ListenAndServe(addr, nil); err != nil {
		Logger.Error("error serving HTTP: ", zap.Error(err))
	}
}

func (s *Server) ServeBiJSON(addr string) {
	if s.birpcSrv == nil {
		return
	}
	lBiJSON, e := net.Listen("tcp", addr)
	if e != nil {
		log.Fatal("ServeBiJSON listen error:", e)
	}
	Logger.Info(fmt.Sprintf("Starting accuRate BiJSON server at %s.", addr))
	s.birpcSrv.Accept(lBiJSON)
}

// rpcRequest represents a RPC request.
// rpcRequest implements the io.ReadWriteCloser interface.
type rpcRequest struct {
	r    io.Reader     // holds the JSON formated RPC request
	rw   io.ReadWriter // holds the JSON formated RPC response
	done chan bool     // signals then end of the RPC request
}

// NewRPCRequest returns a new rpcRequest.
func NewRPCRequest(r io.Reader) *rpcRequest {
	var buf bytes.Buffer
	done := make(chan bool)
	return &rpcRequest{r, &buf, done}
}

func (r *rpcRequest) Read(p []byte) (n int, err error) {
	return r.r.Read(p)
}

func (r *rpcRequest) Write(p []byte) (n int, err error) {
	n, err = r.rw.Write(p)
	r.done <- true
	return
}

func (r *rpcRequest) Close() error {
	//r.done <- true // seem to be called sometimes before the write command finishes!
	return nil
}

// Call invokes the RPC request, waits for it to complete, and returns the results.
func (r *rpcRequest) Call() io.Reader {
	go jsonrpc.ServeConn(r)
	<-r.done
	return r.rw
}
