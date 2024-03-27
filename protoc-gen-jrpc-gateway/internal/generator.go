package internal

import (
	"bytes"
	"fmt"
	"go/format"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
	plugin "google.golang.org/protobuf/types/pluginpb"
	"path/filepath"
	"strings"
)

type Generator interface {
	Generate(request *plugin.CodeGeneratorRequest) ([]*plugin.CodeGeneratorResponse_File, error)
}

type generator struct{}

func NewGenerator() Generator {
	return new(generator)
}

func (g *generator) Generate(request *plugin.CodeGeneratorRequest) ([]*plugin.CodeGeneratorResponse_File, error) {
	var files []*plugin.CodeGeneratorResponse_File

	for _, fileToGenerate := range request.FileToGenerate {
		var file *descriptorpb.FileDescriptorProto
		for _, f := range request.ProtoFile {
			if f.GetName() == fileToGenerate {
				file = f
			}
		}
		code, err := g.generate(file)
		if err != nil {
			return nil, err
		}

		if code == nil {
			continue
		}

		formattedCode, err := format.Source([]byte(*code))
		if err != nil {
			return nil, err
		}

		fileName := g.fileName(file)
		files = append(files, &plugin.CodeGeneratorResponse_File{
			Name:    proto.String(fileName),
			Content: proto.String(string(formattedCode)),
		})
	}

	return files, nil
}

func (g *generator) generate(file *descriptorpb.FileDescriptorProto) (*string, error) {
	if len(file.GetService()) == 0 {
		return nil, nil
	}
	buf := bytes.NewBufferString("")
	err := FileTmpl.Execute(buf, file)
	if err != nil {
		return nil, err
	}

	out := buf.String()
	return &out, nil
}

func (g *generator) fileName(file *descriptorpb.FileDescriptorProto) string {
	name := file.GetName()
	ext := filepath.Ext(name)
	base := strings.TrimSuffix(name, ext)
	output := fmt.Sprintf("%s.pb.jgw.go", base)
	return output
}
