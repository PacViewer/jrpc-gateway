package internal

import (
	"bytes"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
	plugin "google.golang.org/protobuf/types/pluginpb"
	"log"
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
	var reqFiles []*descriptorpb.FileDescriptorProto

	params, err := ParseParams(*request.Parameter)
	if err != nil {
		return nil, err
	}

	for _, fileToGenerate := range request.FileToGenerate {
		for _, f := range request.ProtoFile {
			if f.GetName() == fileToGenerate {
				reqFiles = append(reqFiles, f)
			}
		}
	}
	log.Println(reqFiles[0].Service[0].GetOptions())
	if exists := params.Formats[Markdown]; exists {
		code, err := g.generateMarkdown(reqFiles, *params)
		if err != nil {
			return nil, err
		}

		fileName := g.markdownFileName()
		files = append(files, &plugin.CodeGeneratorResponse_File{
			Name:    proto.String(fileName),
			Content: proto.String(*code),
		})
	}
	if exists := params.Formats[Html]; exists {
		code, err := g.generateHtml(reqFiles, *params)
		if err != nil {
			return nil, err
		}

		fileName := g.htmlFileName()
		files = append(files, &plugin.CodeGeneratorResponse_File{
			Name:    proto.String(fileName),
			Content: proto.String(*code),
		})
	}
	if exists := params.Formats[Postman]; exists {
		code, err := g.generatePostman(reqFiles, *params)
		if err != nil {
			return nil, err
		}

		fileName := g.postmanFileName()
		files = append(files, &plugin.CodeGeneratorResponse_File{
			Name:    proto.String(fileName),
			Content: proto.String(*code),
		})
	}

	return files, nil
}

func (g *generator) generateMarkdown(files []*descriptorpb.FileDescriptorProto, params Params) (*string, error) {
	metadata := map[string]any{
		"Version": params.Version,
	}
	data := filesToData(files)

	return g.generate(data, markdownFile)
}

func (g *generator) generateHtml(files []*descriptorpb.FileDescriptorProto, params Params) (*string, error) {
	metadata := map[string]any{
		"Version": params.Version,
	}
	data := Data{
		Metadata: metadata,
		Files:    files,
	}
	return g.generate(data, htmlFile)
}

func (g *generator) generatePostman(files []*descriptorpb.FileDescriptorProto, params Params) (*string, error) {
	metadata := map[string]any{
		"Version": params.Version,
	}
	data := Data{
		Metadata: metadata,
		Files:    files,
	}
	return g.generate(data, postmanFile)
}

func (g *generator) generate(data Data, file string) (*string, error) {
	buf := bytes.NewBufferString("")
	err := Tmpl(file).Execute(buf, data)
	if err != nil {
		return nil, err
	}

	out := buf.String()
	return &out, nil
}

func (g *generator) htmlFileName() string {
	return "jdoc.html"
}

func (g *generator) postmanFileName() string {
	return "jdoc.json"
}

func (g *generator) markdownFileName() string {
	return "jdoc.md"
}
