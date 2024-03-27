package internal

import (
	"io"

	"github.com/golang/glog"
	"google.golang.org/protobuf/proto"
	plugin "google.golang.org/protobuf/types/pluginpb"
)

// Marshal .
func Marshal(w io.Writer, out []*plugin.CodeGeneratorResponse_File, err error) {
	var response = new(plugin.CodeGeneratorResponse)

	if err != nil {
		response.Error = proto.String(err.Error())
	} else {
		response.File = out
	}

	buf, err := proto.Marshal(response)
	if err != nil {
		glog.Fatal(err)
	}

	if _, err := w.Write(buf); err != nil {
		glog.Fatal(err)
	}
}
