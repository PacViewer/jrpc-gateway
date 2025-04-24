package jrpc

import (
	"github.com/rs/cors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestOptions(t *testing.T) {
	dur := 5 * time.Second

	corsHeaders := []string{"Accept", "Content-Type", "Content-Length", "Accept-Encoding", "X-CSRF-Token"}
	corsOrigins := []string{"*"}
	corsMethods := []string{"GET", "HEAD", "POST", "PUT", "DELETE"}

	opts := []Option{
		WithCustomHeaders("foobar"),
		WithReadHeaderTimeout(dur),
		WithCorsOrigins(&cors.Options{
			AllowedOrigins:      corsOrigins,
			AllowedMethods:      corsMethods,
			AllowedHeaders:      corsHeaders,
			AllowCredentials:    true,
			AllowPrivateNetwork: true,
		}),
	}

	def := defaultOpt()

	for _, opt := range opts {
		opt(def)
	}

	assert.Len(t, def.CustomHeadersKey, 2)
	assert.Equal(t, def.ReadHeaderTimeout.String(), dur.String())
	require.NotNil(t, def.CorsOptions)
	assert.Equal(t, def.CorsOptions.AllowedOrigins, corsOrigins)
	assert.Equal(t, def.CorsOptions.AllowedMethods, corsMethods)
	assert.Equal(t, def.CorsOptions.AllowedHeaders, corsHeaders)
	assert.Equal(t, "*", def.CorsOptions.AllowedOrigins[0])
	assert.Equal(t, def.CorsOptions.AllowCredentials, true)
	assert.Equal(t, def.CorsOptions.AllowPrivateNetwork, true)
}
