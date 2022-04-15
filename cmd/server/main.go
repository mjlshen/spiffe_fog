package main

import (
	"flag"
	"net"

	"github.com/mjlshen/spiffe_fog/pkg/server"
	"github.com/mjlshen/spiffe_fog/proto/agent"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const defaultPort string = "8080"

func main() {
	port := flag.String("port", defaultPort, "Port to listen on")
	flag.Parse()

	listener, err := net.Listen("tcp", ":"+*port)
	if err != nil {
		panic(err)
	}

	s := grpc.NewServer()
	reflection.Register(s)
	agent.RegisterAgentServer(s, &server.Service{})
	if err := s.Serve(listener); err != nil {
		panic(err)
	}
}
