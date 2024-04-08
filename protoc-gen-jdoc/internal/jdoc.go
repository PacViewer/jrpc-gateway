package internal

import (
	"bytes"
	_ "embed"
	"fmt"
	"html/template"
	"strings"
	"unicode"
)

var (
	//go:embed jdoc.md
	markdownFile string
	//go:embed jdoc.html
	htmlFile string
	//go:embed jdoc.json
	postmanFile string
)

var tmplFuncs = map[string]any{
	"rpcMethod":   rpcMethod,
	"methodInput": methodInput,
}

func Tmpl(file string) *template.Template {
	return template.Must(template.New("").Funcs(tmplFuncs).Parse(file))
}

func rpcMethod(pkg, service, method string) string {
	return fmt.Sprintf("%s.%s.%s", camelToSnake(pkg), camelToSnake(service), camelToSnake(method))
}

func camelToSnake(s string) string {
	var buf bytes.Buffer
	for i, r := range s {
		if unicode.IsUpper(r) {
			if i > 0 {
				buf.WriteRune('_')
			}
			buf.WriteRune(unicode.ToLower(r))
		} else {
			buf.WriteRune(r)
		}
	}
	return buf.String()
}

func methodInput(in string) string {
	sep := strings.Split(in, ".")
	return sep[len(sep)-1]
}
