## jrpc-gateway
jrpc-gateway is bridge between [json-rpc 2.0](https://www.jsonrpc.org/specification) and [gRPC](https://grpc.io/)

This repository consists of:
- [jrpc](#jrpc) package 
- [protoc-gen-jrpc-gateway](#protoc-gen-jrpc-gateway) protoc plugin
- [protoc-gen-jrpc-doc](#protoc-gen-jrpc-doc) protoc plugin

<a id="jrpc"></a>
## jrpc
### installation
```bash
go get github.com/pacviewer/jrpc-gateway/jrpc
```

<a id="protoc-gen-jrpc-gateway"></a>
## protoc-gen-jrpc-gateway
protoc-gen-jrpc-gateway generates json-rpc to grpc bridge code based on proto files

### Installation
```bash
go install github.com/pacviewer/jrpc-gateway/protoc-gen-jrpc-gateway@latest
```
### Example
greeting.proto
```go
syntax = "proto3";
package greeting;
option go_package = "/greeting";

service GreetingService {
  rpc Greeting(GreetingReq) returns(GreetingResp) {}
}

message GreetingReq {
  string name = 1;
}

message GreetingResp {
  string message = 2;
}
```
### Create gen directory
```bash
mkdir gen
```
### Generate files
```bash
protoc --go_out=gen --go_opt=paths=source_relative \
    --go-grpc_out=gen --go-grpc_opt=paths=source_relative \
    --jrpc-gateway_out=gen \
    greeting.proto
```
these three files will be created for you in gen directory:
- greeting_grpc.pb.go
- greeting.pb.go
- greeting_jgw.pb.go

### Get dependencies
```bash
go mod tidy
```
### Register JSON-RPC methods
```go
grpcConn, err := grpc.DialContext(
context.Background(),
  "127.0.0.1:8686", // grpc server address
  grpc.WithTransportCredentials(insecure.NewCredentials()),
)

if err != nil {
  log.Fatalln(err)  
}

greeting := pb.NewGreetingServiceClient(grpcConn)
greetingService := pb.NewGreetingServiceJsonRpcService(greeting)

jgw := jrpc.NewServer()
jgw.RegisterServices(&greetingService)

// json-rpc listener
lis, err = net.Listen("tcp", "localhost:8687")
if err != nil {
  log.Fatalln(err)
}
mux := http.NewServeMux()
mux.HandleFunc("/", jgw.HttpHandler)
server := &http.Server{
  Handler:           mux,
  Addr:              lis.Addr().String(),
  ReadHeaderTimeout: 3 * time.Second,
}

if err := server.Serve(lis); err != nil {
  log.Fatalln(err)
}
```
### Test method call
#### Request:
```bash
curl -X POST -H 'Content-Type: application/json' \
     -d '{"jsonrpc":"2.0","id":"1111","method":"greeting.greeting_service.greeting", "params":{"name":"john"}}' \
     http://localhost:8687
```
#### Response:
```bash
{"jsonrpc":"2.0","id":"1111","result":{"message":"Hello john"}}
```

<a id="protoc-gen-jrpc-doc"></a>
## protoc-gen-jrpc-doc
### Installation
```bash
go install github.com/pacviewer/jrpc-gateway/protoc-gen-jrpc-doc/cmd/protoc-gen-jrpc-doc@v0.1.4
```
### Generate doc
```bash
protoc --jrpc-doc_out=gen --jrpc-doc_opt=./json-rpc-md.tmpl,json-rpc.md \
  greeting.proto
```