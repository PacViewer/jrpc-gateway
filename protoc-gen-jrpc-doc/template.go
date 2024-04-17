package jrpc_doc

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/pacviewer/jrpc-gateway/protoc-gen-jrpc-doc/extensions"
	"github.com/pseudomuto/protokit"
)

type Type = string

const (
	ObjectType  Type = "object"
	ArrayType   Type = "array"
	StringType  Type = "string"
	NumericType Type = "numeric"
)

type JsonType = string

const (
	ObjectJsonType         JsonType = "json object"
	ArrayJsonType          JsonType = "json array"
	NumericJsonType        JsonType = "numeric"
	StringJsonType         JsonType = "string"
	BoolJsonType           JsonType = "boolean"
	KeyValueObjectJsonType JsonType = "key:value json object"
)

var (
	Jdict JDict   = JDict{}
	Files []*File = []*File{}
)

// Template is a type for encapsulating all the parsed files, messages, fields, enums, services, extensions, etc. into
// an object that will be supplied to a go template.
type Template struct {
	// The files that were parsed
	Files []*File `json:"files"`
	// Details about the scalar values and their respective types in supported languages.
	Scalars []*ScalarValue `json:"scalarValueTypes"`
	// JDict is fields placeholders
	JDict JDict
}

// NewTemplate creates a Template object from a set of descriptors.
func NewTemplate(descs []*protokit.FileDescriptor, jdict JDict) *Template {
	files := make([]*File, 0, len(descs))
	Jdict = jdict

	for _, f := range descs {
		file := &File{
			Name:          f.GetName(),
			Description:   description(f.GetSyntaxComments().String()),
			Package:       f.GetPackage(),
			HasEnums:      len(f.Enums) > 0,
			HasExtensions: len(f.Extensions) > 0,
			HasMessages:   len(f.Messages) > 0,
			HasServices:   len(f.Services) > 0,
			Enums:         make(orderedEnums, 0, len(f.Enums)),
			Extensions:    make(orderedExtensions, 0, len(f.Extensions)),
			Messages:      make(orderedMessages, 0, len(f.Messages)),
			Services:      make(orderedServices, 0, len(f.Services)),
			Options:       mergeOptions(extractOptions(f.GetOptions()), extensions.Transform(f.OptionExtensions)),
		}

		for _, e := range f.Enums {
			file.Enums = append(file.Enums, parseEnum(e))
		}

		for _, e := range f.Extensions {
			file.Extensions = append(file.Extensions, parseFileExtension(e))
		}

		// Recursively add nested types from messages
		var addFromMessage func(*protokit.Descriptor)
		addFromMessage = func(m *protokit.Descriptor) {
			file.Messages = append(file.Messages, parseMessage(m))
			for _, e := range m.Enums {
				file.Enums = append(file.Enums, parseEnum(e))
			}
			for _, n := range m.Messages {
				addFromMessage(n)
			}
		}
		for _, m := range f.Messages {
			addFromMessage(m)
		}
		sort.Sort(file.Enums)
		sort.Sort(file.Extensions)
		sort.Sort(file.Messages)
		Files = append(Files, file)
		for _, s := range f.Services {
			file.Services = append(file.Services, parseService(s))
		}

		sort.Sort(file.Services)

		files = append(files, file)
		Files = append(Files, file)
	}
	return &Template{Files: files, Scalars: makeScalars(), JDict: jdict}
}

func makeScalars() []*ScalarValue {
	var scalars []*ScalarValue
	json.Unmarshal(scalarsJSON, &scalars)

	return scalars
}

