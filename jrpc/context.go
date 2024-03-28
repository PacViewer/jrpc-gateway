package jrpc

import (
	"context"
	"encoding/json"
)

type contextKey string

const requestID contextKey = "request_id"

func RequestID(ctx context.Context) json.RawMessage {
	raw := ctx.Value(requestID)
	value, ok := raw.(json.RawMessage)
	if !ok {
		return json.RawMessage{}
	}
	return value
}
