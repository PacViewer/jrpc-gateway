### Version {{.Metadata.Version}}

{{$package := (index .Files 0).Package}}

---

### Methods

{{range $file := .Files}}{{range $service := $file.Service}}{{range $method := $service.GetMethod}}
- [{{rpcMethod $package $service.GetName $method.GetName}}](#{{rpcMethod $package $service.GetName $method.GetName}})
{{end}}{{end}}{{end}}

---

{{range $file := .Files}}
{{range $service := $file.Service}}
{{range $method := $service.GetMethod}}

## {{$service.GetName}}
desc of service

<a name="{{rpcMethod $package $service.GetName $method.GetName}}"></a>

### {{rpcMethod $package $service.GetName $method.GetName}}

Creates a new session.

##### Request

```json
{
  "jsonrpc": "2.0",
  "id": "1234567890",
  "method": "{{rpcMethod $package $service.GetName $method.GetName}}",
  "params": {
    "name": "admin",
    "password": "123456"
  }
}
```

#### Response

```json
{
  "jsonrpc": "2.0",
  "id": "1234567890",
  "result": {
    "session_token": "123456890",
    "validity": 3600
  }
}
```
{{end}}
{{end}}
{{end}}