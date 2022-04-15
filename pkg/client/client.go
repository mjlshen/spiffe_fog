package client

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/google/go-attestation/attest"
	"github.com/mjlshen/spiffe_fog/pkg/common"
	"github.com/mjlshen/spiffe_fog/proto/agent"
)

type Client struct {
	agent  agent.Agent_AttestAgentClient
	domain string
}

func generateSpiffeFogDomain(id string) string {
	return fmt.Sprintf("spiffe://spiffe_fog/%s", id)
}

func New(a agent.Agent_AttestAgentClient, id string) Client {
	return Client{
		agent:  a,
		domain: generateSpiffeFogDomain(id),
	}
}

func (c Client) Attest() error {
	tpm, err := attest.OpenTPM(&attest.OpenConfig{
		TPMVersion: attest.TPMVersion20,
	})
	if err != nil {
		log.Fatalf("failed to open TPM: %v", err)
	}
	defer tpm.Close()

	ap, akBlob, err := common.GenerateCredentialActivationData(tpm)
	if err != nil {
		return fmt.Errorf("failed to generate credential activation data: %v", err)
	}

	apBytes, err := json.Marshal(*ap)
	if err != nil {
		return fmt.Errorf("failed to marshal activation parameters into json: %v", err)
	}

	csr, _, err := common.NewCSRTemplate(c.domain)
	if err != nil {
		return fmt.Errorf("failed to generate CSR: %v", err)
	}

	c.agent.Send(&agent.AttestAgentRequest{
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

	challengeReq, err := c.agent.Recv()
	if err != nil {
		return err
	}

	challengeBytes := challengeReq.GetChallenge()
	var challenge attest.EncryptedCredential
	if err := json.Unmarshal(challengeBytes, &challenge); err != nil {
		return fmt.Errorf("failed to unmarshal challenge: %v", err)
	}

	decrypted, err := common.SolveCredentialActivationChallenge(tpm, challenge, akBlob)
	if err != nil {
		return fmt.Errorf("failed to respond to credential activation challenge: %v", err)
	}

	c.agent.Send(&agent.AttestAgentRequest{
		Step: &agent.AttestAgentRequest_ChallengeResponse{
			ChallengeResponse: decrypted,
		},
	})

	svidResp, err := c.agent.Recv()
	if err != nil {
		return err
	}

	log.Print(svidResp.GetResult())
	return nil
}
