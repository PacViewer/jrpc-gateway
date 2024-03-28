package jrpc

import (
	"bytes"
	"errors"
	"net/http"
	"strings"
)

func prepareBody(r *http.Request) (*bytes.Buffer, error) {
	if !strings.HasPrefix(r.Header.Get(contentType), contentTypeJSON) {
		return nil, errors.New("invalid content-type")
	}

	body := bytes.NewBuffer(make([]byte, 0, r.ContentLength))
	if _, err := body.ReadFrom(r.Body); err != nil {
		return nil, errors.New("invalid body")
	}

	return body, nil
}
