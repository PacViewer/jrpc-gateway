package internal

import (
	"google.golang.org/genproto/googleapis/api/annotations"
	proto "google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

type Method struct {
	Name        string
	Description string
	Request     any
	Response    any
}

type Data struct {
	Metadata any
	Methods  []*Method
}

func filesToMethods(files []*descriptorpb.FileDescriptorProto) ([]*Method, error) {
	methods := make([]*Method, 0)
	for _, f := range files {
		for _, s := range f.GetService() {
			for _, m := range s.GetMethod() {
				method := &Method{
					Name: m.GetName(),
				}
				req := map[string]string{}
				resp := map[string]string{}

				if m.GetOptions() != nil {
					if ext, err := proto.GetExtension(m.GetOptions(), annotations.E_Http); err == nil {
						httpRule := ext.(*annotations.HttpRule)
						if httpRule != nil {
							method.Description = httpRule.GetBody()
						}
					}
				}
			}
		}
	}
}
