package common

import (
	"errors"
	"github.com/google/go-attestation/attest"
)

func GenerateCredentialActivationData(tpm *attest.TPM) (*attest.ActivationParameters, error) {
	ek, err := GetEK(tpm)
	if err != nil {
		return nil, err
	}

	ak, err := GetAK(tpm)
	if err != nil {
		return nil, err
	}

	return &attest.ActivationParameters{
		TPMVersion: attest.TPMVersion20,
		EK:         ek.Public,
		AK:         ak.AttestationParameters(),
		Rand:       nil,
	}, nil
}

func GetEK(tpm *attest.TPM) (*attest.EK, error) {
	eks, err := tpm.EKs()
	if err != nil {
		return nil, err
	}

	if len(eks) == 0 {
		return nil, errors.New("no EK available")
	}

	ek := &eks[0]
	return ek, nil
}

func GetAK(tpm *attest.TPM) (*attest.AK, error) {
	ak, err := tpm.NewAK(nil)
	if err != nil {
		return nil, err
	}
	defer ak.Close(tpm)

	return ak, nil
}
