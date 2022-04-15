package main

import (
	"context"
	"crypto/subtle"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"

	"github.com/google/go-attestation/attest"
	"github.com/mjlshen/spiffe_fog/pkg/common"
	"github.com/mjlshen/spiffe_fog/proto/agent"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

type Service struct {
	agent.UnimplementedAgentServer
}

// AttestAgent handles TPM credential activation
func (s *Service) AttestAgent(stream agent.Agent_AttestAgentServer) error {
	ctx := stream.Context()

	req, err := stream.Recv()
	if err != nil {
		return status.Errorf(codes.InvalidArgument, "failed to receive request from stream: %v", err)
	}

	// The first communication with an agent must contain attestation parameters
	params := req.GetParams()
	if err := validateAttestAgentParams(params); err != nil {
		return status.Errorf(codes.InvalidArgument, "malformed attestation param: %v", err)
	}

	// Gather the params, send a challenge and receive a challenge response
	attestResult, err := s.attestChallengeResponse(ctx, stream, params)
	if err != nil {
		return err
	}

	// If there's no error, then this node checks out!
	if err := stream.Send(attestResult); err != nil {
		return status.Errorf(codes.Internal, "failed to send response over stream", err)
	}

	return nil
}

func (s *Service) attestChallengeResponse(ctx context.Context,
	stream agent.Agent_AttestAgentServer,
	params *agent.AttestAgentRequest_Params,
) (*agent.AttestAgentResponse, error) {
	if params.Data.Type != "tpm_activation" {
		return nil, status.Errorf(codes.InvalidArgument, "unsupported type: %s", params.Data.Type)
	}

	log.Println("received attestation request")
	payload := params.Data.GetPayload()
	if payload == nil {
		return nil, status.Error(codes.InvalidArgument, "missing attestation payload")
	}

	var tpmAttestationData common.AttestationData
	if err := json.Unmarshal(payload, &tpmAttestationData); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "malformed activation param: %v", err)
	}

	ek, err := common.DecodeEK(tpmAttestationData.EK)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "malformed EK: %v", err)
	}

	if ok, err := isValidEK(ek); !ok {
		return nil, status.Errorf(codes.InvalidArgument, "invalid EK: %v", err)
	}

	// Collect TPM activation parameters to generate a challenge
	ap := attest.ActivationParameters{
		TPMVersion: attest.TPMVersion20,
		EK:         ek.Public,
		AK:         *tpmAttestationData.AK,
	}

	secret, challenge, err := ap.Generate()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to generate activation challenge: %v", err)
	}

	challengeBytes, err := json.Marshal(challenge)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to marshal challenge: %v", err)
	}

	log.Println("sending attestation challenge")
	if err := stream.Send(&agent.AttestAgentResponse{
		Step: &agent.AttestAgentResponse_Challenge{
			Challenge: challengeBytes,
		},
	}); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to send challenge: %v", err)
	}

	resp, err := stream.Recv()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to receive challenge response: %v", err)
	}

	challengeResp := resp.GetChallengeResponse()
	if challengeResp == nil {
		return nil, status.Error(codes.InvalidArgument, "missing challenge response")
	}

	//Verify the challenge response
	if subtle.ConstantTimeCompare(secret, challengeResp) == 0 {
		return nil, status.Errorf(codes.PermissionDenied, "challenge response does not match")
	}

	cr, err := x509.ParseCertificateRequest(params.Params.Csr)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to parse CSR: %v", err)
	}

	log.Printf("successful attestation for %s", cr.URIs[0].String())
	return &agent.AttestAgentResponse{
		Step: &agent.AttestAgentResponse_Result_{
			Result: &agent.AttestAgentResponse_Result{
				Svid: &agent.X509SVID{
					CertChain: nil,
					Id: &agent.SPIFFEID{
						TrustDomain: "test",
						Path:        cr.URIs[0].String(),
					},
					ExpiresAt: 0,
				},
			},
		},
	}, nil
}

// validEKHashes returns a map of hashes keys with values indicating if they are valid
// TODO: Allow other backing stores
func validEKHashes() map[string]bool {
	return map[string]bool{
		// GCP TPM
		"ae76715da45c546d57473816bb7402b467ac7e11d76ae43205769b65e3821f9d": true,
		// RPi Infineon TPM
		"ae8dec3321f80ab68bdde38e3cf7d59612be0c0a608def2c3d55a63fd875e32c": true,
	}
}

// isValidEK returns true if the provided EK is trusted by comparing the sha256 hash
// of the EK public key after it has been converted to the ASN.1 DER format with
// "valid" EK hashes.
func isValidEK(ek *attest.EK) (bool, error) {
	ekHash, err := common.GetPubHash(ek)
	if err != nil {
		return false, err
	}

	valid := validEKHashes()
	if active, ok := valid[ekHash]; !ok || !active {
		return false, fmt.Errorf("invalid EK hash: %s", ekHash)
	}

	log.Printf("processing EK: %s", ekHash)
	return true, nil
}

func validateAttestAgentParams(params *agent.AttestAgentRequest_Params) error {
	switch {
	case params == nil:
		return errors.New("missing params")
	case params.Data == nil:
		return errors.New("missing attestation data")
	case params.Params == nil:
		return errors.New("missing X509-SVID parameters")
	case len(params.Params.Csr) == 0:
		return errors.New("missing CSR")
	case params.Data.Type == "":
		return errors.New("missing attestation data type")
	case len(params.Data.Payload) == 0:
		return errors.New("missing attestation data payload")
	default:
		return nil
	}
}

func main() {
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		panic(err)
	}

	s := grpc.NewServer()
	reflection.Register(s)
	agent.RegisterAgentServer(s, &Service{})
	if err := s.Serve(listener); err != nil {
		panic(err)
	}
}
