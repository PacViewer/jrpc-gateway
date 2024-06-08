package jrpc

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"time"

	"github.com/creachadair/jrpc2/handler"
	"github.com/creachadair/jrpc2/jhttp"
)

type method = func(ctx context.Context, message json.RawMessage) (any, error)

type Service interface {
	Methods() map[string]method
}

type Server struct {
	sv      *http.Server
	handler http.Handler
}

// NewServer create json rpc server
func NewServer() *Server {
	sv := new(Server)

	mux := http.NewServeMux()
	mux.HandleFunc("/", sv.httpHandler)
	server := &http.Server{
		ReadHeaderTimeout: 3 * time.Second,
		Handler:           mux,
	}

	sv.sv = server

	return sv
}

// GracefulStop stops the JRPC server gracefully. It stops the server from
// accepting new connections.
func (s *Server) GracefulStop(ctx context.Context) error {
	return s.sv.Shutdown(ctx)
}

// Serve accepts incoming connections on the listener lis.
func (s *Server) Serve(listener net.Listener) error {
	s.sv.Addr = listener.Addr().String()
	return s.sv.Serve(listener)
}

// RegisterServices register jgw servers to jrpc handler
func (s *Server) RegisterServices(svs ...Service) {
	hd := handler.Map{}
	for _, sv := range svs {
		for m, h := range sv.Methods() {
			hd[m] = handler.New(h)
		}
	}
	s.handler = jhttp.NewBridge(hd, nil)
}

func (s *Server) httpHandler(w http.ResponseWriter, r *http.Request) {
	s.handler.ServeHTTP(w, r)
}
