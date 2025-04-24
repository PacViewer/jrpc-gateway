package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/format"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
	plugin "google.golang.org/protobuf/types/pluginpb"
	"path/filepath"
	"strings"
	"text/template"
)

type Generator interface {
	Generate(request *plugin.CodeGeneratorRequest,
		generateOpenRPC, mergedOpenRPC bool) ([]*plugin.CodeGeneratorResponse_File, error)
}

type generator struct{}

func NewGenerator() Generator {
	return new(generator)
}

func (g *generator) Generate(request *plugin.CodeGeneratorRequest,
	generateOpenRPC, mergedOpenRPC bool) ([]*plugin.CodeGeneratorResponse_File, error) {
	var files []*plugin.CodeGeneratorResponse_File

	for _, fileToGenerate := range request.FileToGenerate {
		var file *descriptorpb.FileDescriptorProto
		for _, f := range request.ProtoFile {
			if f.GetName() == fileToGenerate {
				file = f
			}
		}

		if file == nil {
			continue
		}

		var (
			code *string
			err  error
		)

		if generateOpenRPC {
			code, err = g.generateOpenRPC(file)
		} else {
			code, err = g.generateJGW(file)
		}
		if err != nil {
			return nil, err
		}
		if code == nil {
			continue
		}

		var output string
		if generateOpenRPC {
			// Output as-is (JSON)
			var prettyJSON bytes.Buffer
			if err := json.Indent(&prettyJSON, []byte(*code), "", "  "); err != nil {
				return nil, fmt.Errorf("invalid JSON output: %w", err)
			}
			output = prettyJSON.String()
		} else {
			// Format Go source
			formattedCode, err := format.Source([]byte(*code))
			if err != nil {
				return nil, fmt.Errorf("failed to format Go code: %w", err)
			}
			output = string(formattedCode)
		}

		fileName := g.fileName(file, generateOpenRPC)
		files = append(files, &plugin.CodeGeneratorResponse_File{
			Name:    proto.String(fileName),
			Content: proto.String(output),
		})
	}

	return files, nil
}

func (g *generator) generateJGW(file *descriptorpb.FileDescriptorProto) (*string, error) {
	return g.generate(file, JGWTmpl)
}

func (g *generator) generateOpenRPC(file *descriptorpb.FileDescriptorProto) (*string, error) {
	return g.generate(file, _openrpcTmpl)
}

func (g *generator) generateMergedOpenRPC(files []*descriptorpb.FileDescriptorProto) ([]*plugin.CodeGeneratorResponse_File, error) {
	doc := OpenRPCDoc{
		OpenRPC: "1.0.0-rc1",
		Info: struct {
			Title       string `json:"title"`
			Description string `json:"description"`
			Version     string `json:"version"`
		}{
			Title:       "Pactus Blockchain API",
			Description: "Merged OpenRPC spec from all protobuf services",
			Version:     "1.0.0",
		},
		Servers: []Server{
			{
				Name:    "Local",
				Summary: "Local JSON-RPC",
				URL:     "http://localhost:8080/",
			},
		},
		Methods:    []OpenRPCMethod{},
		Components: Components{Schemas: map[string]interface{}{}},
	}

	for _, file := range files {
		if len(file.GetService()) == 0 {
			continue
		}

		// Here you would walk through file.Service, file.MessageType, file.EnumType
		// and append to doc.Methods and doc.Components.Schemas

		// For simplicity:
		for _, svc := range file.GetService() {
			for _, m := range svc.GetMethod() {
				doc.Methods = append(doc.Methods, OpenRPCMethod{
					Name:   fmt.Sprintf("%s.%s", svc.GetName(), m.GetName()),
					Params: []Content{}, // You'd parse input type
					Result: Content{},   // You'd parse output type
				})
			}
		}
	}

	jsonData, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return nil, err
	}

	file := &plugin.CodeGeneratorResponse_File{
		Name:    proto.String("openrpc_merged.json"),
		Content: proto.String(string(jsonData)),
	}
	return []*plugin.CodeGeneratorResponse_File{file}, nil
}

func (g *generator) generate(file *descriptorpb.FileDescriptorProto, tmpl *template.Template) (*string, error) {
	if len(file.GetService()) == 0 {
		return nil, nil
	}
	buf := bytes.NewBufferString("")
	err := tmpl.Execute(buf, file)
	if err != nil {
		return nil, err
	}
	out := buf.String()
	return &out, nil
}

func (g *generator) fileName(file *descriptorpb.FileDescriptorProto, isOpenRPC bool) string {
	name := file.GetName()
	ext := filepath.Ext(name)
	base := strings.TrimSuffix(name, ext)

	if isOpenRPC {
		return fmt.Sprintf("%s_openrpc.json", base)
	}
	return fmt.Sprintf("%s_jgw.pb.go", base)
}