func mergeOptions(opts ...map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{})
	for _, opts := range opts {
		for k, v := range opts {
			if _, ok := out[k]; ok {
				continue
			}
			out[k] = v
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// CommonOptions are options common to all descriptor types.
type commonOptions interface {
	GetDeprecated() bool
}

func extractOptions(opts commonOptions) map[string]interface{} {
	out := make(map[string]interface{})
	if opts.GetDeprecated() {
		out["deprecated"] = true
	}
	switch opts := opts.(type) {
	case *descriptor.MethodOptions:
		if opts != nil && opts.IdempotencyLevel != nil {
			out["idempotency_level"] = opts.IdempotencyLevel.String()
		}
	}
	return out
}

// File wraps all the relevant parsed info about a proto file. File objects guarantee that their top-level enums,
// extensions, messages, and services are sorted alphabetically based on their "long name". Other values (enum values,
// fields, service methods) will be in the order that they're defined within their respective proto files.
//
// In the case of proto3 files, HasExtensions will always be false, and Extensions will be empty.
type File struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Package     string `json:"package"`

	HasEnums      bool `json:"hasEnums"`
	HasExtensions bool `json:"hasExtensions"`
	HasMessages   bool `json:"hasMessages"`
	HasServices   bool `json:"hasServices"`

	Enums      orderedEnums      `json:"enums"`
	Extensions orderedExtensions `json:"extensions"`
	Messages   orderedMessages   `json:"messages"`
	Services   orderedServices   `json:"services"`

	Options map[string]interface{} `json:"options,omitempty"`
}

// Option returns the named option.
func (f File) Option(name string) interface{} { return f.Options[name] }

// FileExtension contains details about top-level extensions within a proto(2) file.
type FileExtension struct {
	Name               string `json:"name"`
	LongName           string `json:"longName"`
	FullName           string `json:"fullName"`
	Description        string `json:"description"`
	Label              string `json:"label"`
	Type               string `json:"type"`
	LongType           string `json:"longType"`
	FullType           string `json:"fullType"`
	Number             int    `json:"number"`
	DefaultValue       string `json:"defaultValue"`
	ContainingType     string `json:"containingType"`
	ContainingLongType string `json:"containingLongType"`
	ContainingFullType string `json:"containingFullType"`
}

// Message contains details about a protobuf message.
//
// In the case of proto3 files, HasExtensions will always be false, and Extensions will be empty.
type Message struct {
	Name        string `json:"name"`
	LongName    string `json:"longName"`
	FullName    string `json:"fullName"`
	Description string `json:"description"`

	HasExtensions bool `json:"hasExtensions"`
	HasFields     bool `json:"hasFields"`
	HasOneofs     bool `json:"hasOneofs"`

	Extensions []*MessageExtension `json:"extensions"`
	Fields     []*MessageField     `json:"fields"`

	Options map[string]interface{} `json:"options,omitempty"`
}

// Option returns the named option.
func (m Message) Option(name string) interface{} { return m.Options[name] }

// FieldOptions returns all options that are set on the fields in this message.
func (m Message) FieldOptions() []string {
	optionSet := make(map[string]struct{})
	for _, field := range m.Fields {
		for option := range field.Options {
			optionSet[option] = struct{}{}
		}
	}
	if len(optionSet) == 0 {
		return nil
	}
	options := make([]string, 0, len(optionSet))
	for option := range optionSet {
		options = append(options, option)
	}
	sort.Strings(options)
	return options
}

// FieldsWithOption returns all fields that have the given option set.
// If no single value has the option set, this returns nil.
func (m Message) FieldsWithOption(optionName string) []*MessageField {
	fields := make([]*MessageField, 0, len(m.Fields))
	for _, field := range m.Fields {
		if _, ok := field.Options[optionName]; ok {
			fields = append(fields, field)
		}
	}
	if len(fields) > 0 {
		return fields
	}
	return nil
}

// MessageField contains details about an individual field within a message.
//
// In the case of proto3 files, DefaultValue will always be empty. Similarly, label will be empty unless the field is
// repeated (in which case it'll be "repeated").
type MessageField struct {
	Name         string `json:"name"`
	Description  string `json:"description"`
	Label        string `json:"label"`
	Type         string `json:"type"`
	LongType     string `json:"longType"`
	FullType     string `json:"fullType"`
	IsMap        bool   `json:"ismap"`
	IsRepeated   bool   `json:"isRepeated"`
	IsOneof      bool   `json:"isoneof"`
	OneofDecl    string `json:"oneofdecl"`
	DefaultValue string `json:"defaultValue"`

	Options map[string]interface{} `json:"options,omitempty"`
}

// Option returns the named option.
func (f MessageField) Option(name string) interface{} { return f.Options[name] }

// MessageExtension contains details about message-scoped extensions in proto(2) files.
type MessageExtension struct {
	FileExtension

	ScopeType     string `json:"scopeType"`
	ScopeLongType string `json:"scopeLongType"`
	ScopeFullType string `json:"scopeFullType"`
}

// Enum contains details about enumerations. These can be either top level enums, or nested (defined within a message).
type Enum struct {
	Name        string       `json:"name"`
	LongName    string       `json:"longName"`
	FullName    string       `json:"fullName"`
	Description string       `json:"description"`
	Values      []*EnumValue `json:"values"`

	Options map[string]interface{} `json:"options,omitempty"`
}

// Option returns the named option.
func (e Enum) Option(name string) interface{} { return e.Options[name] }

// ValueOptions returns all options that are set on the values in this enum.
func (e Enum) ValueOptions() []string {
	optionSet := make(map[string]struct{})
	for _, value := range e.Values {
		for option := range value.Options {
			optionSet[option] = struct{}{}
		}
	}
	if len(optionSet) == 0 {
		return nil
	}
	options := make([]string, 0, len(optionSet))
	for option := range optionSet {
		options = append(options, option)
	}
	sort.Strings(options)
	return options
}

// ValuesWithOption returns all values that have the given option set.
// If no single value has the option set, this returns nil.
func (e Enum) ValuesWithOption(optionName string) []*EnumValue {
	values := make([]*EnumValue, 0, len(e.Values))
	for _, value := range e.Values {
		if _, ok := value.Options[optionName]; ok {
			values = append(values, value)
		}
	}
	if len(values) > 0 {
		return values
	}
	return nil
}

// EnumValue contains details about an individual value within an enumeration.
type EnumValue struct {
	Name        string `json:"name"`
	Number      string `json:"number"`
	Description string `json:"description"`

	Options map[string]interface{} `json:"options,omitempty"`
}

// Option returns the named option.
func (v EnumValue) Option(name string) interface{} { return v.Options[name] }

// Service contains details about a service definition within a proto file.
type Service struct {
	Name        string           `json:"name"`
	LongName    string           `json:"longName"`
	FullName    string           `json:"fullName"`
	Description string           `json:"description"`
	Methods     []*ServiceMethod `json:"methods"`

	Options map[string]interface{} `json:"options,omitempty"`
}

// Option returns the named option.
func (s Service) Option(name string) interface{} { return s.Options[name] }

// MethodOptions returns all options that are set on the methods in this service.
func (s Service) MethodOptions() []string {
	optionSet := make(map[string]struct{})
	for _, method := range s.Methods {
		for option := range method.Options {
			optionSet[option] = struct{}{}
		}
	}
	if len(optionSet) == 0 {
		return nil
	}
	options := make([]string, 0, len(optionSet))
	for option := range optionSet {
		options = append(options, option)
	}
	sort.Strings(options)
	return options
}

// MethodsWithOption returns all methods that have the given option set.
// If no single method has the option set, this returns nil.
func (s Service) MethodsWithOption(optionName string) []*ServiceMethod {
	methods := make([]*ServiceMethod, 0, len(s.Methods))
	for _, method := range s.Methods {
		if _, ok := method.Options[optionName]; ok {
			methods = append(methods, method)
		}
	}
	if len(methods) > 0 {
		return methods
	}
	return nil
}

// Param is json-rpc friendly param for service methods.
type Param struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Type        string `json:"type"`
	Constraints string `json:"constraints"`
}

