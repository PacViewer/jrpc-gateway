package jrpc

import "time"

type jrpcOpt struct {
	CustomHeadersKey  []string
	ReadHeaderTimeout time.Duration
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
