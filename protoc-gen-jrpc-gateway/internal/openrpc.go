package internal

import "text/template"

type OpenRPCDoc struct {
	OpenRPC string `json:"openrpc"`
	Info    struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Version     string `json:"version"`
	} `json:"info"`
	Servers    []Server        `json:"servers"`
	Methods    []OpenRPCMethod `json:"methods"`
	Components Components      `json:"components"`
}

type Server struct {
	Name        string `json:"name"`
	Summary     string `json:"summary"`
	Description string `json:"description"`
	URL         string `json:"url"`
}

type OpenRPCMethod struct {
	Name     string    `json:"name"`
	Params   []Content `json:"params"`
	Result   Content   `json:"result"`
	Examples []Example `json:"examples,omitempty"`
}

type Content struct {
	Name   string    `json:"name"`
	Schema SchemaRef `json:"schema"`
}

type SchemaRef struct {
	Ref  string `json:"$ref,omitempty"`
	Type string `json:"type,omitempty"`
}

type Example struct {
	Name   string      `json:"name"`
	Params []SchemaRef `json:"params"`
	Result SchemaRef   `json:"result"`
}

type Components struct {
	Schemas map[string]interface{} `json:"schemas"`
}

var _openrpcTmpl = template.Must(template.New("").Funcs(tmplFuncs).Parse(`
{
  "openrpc": "1.0.0-rc1",
  "info": {
    "title": "{{ .Info.Title }}",
    "description": "{{ .Info.Description }}",
    "version": "{{ .Info.Version }}"
  },
  "servers": [
    {{- range $i, $server := .Servers }}
    {{ if $i }},{{ end }}
    {
      "name": "{{ $server.Name }}",
      "summary": "{{ $server.Summary }}",
      "description": "{{ $server.Description }}",
      "url": "{{ $server.URL }}",
      "variables": {
        {{- range $j, $var := $server.Variables }}
        {{ if $j }},{{ end }}
        "{{ $var.Name }}": {
          "default": "{{ $var.Default }}",
          {{ if $var.Description }}"description": "{{ $var.Description }}",{{ end }}
          {{ if $var.Enum }}
          "enum": [{{ range $k, $e := $var.Enum }}{{ if $k }}, {{ end }}"{{ $e }}"{{ end }}]
          {{ end }}
        }
        {{- end }}
      }
    }
    {{- end }}
  ],
  "methods": [
    {{- range $i, $method := .Methods }}
    {{ if $i }},{{ end }}
    {
      "name": "{{ $method.Name }}",
      "params": [
        {{- range $j, $param := $method.Params }}
        {{ if $j }},{{ end }}
        {
          "name": "{{ $param.Name }}",
          "schema": {
            "$ref": "{{ $param.Schema.Ref }}"
          }
        }
        {{- end }}
      ],
      "result": {
        "$ref": "{{ $method.Result.Ref }}"
      },
      "examples": [
        {{- range $k, $example := $method.Examples }}
        {{ if $k }},{{ end }}
        {
          "name": "{{ $example.Name }}",
          "params": [
            {{- range $l, $exParam := $example.Params }}
            {{ if $l }},{{ end }}
            { "$ref": "{{ $exParam.Ref }}" }
            {{- end }}
          ],
          "result": { "$ref": "{{ $example.Result.Ref }}" }
        }
        {{- end }}
      ],
      "links": [
        {{- range $m, $link := $method.Links }}
        {{ if $m }},{{ end }}
        {
          "name": "{{ $link.Name }}",
          "description": "{{ $link.Description }}",
          "method": "{{ $link.Method }}",
          "params": {
            {{- range $key, $value := $link.Params }}
            "{{ $key }}": "{{ $value }}",
            {{- end }}
            "_end": null
          }
        }
        {{- end }}
      ]
    }
    {{- end }}
  ],
  "components": {
    "contentDescriptors": {
      {{- range $i, $desc := .Components.ContentDescriptors }}
      {{ if $i }},{{ end }}
      "{{ $desc.Name }}": {
        "name": "{{ $desc.Name }}",
        "schema": {
          "type": "{{ $desc.Schema.Type }}"
        }
      }
      {{- end }}
    },
    "schemas": {
      {{- range $i, $schema := .Components.Schemas }}
      {{ if $i }},{{ end }}
      "{{ $schema.Name }}": {
        "type": "{{ $schema.Type }}"
      }
      {{- end }}
    },
    "examples": {
      {{- range $i, $ex := .Components.Examples }}
      {{ if $i }},{{ end }}
      "{{ $ex.Name }}": {
        "name": "{{ $ex.Name }}",
        "summary": "{{ $ex.Summary }}",
        "description": "{{ $ex.Description }}",
        "value": {{ $ex.Value }}
      }
      {{- end }}
    }
  }
}
`))