// ServiceMethod contains details about an individual method within a service.
type ServiceMethod struct {
	Name              string `json:"name"`
	Description       string `json:"description"`
	RequestType       string `json:"requestType"`
	RequestLongType   string `json:"requestLongType"`
	RequestFullType   string `json:"requestFullType"`
	RequestStreaming  bool   `json:"requestStreaming"`
	ResponseType      string `json:"responseType"`
	ResponseLongType  string `json:"responseLongType"`
	ResponseFullType  string `json:"responseFullType"`
	ResponseStreaming bool   `json:"responseStreaming"`

	RequestMessage  *Message `json:"requestMessage"`
	ResponseMessage *Message `json:"responseMessage"`

	// Params []Param `json:"params"`
	// Result []Param `json:"results"`

	// ExampleRequest  string `json:"exampleRequest"`
	// ExampleResponse string `json:"exampleResponse"`

	Params string `json:"parameters"`
	Result string `json:"result"`

	Options map[string]interface{} `json:"options,omitempty"`
}

// Option returns the named option.
func (m ServiceMethod) Option(name string) interface{} { return m.Options[name] }

// ScalarValue contains information about scalar value types in protobuf. The common use case for this type is to know
// which language specific type maps to the protobuf type.
//
// For example, the protobuf type `int64` maps to `long` in C#, and `Bignum` in Ruby. For the full list, take a look at
// https://developers.google.com/protocol-buffers/docs/proto3#scalar
type ScalarValue struct {
	ProtoType  string `json:"protoType"`
	Notes      string `json:"notes"`
	CppType    string `json:"cppType"`
	CSharp     string `json:"csType"`
	GoType     string `json:"goType"`
	JavaType   string `json:"javaType"`
	PhpType    string `json:"phpType"`
	PythonType string `json:"pythonType"`
	RubyType   string `json:"rubyType"`
}

