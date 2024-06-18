package jrpc

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestOptions(t *testing.T) {
	dur := 5 * time.Second

	opts := []Option{
		WithCustomHeaders("foobar"),
		WithReadHeaderTimeout(dur),
	}

	def := defaultOpt()

	for _, opt := range opts {
		opt(def)
	}

	assert.Len(t, def.CustomHeadersKey, 2)
	assert.Equal(t, def.ReadHeaderTimeout.String(), dur.String())
}
