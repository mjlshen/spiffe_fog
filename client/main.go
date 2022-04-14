package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/google/go-attestation/attest"
	"github.com/mjlshen/spiffe_fog/pkg/common"
	"github.com/mjlshen/spiffe_fog/proto/agent"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	conn, err := grpc.Dial("localhost:8080", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	client := agent.NewAgentClient(conn)
	attestClient, err := client.AttestAgent(context.TODO())
	if err != nil {
		log.Fatal(err)
	}

	tpm, err := attest.OpenTPM(&attest.OpenConfig{
		TPMVersion: attest.TPMVersion20,
	})
	if err != nil {
		log.Fatalf("failed to open TPM: %v", err)
	}
	defer tpm.Close()

	ap, akBlob, err := common.GenerateCredentialActivationData(tpm)
	if err != nil {
		log.Fatalf("failed to generate credential activation data: %v", err)
	}

	apBytes, err := json.Marshal(*ap)
	if err != nil {
		log.Fatalf("failed to marshal activation parameters into json: %v", err)
	}

	csr, _, err := common.NewCSRTemplate("spiffe://spiffe_fog.com/raspberrypi")
	if err != nil {
		log.Fatalf("failed to generate CSR: %v", err)
	}

	attestClient.Send(&agent.AttestAgentRequest{
		Step: &agent.AttestAgentRequest_Params_{
			Params: &agent.AttestAgentRequest_Params{
				Data: &agent.AttestationData{
					Type:    "tpm_activation",
					Payload: apBytes,
				},
				Params: &agent.AgentX509SVIDParams{
					Csr: csr,
				},
			},
		}},
	)

	challengeReq, err := attestClient.Recv()
	if err != nil {
		log.Fatal(err)
	}

	challengeBytes := challengeReq.GetChallenge()
	var challenge attest.EncryptedCredential
	if err := json.Unmarshal(challengeBytes, &challenge); err != nil {
		log.Fatalf("failed to unmarshal challenge: %v", err)
	}

	decrypted, err := common.SolveCredentialActivationChallenge(tpm, challenge, akBlob)
	if err != nil {
		log.Fatalf("failed to respond to credential activation challenge: %v", err)
	}

	attestClient.Send(&agent.AttestAgentRequest{
		Step: &agent.AttestAgentRequest_ChallengeResponse{
			ChallengeResponse: decrypted,
		},
	})

	svidResp, err := attestClient.Recv()
	if err != nil {
		log.Fatal(err)
	}

	log.Print(svidResp.GetResult())
	fmt.Println("SUCCESS!!")
}