func parseEnum(pe *protokit.EnumDescriptor) *Enum {
	enum := &Enum{
		Name:        pe.GetName(),
		LongName:    pe.GetLongName(),
		FullName:    pe.GetFullName(),
		Description: description(pe.GetComments().String()),
		Options:     mergeOptions(extractOptions(pe.GetOptions()), extensions.Transform(pe.OptionExtensions)),
	}

	for _, val := range pe.GetValues() {
		enum.Values = append(enum.Values, &EnumValue{
			Name:        val.GetName(),
			Number:      fmt.Sprint(val.GetNumber()),
			Description: description(val.GetComments().String()),
			Options:     mergeOptions(extractOptions(val.GetOptions()), extensions.Transform(val.OptionExtensions)),
		})
	}

	return enum
}

func parseFileExtension(pe *protokit.ExtensionDescriptor) *FileExtension {
	t, lt, ft := parseType(pe)

	return &FileExtension{
		Name:               pe.GetName(),
		LongName:           pe.GetLongName(),
		FullName:           pe.GetFullName(),
		Description:        description(pe.GetComments().String()),
		Label:              labelName(pe.GetLabel(), pe.IsProto3(), pe.GetProto3Optional()),
		Type:               t,
		LongType:           lt,
		FullType:           ft,
		Number:             int(pe.GetNumber()),
		DefaultValue:       pe.GetDefaultValue(),
		ContainingType:     baseName(pe.GetExtendee()),
		ContainingLongType: strings.TrimPrefix(pe.GetExtendee(), "."+pe.GetPackage()+"."),
		ContainingFullType: strings.TrimPrefix(pe.GetExtendee(), "."),
	}
}

func parseMessage(pm *protokit.Descriptor) *Message {
	msg := &Message{
		Name:          pm.GetName(),
		LongName:      pm.GetLongName(),
		FullName:      pm.GetFullName(),
		Description:   description(pm.GetComments().String()),
		HasExtensions: len(pm.GetExtensions()) > 0,
		HasFields:     len(pm.GetMessageFields()) > 0,
		HasOneofs:     len(pm.GetOneofDecl()) > 0,
		Extensions:    make([]*MessageExtension, 0, len(pm.Extensions)),
		Fields:        make([]*MessageField, 0, len(pm.Fields)),
		Options:       mergeOptions(extractOptions(pm.GetOptions()), extensions.Transform(pm.OptionExtensions)),
	}

	for _, ext := range pm.Extensions {
		msg.Extensions = append(msg.Extensions, parseMessageExtension(ext))
	}

	for _, f := range pm.Fields {
		msg.Fields = append(msg.Fields, parseMessageField(f, pm.GetOneofDecl()))
	}

	return msg
}

func parseMessageExtension(pe *protokit.ExtensionDescriptor) *MessageExtension {
	return &MessageExtension{
		FileExtension: *parseFileExtension(pe),
		ScopeType:     pe.GetParent().GetName(),
		ScopeLongType: pe.GetParent().GetLongName(),
		ScopeFullType: pe.GetParent().GetFullName(),
	}
}

func parseMessageField(pf *protokit.FieldDescriptor, oneofDecls []*descriptor.OneofDescriptorProto) *MessageField {
	t, lt, ft := parseType(pf)

	m := &MessageField{
		Name:         pf.GetName(),
		Description:  description(pf.GetComments().String()),
		Label:        labelName(pf.GetLabel(), pf.IsProto3(), pf.GetProto3Optional()),
		Type:         t,
		LongType:     lt,
		FullType:     ft,
		DefaultValue: pf.GetDefaultValue(),
		Options:      mergeOptions(extractOptions(pf.GetOptions()), extensions.Transform(pf.OptionExtensions)),
		IsOneof:      pf.OneofIndex != nil,
	}

	if m.IsOneof {
		m.OneofDecl = oneofDecls[pf.GetOneofIndex()].GetName()
	}

	// Check if this is a map.
	// See https://github.com/golang/protobuf/blob/master/protoc-gen-go/descriptor/descriptor.pb.go#L1556
	// for more information
	if m.Label == "repeated" {
		if strings.Contains(m.LongType, ".") &&
			strings.HasSuffix(m.Type, "Entry") &&
			strings.HasSuffix(m.LongType, "Entry") &&
			strings.HasSuffix(m.FullType, "Entry") {
			m.IsMap = true
		} else {
			m.IsRepeated = true
		}
	}

	return m
}

