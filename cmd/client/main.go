package main

import (
	"context"
	"flag"

	"github.com/mjlshen/spiffe_fog/pkg/client"
	"github.com/mjlshen/spiffe_fog/proto/agent"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	defaultSpiffeId string = "demo"
	defaultServer   string = "localhost:8080"
)

func main() {
	id := flag.String("id", defaultSpiffeId, "The SPIFFE ID to request validation for")
	server := flag.String("server", defaultServer, "The URL to the SPIFFE Fog server")
	flag.Parse()

	conn, err := grpc.Dial(*server, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	agentClient, err := agent.NewAgentClient(conn).AttestAgent(context.TODO())
	if err != nil {
		panic(err)
	}

	agent := client.New(agentClient, *id)
	if err := agent.Attest(); err != nil {
		panic(err)
	}
}
