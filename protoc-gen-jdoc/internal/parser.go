package internal

import (
	"fmt"
	"slices"
	"strings"
)

type RenderType = string

const (
	Version  RenderType = "version"
	Format              = "format"
	Html                = "html"
	Markdown            = "markdown"
	Postman             = "postman"
)

var (
	validParams  = []string{"version", "format"}
	validFormats = []string{"html", "markdown", "postman"}
)

type Params struct {
	Version string
	Formats map[RenderType]bool
}

func ParseParams(params string) (*Params, error) {
	par := &Params{
		Version: "v1.0.0",
		Formats: make(map[RenderType]bool),
	}

	ps := strings.Split(params, "/")
	for _, p := range ps {
		kvs := strings.Split(p, "=")
		if len(kvs) != 2 {
			return nil, fmt.Errorf("invalid request parameter %v", p)
		}
		if !slices.Contains[[]string, string](validParams, kvs[0]) {
			return nil, fmt.Errorf("invalid request parameter %v", kvs[0])
		}
		vs := strings.Split(kvs[1], ",")
		switch kvs[0] {
		case Version:
			par.Version = kvs[1]
		case Format:
			for _, v := range vs {
				if !slices.Contains[[]string, string](validFormats, v) {
					return nil, fmt.Errorf("invalid request parameter %v for %v", v, kvs[0])
				}
				par.Formats[v] = true
			}
		}
	}
	if len(par.Formats) == 0 {
		par.Formats["html"] = true
	}
	return par, nil
}
