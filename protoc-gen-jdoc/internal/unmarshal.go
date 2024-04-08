package internal

import (
	"fmt"
	"io"

	"google.golang.org/protobuf/proto"
	plugin "google.golang.org/protobuf/types/pluginpb"
)

// Unmarshal parses a protoc request to a proto Message
func Unmarshal(r io.Reader) (*plugin.CodeGeneratorRequest, error) {
	input, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read protoc request: %v", err)
	}
	req := new(plugin.CodeGeneratorRequest)
	if err = proto.Unmarshal(input, req); err != nil {
		return nil, fmt.Errorf("failed to unmarshal protoc request: %v", err)
	}
	return req, nil
}
