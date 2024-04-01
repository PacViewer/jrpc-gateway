### jrpc-gateway
A bridge from JSON-RPC to gRPC.

### protoc-gen-jrpc-gateway
JSON-RPC to gRPC protoc plugin.

### Installation
jrpc:
```
go get github.com/pacviewer/jrpc-gateway
```
plguin:
```
go install github.com/pacviewer/jrpc-gateway/protoc-gen-jrpc-gateway@v0.1.3
```
### Example
greeting.proto
```
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
create gen directory:
```
mkdir gen
```
generate files:
```
protoc --go_out=gen --go_opt=paths=source_relative \
    --go-grpc_out=gen --go-grpc_opt=paths=source_relative \
    --jrpc-gateway_out=gen \
    greeting.proto
```
these three files will be created for you in gen directory:
- greeting_grpc.pb.go
- greeting.pb.go
- greeting.pb.jgw.go

get neccessary dependencies
```
go mod tidy
```
### Register JSON-RPC methods
```
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
request:
```
curl -X POST -H 'Content-Type: application/json' \
     -d '{"jsonrpc":"2.0","id":"1111","method":"greeting.greeting_service.greeting", "params":{"name":"john"}}' \
     http://localhost:8687
```
response:
```
{"jsonrpc":"2.0","id":"1111","result":{"message":"Hello john"}}
```
