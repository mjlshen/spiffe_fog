package common

import (
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"fmt"

	"github.com/google/go-attestation/attest"
)

type AttestationData struct {
	EK []byte
	AK *attest.AttestationParameters
}

type Challenge struct {
	EC *attest.EncryptedCredential
}

type ChallengeResponse struct {
	Secret []byte
}

func pubBytes(ek *attest.EK) ([]byte, error) {
	data, err := x509.MarshalPKIXPublicKey(ek.Public)
	if err != nil {
		return nil, fmt.Errorf("error marshaling ec public key: %v", err)
	}
	return data, nil
}

func GetPubHash(ek *attest.EK) (string, error) {
	data, err := pubBytes(ek)
	if err != nil {
		return "", err
	}
	pubHash := sha256.Sum256(data)
	hashEncoded := fmt.Sprintf("%x", pubHash)
	return hashEncoded, nil
}

func EncodeEK(ek *attest.EK) ([]byte, error) {
	if ek.Certificate != nil {
		return pem.EncodeToMemory(&pem.Block{
			Type:  "CERTIFICATE",
			Bytes: ek.Certificate.Raw,
		}), nil
	}

	data, err := pubBytes(ek)
	if err != nil {
		return nil, err
	}

	return pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: data,
	}), nil
}

func DecodeEK(pemBytes []byte) (*attest.EK, error) {
	block, _ := pem.Decode(pemBytes)

	if block == nil {
		return nil, fmt.Errorf("invalid pemBytes")
	}

	switch block.Type {
	case "CERTIFICATE":
		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("error parsing certificate: %v", err)
		}
		return &attest.EK{
			Certificate: cert,
			Public:      cert.PublicKey,
		}, nil

	case "PUBLIC KEY":
		pub, err := x509.ParsePKIXPublicKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("error parsing ecdsa public key: %v", err)
		}

		return &attest.EK{
			Public: pub,
		}, nil
	}

	return nil, fmt.Errorf("invalid pem type: %s", block.Type)
}

// GenerateCredentialActivationData generates TPM2.0 credential activation material so that it can
// be verified with TPM2_ActivateCredential to prove this device has a TPM with both EK and AK and
// an encoded blob of AK so that it can be loaded later.
func GenerateCredentialActivationData(tpm *attest.TPM) (*AttestationData, []byte, error) {
	ak, err := tpm.NewAK(nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create AK: %v", err)
	}
	defer ak.Close(tpm)

	ek, err := GetEK(tpm)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get EK: %v", err)
	}

	ekBytes, err := EncodeEK(ek)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to encode EK: %v", err)
	}

	akBlob, err := ak.Marshal()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal AK into blob: %v", err)
	}

	params := ak.AttestationParameters()
	ap := &AttestationData{
		EK: ekBytes,
		AK: &params,
	}

	return ap, akBlob, nil
}

// SolveCredentialActivationChallenge attempts to decrypt a challenge that should be encrypted with
// the public EK and AK (represented by akBlob), so that only a specific TPM can decrypt it successfully.
func SolveCredentialActivationChallenge(tpm *attest.TPM, challenge attest.EncryptedCredential, akBlob []byte) ([]byte, error) {
	ak, err := tpm.LoadAK(akBlob)
	if err != nil {
		return nil, fmt.Errorf("unable to load AK: %v", err)
	}
	defer ak.Close(tpm)

	return ak.ActivateCredential(tpm, challenge)
}

// GetEK returns the first EK provided, otherwise returns an error
func GetEK(tpm *attest.TPM) (*attest.EK, error) {
	eks, err := tpm.EKs()
	if err != nil {
		return nil, fmt.Errorf("failed to enumerate EKs: %v", err)
	}

	if len(eks) == 0 {
		return nil, fmt.Errorf("no EK available")
	}

	ek := &eks[0]
	return ek, nil
}