func parseService(ps *protokit.ServiceDescriptor) *Service {
	service := &Service{
		Name:        ps.GetName(),
		LongName:    ps.GetLongName(),
		FullName:    ps.GetFullName(),
		Description: description(ps.GetComments().String()),
		Options:     mergeOptions(extractOptions(ps.GetOptions()), extensions.Transform(ps.OptionExtensions)),
	}

	for _, sm := range ps.Methods {
		service.Methods = append(service.Methods, parseServiceMethod(sm))
	}

	return service
}

func parseServiceMethod(pm *protokit.MethodDescriptor) *ServiceMethod {
	reqType := baseName(pm.GetInputType())
	respType := baseName(pm.GetOutputType())
	reqMessage := new(Message)
	respMessage := new(Message)

	for _, msg := range pm.GetService().GetFile().GetMessages() {
		if reqType == msg.GetName() {
			reqMessage = parseMessage(msg)
		}
		if respType == msg.GetName() {
			respMessage = parseMessage(msg)
		}
	}

	p := parseMessageToMap(reqMessage)
	r := parseMessageToMap(respMessage)

	params, err := parseMapToPrettyJson(p)
	if err != nil {
		panic(err.Error())
	}

	result, err := parseMapToPrettyJson(r)
	if err != nil {
		panic(err.Error())
	}

	params = cleanJsonString(parseJsonComments(params))
	result = cleanJsonString(parseJsonComments(result))

	return &ServiceMethod{
		Name:              pm.GetName(),
		Description:       description(pm.GetComments().String()),
		RequestType:       reqType,
		RequestLongType:   strings.TrimPrefix(pm.GetInputType(), "."+pm.GetPackage()+"."),
		RequestFullType:   strings.TrimPrefix(pm.GetInputType(), "."),
		RequestStreaming:  pm.GetClientStreaming(),
		ResponseType:      respType,
		ResponseLongType:  strings.TrimPrefix(pm.GetOutputType(), "."+pm.GetPackage()+"."),
		ResponseFullType:  strings.TrimPrefix(pm.GetOutputType(), "."),
		ResponseStreaming: pm.GetServerStreaming(),
		Options:           mergeOptions(extractOptions(pm.GetOptions()), extensions.Transform(pm.OptionExtensions)),
		RequestMessage:    reqMessage,
		ResponseMessage:   respMessage,
		Params:            params,
		Result:            result,
	}
}

func getMessage(name string) *Message {
	for _, f := range Files {
		for _, m := range f.Messages {
			if m.Name == name || m.LongName == name {
				return m
			}
		}
	}
	return nil
}
func getEnum(name string) *Enum {
	for _, f := range Files {
		for _, m := range f.Enums {
			if m.Name == name || m.LongName == name {
				return m
			}
		}
	}
	return nil
}

