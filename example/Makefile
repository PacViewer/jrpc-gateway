ifneq (,$(filter $(OS),Windows_NT MINGW64))
EXE = .exe
RM = del /q
else
RM = rm -rf
endif

all: devtools proto

devtools:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2
	go install github.com/pacviewer/jrpc-gateway/protoc-gen-jrpc-gateway

proto:
	$(RM) ./gen
	cd ./buf && buf generate --template buf.gen.yaml ../proto

.PHONY: devtools proto