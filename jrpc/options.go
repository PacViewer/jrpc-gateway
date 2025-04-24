package jrpc

import (
	"github.com/rs/cors"
	"time"
)

type jrpcOpt struct {
	CustomHeadersKey  []string
	ReadHeaderTimeout time.Duration
	CorsOptions       *cors.Options
}

type Option func(*jrpcOpt)

func defaultOpt() *jrpcOpt {
	return &jrpcOpt{
		ReadHeaderTimeout: 3 * time.Second,
		CustomHeadersKey:  []string{"Authorization"},
	}
}

// WithCustomHeaders add custom headers to metadata.MD
func WithCustomHeaders(headersKey ...string) Option {
	return func(opt *jrpcOpt) {
		opt.CustomHeadersKey = append(opt.CustomHeadersKey, headersKey...)
	}
}

// WithReadHeaderTimeout set custom read header timeout
func WithReadHeaderTimeout(timeout time.Duration) Option {
	return func(opt *jrpcOpt) {
		opt.ReadHeaderTimeout = timeout
	}
}

// WithCorsOrigins is an Option function that allows setting custom CORS options for the JSON-RPC server.
// It takes a pointer to a cors.Options object and updates the server's configuration to use the provided CORS settings.
// This function enables the configuration of allowed origins, methods, headers, and other CORS-related options.
//
// Usage example:
//
//	server := NewServer(WithCorsOrigins(&cors.Options{
//	    AllowedOrigins: []string{"https://example.com"},
//	}))
func WithCorsOrigins(o *cors.Options) Option {
	return func(opt *jrpcOpt) {
		opt.CorsOptions = o
	}
}