func parseMessageToMap(m *Message) map[string]any {
	result := map[string]any{}
	for _, f := range m.Fields {
		if f.IsRepeated {
			msg := getMessage(f.Type)
			if msg != nil {
				if f.IsMap {
					kvs := []*MessageField{}
					kvs = append(kvs, msg.Fields...)
					ktype, vtype := jTypeToJsonType(jType(kvs[0].Type)), jTypeToJsonType(jType(kvs[1].Type))
					result[f.Name+"@"+KeyValueObjectJsonType+"#"+f.Description] = map[string]string{ktype: vtype, "...": "..."}
					continue
				}
				r := parseMessageToMap(msg)
				result[f.Name+"@"+ArrayJsonType+"#"+f.Description] = []any{r, "..."}
				continue
			}
			enum := getEnum(f.Type)
			if enum != nil {
				val := ""
				for i, v := range enum.Values {
					val += v.Name
					if i != len(enum.Values)-1 {
						val += " or "
					}
				}
				result[f.Name+"@"+ArrayJsonType+"#"+f.Description] = []string{val, "..."}
				continue
			}
			if jType(f.Type) == "numeric" {
				result[f.Name+"@"+ArrayJsonType+"#"+f.Description] = []string{"n", "..."}
				continue
			}
			if jType(f.Type) == "string" {
				result[f.Name+"@"+ArrayJsonType+"#"+f.Description] = []string{"str", "..."}
				continue
			}
			if f.Type == "bool" {
				result[f.Name+"@"+ArrayJsonType+"#"+f.Description] = []string{"true|false", "..."}
				continue
			}
		}
		msg := getMessage(f.Type)
		if msg != nil {
			if f.IsMap {
				kvs := []*MessageField{}
				kvs = append(kvs, msg.Fields...)
				ktype, vtype := jTypeToJsonType(jType(kvs[0].Type)), jTypeToJsonType(jType(kvs[1].Type))
				result[f.Name+"@"+KeyValueObjectJsonType+"#"+f.Description] = map[string]string{ktype: vtype, "...": "..."}
				continue
			}
			result[f.Name+"@"+ObjectJsonType+"#"+f.Description] = parseMessageToMap(msg)
			continue
		}
		enum := getEnum(f.Type)
		if enum != nil {
			val := ""
			for i, v := range enum.Values {
				val += v.Name
				if i != len(enum.Values)-1 {
					val += " or "
				}
			}
			result[f.Name+"@"+StringJsonType+"#"+f.Description] = val
			continue
		}
		if jType(f.Type) == "numeric" {
			result[f.Name+"@"+NumericJsonType+"#"+f.Description] = "n"
			continue
		}
		if jType(f.Type) == "string" {
			result[f.Name+"@"+StringJsonType+"#"+f.Description] = "str"
			continue
		}
		if f.Type == "bool" {
			result[f.Name+"@"+BoolJsonType+"#"+f.Description] = "true|false"
			continue
		}
	}
	return result
}

func jTypeToJsonType(t string) string {
	switch t {
	case "string":
		return "str"
	case "numeric":
		return "n"
	case "boolean":
		return "true|false"
	}
	return "str"
}

func parseJsonComments(j string) string {
	lines := strings.Split(j, "\n")
	for i, line := range lines {
		atSignSep := strings.Split(line, "@")
		if len(atSignSep) > 1 {
			descSep := strings.Split(atSignSep[1], "\"")
			descSep[1] = strings.Replace(descSep[1], ":", "", 1)
			typeDescSep := strings.Split(descSep[0], "#")
			lines[i] = atSignSep[0] + "\":" + strings.Join(descSep[1:], "\"") + "\t// (" + typeDescSep[0] + ") " + NoBrFilter(typeDescSep[1])
		}
	}
	return strings.Join(lines, "\n")
}

func cleanJsonString(s string) string {
	s = strings.ReplaceAll(s, "\"n\"", "n")
	s = strings.ReplaceAll(s, "\"...\"", "...")
	s = strings.ReplaceAll(s, "\"true|false\"", "true|false")
	return s
}

func parseMapToPrettyJson(m map[string]any) (string, error) {
	j, err := json.MarshalIndent(m, "", "\t")
	if err != nil {
		return "", err
	}
	return string(j), nil
}

func parseMessageToJson(msg Message) (string, error) {
	res := map[string]any{}
	for _, f := range msg.Fields {
		n, ok := Jdict[f.Name]
		if !ok {
			panic(fmt.Sprintf("field %s is not defined in dictionary", f.Name))
		}
		tp, ok := n[f.Type]
		if !ok {
			panic(fmt.Sprintf("type %s for field %s is not defined in dictionary", f.Type, f.Name))
		}
		res[f.Name] = tp
	}
	j, err := json.MarshalIndent(res, "", "\t")
	if err != nil {
		return "", err
	}
	return string(j), nil
}

