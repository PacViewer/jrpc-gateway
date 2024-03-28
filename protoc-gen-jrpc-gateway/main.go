package main

import (
	"github.com/golang/glog"
	"github.com/pacviewer/jrpc-gateway/protoc-gen-jrpc-gateway/internal"
	"os"
)

func main() {
	defer glog.Flush()
	in, err := internal.Unmarshal(os.Stdin)
	if err != nil {
		internal.Marshal(os.Stdout, nil, err)
		return
	}

	gen := internal.NewGenerator()
	out, err := gen.Generate(in)
	if err != nil {
		internal.Marshal(os.Stdout, nil, err)
		return
	}

	internal.Marshal(os.Stdout, out, nil)
}
