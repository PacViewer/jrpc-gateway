package jsonrpc

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/creachadair/jrpc2/handler"
	"github.com/creachadair/jrpc2/jhttp"
)

type method = func(ctx context.Context, message json.RawMessage) (any, error)

type Service interface {
	Methods() map[string]method
}

type Server struct {
	handler http.Handler
}

func NewServer() *Server {
	return &Server{nil}
}

func (s *Server) RegisterServices(svs ...Service) {
	hd := handler.Map{}
	for _, sv := range svs {
		for m, h := range sv.Methods() {
			hd[m] = handler.New(h)
		}
	}
	s.handler = jhttp.NewBridge(hd, nil)
}

func (s *Server) AsyncHttpHandle(w http.ResponseWriter, r *http.Request) {
	go s.handler.ServeHTTP(w, r)
}

func (s *Server) HttpHandle(w http.ResponseWriter, r *http.Request) {
	s.handler.ServeHTTP(w, r)
}
