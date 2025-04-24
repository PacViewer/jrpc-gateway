package main

import (
	"flag"
	"github.com/golang/glog"
	"github.com/pacviewer/jrpc-gateway/protoc-gen-jrpc-gateway/internal"
	"os"
)

var (
	generateOpenRPC *bool
	mergedOpenRPC   *bool
)

func init() {
	generateOpenRPC = flag.Bool("generate-openrpc", false, "generate open RPC")
	mergedOpenRPC = flag.Bool("merged-openrpc", false, "merged open RPC")
	flag.Parse()
}

func main() {
	defer glog.Flush()
	in, err := internal.Unmarshal(os.Stdin)
	if err != nil {
		internal.Marshal(os.Stdout, nil, err)
		return
	}

	gen := internal.NewGenerator()
	out, err := gen.Generate(in, *generateOpenRPC, *mergedOpenRPC)
	if err != nil {
		internal.Marshal(os.Stdout, nil, err)
		return
	}

	internal.Marshal(os.Stdout, out, nil)
}
