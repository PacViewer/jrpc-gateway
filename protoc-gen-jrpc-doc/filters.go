package jrpc_doc

import (
	"bytes"
	"fmt"
	"html/template"
	"regexp"
	"strings"
	"unicode"
)

var (
	paraPattern         = regexp.MustCompile(`(\n|\r|\r\n)\s*`)
	spacePattern        = regexp.MustCompile("( )+")
	multiNewlinePattern = regexp.MustCompile(`(\r\n|\r|\n){2,}`)
	specialCharsPattern = regexp.MustCompile(`[^a-zA-Z0-9_-]`)
)

// PFilter splits the content by new lines and wraps each one in a <p> tag.
func PFilter(content string) template.HTML {
	paragraphs := paraPattern.Split(content, -1)
	return template.HTML(fmt.Sprintf("<p>%s</p>", strings.Join(paragraphs, "</p><p>")))
}

// ParaFilter splits the content by new lines and wraps each one in a <para> tag.
func ParaFilter(content string) string {
	paragraphs := paraPattern.Split(content, -1)
	return fmt.Sprintf("<para>%s</para>", strings.Join(paragraphs, "</para><para>"))
}

// NoBrFilter removes single CR and LF from content.
func NoBrFilter(content string) string {
	normalized := strings.Replace(content, "\r\n", "\n", -1)
	paragraphs := multiNewlinePattern.Split(normalized, -1)
	for i, p := range paragraphs {
		withoutCR := strings.Replace(p, "\r", " ", -1)
		withoutLF := strings.Replace(withoutCR, "\n", " ", -1)
		paragraphs[i] = spacePattern.ReplaceAllString(withoutLF, " ")
	}
	return strings.Join(paragraphs, "\n\n")
}

// AnchorFilter replaces all special characters with URL friendly dashes
func AnchorFilter(str string) string {
	return specialCharsPattern.ReplaceAllString(strings.ReplaceAll(str, "/", "_"), "-")
}

func RpcMethod(pkg, service, method string) string {
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

func Minus(a, b int) int {
	return a - b
}

func ToJsonRpc(text string, protoMethod string) string {
	str := strings.Split(text, " ")
	for i, s := range str {
		if s == protoMethod {
			str[i] = camelToSnake(protoMethod)
		}
	}
	return strings.Join(str, " ")
}

// func FindExample(name, t string, jdict JDict) any {
// 	n, ok := jdict[name]
// 	if !ok {
// 		panic(fmt.Sprintf("field %s is not defined in dictionary", name))
// 	}
// 	tp, ok := n[t]
// 	if !ok {
// 		panic(fmt.Sprintf("type %s for field %s is not defined in dictionary", t, name))
// 	}

// 	mar, _ := json.Marshal(tp)

// 	var res any
// 	if err := json.Unmarshal(mar, &res); err != nil {
// 		panic(fmt.Sprintf("invalid value %s for field %s in dictionary", t, name))
// 	}
// 	return res
// }

func FullTypeName(t string, isRepeated, isMap bool) string {
	if isRepeated || isMap {
		return fmt.Sprintf("[]%s", t)
	}
	return t
}
