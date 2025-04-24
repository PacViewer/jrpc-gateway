package jrpc

import (
	"context"
	"encoding/json"
	"github.com/creachadair/jrpc2"
	"github.com/creachadair/jrpc2/handler"
	"github.com/creachadair/jrpc2/jhttp"
	"github.com/rs/cors"
	"google.golang.org/grpc/metadata"
	"io"
	"net"
	"net/http"
	"strings"
)

type method = func(ctx context.Context, message json.RawMessage) (any, error)

type Service interface {
	Methods() map[string]method
}

type Server struct {
	sv            *http.Server
	handler       http.Handler
	customHeaders []string
}

type paramsAndHeaders struct {
	Headers metadata.MD     `json:"headers,omitempty"`
	Params  json.RawMessage `json:"params"`
}

// NewServer create json rpc server
func NewServer(opts ...Option) *Server {
	sv := new(Server)
	opt := defaultOpt()

	mux := http.NewServeMux()
	mux.HandleFunc("/", sv.httpHandler)
	server := &http.Server{
		ReadHeaderTimeout: opt.ReadHeaderTimeout,
		Handler:           mux,
	}

	for _, o := range opts {
		o(opt)
	}

	if opt.CorsOptions != nil {
		c := cors.New(*opt.CorsOptions)
		server.Handler = c.Handler(mux)
	}

	sv.customHeaders = opt.CustomHeadersKey
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
	s.handler = jhttp.NewBridge(hd, &jhttp.BridgeOptions{
		ParseRequest: func(req *http.Request) ([]*jrpc2.ParsedRequest, error) {
			body, err := io.ReadAll(req.Body)
			if err != nil {
				return nil, err
			}
			prs, err := jrpc2.ParseRequests(body)
			if err != nil {
				return nil, err
			}

			// Decorate the incoming request parameters with the headers.
			for _, pr := range prs {
				w, err := json.Marshal(paramsAndHeaders{
					Headers: s.headersToMetadata(req),
					Params:  pr.Params,
				})
				if err != nil {
					return nil, err
				}
				pr.Params = w
			}
			return prs, nil
		},
	})
}

func (s *Server) httpHandler(w http.ResponseWriter, r *http.Request) {
	s.handler.ServeHTTP(w, r)
}

func (s *Server) headersToMetadata(r *http.Request) metadata.MD {
	headersMap := make(map[string]string)

	for _, header := range s.customHeaders {
		canonicalHeader := http.CanonicalHeaderKey(header)
		if v, ok := r.Header[canonicalHeader]; ok {
			if len(v) > 0 {
				headersMap[strings.ToLower(canonicalHeader)] = v[0]
			}
		}
	}

	return metadata.New(headersMap)
}
