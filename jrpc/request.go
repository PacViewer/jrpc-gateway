package jrpc

import (
	"encoding/json"
	"errors"
)

// https://www.jsonrpc.org/specification
// section 4
type request struct {
	ID      json.RawMessage `json:"id,omitempty"`
	Jsonrpc string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

func (r *request) isNotification() bool {
	return r.ID == nil
}

func (r *request) isValidVersion() bool {
	return r.Jsonrpc == jsonrpcVersion
}

func (r *request) isMethodEmpty() bool {
	return len(r.Method) == 0
}

func (r *request) validate() error {
	if !r.isValidVersion() {
		return errors.New("invalid json-rpc version")
	}

	if r.isMethodEmpty() {
		return errors.New("method is empty")
	}

	return nil
}
