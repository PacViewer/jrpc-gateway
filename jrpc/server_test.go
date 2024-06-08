package jrpc

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

type MockService struct{}

func (ms *MockService) Methods() map[string]method {
	return map[string]method{
		"testMethod": ms.testMethod,
	}
}

func (ms *MockService) testMethod(ctx context.Context, message json.RawMessage) (any, error) {
	var params map[string]string
	if err := json.Unmarshal(message, &params); err != nil {
		return nil, err
	}
	return map[string]string{"response": "Hello " + params["name"]}, nil
}

func TestServer(t *testing.T) {
	server := NewServer()

	mockService := &MockService{}
	server.RegisterServices(mockService)

	ts := httptest.NewServer(http.HandlerFunc(server.HttpHandler))
	defer ts.Close()

	requestBody, err := json.Marshal(map[string]any{
		"jsonrpc": "2.0",
		"id":      "1",
		"method":  "testMethod",
		"params":  map[string]string{"name": "javad"},
	})
	assert.NoError(t, err, "Failed to marshal request body")

	resp, err := http.Post(ts.URL, "application/json", bytes.NewBuffer(requestBody))
	assert.NoError(t, err, "Failed to make POST request")
	defer resp.Body.Close()

	var response map[string]any
	err = json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(t, err, "Failed to decode response body")

	assert.Equal(t, "1", response["id"], "Expected id '1'")
	assert.Equal(t, "2.0", response["jsonrpc"], "Expected jsonrpc '2.0'")

	result, ok := response["result"].(map[string]any)
	assert.True(t, ok, "Expected result to be a map")

	expectedResponse := "Hello javad"
	assert.Equal(t, expectedResponse, result["response"], "Expected response '%s'", expectedResponse)
}
