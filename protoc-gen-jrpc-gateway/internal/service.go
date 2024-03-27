package internal

import (
	"context"
	"encoding/json"
)

type Method = func(ctx context.Context, message json.RawMessage) (any, error)

type Service interface {
	Methods() map[string]Method
}
