package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/pacviewer/jrpc-gateway/_example/proto"
	"github.com/pacviewer/jrpc-gateway/jrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
)

type Server struct {
	sv           *grpc.Server
	jrpcSv       *jrpc.Server
	grpcListener net.Listener
	jrpcListener net.Listener
	grpcErr      chan error
	proto.UnimplementedEchoServiceServer
}

var (
	address  string
	jrpcAddr string
)

func main() {
	flag.StringVar(&address, "address", "localhost:8080", "the address ip:port")
	flag.StringVar(&jrpcAddr, "jrpc_addr", "localhost:8081", "the json rpc address ip:port")
	flag.Parse()

	server := &Server{}

	grpcSv, grpcLis, err := newGRPCSv(server)
	if err != nil {
		log.Fatal(err)
	}

	grpcConn, err := grpc.NewClient(
		address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatal(err)
	}

	httpSv, jrpcLis, err := newJrpcSv(proto.RegisterEchoServiceJsonRPC(grpcConn))
	if err != nil {
		log.Fatal(err)
	}

	server.sv = grpcSv
	server.jrpcSv = httpSv
	server.grpcListener = grpcLis
	server.grpcErr = make(chan error)
	server.jrpcListener = jrpcLis

	server.StartGRPC()
	server.StartJsonRPC()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	select {
	case s := <-interrupt:
		server.Stop()
		log.Println("app/run: signal received", "signal", s.String())
	}

}

func (s *Server) StartGRPC() {
	go func() {
		log.Println("GRPC server listening on", address)
		s.grpcErr <- s.sv.Serve(s.grpcListener)
	}()
}

func (s *Server) StartJsonRPC() {
	go func() {
		log.Println("JSON RPC server listening on", jrpcAddr)
		s.grpcErr <- s.jrpcSv.Serve(s.jrpcListener)
	}()
}

func (s *Server) Stop() {
	s.sv.GracefulStop()
	_ = s.jrpcSv.GracefulStop(context.Background())
}

func (s *Server) Echo(ctx context.Context, req *proto.EchoRequest) (*proto.EchoResponse, error) {
	return &proto.EchoResponse{
		Message: fmt.Sprintf("echo %s!!!", req.Name),
	}, nil
}

func newGRPCSv(server *Server) (*grpc.Server, net.Listener, error) {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatal(err)
	}

	sv := grpc.NewServer()
	proto.RegisterEchoServiceServer(sv, server)
	reflection.Register(sv)

	return sv, listener, nil
}

func newJrpcSv(echoClient *proto.EchoServiceJsonRPC) (*jrpc.Server, net.Listener, error) {
	jrpcSv := jrpc.NewServer(jrpc.WithCustomHeaders("x-custom-key"))

	jrpcSv.RegisterServices(echoClient)

	jrpcListener, err := net.Listen("tcp", jrpcAddr)
	if err != nil {
		return nil, nil, err
	}

	return jrpcSv, jrpcListener, nil
}
