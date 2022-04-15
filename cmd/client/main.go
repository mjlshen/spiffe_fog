package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"flag"
	"log"

	"github.com/mjlshen/spiffe_fog/pkg/client"
	"github.com/mjlshen/spiffe_fog/proto/agent"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	defaultSpiffeId string = "demo"
	defaultHost     string = "localhost:8080"
)

func NewConn(host string, ins bool) (*grpc.ClientConn, error) {
	var opts []grpc.DialOption
	if host != "" {
		opts = append(opts, grpc.WithAuthority(host))
	}

	if ins {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	} else {
		systemRoots, err := x509.SystemCertPool()
		if err != nil {
			return nil, err
		}
		cred := credentials.NewTLS(&tls.Config{
			RootCAs: systemRoots,
		})
		opts = append(opts, grpc.WithTransportCredentials(cred))
	}

	return grpc.Dial(host, opts...)
}

func main() {
	id := flag.String("id", defaultSpiffeId, "The SPIFFE ID to request validation for")
	host := flag.String("host", defaultHost, "The host in the form domain:port to the SPIFFE Fog server")
	ins := flag.Bool("insecure", false, "Use an insecure gRPC connection")
	flag.Parse()

	log.Printf("Requesting SPIFFE ID: %s, host: %s, insecure: %v", *id, *host, *ins)
	conn, err := NewConn(*host, *ins)
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