func messageToParams(m *Message, file *protokit.FileDescriptor, prefix string, params *[]Param, f bool) {
	for _, f := range m.Fields {
		p := Param{
			Description: f.Description,
			Constraints: "",
			Name:        prefix + "." + f.Name,
		}
		if f.IsMap {
			p.Type = ObjectType
			*params = append(*params, p)
			pre := prefix + "." + f.Name
			*params = append(*params, Param{
				Name: pre + ".key",
			})
			*params = append(*params, Param{
				Name: pre + ".value",
			})
			continue
		}
		if f.IsRepeated {
			p.Type = ArrayType
			*params = append(*params, p)
			pre := prefix + "." + f.Name + "[0]"
			msg := file.GetMessage(f.Type)
			if msg != nil {
				messageToParams(parseMessage(msg), file, pre, params, false)
				continue
			}
			p.Name = pre
			enum := file.GetEnum(f.Type)
			if enum != nil {
				p.Type = StringType
				p.Constraints = "enum=\""
				for i, val := range enum.GetValues() {
					p.Constraints += val.GetName()
					if i != len(enum.GetValues())-1 {
						p.Constraints += ", "
					}
				}
				p.Constraints += "\""
				*params = append(*params, p)

				continue
			}

			p.Type = jType(f.Type)
			*params = append(*params, p)
			continue
		}
		msg := file.GetMessage(f.Type)
		if msg != nil {
			pre := prefix + "." + f.Name
			p.Type = ObjectType
			*params = append(*params, p)
			messageToParams(parseMessage(msg), file, pre, params, false)
			continue
		}
		enum := file.GetEnum(f.Type)
		if enum != nil {
			p.Type = StringType
			p.Constraints = "enum=\""
			for i, val := range enum.GetValues() {
				p.Constraints += val.GetName()
				if i != len(enum.GetValues())-1 {
					p.Constraints += ", "
				}
			}
			p.Constraints += "\""
			*params = append(*params, p)

			continue
		}
		p.Type = jType(f.Type)
		*params = append(*params, p)
	}
}

func jType(ftype string) string {
	switch ftype {
	case "string", "bytes":
		return "string"
	case "int", "int8", "int16", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64":
		return "numeric"
	case "float32", "float64", "double", "float":
		return "numeric"
	case "bool":
		return "boolean"
	default:
		return "unknown"
	}
}

func baseName(name string) string {
	parts := strings.Split(name, ".")
	return parts[len(parts)-1]
}

func labelName(lbl descriptor.FieldDescriptorProto_Label, proto3 bool, proto3Opt bool) string {
	if proto3 && !proto3Opt && lbl != descriptor.FieldDescriptorProto_LABEL_REPEATED {
		return ""
	}

	return strings.ToLower(strings.TrimPrefix(lbl.String(), "LABEL_"))
}

type typeContainer interface {
	GetType() descriptor.FieldDescriptorProto_Type
	GetTypeName() string
	GetPackage() string
}

func parseType(tc typeContainer) (string, string, string) {
	name := tc.GetTypeName()

	if strings.HasPrefix(name, ".") {
		name = strings.TrimPrefix(name, ".")
		return baseName(name), strings.TrimPrefix(name, tc.GetPackage()+"."), name
	}

	name = strings.ToLower(strings.TrimPrefix(tc.GetType().String(), "TYPE_"))
	return name, name, name
}

func description(comment string) string {
	val := strings.TrimLeft(comment, "*/\n ")
	if strings.HasPrefix(val, "@exclude") {
		return ""
	}

	return val
}

type orderedEnums []*Enum

func (oe orderedEnums) Len() int           { return len(oe) }
func (oe orderedEnums) Swap(i, j int)      { oe[i], oe[j] = oe[j], oe[i] }
func (oe orderedEnums) Less(i, j int) bool { return oe[i].LongName < oe[j].LongName }

type orderedExtensions []*FileExtension

func (oe orderedExtensions) Len() int           { return len(oe) }
func (oe orderedExtensions) Swap(i, j int)      { oe[i], oe[j] = oe[j], oe[i] }
func (oe orderedExtensions) Less(i, j int) bool { return oe[i].LongName < oe[j].LongName }

type orderedMessages []*Message

func (om orderedMessages) Len() int           { return len(om) }
func (om orderedMessages) Swap(i, j int)      { om[i], om[j] = om[j], om[i] }
func (om orderedMessages) Less(i, j int) bool { return om[i].LongName < om[j].LongName }

type orderedServices []*Service

func (os orderedServices) Len() int           { return len(os) }
func (os orderedServices) Swap(i, j int)      { os[i], os[j] = os[j], os[i] }
func (os orderedServices) Less(i, j int) bool { return os[i].LongName < os[j].LongName }
