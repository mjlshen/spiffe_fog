package main

import (
	"context"
	"fmt"
	"log"

	"github.com/google/go-attestation/attest"
	"github.com/mjlshen/spiffe_fog/pkg/common"
	"github.com/mjlshen/spiffe_fog/workload"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	conn, err := grpc.Dial("localhost:8080", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	client := workload.NewSpiffeWorkloadAPIClient(conn)
	resp, err := client.FetchJWTSVID(context.TODO(), &workload.JWTSVIDRequest{
		Audience: []string{"ayylmao"},
		SpiffeId: "ayylmao?",
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(resp)

	tpm, err := attest.OpenTPM(&attest.OpenConfig{
		TPMVersion: attest.TPMVersion20,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer tpm.Close()

	activation, err := common.GenerateCredentialActivationData(tpm)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(*activation)
}
