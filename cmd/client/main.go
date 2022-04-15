package main

import (
	"context"
	"flag"

	"github.com/mjlshen/spiffe_fog/pkg/client"
	"github.com/mjlshen/spiffe_fog/proto/agent"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const defaultSpiffeId string = "demo"

func main() {
	id := flag.String("id", defaultSpiffeId, "The SPIFFE ID to request validation for")
	flag.Parse()

	conn, err := grpc.Dial("localhost:8080", grpc.WithTransportCredentials(insecure.NewCredentials()))
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
